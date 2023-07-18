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
	"encoding/json"
	"github.com/eliona-smart-building-assistant/go-eliona/utils"
	"fmt"
	"kentix-one/apiserver"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
)

const (
	AccessManagerAssetType = "kentix_one_access_manager"
	AlarmManagerAssetType  = "kentix_one_alarm_manager"
	MultiSensorAssetType   = "kentix_one_multi_sensor"
	DoorlockAssetType      = "kentix_one_doorlock"
)

type systemValuesResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

type DeviceInfo struct {
	Name         string        `json:"name" eliona:"name,filterable"`
	IPAddress    string        `json:"address" eliona:"ip_address,filterable"`
	MacAddress   string        `json:"mac_address" eliona:"mac_address,filterable"`
	Type         int           `json:"type" eliona:"type,filterable"`
	AssetType    string        `json:"-" eliona:"asset_type,filterable"`
	UUID         string        `json:"uuid"`
	Version      string        `json:"version" eliona:"firmware_version,filterable"`
	Measurements *Measurements `json:"measurements"`
}

func (doorlock *DeviceInfo) AdheresToFilter(config apiserver.Configuration) (bool, error) {
	f := apiFilterToCommonFilter(config.AssetFilter)
	fp, err := utils.StructToMap(doorlock)
	if err != nil {
		return false, fmt.Errorf("converting struct to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
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
	Min        *jsonAnythingString `json:"min"`
	Max        *jsonAnythingString `json:"max"`
	Value      *jsonAnythingString `json:"value"`
	Assignment string              `json:"assignment"`
	Status     string              `json:"status"`
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
	var devices []DeviceInfo
	for _, d := range systemValuesResponse.Devices {
		d.AssetType, err = inferAssetType(d.Type)
		if err != nil {
			return nil, fmt.Errorf("inferring asset type for %s: %v", d.IPAddress, err)
		}
		if d.AssetType == DoorlockAssetType {
			// Doorlocks are added differently
			continue
		}
		devices = append(devices, d)
	}
	return devices, nil
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
	ID           int    `json:"id" eliona:"id"`
	DeviceID     int    `json:"device_id" eliona:"device_id"`
	Active       bool   `json:"is_active"`
	Name         string `json:"name" eliona:"name"`
	BatteryLevel int    `json:"battery_level" eliona:"battery"`
}

func (doorlock *DoorLock) AdheresToFilter(config apiserver.Configuration) (bool, error) {
	f := apiFilterToCommonFilter(config.AssetFilter)
	fp, err := utils.StructToMap(doorlock)
	if err != nil {
		return false, fmt.Errorf("converting struct to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
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

//

// KentixONE API for some reason sends the numeric values sometimes as integer,
// sometimes float, sometime string. It seems to be non-deterministic, but it
// certainly is not documented.
// We don't care about the type they send it in that case, so let's make
// everything a string.
type jsonAnythingString string

func (j *jsonAnythingString) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*j = jsonAnythingString(fmt.Sprintf("%v", v))

	return nil
}

func (j *jsonAnythingString) ToStringPtr() *string {
	if j != nil {
		s := string(*j)
		return &s
	}
	return nil
}

func OpenDoorlock(id int, conf apiserver.Configuration) error {
	url, err := url.JoinPath(conf.Address, "api/doorlocks", fmt.Sprint(id), "open")
	if err != nil {
		return fmt.Errorf("appending endpoint to URL: %v", err)
	}
	r, err := http.NewPutRequestWithApiKey(url, nil, "Authorization", "Bearer "+conf.ApiKey)
	if err != nil {
		return fmt.Errorf("creating request to %s: %v", url, err)
	}
	if _, err := http.Read[any](r, time.Duration(*conf.RequestTimeout)*time.Second, false); err != nil {
		return fmt.Errorf("reading response from %s: %v", url, err)
	}
	return nil
}


// To be moved to go-utils.

func structToMap(input interface{}) (map[string]string, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	inputValue := reflect.ValueOf(input)
	inputType := reflect.TypeOf(input)

	if inputValue.Kind() == reflect.Ptr {
		inputValue = inputValue.Elem()
		inputType = inputType.Elem()
	}

	if inputValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}

	output := make(map[string]string)
	for i := 0; i < inputValue.NumField(); i++ {
		fieldType := inputType.Field(i)

		fieldTag, err := parseElionaTag(fieldType)
		if err != nil {
			return nil, err
		}

		if !fieldTag.Filterable {
			continue
		}

		fieldValue := inputValue.Field(i)
		output[fieldTag.ParamName] = fieldValue.String()
	}

	return output, nil
}

type SubType string

const (
	Status SubType = "status"
	Info   SubType = "info"
	Input  SubType = "input"
	Output SubType = "output"
)

type FieldTag struct {
	ParamName  string
	SubType    SubType
	Filterable bool
}

func parseElionaTag(field reflect.StructField) (*FieldTag, error) {
	elionaTag := field.Tag.Get("eliona")
	subtypeTag := field.Tag.Get("subtype")

	elionaTagParts := strings.Split(elionaTag, ",")
	if len(elionaTagParts) < 1 {
		return nil, fmt.Errorf("invalid eliona tag on field %s", field.Name)
	}

	paramName := elionaTagParts[0]
	filterable := len(elionaTagParts) > 1 && elionaTagParts[1] == "filterable"

	var subType SubType
	if subtypeTag != "" {
		subType = SubType(subtypeTag)
		switch subType {
		case Status, Info, Input, Output:
			// valid subtype
		default:
			return nil, fmt.Errorf("invalid subtype in eliona tag on field %s", field.Name)
		}
	}

	return &FieldTag{
		ParamName:  paramName,
		SubType:    subType,
		Filterable: filterable,
	}, nil
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
