// Code generated by go-swagger; DO NOT EDIT.

package bundle_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// ListSkuBundlesOKCode is the HTTP code returned for type ListSkuBundlesOK
const ListSkuBundlesOKCode int = 200

/*ListSkuBundlesOK list of skus bundles returned

swagger:response listSkuBundlesOK
*/
type ListSkuBundlesOK struct {

	/*
	  In: Body
	*/
	Payload []*models.SkuBundle `json:"body,omitempty"`
}

// NewListSkuBundlesOK creates ListSkuBundlesOK with default headers values
func NewListSkuBundlesOK() *ListSkuBundlesOK {

	return &ListSkuBundlesOK{}
}

// WithPayload adds the payload to the list sku bundles o k response
func (o *ListSkuBundlesOK) WithPayload(payload []*models.SkuBundle) *ListSkuBundlesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list sku bundles o k response
func (o *ListSkuBundlesOK) SetPayload(payload []*models.SkuBundle) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListSkuBundlesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.SkuBundle, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// ListSkuBundlesInternalServerErrorCode is the HTTP code returned for type ListSkuBundlesInternalServerError
const ListSkuBundlesInternalServerErrorCode int = 500

/*ListSkuBundlesInternalServerError unexpected error

swagger:response listSkuBundlesInternalServerError
*/
type ListSkuBundlesInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewListSkuBundlesInternalServerError creates ListSkuBundlesInternalServerError with default headers values
func NewListSkuBundlesInternalServerError() *ListSkuBundlesInternalServerError {

	return &ListSkuBundlesInternalServerError{}
}

// WithPayload adds the payload to the list sku bundles internal server error response
func (o *ListSkuBundlesInternalServerError) WithPayload(payload *models.ErrorResponse) *ListSkuBundlesInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list sku bundles internal server error response
func (o *ListSkuBundlesInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListSkuBundlesInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
