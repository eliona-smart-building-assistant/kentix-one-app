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
	"fmt"
	"kentix-one/kentix"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

func KentixDevicesDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "KentixONE devices"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	multiSensors, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(kentix.MultiSensorAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching MulitSensor assets: %v", err)
	}

	doorlocks, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(kentix.DoorlockAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching Doorlock assets: %v", err)
	}

	var doorlockButtons []api.WidgetData
	var doorlockBatteries []api.WidgetData
	for _, doorlock := range doorlocks {
		btn := api.WidgetData{
			ElementSequence: nullableInt32(1),
			AssetId:         doorlock.Id,
			Data: map[string]interface{}{
				"aggregatedDataType": "heap",
				"attribute":          "open",
				"description":        doorlock.Name,
				"subtype":            "output",
			},
		}
		doorlockButtons = append(doorlockButtons, btn)

		bat := api.WidgetData{
			ElementSequence: nullableInt32(1),
			AssetId:         doorlock.Id,
			Data: map[string]interface{}{
				"aggregatedDataType": "heap",
				"attribute":          "battery",
				"description":        fmt.Sprintf("%s - battery", doorlock.Name),
				"subtype":            "input",
			},
		}
		doorlockBatteries = append(doorlockBatteries, bat)
	}
	doorsWidget := api.Widget{
		WidgetTypeName: "KentixONE doorlock",
		Details: map[string]interface{}{
			"size":     1,
			"timespan": 7,
		},
		Data: doorlockButtons,
	}
	dashboard.Widgets = append(dashboard.Widgets, doorsWidget)

	doorsBatteryWidget := api.Widget{
		WidgetTypeName: "KentixONE Multisensor",
		Details: map[string]interface{}{
			"size":     1,
			"timespan": 7,
		},
		Data: doorlockBatteries,
	}
	dashboard.Widgets = append(dashboard.Widgets, doorsBatteryWidget)

	for _, multiSensor := range multiSensors {
		widget := api.Widget{
			WidgetTypeName: "KentixONE Multisensor",
			AssetId:        multiSensor.Id,
			Details: map[string]interface{}{
				"size":     2,
				"timespan": 7,
			},
			Data: []api.WidgetData{
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "humidity",
						"description":         "Humidity",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "air_quality",
						"description":         "Air quality",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "co2",
						"description":         "CO₂",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "temperature",
						"description":         "Temperature",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
			},
		}
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}

	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr(val))
}
