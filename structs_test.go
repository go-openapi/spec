// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestSerialization_SerializeJSON(t *testing.T) {
	assert.JSONMarshalAsT(t, `["hello"]`, []string{"hello"})
	assert.JSONMarshalAsT(t, `["hello","world","and","stuff"]`, []string{"hello", "world", "and", "stuff"})
	assert.JSONMarshalAsT(t, `null`, StringOrArray(nil))
	assert.JSONMarshalAsT(t, `[{"type":"string"}]`, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
	})
	assert.JSONMarshalAsT(t, `[{"type":"string"},{"type":"string"}]`, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
	})
	assert.JSONMarshalAsT(t, `null`, SchemaOrArray{})
}

func TestSerialization_DeserializeJSON(t *testing.T) {
	// String
	assert.JSONUnmarshalAsT(t, StringOrArray([]string{"hello"}), `"hello"`)
	assert.JSONUnmarshalAsT(t,
		StringOrArray([]string{"hello", "world", "and", "stuff"}),
		`["hello","world","and","stuff"]`)
	assert.JSONUnmarshalAsT(t,
		StringOrArray([]string{"hello", "world", "", "stuff"}),
		`["hello","world",null,"stuff"]`)
	assert.JSONUnmarshalAsT(t, StringOrArray(nil), `null`)

	// Schema
	assert.JSONUnmarshalAsT(t, SchemaOrArray{
		Schema: &Schema{
			SchemaProps: SchemaProps{Type: []string{"string"}},
		},
	}, `{"type":"string"}`)
	assert.JSONUnmarshalAsT(t, &SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
	}, `[{"type":"string"},{"type":"string"}]`)
	assert.JSONUnmarshalAsT(t, SchemaOrArray{}, `null`)
}
