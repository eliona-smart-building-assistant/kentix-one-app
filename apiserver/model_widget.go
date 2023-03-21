/*
 * Kentix One app API
 *
 * API to access and configure the Kentix One app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

// Widget - A widget on a frontend dashboard
type Widget struct {

	// The internal Id of widget
	Id *int32 `json:"id,omitempty"`

	// The name for the type of this widget
	WidgetTypeName string `json:"widgetTypeName"`

	// Detailed configuration depending on the widget type
	Details *map[string]interface{} `json:"details,omitempty"`

	// The master asset id of this widget
	AssetId *int32 `json:"assetId,omitempty"`

	// Placement order on dashboard; if not set the index in array is taken
	Sequence *int32 `json:"sequence,omitempty"`

	// List of data for the elements of widget
	Data *[]WidgetData `json:"data,omitempty"`
}

// AssertWidgetRequired checks if the required fields are not zero-ed
func AssertWidgetRequired(obj Widget) error {
	elements := map[string]interface{}{
		"widgetTypeName": obj.WidgetTypeName,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	if obj.Data != nil {
		for _, el := range *obj.Data {
			if err := AssertWidgetDataRequired(el); err != nil {
				return err
			}
		}
	}
	return nil
}

// AssertRecurseWidgetRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Widget (e.g. [][]Widget), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseWidgetRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aWidget, ok := obj.(Widget)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertWidgetRequired(aWidget)
	})
}
