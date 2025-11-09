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

var testItems = Items{
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
		Type:   "string",
		Format: "date",
		Items: &Items{
			Refable: Refable{Ref: MustCreateRef("Cat")},
		},
		CollectionFormat: "csv",
		Default:          "8",
	},
}

const itemsJSON = `{
	"items": {
		"$ref": "Cat"
	},
  "$ref": "Dog",
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
	"collectionFormat": "csv",
	"default": "8"
}`

func TestIntegrationItems(t *testing.T) {
	var actual Items
	require.NoError(t, json.Unmarshal([]byte(itemsJSON), &actual))
	assert.Equal(t, actual, testItems)

	assertParsesJSON(t, itemsJSON, testItems)
}

func TestTypeNameItems(t *testing.T) {
	var nilItems Items
	assert.Empty(t, nilItems.TypeName())

	assert.Equal(t, "date", testItems.TypeName())
	assert.Empty(t, testItems.ItemsTypeName())

	nested := Items{
		SimpleSchema: SimpleSchema{
			Type: "array",
			Items: &Items{
				SimpleSchema: SimpleSchema{
					Type:   "integer",
					Format: "int32",
				},
			},
			CollectionFormat: "csv",
		},
	}

	assert.Equal(t, "array", nested.TypeName())
	assert.Equal(t, "int32", nested.ItemsTypeName())

	simple := SimpleSchema{
		Type:  "string",
		Items: nil,
	}

	assert.Equal(t, "string", simple.TypeName())
	assert.Empty(t, simple.ItemsTypeName())

	simple.Items = NewItems()
	simple.Type = "array"
	simple.Items.Type = "string"

	assert.Equal(t, "array", simple.TypeName())
	assert.Equal(t, "string", simple.ItemsTypeName())
}

func TestItemsBuilder(t *testing.T) {
	simple := SimpleSchema{
		Type: "array",
		Items: NewItems().
			Typed("string", "uuid").
			WithDefault([]string{"default-value"}).
			WithEnum([]string{"abc", "efg"}, []string{"hij"}).
			WithMaxItems(4).
			WithMinItems(1).
			UniqueValues(),
	}

	assert.Equal(t, SimpleSchema{
		Type: "array",
		Items: &Items{
			SimpleSchema: SimpleSchema{
				Type:    "string",
				Format:  "uuid",
				Default: []string{"default-value"},
			},
			CommonValidations: CommonValidations{
				Enum:        []any{[]string{"abc", "efg"}, []string{"hij"}},
				MinItems:    conv.Pointer(int64(1)),
				MaxItems:    conv.Pointer(int64(4)),
				UniqueItems: true,
			},
		},
	}, simple)
}

func TestJSONLookupItems(t *testing.T) {
	t.Run(`lookup should find "$ref"`, func(t *testing.T) {
		res, err := testItems.JSONLookup("$ref")
		require.NoError(t, err)
		require.NotNil(t, res)
		require.IsType(t, &Ref{}, res)

		ref, ok := res.(*Ref)
		require.True(t, ok)
		assert.Equal(t, MustCreateRef("Dog"), *ref)
	})

	t.Run(`lookup should find "maximum"`, func(t *testing.T) {
		var maximum *float64
		res, err := testItems.JSONLookup("maximum")
		require.NoError(t, err)
		require.NotNil(t, res)
		require.IsType(t, maximum, res)

		var ok bool
		maximum, ok = res.(*float64)
		require.True(t, ok)
		assert.InDelta(t, float64(100), *maximum, epsilon)
	})

	t.Run(`lookup should find "collectionFormat"`, func(t *testing.T) {
		var f string
		res, err := testItems.JSONLookup("collectionFormat")
		require.NoError(t, err)
		require.NotNil(t, res)
		require.IsType(t, f, res)

		f, ok := res.(string)
		require.True(t, ok)
		assert.Equal(t, "csv", f)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := testItems.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestItemsWithValidation(t *testing.T) {
	i := new(Items).WithValidations(CommonValidations{MaxLength: conv.Pointer(int64(15))})
	assert.Equal(t, conv.Pointer(int64(15)), i.MaxLength)
}
