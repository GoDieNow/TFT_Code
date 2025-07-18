// Code generated by go-swagger; DO NOT EDIT.

package credit_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
)

// IncreaseCreditReader is a Reader for the IncreaseCredit structure.
type IncreaseCreditReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *IncreaseCreditReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewIncreaseCreditOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewIncreaseCreditNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewIncreaseCreditInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewIncreaseCreditOK creates a IncreaseCreditOK with default headers values
func NewIncreaseCreditOK() *IncreaseCreditOK {
	return &IncreaseCreditOK{}
}

/*IncreaseCreditOK handles this case with default header values.

Credit status of the account with the provided id
*/
type IncreaseCreditOK struct {
	Payload *models.CreditStatus
}

func (o *IncreaseCreditOK) Error() string {
	return fmt.Sprintf("[POST /{medium}/available/increase/{id}][%d] increaseCreditOK  %+v", 200, o.Payload)
}

func (o *IncreaseCreditOK) GetPayload() *models.CreditStatus {
	return o.Payload
}

func (o *IncreaseCreditOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.CreditStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIncreaseCreditNotFound creates a IncreaseCreditNotFound with default headers values
func NewIncreaseCreditNotFound() *IncreaseCreditNotFound {
	return &IncreaseCreditNotFound{}
}

/*IncreaseCreditNotFound handles this case with default header values.

The account with the id provided doesn't exist
*/
type IncreaseCreditNotFound struct {
	Payload *models.ErrorResponse
}

func (o *IncreaseCreditNotFound) Error() string {
	return fmt.Sprintf("[POST /{medium}/available/increase/{id}][%d] increaseCreditNotFound  %+v", 404, o.Payload)
}

func (o *IncreaseCreditNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *IncreaseCreditNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIncreaseCreditInternalServerError creates a IncreaseCreditInternalServerError with default headers values
func NewIncreaseCreditInternalServerError() *IncreaseCreditInternalServerError {
	return &IncreaseCreditInternalServerError{}
}

/*IncreaseCreditInternalServerError handles this case with default header values.

Something unexpected happend, error raised
*/
type IncreaseCreditInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *IncreaseCreditInternalServerError) Error() string {
	return fmt.Sprintf("[POST /{medium}/available/increase/{id}][%d] increaseCreditInternalServerError  %+v", 500, o.Payload)
}

func (o *IncreaseCreditInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *IncreaseCreditInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
