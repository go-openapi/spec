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

func TestParameterConstructors(t *testing.T) {
	t.Run("FormDataParam sits in formData", func(t *testing.T) {
		p := FormDataParam("file-name")
		assert.EqualT(t, "file-name", p.Name)
		assert.EqualT(t, "formData", p.In)
	})

	t.Run("FileParam is a typed formData param", func(t *testing.T) {
		p := FileParam("upload")
		assert.EqualT(t, "upload", p.Name)
		assert.EqualT(t, "formData", p.In)
		assert.EqualT(t, "file", p.Type)
	})

	t.Run("SimpleArrayParam defaults to csv", func(t *testing.T) {
		p := SimpleArrayParam("tags", "string", "date")
		assert.EqualT(t, "tags", p.Name)
		assert.EqualT(t, "array", p.Type)
		assert.EqualT(t, "csv", p.CollectionFormat)
		require.NotNil(t, p.Items)
		assert.EqualT(t, "string", p.Items.Type)
		assert.EqualT(t, "date", p.Items.Format)
	})

	t.Run("ParamRef holds only a $ref", func(t *testing.T) {
		p := ParamRef("Dog")
		assert.Equal(t, MustCreateRef("Dog"), p.Ref)
		assert.JSONMarshalAsT(t, `{"$ref":"Dog"}`, p)
	})
}

func TestParameterBuilder(t *testing.T) {
	t.Run("naming and location", func(t *testing.T) {
		p := QueryParam("q").
			Named("query-name").
			WithLocation("header").
			WithDescription("the description of this parameter")

		assert.EqualT(t, "query-name", p.Name)
		assert.EqualT(t, "header", p.In)
		assert.EqualT(t, "the description of this parameter", p.Description)
	})

	t.Run("empty values", func(t *testing.T) {
		p := QueryParam("q").AllowsEmptyValues()
		assert.TrueT(t, p.AllowEmptyValue)

		p = p.NoEmptyValues()
		assert.FalseT(t, p.AllowEmptyValue)
	})

	t.Run("a default makes the parameter optional", func(t *testing.T) {
		p := PathParam("id").WithDefault("a default")
		assert.Equal(t, "a default", p.Default)
		assert.FalseT(t, p.Required)

		// AsRequired is a no-op once a default is set
		p = p.AsRequired()
		assert.FalseT(t, p.Required)
	})

	t.Run("required and optional", func(t *testing.T) {
		p := QueryParam("q").AsRequired()
		assert.TrueT(t, p.Required)

		p = p.AsOptional()
		assert.FalseT(t, p.Required)
	})

	t.Run("validations", func(t *testing.T) {
		p := QueryParam("q").
			WithMaxLength(100).
			WithMinLength(5).
			WithPattern(`\w+`).
			WithMultipleOf(5).
			WithMaximum(100, true).
			WithMinimum(5, true).
			WithEnum("hello", "world").
			WithMaxItems(100).
			WithMinItems(5).
			UniqueValues()

		assert.Equal(t, int64Ptr(100), p.MaxLength)
		assert.Equal(t, int64Ptr(5), p.MinLength)
		assert.EqualT(t, `\w+`, p.Pattern)
		assert.Equal(t, float64Ptr(5), p.MultipleOf)
		assert.Equal(t, float64Ptr(100), p.Maximum)
		assert.TrueT(t, p.ExclusiveMaximum)
		assert.Equal(t, float64Ptr(5), p.Minimum)
		assert.TrueT(t, p.ExclusiveMinimum)
		assert.Equal(t, []any{"hello", "world"}, p.Enum)
		assert.Equal(t, int64Ptr(100), p.MaxItems)
		assert.Equal(t, int64Ptr(5), p.MinItems)
		assert.TrueT(t, p.UniqueItems)

		p = p.AllowDuplicates().WithMaximum(100, false).WithMinimum(5, false)
		assert.FalseT(t, p.UniqueItems)
		assert.FalseT(t, p.ExclusiveMaximum)
		assert.FalseT(t, p.ExclusiveMinimum)
	})
}

func TestJSONLookupParameter(t *testing.T) {
	t.Run(`lookup should find an extension`, func(t *testing.T) {
		res, err := parameter.JSONLookup("x-framework")
		require.NoError(t, err)
		require.NotNil(t, res)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, "swagger-go", *ext)
	})

	t.Run(`lookup should find "$ref"`, func(t *testing.T) {
		res, err := parameter.JSONLookup("$ref")
		require.NoError(t, err)

		ref, ok := res.(*Ref)
		require.TrueT(t, ok)
		assert.Equal(t, MustCreateRef("Dog"), *ref)
	})

	t.Run(`lookup should find "maximum" in CommonValidations`, func(t *testing.T) {
		res, err := parameter.JSONLookup("maximum")
		require.NoError(t, err)

		maximum, ok := res.(*float64)
		require.TrueT(t, ok)
		assert.InDeltaT(t, float64(100), *maximum, epsilon)
	})

	t.Run(`lookup should find "collectionFormat" in SimpleSchema`, func(t *testing.T) {
		res, err := parameter.JSONLookup("collectionFormat")
		require.NoError(t, err)

		f, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "csv", f)
	})

	t.Run(`lookup should find "in" in ParamProps`, func(t *testing.T) {
		res, err := parameter.JSONLookup("in")
		require.NoError(t, err)

		in, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "header", in)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := parameter.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}
