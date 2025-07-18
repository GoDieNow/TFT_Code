// Code generated by go-swagger; DO NOT EDIT.

package reseller_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
)

// AddResellerCreatedCode is the HTTP code returned for type AddResellerCreated
const AddResellerCreatedCode int = 201

/*AddResellerCreated New reseller was added successfully

swagger:response addResellerCreated
*/
type AddResellerCreated struct {

	/*
	  In: Body
	*/
	Payload *models.ItemCreatedResponse `json:"body,omitempty"`
}

// NewAddResellerCreated creates AddResellerCreated with default headers values
func NewAddResellerCreated() *AddResellerCreated {

	return &AddResellerCreated{}
}

// WithPayload adds the payload to the add reseller created response
func (o *AddResellerCreated) WithPayload(payload *models.ItemCreatedResponse) *AddResellerCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the add reseller created response
func (o *AddResellerCreated) SetPayload(payload *models.ItemCreatedResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AddResellerCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// AddResellerAcceptedCode is the HTTP code returned for type AddResellerAccepted
const AddResellerAcceptedCode int = 202

/*AddResellerAccepted The new reseller was added but there might have been some fails when adding part of the data

swagger:response addResellerAccepted
*/
type AddResellerAccepted struct {

	/*
	  In: Body
	*/
	Payload *models.ItemCreatedResponse `json:"body,omitempty"`
}

// NewAddResellerAccepted creates AddResellerAccepted with default headers values
func NewAddResellerAccepted() *AddResellerAccepted {

	return &AddResellerAccepted{}
}

// WithPayload adds the payload to the add reseller accepted response
func (o *AddResellerAccepted) WithPayload(payload *models.ItemCreatedResponse) *AddResellerAccepted {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the add reseller accepted response
func (o *AddResellerAccepted) SetPayload(payload *models.ItemCreatedResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AddResellerAccepted) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(202)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// AddResellerBadRequestCode is the HTTP code returned for type AddResellerBadRequest
const AddResellerBadRequestCode int = 400

/*AddResellerBadRequest Invalid input, object invalid

swagger:response addResellerBadRequest
*/
type AddResellerBadRequest struct {
}

// NewAddResellerBadRequest creates AddResellerBadRequest with default headers values
func NewAddResellerBadRequest() *AddResellerBadRequest {

	return &AddResellerBadRequest{}
}

// WriteResponse to the client
func (o *AddResellerBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// AddResellerConflictCode is the HTTP code returned for type AddResellerConflict
const AddResellerConflictCode int = 409

/*AddResellerConflict The given item already exists

swagger:response addResellerConflict
*/
type AddResellerConflict struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewAddResellerConflict creates AddResellerConflict with default headers values
func NewAddResellerConflict() *AddResellerConflict {

	return &AddResellerConflict{}
}

// WithPayload adds the payload to the add reseller conflict response
func (o *AddResellerConflict) WithPayload(payload *models.ErrorResponse) *AddResellerConflict {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the add reseller conflict response
func (o *AddResellerConflict) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AddResellerConflict) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(409)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// AddResellerInternalServerErrorCode is the HTTP code returned for type AddResellerInternalServerError
const AddResellerInternalServerErrorCode int = 500

/*AddResellerInternalServerError Something unexpected happend, error raised

swagger:response addResellerInternalServerError
*/
type AddResellerInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewAddResellerInternalServerError creates AddResellerInternalServerError with default headers values
func NewAddResellerInternalServerError() *AddResellerInternalServerError {

	return &AddResellerInternalServerError{}
}

// WithPayload adds the payload to the add reseller internal server error response
func (o *AddResellerInternalServerError) WithPayload(payload *models.ErrorResponse) *AddResellerInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the add reseller internal server error response
func (o *AddResellerInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *AddResellerInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
