// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag/jsonutils"
)

// Encoding represents a single encoding definition applied to a single schema property.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#encoding-object
type Encoding struct {
	VendorExtensible
	EncodingProps
}

// EncodingProps describes the properties of an Encoding
type EncodingProps struct {
	ContentType   string            `json:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty"`
	Style         string            `json:"style,omitempty"`
	Explode       *bool             `json:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty"`
}

// JSONLookup look up a value by the json property name
func (e Encoding) JSONLookup(token string) (any, error) {
	if ex, ok := e.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(e.EncodingProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (e Encoding) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(e.EncodingProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(e.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals this from JSON
func (e *Encoding) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.EncodingProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &e.VendorExtensible)
}

// MediaType provides schema and examples for the media type identified by its key.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#media-type-object
type MediaType struct {
	VendorExtensible
	MediaTypeProps
}

// MediaTypeProps describes the properties of a MediaType
type MediaTypeProps struct {
	Schema   *Schema             `json:"schema,omitempty"`
	Example  any                 `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// JSONLookup look up a value by the json property name
func (m MediaType) JSONLookup(token string) (any, error) {
	if ex, ok := m.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(m.MediaTypeProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (m MediaType) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(m.MediaTypeProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(m.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals this from JSON
func (m *MediaType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &m.MediaTypeProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &m.VendorExtensible)
}

// Example represents an example value.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#example-object
type Example struct {
	Refable
	VendorExtensible
	ExampleProps
}

// ExampleProps describes the properties of an Example
type ExampleProps struct {
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// JSONLookup look up a value by the json property name
func (e Example) JSONLookup(token string) (any, error) {
	if ex, ok := e.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(e.ExampleProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (e Example) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(e.Refable)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(e.ExampleProps)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(e.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2, b3), nil
}

// UnmarshalJSON unmarshals this from JSON
func (e *Example) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &e.Refable); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &e.ExampleProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &e.VendorExtensible)
}
