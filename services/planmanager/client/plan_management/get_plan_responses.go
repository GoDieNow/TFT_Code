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

// GetPlanReader is a Reader for the GetPlan structure.
type GetPlanReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPlanReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPlanOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetPlanNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetPlanInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetPlanOK creates a GetPlanOK with default headers values
func NewGetPlanOK() *GetPlanOK {
	return &GetPlanOK{}
}

/*GetPlanOK handles this case with default header values.

plan returned
*/
type GetPlanOK struct {
	Payload *models.Plan
}

func (o *GetPlanOK) Error() string {
	return fmt.Sprintf("[GET /plan/{id}][%d] getPlanOK  %+v", 200, o.Payload)
}

func (o *GetPlanOK) GetPayload() *models.Plan {
	return o.Payload
}

func (o *GetPlanOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Plan)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPlanNotFound creates a GetPlanNotFound with default headers values
func NewGetPlanNotFound() *GetPlanNotFound {
	return &GetPlanNotFound{}
}

/*GetPlanNotFound handles this case with default header values.

plan with planid not found
*/
type GetPlanNotFound struct {
	Payload *models.ErrorResponse
}

func (o *GetPlanNotFound) Error() string {
	return fmt.Sprintf("[GET /plan/{id}][%d] getPlanNotFound  %+v", 404, o.Payload)
}

func (o *GetPlanNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetPlanNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPlanInternalServerError creates a GetPlanInternalServerError with default headers values
func NewGetPlanInternalServerError() *GetPlanInternalServerError {
	return &GetPlanInternalServerError{}
}

/*GetPlanInternalServerError handles this case with default header values.

unexpected error
*/
type GetPlanInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *GetPlanInternalServerError) Error() string {
	return fmt.Sprintf("[GET /plan/{id}][%d] getPlanInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPlanInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetPlanInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
