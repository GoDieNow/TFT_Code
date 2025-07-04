// Code generated by go-swagger; DO NOT EDIT.

package account_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
)

// ListAccountsReader is a Reader for the ListAccounts structure.
type ListAccountsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListAccountsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListAccountsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewListAccountsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListAccountsOK creates a ListAccountsOK with default headers values
func NewListAccountsOK() *ListAccountsOK {
	return &ListAccountsOK{}
}

/*ListAccountsOK handles this case with default header values.

List of accounts in the system
*/
type ListAccountsOK struct {
	Payload []*models.AccountStatus
}

func (o *ListAccountsOK) Error() string {
	return fmt.Sprintf("[GET /account/list][%d] listAccountsOK  %+v", 200, o.Payload)
}

func (o *ListAccountsOK) GetPayload() []*models.AccountStatus {
	return o.Payload
}

func (o *ListAccountsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListAccountsInternalServerError creates a ListAccountsInternalServerError with default headers values
func NewListAccountsInternalServerError() *ListAccountsInternalServerError {
	return &ListAccountsInternalServerError{}
}

/*ListAccountsInternalServerError handles this case with default header values.

Something unexpected happend, error raised
*/
type ListAccountsInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *ListAccountsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /account/list][%d] listAccountsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListAccountsInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ListAccountsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
