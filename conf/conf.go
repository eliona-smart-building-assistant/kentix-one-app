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
	"errors"
	"fmt"
	"kentix-one/apiserver"
	"kentix-one/appdb"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrBadRequest = errors.New("bad request")

// InsertConfig inserts or updates
func InsertConfig(ctx context.Context, config apiserver.Configuration) (apiserver.Configuration, error) {
	dbConfig := dbConfigFromApiConfig(config)
	err := dbConfig.InsertG(ctx, boil.Infer())
	if err != nil {
		return apiserver.Configuration{}, err
	}
	return config, err
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
	apiConfig := apiConfigFromDbConfig(dbConfig)
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

func dbConfigFromApiConfig(apiConfig apiserver.Configuration) (dbConfig appdb.Configuration) {
	dbConfig.ID = null.Int64FromPtr(apiConfig.Id).Int64
	dbConfig.Address = null.StringFrom(apiConfig.Address)
	dbConfig.APIKey = null.StringFrom(apiConfig.ApiKey)
	dbConfig.Enable = null.BoolFromPtr(apiConfig.Enable)
	dbConfig.RefreshInterval = apiConfig.RefreshInterval
	if apiConfig.RequestTimeout != nil {
		dbConfig.RequestTimeout = *apiConfig.RequestTimeout
	}
	dbConfig.Active = null.BoolFromPtr(apiConfig.Active)
	if apiConfig.ProjectIDs != nil {
		dbConfig.ProjectIds = *apiConfig.ProjectIDs
	}
	return dbConfig
}

func apiConfigFromDbConfig(dbConfig *appdb.Configuration) (apiConfig apiserver.Configuration) {
	apiConfig.Id = &dbConfig.ID
	apiConfig.Address = dbConfig.Address.String
	apiConfig.ApiKey = dbConfig.APIKey.String
	apiConfig.Enable = dbConfig.Enable.Ptr()
	apiConfig.RefreshInterval = dbConfig.RefreshInterval
	apiConfig.RequestTimeout = &dbConfig.RequestTimeout
	apiConfig.Active = dbConfig.Active.Ptr()
	apiConfig.ProjectIDs = common.Ptr[[]string](dbConfig.ProjectIds)
	return apiConfig
}

func apiDeviceFromDbDevice(ctx context.Context, dbDevice *appdb.Device) (apiDevice apiserver.Device, err error) {
	apiDevice.AssetID = dbDevice.AssetID.Ptr()
	dbConfiguration, err := dbDevice.Configuration().OneG(ctx)
	if err != nil {
		return apiDevice, fmt.Errorf("fetching configuration for sensor: %v", err)
	}
	apiDevice.Configuration = apiConfigFromDbConfig(dbConfiguration)
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
		apiConfigs = append(apiConfigs, apiConfigFromDbConfig(dbConfig))
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
