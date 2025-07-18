// Code generated by go-swagger; DO NOT EDIT.

package account_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
)

// CreateAccountCreatedCode is the HTTP code returned for type CreateAccountCreated
const CreateAccountCreatedCode int = 201

/*CreateAccountCreated Account created, provided information of the new item created

swagger:response createAccountCreated
*/
type CreateAccountCreated struct {

	/*
	  In: Body
	*/
	Payload *models.AccountStatus `json:"body,omitempty"`
}

// NewCreateAccountCreated creates CreateAccountCreated with default headers values
func NewCreateAccountCreated() *CreateAccountCreated {

	return &CreateAccountCreated{}
}

// WithPayload adds the payload to the create account created response
func (o *CreateAccountCreated) WithPayload(payload *models.AccountStatus) *CreateAccountCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create account created response
func (o *CreateAccountCreated) SetPayload(payload *models.AccountStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAccountCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAccountConflictCode is the HTTP code returned for type CreateAccountConflict
const CreateAccountConflictCode int = 409

/*CreateAccountConflict The account with the id provided already exist

swagger:response createAccountConflict
*/
type CreateAccountConflict struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewCreateAccountConflict creates CreateAccountConflict with default headers values
func NewCreateAccountConflict() *CreateAccountConflict {

	return &CreateAccountConflict{}
}

// WithPayload adds the payload to the create account conflict response
func (o *CreateAccountConflict) WithPayload(payload *models.ErrorResponse) *CreateAccountConflict {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create account conflict response
func (o *CreateAccountConflict) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAccountConflict) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(409)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAccountInternalServerErrorCode is the HTTP code returned for type CreateAccountInternalServerError
const CreateAccountInternalServerErrorCode int = 500

/*CreateAccountInternalServerError Something unexpected happend, error raised

swagger:response createAccountInternalServerError
*/
type CreateAccountInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.ErrorResponse `json:"body,omitempty"`
}

// NewCreateAccountInternalServerError creates CreateAccountInternalServerError with default headers values
func NewCreateAccountInternalServerError() *CreateAccountInternalServerError {

	return &CreateAccountInternalServerError{}
}

// WithPayload adds the payload to the create account internal server error response
func (o *CreateAccountInternalServerError) WithPayload(payload *models.ErrorResponse) *CreateAccountInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create account internal server error response
func (o *CreateAccountInternalServerError) SetPayload(payload *models.ErrorResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAccountInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
