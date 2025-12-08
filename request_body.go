// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag/jsonutils"
)

// RequestBody describes a single request body.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#request-body-object
type RequestBody struct {
	Refable
	VendorExtensible
	RequestBodyProps
}

// RequestBodyProps describes the properties of a RequestBody
type RequestBodyProps struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// JSONLookup look up a value by the json property name
func (r RequestBody) JSONLookup(token string) (any, error) {
	if ex, ok := r.Extensions[token]; ok {
		return &ex, nil
	}
	r2, _, err := jsonpointer.GetForToken(r.RequestBodyProps, token)
	return r2, err
}

// MarshalJSON marshals this to JSON
func (r RequestBody) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.RequestBodyProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(r.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2, b3), nil
}

// UnmarshalJSON unmarshals this from JSON
func (r *RequestBody) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.RequestBodyProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &r.VendorExtensible)
}
