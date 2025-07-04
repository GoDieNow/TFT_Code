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

// ListPlansReader is a Reader for the ListPlans structure.
type ListPlansReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListPlansReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListPlansOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewListPlansInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListPlansOK creates a ListPlansOK with default headers values
func NewListPlansOK() *ListPlansOK {
	return &ListPlansOK{}
}

/*ListPlansOK handles this case with default header values.

list of plans returned
*/
type ListPlansOK struct {
	Payload []*models.Plan
}

func (o *ListPlansOK) Error() string {
	return fmt.Sprintf("[GET /plan][%d] listPlansOK  %+v", 200, o.Payload)
}

func (o *ListPlansOK) GetPayload() []*models.Plan {
	return o.Payload
}

func (o *ListPlansOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListPlansInternalServerError creates a ListPlansInternalServerError with default headers values
func NewListPlansInternalServerError() *ListPlansInternalServerError {
	return &ListPlansInternalServerError{}
}

/*ListPlansInternalServerError handles this case with default header values.

unexpected error
*/
type ListPlansInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *ListPlansInternalServerError) Error() string {
	return fmt.Sprintf("[GET /plan][%d] listPlansInternalServerError  %+v", 500, o.Payload)
}

func (o *ListPlansInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListPlansInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
