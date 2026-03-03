// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/swag/conv"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var parameter = Parameter{ //nolint:gochecknoglobals // test fixture
	VendorExtensible: VendorExtensible{Extensions: map[string]any{
		"x-framework": "swagger-go",
	}},
	Refable: Refable{Ref: MustCreateRef("Dog")},
	CommonValidations: CommonValidations{
		Maximum:          float64Ptr(100),
		ExclusiveMaximum: true,
		ExclusiveMinimum: true,
		Minimum:          float64Ptr(5),
		MaxLength:        int64Ptr(100),
		MinLength:        int64Ptr(5),
		Pattern:          "\\w{1,5}\\w+",
		MaxItems:         int64Ptr(100),
		MinItems:         int64Ptr(5),
		UniqueItems:      true,
		MultipleOf:       float64Ptr(5),
		Enum:             []any{"hello", "world"},
	},
	SimpleSchema: SimpleSchema{
		Type:             "string",
		Format:           "date",
		CollectionFormat: "csv",
		Items: &Items{
			Refable: Refable{Ref: MustCreateRef("Cat")},
		},
		Default: "8",
	},
	ParamProps: ParamProps{
		Name:        "param-name",
		In:          "header",
		Required:    true,
		Schema:      &Schema{SchemaProps: SchemaProps{Type: []string{"string"}}},
		Description: "the description of this parameter",
	},
}

//nolint:gochecknoglobals // test fixture
var parameterJSON = `{
	"items": {
		"$ref": "Cat"
	},
	"x-framework": "swagger-go",
  "$ref": "Dog",
  "description": "the description of this parameter",
  "maximum": 100,
  "minimum": 5,
  "exclusiveMaximum": true,
  "exclusiveMinimum": true,
  "maxLength": 100,
  "minLength": 5,
  "pattern": "\\w{1,5}\\w+",
  "maxItems": 100,
  "minItems": 5,
  "uniqueItems": true,
  "multipleOf": 5,
  "enum": ["hello", "world"],
  "type": "string",
  "format": "date",
	"name": "param-name",
	"in": "header",
	"required": true,
	"schema": {
		"type": "string"
	},
	"collectionFormat": "csv",
	"default": "8"
}`

func TestIntegrationParameter(t *testing.T) {
	assert.JSONUnmarshalAsT(t, parameter, parameterJSON)
}

func TestParameterSerialization(t *testing.T) {
	items := &Items{
		SimpleSchema: SimpleSchema{Type: "string"},
	}

	intItems := &Items{
		SimpleSchema: SimpleSchema{Type: "int", Format: "int32"},
	}

	assert.JSONMarshalAsT(t, `{"type":"string","in":"query"}`, QueryParam("").Typed("string", ""))

	assert.JSONMarshalAsT(t,
		`{"type":"array","items":{"type":"string"},"collectionFormat":"multi","in":"query"}`,
		QueryParam("").CollectionOf(items, "multi"))

	assert.JSONMarshalAsT(t, `{"type":"string","in":"path","required":true}`, PathParam("").Typed("string", ""))

	assert.JSONMarshalAsT(t,
		`{"type":"array","items":{"type":"string"},"collectionFormat":"multi","in":"path","required":true}`,
		PathParam("").CollectionOf(items, "multi"))

	assert.JSONMarshalAsT(t,
		`{"type":"array","items":{"type":"int","format":"int32"},"collectionFormat":"multi","in":"path","required":true}`,
		PathParam("").CollectionOf(intItems, "multi"))

	assert.JSONMarshalAsT(t, `{"type":"string","in":"header","required":true}`, HeaderParam("").Typed("string", ""))

	assert.JSONMarshalAsT(t,
		`{"type":"array","items":{"type":"string"},"collectionFormat":"multi","in":"header","required":true}`,
		HeaderParam("").CollectionOf(items, "multi"))

	schema := &Schema{SchemaProps: SchemaProps{
		Properties: map[string]Schema{
			"name": {SchemaProps: SchemaProps{
				Type: []string{"string"},
			}},
		},
	}}

	refSchema := &Schema{
		SchemaProps: SchemaProps{Ref: MustCreateRef("Cat")},
	}

	assert.JSONMarshalAsT(t,
		`{"in":"body","schema":{"properties":{"name":{"type":"string"}}}}`,
		BodyParam("", schema))

	assert.JSONMarshalAsT(t,
		`{"in":"body","schema":{"$ref":"Cat"}}`,
		BodyParam("", refSchema))

	// array body param
	assert.JSONMarshalAsT(t,
		`{"in":"body","schema":{"type":"array","items":{"$ref":"Cat"}}}`,
		BodyParam("", ArrayProperty(RefProperty("Cat"))))
}

func TestParameterGobEncoding(t *testing.T) {
	var src, dst Parameter
	require.NoError(t, json.Unmarshal([]byte(parameterJSON), &src))
	doTestAnyGobEncoding(t, &src, &dst)
}

func TestParametersWithValidation(t *testing.T) {
	p := new(Parameter).WithValidations(CommonValidations{MaxLength: conv.Pointer(int64(15))})
	assert.Equal(t, conv.Pointer(int64(15)), p.MaxLength)
}
