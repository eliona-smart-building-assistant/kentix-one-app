//  This file is part of the eliona project.
//  Copyright © 2023 LEICOM iTEC AG. All Rights Reserved.
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

package kentix

import (
	"fmt"
	"kentix-one/apiserver"
	"net/url"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/http"
)

const (
	AccessManagerAssetType = "kentixone_access_manager"
	AlarmManagerAssetType  = "kentixone_alarm_manager"
	MultiSensorAssetType   = "kentixone_multi_sensor"
	DoorlockAssetType      = "kentixone_doorlock"
)

type systemValuesResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

type DeviceInfo struct {
	Name         string        `json:"name"`
	IPAddress    string        `json:"address"`
	MacAddress   string        `json:"mac_address"`
	Type         int           `json:"type"`
	AssetType    string        `json:"-"`
	UUID         string        `json:"uuid"`
	Version      string        `json:"version"`
	Measurements *Measurements `json:"measurements"`
}

type Measurements struct {
	Temperature MeasurementValue `json:"temperature"`
	Humidity    MeasurementValue `json:"humidity"`
	Dewpoint    MeasurementValue `json:"dewpoint"`
	AirQuality  MeasurementValue `json:"air_quality"`
	CO2         MeasurementValue `json:"co2"`
	Heat        MeasurementValue `json:"heat"`
	CO          MeasurementValue `json:"co"`
	Motion      MeasurementValue `json:"motion"`
	Vibration   MeasurementValue `json:"vibration"`
	PeopleCount MeasurementValue `json:"people_count"`
}

type MeasurementValue struct {
	Min        string `json:"min"`
	Max        string `json:"max"`
	Assignment string `json:"assignment"`
	Value      string `json:"value"`
	Status     string `json:"status"`
}

func GetDevices(conf apiserver.Configuration) ([]DeviceInfo, error) {
	url, err := url.JoinPath(conf.Address, "api/systemvalues")
	if err != nil {
		return nil, fmt.Errorf("creating URL: %v", err)
	}
	r, err := http.NewRequestWithApiKey(url, "Authorization", "Bearer "+conf.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", url, err)
	}
	systemValuesResponse, err := http.Read[systemValuesResponse](r, time.Duration(*conf.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", url, err)
	}
	for _, d := range systemValuesResponse.Devices {
		d.AssetType, err = inferAssetType(d.Type)
		if err != nil {
			return nil, fmt.Errorf("inferring asset type for %s: %v", d.IPAddress, err)
		}
	}
	return systemValuesResponse.Devices, nil
}

func inferAssetType(typ int) (string, error) {
	// TODO: Verify that this is the correct property to determine device type.
	switch typ {
	case 105:
		return AccessManagerAssetType, nil
	case 101:
		return AlarmManagerAssetType, nil
	case 110:
		return MultiSensorAssetType, nil
	case 21:
		return DoorlockAssetType, nil
	}
	return "", fmt.Errorf("unknown device type: %v", typ)
}

type accessPointResponse struct {
	Data  []DoorLock     `json:"data"`
	Links PaginationLink `json:"links"`
}

type DoorLock struct {
	ID           int    `json:"id"`
	DeviceID     int    `json:"device_id"`
	Active       bool   `json:"is_active"`
	Name         string `json:"name"`
	BatteryLevel int    `json:"battery_level"`
}

type PaginationLink struct {
	Next string `json:"next"`
}

func GetDoorlocks(conf apiserver.Configuration) ([]DoorLock, error) {
	url, err := url.JoinPath(conf.Address, "api/doorlocks")
	if err != nil {
		return nil, fmt.Errorf("appending endpoint to URL: %v", err)
	}
	return fetchDoorlocks(url, conf)
}

func fetchDoorlocks(url string, conf apiserver.Configuration) ([]DoorLock, error) {
	r, err := http.NewRequestWithApiKey(url, "Authorization", "Bearer "+conf.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", url, err)
	}
	accessPointResponse, err := http.Read[accessPointResponse](r, time.Duration(*conf.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", url, err)
	}
	doorlocks := accessPointResponse.Data
	if accessPointResponse.Links.Next != "" {
		dl, err := fetchDoorlocks(accessPointResponse.Links.Next, conf)
		if err != nil {
			return nil, err
		}
		doorlocks = append(doorlocks, dl...)
	}
	return doorlocks, nil
}
