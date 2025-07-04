// Code generated by go-swagger; DO NOT EDIT.

package sku_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
)

// GetSkuReader is a Reader for the GetSku structure.
type GetSkuReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetSkuReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetSkuOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetSkuNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetSkuInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetSkuOK creates a GetSkuOK with default headers values
func NewGetSkuOK() *GetSkuOK {
	return &GetSkuOK{}
}

/*GetSkuOK handles this case with default header values.

sku returned
*/
type GetSkuOK struct {
	Payload *models.Sku
}

func (o *GetSkuOK) Error() string {
	return fmt.Sprintf("[GET /sku/{id}][%d] getSkuOK  %+v", 200, o.Payload)
}

func (o *GetSkuOK) GetPayload() *models.Sku {
	return o.Payload
}

func (o *GetSkuOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Sku)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSkuNotFound creates a GetSkuNotFound with default headers values
func NewGetSkuNotFound() *GetSkuNotFound {
	return &GetSkuNotFound{}
}

/*GetSkuNotFound handles this case with default header values.

sku with skuid not found
*/
type GetSkuNotFound struct {
	Payload *models.ErrorResponse
}

func (o *GetSkuNotFound) Error() string {
	return fmt.Sprintf("[GET /sku/{id}][%d] getSkuNotFound  %+v", 404, o.Payload)
}

func (o *GetSkuNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetSkuNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSkuInternalServerError creates a GetSkuInternalServerError with default headers values
func NewGetSkuInternalServerError() *GetSkuInternalServerError {
	return &GetSkuInternalServerError{}
}

/*GetSkuInternalServerError handles this case with default header values.

unexpected error
*/
type GetSkuInternalServerError struct {
	Payload *models.ErrorResponse
}

func (o *GetSkuInternalServerError) Error() string {
	return fmt.Sprintf("[GET /sku/{id}][%d] getSkuInternalServerError  %+v", 500, o.Payload)
}

func (o *GetSkuInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetSkuInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
