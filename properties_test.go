// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"
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

	var propSerData = []struct {
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
		assertSerializeJSON(t, v.Schema, v.JSON)
		assertParsesJSON(t, v.JSON, v.Schema)
	}

}
