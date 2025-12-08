// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/swag/jsonutils"
)

// Swagger this is the root document object for the API specification.
// It combines what previously was the Resource Listing and API Declaration (version 1.2 and earlier)
// together into one document.
//
// For more information: http://goo.gl/8us55a#swagger-object-
type Swagger struct {
	VendorExtensible
	SwaggerProps
}

// JSONLookup look up a value by the json property name
func (s Swagger) JSONLookup(token string) (any, error) {
	if ex, ok := s.Extensions[token]; ok {
		return &ex, nil
	}

	// Handle backward compatibility for v2 field names
	// When resolving JSON pointers, support both v2 and v3 paths
	switch token {
	case "definitions":
		// First try Components.Schemas (OpenAPI 3.x)
		if s.Components != nil && len(s.Components.Schemas) > 0 {
			return s.Components.Schemas, nil
		}
		// Fall back to Definitions (Swagger 2.0)
		if len(s.Definitions) > 0 {
			return s.Definitions, nil
		}
	case "parameters":
		// First try Components.Parameters (OpenAPI 3.x)
		if s.Components != nil && len(s.Components.Parameters) > 0 {
			return s.Components.Parameters, nil
		}
		// Fall back to Parameters (Swagger 2.0)
		if len(s.Parameters) > 0 {
			return s.Parameters, nil
		}
	case "responses":
		// First try Components.Responses (OpenAPI 3.x)
		if s.Components != nil && len(s.Components.Responses) > 0 {
			return s.Components.Responses, nil
		}
		// Fall back to Responses (Swagger 2.0)
		if len(s.Responses) > 0 {
			return s.Responses, nil
		}
	case "securityDefinitions":
		// First try Components.SecuritySchemes (OpenAPI 3.x)
		if s.Components != nil && len(s.Components.SecuritySchemes) > 0 {
			return s.Components.SecuritySchemes, nil
		}
		// Fall back to SecurityDefinitions (Swagger 2.0)
		if len(s.SecurityDefinitions) > 0 {
			return s.SecurityDefinitions, nil
		}
	}

	r, _, err := jsonpointer.GetForToken(s.SwaggerProps, token)
	return r, err
}

// MarshalJSON marshals this swagger structure to json
func (s Swagger) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(s.SwaggerProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(s.VendorExtensible)
	if err != nil {
		return nil, err
	}
	return jsonutils.ConcatJSON(b1, b2), nil
}

// UnmarshalJSON unmarshals a swagger spec from json
func (s *Swagger) UnmarshalJSON(data []byte) error {
	var sw Swagger
	if err := json.Unmarshal(data, &sw.SwaggerProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &sw.VendorExtensible); err != nil {
		return err
	}

	// Sync Swagger 2.0 fields with OpenAPI 3.x components for backward compatibility
	if sw.Components != nil {
		sw.Definitions = sw.Components.Schemas
		sw.Parameters = sw.Components.Parameters
		sw.Responses = sw.Components.Responses
		// Convert SecuritySchemes (value) to SecurityDefinitions (pointer)
		if sw.Components.SecuritySchemes != nil {
			sw.SecurityDefinitions = make(SecurityDefinitions, len(sw.Components.SecuritySchemes))
			for k, v := range sw.Components.SecuritySchemes {
				scheme := v
				sw.SecurityDefinitions[k] = &scheme
			}
		}
	}

	*s = sw
	return nil
}

// GobEncode provides a safe gob encoder for Swagger, including extensions
func (s Swagger) GobEncode() ([]byte, error) {
	var b bytes.Buffer
	raw := struct {
		Props SwaggerProps
		Ext   VendorExtensible
	}{
		Props: s.SwaggerProps,
		Ext:   s.VendorExtensible,
	}
	err := gob.NewEncoder(&b).Encode(raw)
	return b.Bytes(), err
}

// GobDecode provides a safe gob decoder for Swagger, including extensions
func (s *Swagger) GobDecode(b []byte) error {
	var raw struct {
		Props SwaggerProps
		Ext   VendorExtensible
	}
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&raw)
	if err != nil {
		return err
	}
	s.SwaggerProps = raw.Props
	s.VendorExtensible = raw.Ext
	return nil
}

