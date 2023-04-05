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

package conf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kentix-one/apiserver"
	"kentix-one/appdb"
	"kentix-one/kentix"
	"strconv"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrBadRequest = errors.New("bad request")

// InsertConfig inserts or updates
func InsertConfig(ctx context.Context, config apiserver.Configuration) (apiserver.Configuration, error) {
	dbConfig, err := dbConfigFromApiConfig(config)
	if err != nil {
		return apiserver.Configuration{}, fmt.Errorf("creating DB config from API config: %v", err)
	}
	if err := dbConfig.InsertG(ctx, boil.Infer()); err != nil {
		return apiserver.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	return config, nil
}

func GetConfig(ctx context.Context, configID int64) (*apiserver.Configuration, error) {
	dbConfig, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).OneG(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching config from database")
	}
	if dbConfig == nil {
		return nil, ErrBadRequest
	}
	apiConfig, err := apiConfigFromDbConfig(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("creating API config from DB config: %v", err)
	}
	return &apiConfig, nil
}

func DeleteConfig(ctx context.Context, configID int64) error {
	count, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("fetching config from database")
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID", count)
	}
	if count == 0 {
		return ErrBadRequest
	}
	return nil
}

func dbConfigFromApiConfig(apiConfig apiserver.Configuration) (dbConfig appdb.Configuration, err error) {
	dbConfig.ID = null.Int64FromPtr(apiConfig.Id).Int64
	dbConfig.Address = null.StringFrom(apiConfig.Address)
	dbConfig.APIKey = null.StringFrom(apiConfig.ApiKey)
	dbConfig.Enable = null.BoolFromPtr(apiConfig.Enable)
	dbConfig.RefreshInterval = apiConfig.RefreshInterval
	if apiConfig.RequestTimeout != nil {
		dbConfig.RequestTimeout = *apiConfig.RequestTimeout
	}
	af, err := json.Marshal(apiConfig.AssetFilter)
	if err != nil {
		return appdb.Configuration{}, fmt.Errorf("marshalling assetFilter: %v", err)
	}
	dbConfig.AssetFilter = null.JSONFrom(af)
	dbConfig.Active = null.BoolFromPtr(apiConfig.Active)
	if apiConfig.ProjectIDs != nil {
		dbConfig.ProjectIds = *apiConfig.ProjectIDs
	}
	return dbConfig, nil
}

func apiConfigFromDbConfig(dbConfig *appdb.Configuration) (apiConfig apiserver.Configuration, err error) {
	apiConfig.Id = &dbConfig.ID
	apiConfig.Address = dbConfig.Address.String
	apiConfig.ApiKey = dbConfig.APIKey.String
	apiConfig.Enable = dbConfig.Enable.Ptr()
	apiConfig.RefreshInterval = dbConfig.RefreshInterval
	apiConfig.RequestTimeout = &dbConfig.RequestTimeout
	if dbConfig.AssetFilter.Valid {
		var af [][]apiserver.FilterRule
		if err := json.Unmarshal(dbConfig.AssetFilter.JSON, &af); err != nil {
			return apiserver.Configuration{}, fmt.Errorf("unmarshalling assetFilter: %v", err)
		}
		apiConfig.AssetFilter = af
	}
	apiConfig.Active = dbConfig.Active.Ptr()
	apiConfig.ProjectIDs = common.Ptr[[]string](dbConfig.ProjectIds)
	return apiConfig, nil
}

func apiDeviceFromDbDevice(ctx context.Context, dbDevice *appdb.Device) (apiDevice apiserver.Device, err error) {
	apiDevice.AssetID = dbDevice.AssetID.Ptr()
	dbConfiguration, err := dbDevice.Configuration().OneG(ctx)
	if err != nil {
		return apiDevice, fmt.Errorf("fetching configuration for sensor: %v", err)
	}
	apiDevice.Configuration, err = apiConfigFromDbConfig(dbConfiguration)
	if err != nil {
		return apiDevice, fmt.Errorf("creating API config from DB config: %v", err)
	}
	apiDevice.ProjectID = dbDevice.ProjectID
	apiDevice.SerialNumber = dbDevice.SerialNumber
	return apiDevice, nil
}

