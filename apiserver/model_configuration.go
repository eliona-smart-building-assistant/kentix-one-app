/*
 * KentixONE app API
 *
 * API to access and configure the KentixONE app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

// Configuration - Each configuration defines access to one KentixONE manager device and its satellites.
type Configuration struct {

	// Internal identifier for the configured device (created automatically).
	Id *int64 `json:"id,omitempty"`

	// IP or hostname of the KentixONE device
	Address string `json:"address,omitempty"`

	// KentixONE API key
	ApiKey string `json:"apiKey,omitempty"`

	// Flag to enable or disable fetching from this device
	Enable *bool `json:"enable,omitempty"`

	// Interval in seconds for collecting data from device
	RefreshInterval int32 `json:"refreshInterval,omitempty"`

	// Timeout in seconds
	RequestTimeout *int32 `json:"requestTimeout,omitempty"`

	// Array of rules combined by logical OR
	AssetFilter [][]FilterRule `json:"assetFilter,omitempty"`

	// Set to `true` by the app when running and to `false` when app is stopped
	Active *bool `json:"active,omitempty"`

	// List of Eliona project ids for which this device should collect data. For each project id all smart devices are automatically created as an asset in Eliona. The mapping between Eliona is stored as an asset mapping in the KentixONE app.
	ProjectIDs *[]string `json:"projectIDs,omitempty"`
}

// AssertConfigurationRequired checks if the required fields are not zero-ed
func AssertConfigurationRequired(obj Configuration) error {
	if err := AssertRecurseFilterRuleRequired(obj.AssetFilter); err != nil {
		return err
	}
	return nil
}

// AssertRecurseConfigurationRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Configuration (e.g. [][]Configuration), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseConfigurationRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aConfiguration, ok := obj.(Configuration)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertConfigurationRequired(aConfiguration)
	})
}
