// Code generated by go-swagger; DO NOT EDIT.

package sku_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// NewUpdateSkuParams creates a new UpdateSkuParams object
// no default values defined in spec.
func NewUpdateSkuParams() UpdateSkuParams {

	return UpdateSkuParams{}
}

// UpdateSkuParams contains all the bound params for the update sku operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateSku
type UpdateSkuParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Id of sku to be obtained
	  Required: true
	  In: path
	*/
	ID string
	/*updated sku containing all parameters except id
	  Required: true
	  In: body
	*/
	Sku *models.Sku
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUpdateSkuParams() beforehand.
func (o *UpdateSkuParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.Sku
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("sku", "body", ""))
			} else {
				res = append(res, errors.NewParseError("sku", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Sku = &body
			}
		}
	} else {
		res = append(res, errors.Required("sku", "body", ""))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindID binds and validates parameter ID from path.
func (o *UpdateSkuParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.ID = raw

	return nil
}
