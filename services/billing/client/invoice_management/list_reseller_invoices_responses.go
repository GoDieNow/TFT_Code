// Code generated by go-swagger; DO NOT EDIT.

package invoice_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/billing/models"
)

// ListResellerInvoicesReader is a Reader for the ListResellerInvoices structure.
type ListResellerInvoicesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListResellerInvoicesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListResellerInvoicesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewListResellerInvoicesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListResellerInvoicesOK creates a ListResellerInvoicesOK with default headers values
func NewListResellerInvoicesOK() *ListResellerInvoicesOK {
	return &ListResellerInvoicesOK{}
}

/*ListResellerInvoicesOK handles this case with default header values.

Description of a successfully operation
*/
type ListResellerInvoicesOK struct {
	Payload []*models.Invoice
}

func (o *ListResellerInvoicesOK) Error() string {
	return fmt.Sprintf("[GET /invoice/reseller][%d] listResellerInvoicesOK  %+v", 200, o.Payload)
}

func (o *ListResellerInvoicesOK) GetPayload() []*models.Invoice {
	return o.Payload
}

func (o *ListResellerInvoicesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListResellerInvoicesInternalServerError creates a ListResellerInvoicesInternalServerError with default headers values
func NewListResellerInvoicesInternalServerError() *ListResellerInvoicesInternalServerError {
	return &ListResellerInvoicesInternalServerError{}
}

/*ListResellerInvoicesInternalServerError handles this case with default header values.

Something unexpected happend, error raised
*/
type ListResellerInvoicesInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *ListResellerInvoicesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /invoice/reseller][%d] listResellerInvoicesInternalServerError  %+v", 500, o.Payload)
}

func (o *ListResellerInvoicesInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListResellerInvoicesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
