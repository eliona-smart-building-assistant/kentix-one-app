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

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func CreateAssetsIfNecessary(config apiserver.Configuration, device kentix.DeviceInfo) error {
	for _, projectId := range conf.ProjIds(config) {
		if err := createDeviceAssetIfNecessary(config, projectId, device); err != nil {
			return fmt.Errorf("creating assets for device %v: %v", device.UUID, err)
		}
	}
	return nil
}

func createDeviceAssetIfNecessary(config apiserver.Configuration, projectId string, device kentix.DeviceInfo) error {
	assetData := assetData{
		config:        config,
		projectId:     projectId,
		parentAssetId: nil,
		identifier:    device.UUID,
		assetType:     device.AssetType,
		name:          fmt.Sprintf("%s (%s)", device.Name, device.IPAddress),
		description:   fmt.Sprintf("%s (%s)", device.Name, device.UUID),
	}
	_, _, err := createAssetIfNecessary(assetData)
	return err
}

func CreateDoorlockAssetsIfNecessary(config apiserver.Configuration, doorlock kentix.DoorLock) error {
	for _, projectId := range conf.ProjIds(config) {
		// TODO: Add parent asset ID?
		if err := createDoorlockAssetIfNecessary(config, projectId, nil, doorlock); err != nil {
			return fmt.Errorf("creating assets for doorlock %s: %v", doorlock.Name, err)
		}
	}
	return nil
}

func createDoorlockAssetIfNecessary(config apiserver.Configuration, projectId string, parentAssetId *int32, doorlock kentix.DoorLock) error {
	assetData := assetData{
		config:        config,
		projectId:     projectId,
		parentAssetId: parentAssetId,
		identifier:    fmt.Sprint(doorlock.ID),
		assetType:     kentix.DoorlockAssetType,
		name:          fmt.Sprintf("%s", doorlock.Name),
		description:   fmt.Sprintf("%s (%v)", doorlock.Name, doorlock.ID),
	}
	created, assetID, err := createAssetIfNecessary(assetData)
	if err != nil {
		return fmt.Errorf("creating doorlock asset: %v", err)
	}
	if !created {
		return nil
	}

	if err := upsertData(
		api.SUBTYPE_OUTPUT,
		assetID,
		doorlockOutputDataPayload{
			Open: 0,
		},
	); err != nil {
		return fmt.Errorf("upserting sample output data: %v", err)
	}
	return nil
}

type assetData struct {
	config        apiserver.Configuration
	projectId     string
	parentAssetId *int32
	identifier    string
	assetType     string
	name          string
	description   string
}

func createAssetIfNecessary(d assetData) (created bool, assetID int32, err error) {
	// Get known asset id from configuration
	currentAssetID, err := conf.GetAssetId(context.Background(), d.config, d.projectId, d.identifier)
	if err != nil {
		return false, 0, fmt.Errorf("finding asset ID: %v", err)
	}
	if currentAssetID != nil {
		return false, *currentAssetID, nil
	}

	a := api.Asset{
		ProjectId:               d.projectId,
		GlobalAssetIdentifier:   d.identifier,
		Name:                    *api.NewNullableString(common.Ptr(d.name)),
		AssetType:               d.assetType,
		Description:             *api.NewNullableString(common.Ptr(d.description)),
		ParentFunctionalAssetId: *api.NewNullableInt32(d.parentAssetId),
	}
	newID, err := asset.UpsertAsset(a)
	if err != nil {
		return false, 0, fmt.Errorf("upserting asset %+v into Eliona: %v", a, err)
	}
	if newID == nil {
		return false, 0, fmt.Errorf("cannot create asset %s", d.name)
	}

	// Remember the asset id for further usage
	if err := conf.InsertDevice(context.Background(), d.config, d.projectId, d.identifier, *newID); err != nil {
		return false, 0, fmt.Errorf("inserting asset to config db: %v", err)
	}

	log.Debug("eliona", "Created new asset for project %s and device %s.", d.projectId, d.identifier)

	return true, *newID, nil
}
