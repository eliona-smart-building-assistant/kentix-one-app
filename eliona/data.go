//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	"context"
	"fmt"
	"kentix-one/apiserver"
	"kentix-one/conf"
	"kentix-one/kentix"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func UpsertDeviceInfo(config apiserver.Configuration, device kentix.DeviceInfo) error {
	for _, projectId := range conf.ProjIds(config) {
		err := upsertDeviceInfo(config, projectId, device)
		if err != nil {
			return err
		}
	}
	return nil
}

type deviceInfoPayload struct {
	Name            string `json:"name"`
	IPAddress       string `json:"ip_address"`
	MACAddress      string `json:"mac_address"`
	FirmwareVersion string `json:"firmware_version"`
}

func upsertDeviceInfo(config apiserver.Configuration, projectId string, device kentix.DeviceInfo) error {
	log.Debug("Eliona", "upserting data for device: config %d and device '%s'", config.Id, device.UUID)
	assetId, err := conf.GetAssetId(context.Background(), config, projectId, device.UUID)
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	return upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		deviceInfoPayload{
			Name:            device.Name,
			IPAddress:       device.IPAddress,
			MACAddress:      device.MacAddress,
			FirmwareVersion: device.Version,
		},
	)
}

func UpsertDoorlockData(config apiserver.Configuration, doorlock kentix.DoorLock) error {
	for _, projectId := range conf.ProjIds(config) {
		err := upsertDoorlockData(config, projectId, doorlock)
		if err != nil {
			return err
		}
	}
	return nil
}

type doorlockInfoDataPayload struct {
	ID       int    `json:"id"`
	DeviceID int    `json:"device_id"`
	Name     string `json:"name"`
}

type doorlockInputDataPayload struct {
	Battery int `json:"battery"`
}

type doorlockOutputDataPayload struct {
	Open int `json:"open"`
}

func upsertDoorlockData(config apiserver.Configuration, projectId string, doorlock kentix.DoorLock) error {
	log.Debug("Eliona", "upserting data for doorlock: config %d and doorlock '%s'", config.Id, doorlock.ID)
	assetId, err := conf.GetAssetId(context.Background(), config, projectId, fmt.Sprint(doorlock.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		doorlockInfoDataPayload{
			ID:       doorlock.ID,
			DeviceID: doorlock.DeviceID,
			Name:     doorlock.Name,
		},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}

	if err := upsertData(
		api.SUBTYPE_INPUT,
		*assetId,
		doorlockInputDataPayload{
			Battery: doorlock.BatteryLevel,
		},
	); err != nil {
		return fmt.Errorf("upserting input data: %v", err)
	}
	return nil
}

func UpsertMultiSensorData(config apiserver.Configuration, deviceId string, measurements kentix.Measurements) error {
	sensors, err := conf.GetConfigDevices(context.Background(), config, deviceId)
	if err != nil {
		return fmt.Errorf("getting config sensors: %v", err)
	}
	for _, projectId := range conf.ProjIds(config) {
		for _, sensor := range sensors {
			if err := upsertMultiSensorData(sensor, projectId, measurements); err != nil {
				return fmt.Errorf("upserting MultiSensor data: %v", err)
			}
		}
	}
	return nil
}

type measurementsPayload struct {
	Temperature *string `json:"temperature"`
	Humidity    *string `json:"humidity"`
	DewPoint    *string `json:"dew_point"`
	AirQuality  *string `json:"air_quality"`
	CO2         *string `json:"co2"`
	CO          *string `json:"co"`
	Heat        *string `json:"heat"`
	Motion      *string `json:"motion"`
	Vibration   *string `json:"vibration"`
	PeopleCount *string `json:"people_count"`
}

func upsertMultiSensorData(sensor apiserver.Device, projectId string, measurements kentix.Measurements) error {
	log.Debug("Eliona", "upserting data for MultiSensor: sensor %s", sensor.SerialNumber)

	return upsertData(
		api.SUBTYPE_INPUT,
		*sensor.AssetID,
		measurementsPayload{
			Temperature: measurements.Temperature.Value.ToStringPtr(),
			Humidity:    measurements.Humidity.Value.ToStringPtr(),
			DewPoint:    measurements.Dewpoint.Value.ToStringPtr(),
			AirQuality:  measurements.AirQuality.Value.ToStringPtr(),
			CO2:         measurements.CO2.Value.ToStringPtr(),
			CO:          measurements.CO.Value.ToStringPtr(),
			Heat:        measurements.Heat.Value.ToStringPtr(),
			Motion:      measurements.Motion.Value.ToStringPtr(),
			Vibration:   measurements.Vibration.Value.ToStringPtr(),
			PeopleCount: measurements.PeopleCount.Value.ToStringPtr(),
		},
	)
}

//

func upsertData(subtype api.DataSubtype, assetId int32, payload any) error {
	var statusData api.Data
	statusData.Subtype = subtype
	now := time.Now()
	statusData.Timestamp = *api.NewNullableTime(&now)
	statusData.AssetId = assetId
	statusData.Data = common.StructToMap(payload)
	if err := asset.UpsertDataIfAssetExists[any](statusData); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}
