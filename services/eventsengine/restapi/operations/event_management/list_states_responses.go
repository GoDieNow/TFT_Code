// Code generated by go-swagger; DO NOT EDIT.

package event_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/eventsengine/models"
)

// ListStatesOKCode is the HTTP code returned for type ListStatesOK
const ListStatesOKCode int = 200

/*ListStatesOK Description of a successfully operation

swagger:response listStatesOK
*/
type ListStatesOK struct {

	/*
	  In: Body
	*/
	Payload []*models.MinimalState `json:"body,omitempty"`
}

// NewListStatesOK creates ListStatesOK with default headers values
func NewListStatesOK() *ListStatesOK {

	return &ListStatesOK{}
}

// WithPayload adds the payload to the list states o k response
func (o *ListStatesOK) WithPayload(payload []*models.MinimalState) *ListStatesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list states o k response
func (o *ListStatesOK) SetPayload(payload []*models.MinimalState) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListStatesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.MinimalState, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// ListStatesInternalServerErrorCode is the HTTP code returned for type ListStatesInternalServerError
const ListStatesInternalServerErrorCode int = 500

/*ListStatesInternalServerError Something unexpected happend, error raised

swagger:response listStatesInternalServerError
*/
type ListStatesInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewListStatesInternalServerError creates ListStatesInternalServerError with default headers values
func NewListStatesInternalServerError() *ListStatesInternalServerError {

	return &ListStatesInternalServerError{}
}

// WithPayload adds the payload to the list states internal server error response
func (o *ListStatesInternalServerError) WithPayload(payload *models.ErrorResponse) *ListStatesInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list states internal server error response
func (o *ListStatesInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListStatesInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
