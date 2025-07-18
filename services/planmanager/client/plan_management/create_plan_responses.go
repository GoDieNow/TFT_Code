// Code generated by go-swagger; DO NOT EDIT.

package plan_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// CreatePlanReader is a Reader for the CreatePlan structure.
type CreatePlanReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreatePlanReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreatePlanCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreatePlanBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewCreatePlanConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewCreatePlanInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewCreatePlanCreated creates a CreatePlanCreated with default headers values
func NewCreatePlanCreated() *CreatePlanCreated {
	return &CreatePlanCreated{}
}

/*CreatePlanCreated handles this case with default header values.

item created
*/
type CreatePlanCreated struct {
	Payload *models.ItemCreatedResponse
}

func (o *CreatePlanCreated) Error() string {
	return fmt.Sprintf("[POST /plan][%d] createPlanCreated  %+v", 201, o.Payload)
}

func (o *CreatePlanCreated) GetPayload() *models.ItemCreatedResponse {
	return o.Payload
}

func (o *CreatePlanCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ItemCreatedResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreatePlanBadRequest creates a CreatePlanBadRequest with default headers values
func NewCreatePlanBadRequest() *CreatePlanBadRequest {
	return &CreatePlanBadRequest{}
}

/*CreatePlanBadRequest handles this case with default header values.

invalid input, object invalid
*/
type CreatePlanBadRequest struct {
}

func (o *CreatePlanBadRequest) Error() string {
	return fmt.Sprintf("[POST /plan][%d] createPlanBadRequest ", 400)
}

func (o *CreatePlanBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewCreatePlanConflict creates a CreatePlanConflict with default headers values
func NewCreatePlanConflict() *CreatePlanConflict {
	return &CreatePlanConflict{}
}

/*CreatePlanConflict handles this case with default header values.

an existing item already exists
*/
type CreatePlanConflict struct {
	Payload *models.ErrorResponse
}

func (o *CreatePlanConflict) Error() string {
	return fmt.Sprintf("[POST /plan][%d] createPlanConflict  %+v", 409, o.Payload)
}

func (o *CreatePlanConflict) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreatePlanConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreatePlanInternalServerError creates a CreatePlanInternalServerError with default headers values
func NewCreatePlanInternalServerError() *CreatePlanInternalServerError {
	return &CreatePlanInternalServerError{}
}

/*CreatePlanInternalServerError handles this case with default header values.

unexpected error
*/
type CreatePlanInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *CreatePlanInternalServerError) Error() string {
	return fmt.Sprintf("[POST /plan][%d] createPlanInternalServerError  %+v", 500, o.Payload)
}

func (o *CreatePlanInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreatePlanInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
