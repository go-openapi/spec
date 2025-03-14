// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag"
)

const (
	jsonArray = "array"
)

// HeaderProps describes a response header
type HeaderProps struct {
	Description string `json:"description,omitempty"`
}

// Header describes a header for a response of the API
//
// For more information: http://goo.gl/8us55a#headerObject
type Header struct {
	CommonValidations
	SimpleSchema
	VendorExtensible
	HeaderProps
}

// ResponseHeader creates a new header instance for use in a response
func ResponseHeader() *Header {
	return new(Header)
}

// WithDescription sets the description on this response, allows for chaining
func (h *Header) WithDescription(description string) *Header {
	h.Description = description
	return h
}

// Typed a fluent builder method for the type of parameter
func (h *Header) Typed(tpe, format string) *Header {
	h.Type = tpe
	h.Format = format
	return h
}

// CollectionOf a fluent builder method for an array item
func (h *Header) CollectionOf(items *Items, format string) *Header {
	h.Type = jsonArray
	h.Items = items
	h.CollectionFormat = format
	return h
}

// WithDefault sets the default value on this item
func (h *Header) WithDefault(defaultValue interface{}) *Header {
	h.Default = defaultValue
	return h
}

// WithMaxLength sets a max length value
func (h *Header) WithMaxLength(maximum int64) *Header {
	h.MaxLength = &maximum
	return h
}

// WithMinLength sets a min length value
func (h *Header) WithMinLength(minimum int64) *Header {
	h.MinLength = &minimum
	return h
}

// WithPattern sets a pattern value
func (h *Header) WithPattern(pattern string) *Header {
	h.Pattern = pattern
	return h
}

// WithMultipleOf sets a multiple of value
func (h *Header) WithMultipleOf(number float64) *Header {
	h.MultipleOf = &number
	return h
}

// WithMaximum sets a maximum number value
func (h *Header) WithMaximum(maximum float64, exclusive bool) *Header {
	h.Maximum = &maximum
	h.ExclusiveMaximum = exclusive
	return h
}

// WithMinimum sets a minimum number value
func (h *Header) WithMinimum(minimum float64, exclusive bool) *Header {
	h.Minimum = &minimum
	h.ExclusiveMinimum = exclusive
	return h
}

// WithEnum sets a the enum values (replace)
func (h *Header) WithEnum(values ...interface{}) *Header {
	h.Enum = append([]interface{}{}, values...)
	return h
}

// WithMaxItems sets the max items
func (h *Header) WithMaxItems(size int64) *Header {
	h.MaxItems = &size
	return h
}

// WithMinItems sets the min items
func (h *Header) WithMinItems(size int64) *Header {
	h.MinItems = &size
	return h
}

// UniqueValues dictates that this array can only have unique items
func (h *Header) UniqueValues() *Header {
	h.UniqueItems = true
	return h
}

// AllowDuplicates this array can have duplicates
func (h *Header) AllowDuplicates() *Header {
	h.UniqueItems = false
	return h
}

// WithValidations is a fluent method to set header validations
func (h *Header) WithValidations(val CommonValidations) *Header {
	h.SetValidations(SchemaValidations{CommonValidations: val})
	return h
}

// MarshalJSON marshal this to JSON
func (h Header) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(h.CommonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(h.SimpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(h.HeaderProps)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b1, b2, b3), nil
}

// UnmarshalJSON unmarshals this header from JSON
func (h *Header) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &h.CommonValidations); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &h.SimpleSchema); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &h.VendorExtensible); err != nil {
		return err
	}
	return json.Unmarshal(data, &h.HeaderProps)
}

// JSONLookup look up a value by the json property name
func (h Header) JSONLookup(token string) (interface{}, error) {
	if ex, ok := h.Extensions[token]; ok {
		return &ex, nil
	}

	r, _, err := jsonpointer.GetForToken(h.CommonValidations, token)
	if err != nil && !strings.HasPrefix(err.Error(), "object has no field") {
		return nil, err
	}
	if r != nil {
		return r, nil
	}
	r, _, err = jsonpointer.GetForToken(h.SimpleSchema, token)
	if err != nil && !strings.HasPrefix(err.Error(), "object has no field") {
		return nil, err
	}
	if r != nil {
		return r, nil
	}
	r, _, err = jsonpointer.GetForToken(h.HeaderProps, token)
	return r, err
}
