/*
 * KentixONE app API
 *
 * API to access and configure the KentixONE app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

// Device - Each device represents one asset in Eliona.
type Device struct {

	// The project ID this asset is assigned to
	ProjectID string `json:"projectID,omitempty"`

	Configuration Configuration `json:"configuration,omitempty"`

	// Eliona asset ID
	AssetID *int32 `json:"assetID,omitempty"`

	// Serial number reported by the KentixONE device
	SerialNumber string `json:"serialNumber,omitempty"`
}

// AssertDeviceRequired checks if the required fields are not zero-ed
func AssertDeviceRequired(obj Device) error {
	if err := AssertConfigurationRequired(obj.Configuration); err != nil {
		return err
	}
	return nil
}

// AssertRecurseDeviceRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Device (e.g. [][]Device), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseDeviceRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aDevice, ok := obj.(Device)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertDeviceRequired(aDevice)
	})
}
