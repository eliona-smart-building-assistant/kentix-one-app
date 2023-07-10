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
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/eliona-smart-building-assistant/app-integration-tests/assert"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestApp(t *testing.T) {
	app.StartApp()
	test.AppWorks(t)
	t.Run("TestAssetTypes", assetTypes)
	t.Run("TestWidgetTypes", widgetTypes)
	t.Run("TestSchema", schema)
	app.StopApp()
}

func assetTypes(t *testing.T) {
	t.Parallel()

	assert.AssetTypeExists(t, "kentix_one_access_manager", []string{"name", "ip_address", "mac_address", "firmware_version"})
	assert.AssetTypeExists(t, "kentix_one_alarm_manager", []string{"name", "ip_address", "mac_address", "firmware_version"})
	assert.AssetTypeExists(t, "kentix_one_doorlock", []string{"id", "device_id", "name", "battery", "open"})
	assert.AssetTypeExists(t, "kentix_one_multi_sensor", []string{"name", "ip_address", "mac_address", "firmware_version", "last_updated", "temperature", "humidity", "dew_point", "air_quality", "co2", "co", "heat", "motion", "vibration", "people_count"})
}

func widgetTypes(t *testing.T) {
	t.Parallel()

	assert.WidgetTypeExists(t, "kentix_one_doolock")
	assert.WidgetTypeExists(t, "kentix_one_multi_sensor")
}

func schema(t *testing.T) {
	t.Parallel()

	assert.SchemaExists(t, "kentix_one", []string{"configuration", "device"})
}