// SwaggerProps captures the top-level properties of an Api specification
//
// NOTE: validation rules
// - the scheme, when present must be from [http, https, ws, wss]
// - BasePath must start with a leading "/"
// - Paths is required
type SwaggerProps struct {
	ID           string                 `json:"id,omitempty"`
	Info         *Info                  `json:"info,omitempty"`
	Paths        *Paths                 `json:"paths"`
	Security     []map[string][]string  `json:"security,omitempty"`
	Tags         []Tag                  `json:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	OpenAPI           string              `json:"openapi,omitempty"` // for OpenAPI 3 document
	JSONSchemaDialect string              `json:"jsonSchemaDialect,omitempty"`
	Servers           []Server            `json:"servers,omitempty"`
	Webhooks          map[string]PathItem `json:"webhooks,omitempty"`
	Components        *Components         `json:"components,omitempty"`

	// In OpenAPI 3.x these map to different locations:
	// - Definitions -> Components.Schemas
	// - Parameters -> Components.Parameters
	// - Responses -> Components.Responses
	// - SecurityDefinitions -> Components.SecuritySchemes
	// - Consumes/Produces are now in individual operations or media types
	// - Host/BasePath/Schemes -> Servers
	// These fields are kept for backward compatibility with Swagger 2.0 specs
	Swagger             string               `json:"swagger,omitempty"` // for Swagger 2 document
	Consumes            []string             `json:"consumes,omitempty"`
	Produces            []string             `json:"produces,omitempty"`
	Schemes             []string             `json:"schemes,omitempty"`
	Host                string               `json:"host,omitempty"`
	BasePath            string               `json:"basePath,omitempty"`
	Definitions         Definitions          `json:"definitions,omitempty"`
	Parameters          map[string]Parameter `json:"parameters,omitempty"`
	Responses           map[string]Response  `json:"responses,omitempty"`
	SecurityDefinitions SecurityDefinitions  `json:"securityDefinitions,omitempty"`
}

type swaggerPropsAlias SwaggerProps

type gobSwaggerPropsAlias struct {
	Security []map[string]struct {
		List []string
		Pad  bool
	}
	Alias           *swaggerPropsAlias
	SecurityIsEmpty bool
}

// GobEncode provides a safe gob encoder for SwaggerProps, including empty security requirements
func (o SwaggerProps) GobEncode() ([]byte, error) {
	raw := gobSwaggerPropsAlias{
		Alias: (*swaggerPropsAlias)(&o),
	}

	var b bytes.Buffer
	if o.Security == nil {
		// nil security requirement
		err := gob.NewEncoder(&b).Encode(raw)
		return b.Bytes(), err
	}

	if len(o.Security) == 0 {
		// empty, but non-nil security requirement
		raw.SecurityIsEmpty = true
		raw.Alias.Security = nil
		err := gob.NewEncoder(&b).Encode(raw)
		return b.Bytes(), err
	}

	raw.Security = make([]map[string]struct {
		List []string
		Pad  bool
	}, 0, len(o.Security))
	for _, req := range o.Security {
		v := make(map[string]struct {
			List []string
			Pad  bool
		}, len(req))
		for k, val := range req {
			v[k] = struct {
				List []string
				Pad  bool
			}{
				List: val,
			}
		}
		raw.Security = append(raw.Security, v)
	}

	err := gob.NewEncoder(&b).Encode(raw)
	return b.Bytes(), err
}

// GobDecode provides a safe gob decoder for SwaggerProps, including empty security requirements
func (o *SwaggerProps) GobDecode(b []byte) error {
	var raw gobSwaggerPropsAlias

	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&raw)
	if err != nil {
		return err
	}
	if raw.Alias == nil {
		return nil
	}

	switch {
	case raw.SecurityIsEmpty:
		// empty, but non-nil security requirement
		raw.Alias.Security = []map[string][]string{}
	case len(raw.Alias.Security) == 0:
		// nil security requirement
		raw.Alias.Security = nil
	default:
		raw.Alias.Security = make([]map[string][]string, 0, len(raw.Security))
		for _, req := range raw.Security {
			v := make(map[string][]string, len(req))
			for k, val := range req {
				v[k] = make([]string, 0, len(val.List))
				v[k] = append(v[k], val.List...)
			}
			raw.Alias.Security = append(raw.Alias.Security, v)
		}
	}

	*o = *(*SwaggerProps)(raw.Alias)
	return nil
}

// Dependencies represent a dependencies property
type Dependencies map[string]SchemaOrStringArray

// SchemaOrBool represents a schema or boolean value, is biased towards true for the boolean property
type SchemaOrBool struct {
	Allows bool
	Schema *Schema
}

// JSONLookup implements an interface to customize json pointer lookup
func (s SchemaOrBool) JSONLookup(token string) (any, error) {
	if token == "allows" {
		return s.Allows, nil
	}
	r, _, err := jsonpointer.GetForToken(s.Schema, token)
	return r, err
}

var jsTrue = []byte("true")
var jsFalse = []byte("false")

// MarshalJSON convert this object to JSON
func (s SchemaOrBool) MarshalJSON() ([]byte, error) {
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}

	if s.Schema == nil && !s.Allows {
		return jsFalse, nil
	}
	return jsTrue, nil
}

// UnmarshalJSON converts this bool or schema object from a JSON structure
func (s *SchemaOrBool) UnmarshalJSON(data []byte) error {
	var nw SchemaOrBool
	if len(data) > 0 {
		if data[0] == '{' {
			var sch Schema
			if err := json.Unmarshal(data, &sch); err != nil {
				return err
			}
			nw.Schema = &sch
		}
		nw.Allows = !bytes.Equal(data, []byte("false"))
	}
	*s = nw
	return nil
}

// SchemaOrStringArray represents a schema or a string array
type SchemaOrStringArray struct {
	Schema   *Schema
	Property []string
}

// JSONLookup implements an interface to customize json pointer lookup
func (s SchemaOrStringArray) JSONLookup(token string) (any, error) {
	r, _, err := jsonpointer.GetForToken(s.Schema, token)
	return r, err
}

// MarshalJSON converts this schema object or array into JSON structure
func (s SchemaOrStringArray) MarshalJSON() ([]byte, error) {
	if len(s.Property) > 0 {
		return json.Marshal(s.Property)
	}
	if s.Schema != nil {
		return json.Marshal(s.Schema)
	}
	return []byte("null"), nil
}

// UnmarshalJSON converts this schema object or array from a JSON structure
func (s *SchemaOrStringArray) UnmarshalJSON(data []byte) error {
	var first byte
	if len(data) > 1 {
		first = data[0]
	}
	var nw SchemaOrStringArray
	if first == '{' {
		var sch Schema
		if err := json.Unmarshal(data, &sch); err != nil {
			return err
		}
		nw.Schema = &sch
	}
	if first == '[' {
		if err := json.Unmarshal(data, &nw.Property); err != nil {
			return err
		}
	}
	*s = nw
	return nil
}

// Definitions contains the models explicitly defined in this spec
// An object to hold data types that can be consumed and produced by operations.
// These data types can be primitives, arrays or models.
//
// For more information: http://goo.gl/8us55a#definitionsObject
type Definitions map[string]Schema

// SecurityDefinitions a declaration of the security schemes available to be used in the specification.
// This does not enforce the security schemes on the operations and only serves to provide
// the relevant details for each scheme.
//
// For more information: http://goo.gl/8us55a#securityDefinitionsObject
type SecurityDefinitions map[string]*SecurityScheme

// StringOrArray represents a value that can either be a string
// or an array of strings. Mainly here for serialization purposes
type StringOrArray []string

// Contains returns true when the value is contained in the slice
func (s StringOrArray) Contains(value string) bool {
	return slices.Contains(s, value)
}

// JSONLookup implements an interface to customize json pointer lookup
func (s SchemaOrArray) JSONLookup(token string) (any, error) {
	if _, err := strconv.Atoi(token); err == nil {
		r, _, err := jsonpointer.GetForToken(s.Schemas, token)
		return r, err
	}
	r, _, err := jsonpointer.GetForToken(s.Schema, token)
	return r, err
}

// UnmarshalJSON unmarshals this string or array object from a JSON array or JSON string
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var first byte
	if len(data) > 1 {
		first = data[0]
	}

	if first == '[' {
		var parsed []string
		if err := json.Unmarshal(data, &parsed); err != nil {
			return err
		}
		*s = StringOrArray(parsed)
		return nil
	}

	var single any
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	if single == nil {
		return nil
	}
	switch v := single.(type) {
	case string:
		*s = StringOrArray([]string{v})
		return nil
	default:
		return fmt.Errorf("only string or array is allowed, not %T: %w", single, ErrSpec)
	}
}

// MarshalJSON converts this string or array to a JSON array or JSON string
func (s StringOrArray) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal([]string(s)[0])
	}
	return json.Marshal([]string(s))
}

// SchemaOrArray represents a value that can either be a Schema
// or an array of Schema. Mainly here for serialization purposes
type SchemaOrArray struct {
	Schema  *Schema
	Schemas []Schema
}

// Len returns the number of schemas in this property
func (s SchemaOrArray) Len() int {
	if s.Schema != nil {
		return 1
	}
	return len(s.Schemas)
}

// ContainsType returns true when one of the schemas is of the specified type
func (s *SchemaOrArray) ContainsType(name string) bool {
	if s.Schema != nil {
		return s.Schema.Type != nil && s.Schema.Type.Contains(name)
	}
	return false
}

// MarshalJSON converts this schema object or array into JSON structure
func (s SchemaOrArray) MarshalJSON() ([]byte, error) {
	if len(s.Schemas) > 0 {
		return json.Marshal(s.Schemas)
	}
	return json.Marshal(s.Schema)
}

// UnmarshalJSON converts this schema object or array from a JSON structure
func (s *SchemaOrArray) UnmarshalJSON(data []byte) error {
	var nw SchemaOrArray
	var first byte
	if len(data) > 1 {
		first = data[0]
	}
	if first == '{' {
		var sch Schema
		if err := json.Unmarshal(data, &sch); err != nil {
			return err
		}
		nw.Schema = &sch
	}
	if first == '[' {
		if err := json.Unmarshal(data, &nw.Schemas); err != nil {
			return err
		}
	}
	*s = nw
	return nil
}

// vim:set ft=go noet sts=2 sw=2 ts=2:
