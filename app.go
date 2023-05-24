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

package main

import (
	"context"
	"kentix-one/apiserver"
	"kentix-one/apiservices"
	"kentix-one/conf"
	"kentix-one/eliona"
	"kentix-one/kentix"
	"net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		log.Info("conf", "No configs in DB")
		return
	}

	for _, config := range configs {
		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				conf.SetConfigActiveState(context.Background(), config, false)
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			conf.SetConfigActiveState(context.Background(), config, true)
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
				"Address: %s\n"+
				"API Key: %s\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Active: %t\n"+
				"Project IDs: %v\n",
				*config.Id,
				config.Address,
				"**************",
				*config.Enable,
				config.RefreshInterval,
				*config.RequestTimeout,
				*config.Active,
				*config.ProjectIDs)
		}

		// Runs the ReadNode. If the current node is currently running, skip the execution
		// After the execution sleeps the configured timeout. During this timeout no further
		// process for this config is started to read the data.
		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Collecting %d started", *config.Id)

			collectDataForConfig(config)

			log.Info("main", "Collecting %d finished", *config.Id)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}
}

func collectDataForConfig(config apiserver.Configuration) {
	devices, err := kentix.GetDevices(config)
	if err != nil {
		log.Error("kentix", "getting devices info: %v", err)
		return
	}
	for _, device := range devices {
		if config.AssetFilter != nil {
			shouldCreate, err := conf.DoesDeviceAdhereToFilter(context.Background(), config, device)
			if err != nil {
				log.Error("config", "determining whether device adheres to filter: %v", err)
				return
			}
			if !shouldCreate {
				continue
			}
		}

		if err := eliona.CreateAssetsIfNecessary(config, device); err != nil {
			log.Error("eliona", "creating assets: %v", err)
			return
		}

		if err := eliona.UpsertDeviceInfo(config, device); err != nil {
			log.Error("eliona", "inserting device info: %v", err)
			return
		}

		if device.Measurements != nil {
			if err := eliona.UpsertMultiSensorData(config, device.UUID, *device.Measurements); err != nil {
				log.Error("eliona", "upserting MultiSensor data: %v", err)
				return
			}
		}
	}

	doorlocks, err := kentix.GetDoorlocks(config)
	if err != nil {
		log.Error("kentix", "getting doorlocks info: %v", err)
		return
	}
	for _, doorlock := range doorlocks {
		if config.AssetFilter != nil {
			shouldCreate, err := conf.DoesDoorlockAdhereToFilter(context.Background(), config, doorlock)
			if err != nil {
				log.Error("config", "determining whether doorlock adheres to filter: %v", err)
				return
			}
			if !shouldCreate {
				continue
			}
		}
		if err := eliona.CreateDoorlockAssetsIfNecessary(config, doorlock); err != nil {
			log.Error("eliona", "creating assets: %v", err)
			return
		}

		if err := eliona.UpsertDoorlockData(config, doorlock); err != nil {
			log.Error("eliona", "inserting doorlock data: %v", err)
			return
		}
	}
}

func listenForOutputChanges() {
	outputs, err := eliona.ListenForOutputChanges()
	if err != nil {
		log.Error("eliona", "listening for output changes: %v", err)
		return
	}
	for output := range outputs {
		open, ok := output.Data["open"]
		if !ok {
			log.Error("eliona", "no 'open' attribute in data")
			return
		}
		switch open.(float64) {
		case 0:
			continue
		case 1:
			openDoorlock(output.AssetId)
		default:
			log.Error("eliona", "invalid value '%v' in 'open' attribute", open)
		}
	}
}

func openDoorlock(assetID int32) {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		log.Info("conf", "No configs in DB")
		return
	}
	for _, config := range configs {
		doorlockID, err := conf.GetDeviceId(context.Background(), assetID)
		if err != nil {
			log.Error("conf", "getting device ID from asset (%v) id: %v", assetID, err)
			return
		}
		log.Debug("kentix", "opening doorlock %v", doorlockID)
		if err := kentix.OpenDoorlock(doorlockID, config); err != nil {
			log.Error("kentix", "opening doorlock %v: %v", doorlockID, err)
			return
		}
	}
}

// listenApiRequests starts an API server and listen for API requests
// The API endpoints are defined in the openapi.yaml file
func listenApiRequests() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
		apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
		apiserver.NewCustomizationApiController(apiservices.NewCustomizationApiService()),
	))
	log.Fatal("main", "API server: %v", err)
}
