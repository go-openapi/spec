// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/swag/conv"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

const epsilon = 1e-9

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(f int64) *int64 {
	return &f
}

var header = Header{ //nolint:gochecknoglobals // test fixture
	VendorExtensible: VendorExtensible{Extensions: map[string]any{
		"x-framework": "swagger-go",
	}},
	HeaderProps: HeaderProps{Description: "the description of this header"},
	SimpleSchema: SimpleSchema{
		Items: &Items{
			Refable: Refable{Ref: MustCreateRef("Cat")},
		},
		Type:    "string",
		Format:  "date",
		Default: "8",
	},
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
}

const headerJSON = `{
  "items": {
    "$ref": "Cat"
  },
  "x-framework": "swagger-go",
  "description": "the description of this header",
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
  "default": "8"
}`

func TestIntegrationHeader(t *testing.T) {
	assert.JSONUnmarshalAsT(t, header, headerJSON)
}

func TestJSONLookupHeader(t *testing.T) {
	var def string
	res, err := header.JSONLookup("default")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, def, res)

	var ok bool
	def, ok = res.(string)
	require.TrueT(t, ok)
	assert.EqualT(t, "8", def)

	var x *any
	res, err = header.JSONLookup("x-framework")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, x, res)

	x, ok = res.(*any)
	require.TrueT(t, ok)
	assert.EqualValues(t, "swagger-go", *x)

	res, err = header.JSONLookup("unknown")
	require.Error(t, err)
	require.Nil(t, res)

	var maximum *float64
	res, err = header.JSONLookup("maximum")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, maximum, res)

	maximum, ok = res.(*float64)
	require.TrueT(t, ok)
	assert.InDeltaT(t, float64(100), *maximum, epsilon)
}

func TestResponseHeaueder(t *testing.T) {
	var expectedHeader *Header
	h := ResponseHeader()
	assert.IsType(t, expectedHeader, h)
}

func TestWithHeader(t *testing.T) {
	h := new(Header).WithDescription("header description").Typed("integer", "int32")
	assert.EqualT(t, "header description", h.Description)
	assert.EqualT(t, "integer", h.Type)
	assert.EqualT(t, "int32", h.Format)

	i := new(Items).Typed("string", "date")
	h = new(Header).CollectionOf(i, "pipe")

	assert.Equal(t, *i, *h.Items)
	assert.EqualT(t, "pipe", h.CollectionFormat)

	h = new(Header).WithDefault([]string{"a", "b", "c"}).WithMaxLength(10).WithMinLength(3)

	assert.EqualT(t, int64(10), *h.MaxLength)
	assert.EqualT(t, int64(3), *h.MinLength)
	assert.EqualValues(t, []string{"a", "b", "c"}, h.Default)

	h = new(Header).WithPattern("^abc$")
	assert.Equal(t, Header{
		CommonValidations: CommonValidations{
			Pattern: "^abc$",
		},
	}, *h)
	h = new(Header).WithEnum("a", "b", "c")
	assert.Equal(t, Header{
		CommonValidations: CommonValidations{
			Enum: []any{
				"a",
				"b",
				"c",
			},
		},
	}, *h)
}

func TestHeaderWithValidation(t *testing.T) {
	h := new(Header).WithValidations(CommonValidations{MaxLength: conv.Pointer(int64(15))})
	assert.Equal(t, conv.Pointer(int64(15)), h.MaxLength)
}
