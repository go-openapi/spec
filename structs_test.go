// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	yaml "go.yaml.in/yaml/v3"
)

func assertSerializeJSON(t testing.TB, actual any, expected string) bool {
	ser, err := json.Marshal(actual)
	if err != nil {
		return assert.Failf(t, "unable to marshal to json", "got: %v: %#v", err, actual)
	}

	return assert.Equal(t, expected, string(ser))
}

func assertSerializeYAML(t testing.TB, actual any, expected string) bool {
	ser, err := yaml.Marshal(actual)
	if err != nil {
		return assert.Failf(t, "unable to marshal to yaml", "got: %v: %#v", err, actual)
	}
	return assert.Equal(t, expected, string(ser))
}

func derefTypeOf(expected any) (tpe reflect.Type) {
	tpe = reflect.TypeOf(expected)
	if tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}
	return
}

func isPointed(expected any) (pointed bool) {
	tpe := reflect.TypeOf(expected)
	if tpe.Kind() == reflect.Ptr {
		pointed = true
	}
	return
}

func assertParsesJSON(t testing.TB, actual string, expected any) bool {
	parsed := reflect.New(derefTypeOf(expected))
	err := json.Unmarshal([]byte(actual), parsed.Interface())
	if err != nil {
		return assert.Failf(t, "unable to unmarshal from json", "got: %v: %s", err, actual)
	}
	act := parsed.Interface()
	if !isPointed(expected) {
		act = reflect.Indirect(parsed).Interface()
	}
	return assert.Equal(t, expected, act)
}

func assertParsesYAML(t testing.TB, actual string, expected any) bool {
	parsed := reflect.New(derefTypeOf(expected))
	err := yaml.Unmarshal([]byte(actual), parsed.Interface())
	if err != nil {
		return assert.Failf(t, "unable to unmarshal from yaml", "got: %v: %s", err, actual)
	}
	act := parsed.Interface()
	if !isPointed(expected) {
		act = reflect.Indirect(parsed).Interface()
	}
	return assert.Equal(t, expected, act)
}

func TestSerialization_SerializeJSON(t *testing.T) {
	assertSerializeJSON(t, []string{"hello"}, "[\"hello\"]")
	assertSerializeJSON(t, []string{"hello", "world", "and", "stuff"}, "[\"hello\",\"world\",\"and\",\"stuff\"]")
	assertSerializeJSON(t, StringOrArray(nil), "null")
	assertSerializeJSON(t, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}}},
	}, "[{\"type\":\"string\"}]")
	assertSerializeJSON(t, SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		}}, "[{\"type\":\"string\"},{\"type\":\"string\"}]")
	assertSerializeJSON(t, SchemaOrArray{}, "null")
}

func TestSerialization_DeserializeJSON(t *testing.T) {
	// String
	assertParsesJSON(t, "\"hello\"", StringOrArray([]string{"hello"}))
	assertParsesJSON(t, "[\"hello\",\"world\",\"and\",\"stuff\"]",
		StringOrArray([]string{"hello", "world", "and", "stuff"}))
	assertParsesJSON(t, "[\"hello\",\"world\",null,\"stuff\"]", StringOrArray([]string{"hello", "world", "", "stuff"}))
	assertParsesJSON(t, "null", StringOrArray(nil))

	// Schema
	assertParsesJSON(t, "{\"type\":\"string\"}", SchemaOrArray{Schema: &Schema{
		SchemaProps: SchemaProps{Type: []string{"string"}}},
	})
	assertParsesJSON(t, "[{\"type\":\"string\"},{\"type\":\"string\"}]", &SchemaOrArray{
		Schemas: []Schema{
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
			{SchemaProps: SchemaProps{Type: []string{"string"}}},
		},
	})
	assertParsesJSON(t, "null", SchemaOrArray{})
}
