// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
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
	// an empty tuple is still a tuple: "null" would not be a schema
	assert.JSONMarshalAsT(t, `[]`, SchemaOrArray{Schemas: []Schema{}})
	assert.JSONMarshalAsT(t, `null`, SchemaOrArray{})
}

func TestSerialization_RoundTripEmptyItems(t *testing.T) {
	const document = `{"definitions":{"A":{"items":[]}}}`

	var doc Swagger
	require.NoError(t, json.Unmarshal([]byte(document), &doc))

	out, err := json.Marshal(doc)
	require.NoError(t, err)
	assert.JSONEqT(t, `{"paths":null,"definitions":{"A":{"items":[]}}}`, string(out))
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
