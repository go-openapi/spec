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
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const epsilon = 1e-9

func float64Ptr(f float64) *float64 {
	return &f
}
func int64Ptr(f int64) *int64 {
	return &f
}

var header = Header{
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{
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
		Enum:             []interface{}{"hello", "world"},
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
	var actual Header
	require.NoError(t, json.Unmarshal([]byte(headerJSON), &actual))
	assert.EqualValues(t, actual, header)

	assertParsesJSON(t, headerJSON, header)
}

func TestJSONLookupHeader(t *testing.T) {
	var def string
	res, err := header.JSONLookup("default")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, def, res)

	var ok bool
	def, ok = res.(string)
	require.True(t, ok)
	assert.Equal(t, "8", def)

	var x *interface{}
	res, err = header.JSONLookup("x-framework")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, x, res)

	x, ok = res.(*interface{})
	require.True(t, ok)
	assert.EqualValues(t, "swagger-go", *x)

	res, err = header.JSONLookup("unknown")
	require.Error(t, err)
	require.Nil(t, res)

	var max *float64
	res, err = header.JSONLookup("maximum")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, max, res)

	max, ok = res.(*float64)
	require.True(t, ok)
	assert.InDelta(t, float64(100), *max, epsilon)
}

func TestResponseHeaueder(t *testing.T) {
	var expectedHeader *Header
	h := ResponseHeader()
	assert.IsType(t, expectedHeader, h)
}

func TestWithHeader(t *testing.T) {
	h := new(Header).WithDescription("header description").Typed("integer", "int32")
	assert.Equal(t, "header description", h.Description)
	assert.Equal(t, "integer", h.Type)
	assert.Equal(t, "int32", h.Format)

	i := new(Items).Typed("string", "date")
	h = new(Header).CollectionOf(i, "pipe")

	assert.EqualValues(t, *i, *h.Items)
	assert.Equal(t, "pipe", h.CollectionFormat)

	h = new(Header).WithDefault([]string{"a", "b", "c"}).WithMaxLength(10).WithMinLength(3)

	assert.Equal(t, int64(10), *h.MaxLength)
	assert.Equal(t, int64(3), *h.MinLength)
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
			Enum: []interface{}{
				"a",
				"b",
				"c",
			},
		},
	}, *h)
}

func TestHeaderWithValidation(t *testing.T) {
	h := new(Header).WithValidations(CommonValidations{MaxLength: swag.Int64(15)})
	assert.EqualValues(t, swag.Int64(15), h.MaxLength)
}
