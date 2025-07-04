// Code generated by go-swagger; DO NOT EDIT.

package bundle_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// NewCreateSkuBundleParams creates a new CreateSkuBundleParams object
// no default values defined in spec.
func NewCreateSkuBundleParams() CreateSkuBundleParams {

	return CreateSkuBundleParams{}
}

// CreateSkuBundleParams contains all the bound params for the create sku bundle operation
// typically these are obtained from a http.Request
//
// swagger:parameters createSkuBundle
type CreateSkuBundleParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*SKU bundle to be added
	  In: body
	*/
	Bundle *models.SkuBundle
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewCreateSkuBundleParams() beforehand.
func (o *CreateSkuBundleParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.SkuBundle
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("bundle", "body", "", err))
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Bundle = &body
			}
		}
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
