// Code generated by go-swagger; DO NOT EDIT.

package reseller_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
)

// UpdateResellerOKCode is the HTTP code returned for type UpdateResellerOK
const UpdateResellerOKCode int = 200

/*UpdateResellerOK Reseller with the given id was updated

swagger:response updateResellerOK
*/
type UpdateResellerOK struct {

	/*
	  In: Body
	*/
	Payload *models.ItemCreatedResponse `json:"body,omitempty"`
}

// NewUpdateResellerOK creates UpdateResellerOK with default headers values
func NewUpdateResellerOK() *UpdateResellerOK {

	return &UpdateResellerOK{}
}

// WithPayload adds the payload to the update reseller o k response
func (o *UpdateResellerOK) WithPayload(payload *models.ItemCreatedResponse) *UpdateResellerOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update reseller o k response
func (o *UpdateResellerOK) SetPayload(payload *models.ItemCreatedResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateResellerOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateResellerAcceptedCode is the HTTP code returned for type UpdateResellerAccepted
const UpdateResellerAcceptedCode int = 202

/*UpdateResellerAccepted The reseller was updated but there might have been some fails when adding part of the data

swagger:response updateResellerAccepted
*/
type UpdateResellerAccepted struct {

	/*
	  In: Body
	*/
	Payload *models.ItemCreatedResponse `json:"body,omitempty"`
}

// NewUpdateResellerAccepted creates UpdateResellerAccepted with default headers values
func NewUpdateResellerAccepted() *UpdateResellerAccepted {

	return &UpdateResellerAccepted{}
}

// WithPayload adds the payload to the update reseller accepted response
func (o *UpdateResellerAccepted) WithPayload(payload *models.ItemCreatedResponse) *UpdateResellerAccepted {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update reseller accepted response
func (o *UpdateResellerAccepted) SetPayload(payload *models.ItemCreatedResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateResellerAccepted) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(202)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateResellerNotFoundCode is the HTTP code returned for type UpdateResellerNotFound
const UpdateResellerNotFoundCode int = 404

/*UpdateResellerNotFound The reseller with the given id wasn't found

swagger:response updateResellerNotFound
*/
type UpdateResellerNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewUpdateResellerNotFound creates UpdateResellerNotFound with default headers values
func NewUpdateResellerNotFound() *UpdateResellerNotFound {

	return &UpdateResellerNotFound{}
}

// WithPayload adds the payload to the update reseller not found response
func (o *UpdateResellerNotFound) WithPayload(payload *models.ErrorResponse) *UpdateResellerNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update reseller not found response
func (o *UpdateResellerNotFound) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateResellerNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UpdateResellerInternalServerErrorCode is the HTTP code returned for type UpdateResellerInternalServerError
const UpdateResellerInternalServerErrorCode int = 500

/*UpdateResellerInternalServerError Something unexpected happend, error raised

swagger:response updateResellerInternalServerError
*/
type UpdateResellerInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewUpdateResellerInternalServerError creates UpdateResellerInternalServerError with default headers values
func NewUpdateResellerInternalServerError() *UpdateResellerInternalServerError {

	return &UpdateResellerInternalServerError{}
}

// WithPayload adds the payload to the update reseller internal server error response
func (o *UpdateResellerInternalServerError) WithPayload(payload *models.ErrorResponse) *UpdateResellerInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the update reseller internal server error response
func (o *UpdateResellerInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UpdateResellerInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
