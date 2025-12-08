// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag/jsonutils"
)

// ServerVariable represents a Server Variable for server URL template substitution.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#server-variable-object
type ServerVariable struct {
	VendorExtensible
	ServerVariableProps
}

// ServerVariableProps describes the properties of a Server Variable
type ServerVariableProps struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// JSONLookup look up a value by the json property name
func (s ServerVariable) JSONLookup(token string) (any, error) {
	if ex, ok := s.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(s.ServerVariableProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (s ServerVariable) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.ServerVariableProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals this from JSON
func (s *ServerVariable) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.ServerVariableProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &s.VendorExtensible)
}

// Server represents a Server.
//
// For more information: https://spec.openapis.org/oas/v3.1.0#server-object
type Server struct {
	VendorExtensible
	ServerProps
}

// ServerProps describes the properties of a Server
type ServerProps struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// JSONLookup look up a value by the json property name
func (s Server) JSONLookup(token string) (any, error) {
	if ex, ok := s.Extensions[token]; ok {
		return &ex, nil
	}
	r, _, err := jsonpointer.GetForToken(s.ServerProps, token)
	return r, err
}

// MarshalJSON marshals this to JSON
func (s Server) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.ServerProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals this from JSON
func (s *Server) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.ServerProps); err != nil {
		return err
	}
	return json.Unmarshal(data, &s.VendorExtensible)
}
