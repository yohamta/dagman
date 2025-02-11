// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// APIError Api error
//
// swagger:model ApiError
type APIError struct {

	// Detailed error description.
	// Required: true
	DetailedMessage *string `json:"detailedMessage"`

	// Short error message.
	// Required: true
	Message *string `json:"message"`
}

// Validate validates this Api error
func (m *APIError) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDetailedMessage(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMessage(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIError) validateDetailedMessage(formats strfmt.Registry) error {

	if err := validate.Required("detailedMessage", "body", m.DetailedMessage); err != nil {
		return err
	}

	return nil
}

func (m *APIError) validateMessage(formats strfmt.Registry) error {

	if err := validate.Required("message", "body", m.Message); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this Api error based on context it is used
func (m *APIError) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *APIError) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIError) UnmarshalBinary(b []byte) error {
	var res APIError
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
