// Code generated by go-swagger; DO NOT EDIT.

package dags

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/dagu-org/dagu/internal/frontend/gen/models"
)

// PostDagActionOKCode is the HTTP code returned for type PostDagActionOK
const PostDagActionOKCode int = 200

/*
PostDagActionOK A successful response.

swagger:response postDagActionOK
*/
type PostDagActionOK struct {

	/*
	  In: Body
	*/
	Payload *models.PostDagActionResponse `json:"body,omitempty"`
}

// NewPostDagActionOK creates PostDagActionOK with default headers values
func NewPostDagActionOK() *PostDagActionOK {

	return &PostDagActionOK{}
}

// WithPayload adds the payload to the post dag action o k response
func (o *PostDagActionOK) WithPayload(payload *models.PostDagActionResponse) *PostDagActionOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post dag action o k response
func (o *PostDagActionOK) SetPayload(payload *models.PostDagActionResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostDagActionOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*
PostDagActionDefault Generic error response.

swagger:response postDagActionDefault
*/
type PostDagActionDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostDagActionDefault creates PostDagActionDefault with default headers values
func NewPostDagActionDefault(code int) *PostDagActionDefault {
	if code <= 0 {
		code = 500
	}

	return &PostDagActionDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the post dag action default response
func (o *PostDagActionDefault) WithStatusCode(code int) *PostDagActionDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the post dag action default response
func (o *PostDagActionDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the post dag action default response
func (o *PostDagActionDefault) WithPayload(payload *models.Error) *PostDagActionDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post dag action default response
func (o *PostDagActionDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostDagActionDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
