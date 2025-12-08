// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag/jsonutils"
)

// Components holds a set of reusable objects for different aspects of the OAS.
// All objects defined within the components object will have no effect on the API
// unless they are explicitly referenced from properties outside the components object.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#components-object
type Components struct {
	VendorExtensible
	ComponentsProps
}

// ComponentsProps describes the properties of a Components object
type ComponentsProps struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty"`
	Examples        map[string]Example        `json:"examples,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]Header         `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	Links           map[string]Link           `json:"links,omitempty"`
	Callbacks       map[string]Callback       `json:"callbacks,omitempty"`
}

// JSONLookup look up a value by the json property name
func (c Components) JSONLookup(token string) (any, error) {
	if ex, ok := c.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(c.ComponentsProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (c Components) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(c.ComponentsProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(c.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals this from JSON
func (c *Components) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.ComponentsProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &c.VendorExtensible)
}

// Link represents a possible design-time link for a response.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#link-object
type Link struct {
	Refable
	VendorExtensible
	LinkProps
}

// LinkProps describes the properties of a Link
type LinkProps struct {
	OperationRef string         `json:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty"`
	Server       *Server        `json:"server,omitempty"`
}

// JSONLookup look up a value by the json property name
func (l Link) JSONLookup(token string) (any, error) {
	if ex, ok := l.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(l.LinkProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (l Link) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(l.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(l.LinkProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(l.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2, b3), nil
}

// UnmarshalJSON unmarshals this from JSON
func (l *Link) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &l.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &l.LinkProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &l.VendorExtensible)
}

// Callback is a map of possible out-of band callbacks related to the parent operation.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#callback-object
type Callback struct {
	Refable
	VendorExtensible
	CallbackProps
}

// CallbackProps describes the properties of a Callback
type CallbackProps struct {
	Expressions map[string]PathItem `json:"-"`
}

// JSONLookup look up a value by the json property name
func (c Callback) JSONLookup(token string) (any, error) {
	if ex, ok := c.Extensions[token]; ok {
		return &ex, nil
	}
	if pi, ok := c.Expressions[token]; ok {
		return pi, nil
	}
	return nil, nil
}

// MarshalJSON marshals this to JSON
func (c Callback) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(c.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(c.Expressions)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(c.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2, b3), nil
}

// UnmarshalJSON unmarshals this from JSON
func (c *Callback) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &c.Expressions); err != nil {
		return err
	}
	return json.Unmarshal(data, &c.VendorExtensible)
}
