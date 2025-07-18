// Code generated by go-swagger; DO NOT EDIT.

package sku_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// NewCreateSkuParams creates a new CreateSkuParams object
// no default values defined in spec.
func NewCreateSkuParams() CreateSkuParams {

	return CreateSkuParams{}
}

// CreateSkuParams contains all the bound params for the create sku operation
// typically these are obtained from a http.Request
//
// swagger:parameters createSku
type CreateSkuParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*SKU to be added
	  In: body
	*/
	Sku *models.Sku
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewCreateSkuParams() beforehand.
func (o *CreateSkuParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.Sku
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("sku", "body", "", err))
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Sku = &body
			}
		}
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