func GetConfigs(ctx context.Context) ([]apiserver.Configuration, error) {
	dbConfigs, err := appdb.Configurations().AllG(ctx)
	if err != nil {
		return nil, err
	}
	var apiConfigs []apiserver.Configuration
	for _, dbConfig := range dbConfigs {
		dbConfig.R.GetDevices()
		ac, err := apiConfigFromDbConfig(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("creating API config from DB config: %v", err)
		}
		apiConfigs = append(apiConfigs, ac)
	}
	return apiConfigs, nil
}

func GetConfigDevices(ctx context.Context, config apiserver.Configuration) ([]apiserver.Device, error) {
	if config.Id == nil {
		return nil, fmt.Errorf("shouldn't happen: config ID is null")
	}
	dbDevices, err := appdb.Devices(
		appdb.DeviceWhere.ConfigurationID.EQ(*config.Id),
	).AllG(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up sensors in DB: %v", err)
	}
	if len(dbDevices) == 0 {
		return nil, fmt.Errorf("no sensor found for config %v", config.Id)
	}
	var apiDevices []apiserver.Device
	for _, dbDevice := range dbDevices {
		s, err := apiDeviceFromDbDevice(ctx, dbDevice)
		if err != nil {
			return nil, fmt.Errorf("creating API sensor from DB sensor: %v", err)
		}
		apiDevices = append(apiDevices, s)
	}
	return apiDevices, nil
}

func GetDeviceId(ctx context.Context, assetID int32) (int, error) {
	dbDevice, err := appdb.Devices(
		appdb.DeviceWhere.AssetID.EQ(null.Int32From(assetID)),
	).OneG(ctx)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(dbDevice.SerialNumber)
	if err != nil {
		return 0, fmt.Errorf("parsing id %s: %v", dbDevice.SerialNumber, err)
	}
	return id, nil
}

func GetAssetId(ctx context.Context, config apiserver.Configuration, projId string, deviceId string) (*int32, error) {
	dbDevices, err := appdb.Devices(
		appdb.DeviceWhere.ConfigurationID.EQ(null.Int64FromPtr(config.Id).Int64),
		appdb.DeviceWhere.ProjectID.EQ(projId),
		appdb.DeviceWhere.SerialNumber.EQ(deviceId),
	).AllG(ctx)
	if err != nil || len(dbDevices) == 0 {
		return nil, err
	}
	return common.Ptr(dbDevices[0].AssetID.Int32), nil
}

func InsertDevice(ctx context.Context, config apiserver.Configuration, projId string, SerialNumber string, assetId int32) error {
	var dbDevice appdb.Device
	dbDevice.ConfigurationID = null.Int64FromPtr(config.Id).Int64
	dbDevice.ProjectID = projId
	dbDevice.SerialNumber = SerialNumber
	dbDevice.AssetID = null.Int32From(assetId)
	return dbDevice.InsertG(ctx, boil.Infer())
}

func SetConfigActiveState(ctx context.Context, config apiserver.Configuration, state bool) (int64, error) {
	return appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(null.Int64FromPtr(config.Id).Int64),
	).UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: state,
	})
}

func ProjIds(config apiserver.Configuration) []string {
	if config.ProjectIDs == nil {
		return []string{}
	}
	return *config.ProjectIDs
}

func IsConfigActive(config apiserver.Configuration) bool {
	return config.Active == nil || *config.Active
}

func IsConfigEnabled(config apiserver.Configuration) bool {
	return config.Enable == nil || *config.Enable
}

func SetAllConfigsInactive(ctx context.Context) (int64, error) {
	return appdb.Configurations().UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: false,
	})
}

func DoesDeviceAdhereToFilter(ctx context.Context, config apiserver.Configuration, device kentix.DeviceInfo) (bool, error) {
	fp := map[string]string{
		"deviceName":    device.Name,
		"ipAddress":     device.IPAddress,
		"macAddress":    device.MacAddress,
		"deviceSerial":  device.UUID,
		"deviceType":    fmt.Sprint(device.Type),
		"deviceVersion": device.Version,
	}
	adheres, err := common.Filter(apiFilterToCommonFilter(config.AssetFilter), fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func DoesDoorlockAdhereToFilter(ctx context.Context, config apiserver.Configuration, doorlock kentix.DoorLock) (bool, error) {
	fp := map[string]string{
		"doorlockName":   doorlock.Name,
		"doorlockSerial": fmt.Sprint(doorlock.ID),
	}
	adheres, err := common.Filter(apiFilterToCommonFilter(config.AssetFilter), fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func apiFilterToCommonFilter(input [][]apiserver.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}
