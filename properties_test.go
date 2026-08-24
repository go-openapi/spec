// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"sort"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestPropertySerialization(t *testing.T) {
	strProp := StringProperty()
	strProp.Enum = append(strProp.Enum, "a", "b")

	prop := &Schema{SchemaProps: SchemaProps{
		Items: &SchemaOrArray{Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		}},
	}}

	propSerData := []struct {
		Schema *Schema
		JSON   string
	}{
		{BooleanProperty(), `{"type":"boolean"}`},
		{DateProperty(), `{"type":"string","format":"date"}`},
		{DateTimeProperty(), `{"type":"string","format":"date-time"}`},
		{Float64Property(), `{"type":"number","format":"double"}`},
		{Float32Property(), `{"type":"number","format":"float"}`},
		{Int32Property(), `{"type":"integer","format":"int32"}`},
		{Int64Property(), `{"type":"integer","format":"int64"}`},
		{MapProperty(StringProperty()), `{"type":"object","additionalProperties":{"type":"string"}}`},
		{MapProperty(Int32Property()), `{"type":"object","additionalProperties":{"type":"integer","format":"int32"}}`},
		{RefProperty("Dog"), `{"$ref":"Dog"}`},
		{StringProperty(), `{"type":"string"}`},
		{strProp, `{"type":"string","enum":["a","b"]}`},
		{ArrayProperty(StringProperty()), `{"type":"array","items":{"type":"string"}}`},
		{prop, `{"items":[{"type":"string"},{"type":"string"}]}`},
	}

	for _, v := range propSerData {
		t.Log("roundtripping for", v.JSON)
		assert.JSONMarshalAsT(t, v.JSON, v.Schema)
		assert.JSONUnmarshalAsT(t, v.Schema, v.JSON)
	}
}

func TestOrderedSchemaItem_Issue216(t *testing.T) {
	stringSchema := new(Schema).Typed("string", "")
	items := OrderSchemaItems{
		{
			Name:   "emails\n", // Key contains newline character
			Schema: *stringSchema,
		},
		{
			Name:   "regular",
			Schema: *stringSchema,
		},
	}

	jazon, err := items.MarshalJSON()
	require.NoError(t, err)

	require.JSONEqBytes(t,
		[]byte(`{"emails\n":{"type":"string"},"regular":{"type":"string"}}`),
		jazon,
	)
}

func TestOrderSchemaItemsLess(t *testing.T) {
	withOrder := func(name string, order any) OrderSchemaItem {
		s := new(Schema).Typed("string", "")
		s.AddExtension("x-order", order)

		return OrderSchemaItem{Name: name, Schema: *s}
	}
	plain := func(name string) OrderSchemaItem {
		return OrderSchemaItem{Name: name, Schema: *new(Schema).Typed("string", "")}
	}

	t.Run("x-order should rank both items", func(t *testing.T) {
		items := OrderSchemaItems{withOrder("second", float64(2)), withOrder("first", float64(1))}
		sort.Sort(items)
		assert.EqualT(t, "first", items[0].Name)
		assert.EqualT(t, "second", items[1].Name)
	})

	t.Run("an ordered item should sort before an unordered one", func(t *testing.T) {
		items := OrderSchemaItems{plain("aaa"), withOrder("zzz", float64(1))}
		sort.Sort(items)
		assert.EqualT(t, "zzz", items[0].Name)
		assert.EqualT(t, "aaa", items[1].Name)
	})

	t.Run("x-order should be read from a string too", func(t *testing.T) {
		items := OrderSchemaItems{withOrder("second", "2"), withOrder("first", "1")}
		sort.Sort(items)
		assert.EqualT(t, "first", items[0].Name)
		assert.EqualT(t, "second", items[1].Name)
	})

	t.Run("unordered items should sort by name", func(t *testing.T) {
		items := OrderSchemaItems{plain("zzz"), plain("aaa")}
		sort.Sort(items)
		assert.EqualT(t, "aaa", items[0].Name)
		assert.EqualT(t, "zzz", items[1].Name)
	})
}

func TestOrderSchemaItemsMarshalJSON(t *testing.T) {
	t.Run("no items should marshal to an empty object", func(t *testing.T) {
		jazon, err := OrderSchemaItems{}.MarshalJSON()
		require.NoError(t, err)
		assert.EqualT(t, "{}", string(jazon))
	})

	t.Run("an unmarshalable schema should report the error", func(t *testing.T) {
		broken := new(Schema).Typed("string", "")
		broken.AddExtension("x-broken", func() {}) // funcs cannot be marshaled

		t.Run("as the first item", func(t *testing.T) {
			items := OrderSchemaItems{{Name: "broken", Schema: *broken}}
			_, err := items.MarshalJSON()
			require.Error(t, err)
		})

		t.Run("as a later item", func(t *testing.T) {
			items := OrderSchemaItems{
				{Name: "ok", Schema: *new(Schema).Typed("string", "")},
				{Name: "broken", Schema: *broken},
			}
			_, err := items.MarshalJSON()
			require.Error(t, err)
		})
	})
}

func TestSchemaPropertiesMarshalJSON(t *testing.T) {
	t.Run("nil properties should marshal to null", func(t *testing.T) {
		var properties SchemaProperties
		jazon, err := properties.MarshalJSON()
		require.NoError(t, err)
		assert.EqualT(t, "null", string(jazon))
	})

	t.Run("properties should marshal in x-order", func(t *testing.T) {
		first := new(Schema).Typed("string", "")
		first.AddExtension("x-order", float64(1))
		second := new(Schema).Typed("integer", "int32")
		second.AddExtension("x-order", float64(2))

		properties := SchemaProperties{"zzz": *first, "aaa": *second}
		jazon, err := properties.MarshalJSON()
		require.NoError(t, err)
		assert.EqualT(t,
			`{"zzz":{"type":"string","x-order":1},"aaa":{"type":"integer","format":"int32","x-order":2}}`,
			string(jazon),
		)
	})
}
