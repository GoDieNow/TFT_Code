// Code generated by go-swagger; DO NOT EDIT.

package status_management

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/billing/models"
)

// ShowStatusReader is a Reader for the ShowStatus structure.
type ShowStatusReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ShowStatusReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewShowStatusOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewShowStatusOK creates a ShowStatusOK with default headers values
func NewShowStatusOK() *ShowStatusOK {
	return &ShowStatusOK{}
}

/*ShowStatusOK handles this case with default header values.

Status information of the system
*/
type ShowStatusOK struct {
	Payload *models.Status
}

func (o *ShowStatusOK) Error() string {
	return fmt.Sprintf("[GET /status][%d] showStatusOK  %+v", 200, o.Payload)
}

func (o *ShowStatusOK) GetPayload() *models.Status {
	return o.Payload
}

func (o *ShowStatusOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Status)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
