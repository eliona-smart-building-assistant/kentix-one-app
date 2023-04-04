/*
 * KentixONE app API
 *
 * API to access and configure the KentixONE app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

import (
	"net/http"
	"strings"
)

// VersionApiController binds http requests to an api service and writes the service results to the http response
type VersionApiController struct {
	service      VersionApiServicer
	errorHandler ErrorHandler
}

// VersionApiOption for how the controller is set up.
type VersionApiOption func(*VersionApiController)

// WithVersionApiErrorHandler inject ErrorHandler into controller
func WithVersionApiErrorHandler(h ErrorHandler) VersionApiOption {
	return func(c *VersionApiController) {
		c.errorHandler = h
	}
}

// NewVersionApiController creates a default api controller
func NewVersionApiController(s VersionApiServicer, opts ...VersionApiOption) Router {
	controller := &VersionApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the VersionApiController
func (c *VersionApiController) Routes() Routes {
	return Routes{
		{
			"GetOpenAPI",
			strings.ToUpper("Get"),
			"/v1/version/openapi.json",
			c.GetOpenAPI,
		},
		{
			"GetVersion",
			strings.ToUpper("Get"),
			"/v1/version",
			c.GetVersion,
		},
	}
}

// GetOpenAPI - OpenAPI specification for this API version
func (c *VersionApiController) GetOpenAPI(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.GetOpenAPI(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetVersion - Version of the API
func (c *VersionApiController) GetVersion(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.GetVersion(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
