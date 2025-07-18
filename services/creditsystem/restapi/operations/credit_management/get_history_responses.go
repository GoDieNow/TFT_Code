// Code generated by go-swagger; DO NOT EDIT.

package credit_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
)

// GetHistoryOKCode is the HTTP code returned for type GetHistoryOK
const GetHistoryOKCode int = 200

/*GetHistoryOK Credit status history of the account with the provided id

swagger:response getHistoryOK
*/
type GetHistoryOK struct {

	/*
	  In: Body
	*/
	Payload *models.CreditHistory `json:"body,omitempty"`
}

// NewGetHistoryOK creates GetHistoryOK with default headers values
func NewGetHistoryOK() *GetHistoryOK {

	return &GetHistoryOK{}
}

// WithPayload adds the payload to the get history o k response
func (o *GetHistoryOK) WithPayload(payload *models.CreditHistory) *GetHistoryOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get history o k response
func (o *GetHistoryOK) SetPayload(payload *models.CreditHistory) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetHistoryOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetHistoryNotFoundCode is the HTTP code returned for type GetHistoryNotFound
const GetHistoryNotFoundCode int = 404

/*GetHistoryNotFound The endpoint provided doesn't exist

swagger:response getHistoryNotFound
*/
type GetHistoryNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewGetHistoryNotFound creates GetHistoryNotFound with default headers values
func NewGetHistoryNotFound() *GetHistoryNotFound {

	return &GetHistoryNotFound{}
}

// WithPayload adds the payload to the get history not found response
func (o *GetHistoryNotFound) WithPayload(payload *models.ErrorResponse) *GetHistoryNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get history not found response
func (o *GetHistoryNotFound) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetHistoryNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetHistoryInternalServerErrorCode is the HTTP code returned for type GetHistoryInternalServerError
const GetHistoryInternalServerErrorCode int = 500

/*GetHistoryInternalServerError Something unexpected happend, error raised

swagger:response getHistoryInternalServerError
*/
type GetHistoryInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewGetHistoryInternalServerError creates GetHistoryInternalServerError with default headers values
func NewGetHistoryInternalServerError() *GetHistoryInternalServerError {

	return &GetHistoryInternalServerError{}
}

// WithPayload adds the payload to the get history internal server error response
func (o *GetHistoryInternalServerError) WithPayload(payload *models.ErrorResponse) *GetHistoryInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get history internal server error response
func (o *GetHistoryInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetHistoryInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
