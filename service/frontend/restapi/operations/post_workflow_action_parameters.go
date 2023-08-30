// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// NewPostWorkflowActionParams creates a new PostWorkflowActionParams object
//
// There are no default values defined in the spec.
func NewPostWorkflowActionParams() PostWorkflowActionParams {

	return PostWorkflowActionParams{}
}

// PostWorkflowActionParams contains all the bound params for the post workflow action operation
// typically these are obtained from a http.Request
//
// swagger:parameters postWorkflowAction
type PostWorkflowActionParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: query
	*/
	Action string
	/*
	  In: query
	*/
	Params *string
	/*
	  In: query
	*/
	RequestID *string
	/*
	  In: query
	*/
	Step *string
	/*
	  In: query
	*/
	Value *string
	/*
	  Required: true
	  In: path
	*/
	WorkflowID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostWorkflowActionParams() beforehand.
func (o *PostWorkflowActionParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qAction, qhkAction, _ := qs.GetOK("action")
	if err := o.bindAction(qAction, qhkAction, route.Formats); err != nil {
		res = append(res, err)
	}

	qParams, qhkParams, _ := qs.GetOK("params")
	if err := o.bindParams(qParams, qhkParams, route.Formats); err != nil {
		res = append(res, err)
	}

	qRequestID, qhkRequestID, _ := qs.GetOK("request-id")
	if err := o.bindRequestID(qRequestID, qhkRequestID, route.Formats); err != nil {
		res = append(res, err)
	}

	qStep, qhkStep, _ := qs.GetOK("step")
	if err := o.bindStep(qStep, qhkStep, route.Formats); err != nil {
		res = append(res, err)
	}

	qValue, qhkValue, _ := qs.GetOK("value")
	if err := o.bindValue(qValue, qhkValue, route.Formats); err != nil {
		res = append(res, err)
	}

	rWorkflowID, rhkWorkflowID, _ := route.Params.GetOK("workflowId")
	if err := o.bindWorkflowID(rWorkflowID, rhkWorkflowID, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAction binds and validates parameter Action from query.
func (o *PostWorkflowActionParams) bindAction(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("action", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("action", "query", raw); err != nil {
		return err
	}
	o.Action = raw

	return nil
}

// bindParams binds and validates parameter Params from query.
func (o *PostWorkflowActionParams) bindParams(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Params = &raw

	return nil
}

// bindRequestID binds and validates parameter RequestID from query.
func (o *PostWorkflowActionParams) bindRequestID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.RequestID = &raw

	return nil
}

// bindStep binds and validates parameter Step from query.
func (o *PostWorkflowActionParams) bindStep(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Step = &raw

	return nil
}

// bindValue binds and validates parameter Value from query.
func (o *PostWorkflowActionParams) bindValue(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Value = &raw

	return nil
}

// bindWorkflowID binds and validates parameter WorkflowID from path.
func (o *PostWorkflowActionParams) bindWorkflowID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.WorkflowID = raw

	return nil
}
