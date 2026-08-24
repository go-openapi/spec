// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// TestGob_KeepsZeroValuedBounds pins the whole document against gob's zero-value elision.
//
// gob omits a struct field holding the zero value and flattens a pointer to what it points at,
// so an optional number that is present and zero used to come back nil.
func TestGob_KeepsZeroValuedBounds(t *testing.T) {
	const doc = `{
	  "swagger": "2.0",
	  "info": {"title": "zero-valued bounds", "version": "1.0.0"},
	  "paths": {
	    "/x": {
	      "get": {
	        "parameters": [
	          {"name": "q", "in": "query", "type": "integer", "maximum": 0, "minimum": 0,
	           "multipleOf": 0, "maxLength": 0, "minLength": 0},
	          {"name": "a", "in": "query", "type": "array", "maxItems": 0, "minItems": 0,
	           "items": {"type": "integer", "maximum": 0, "minItems": 0}}
	        ],
	        "responses": {
	          "200": {"description": "ok", "headers": {"X-Count": {"type": "integer", "minimum": 0, "maxItems": 0}}}
	        }
	      }
	    }
	  },
	  "definitions": {
	    "A": {"type": "object", "maxProperties": 0, "minProperties": 0,
	      "properties": {"p": {"type": "number", "minimum": 0, "maximum": 0, "maxLength": 0}},
	      "items": {"type": "integer", "minItems": 0}}
	  }
	}`

	original := new(Swagger)
	require.NoError(t, json.Unmarshal([]byte(doc), original))

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(original))

	decoded := new(Swagger)
	require.NoError(t, gob.NewDecoder(&buf).Decode(decoded))

	before, err := json.Marshal(original)
	require.NoError(t, err)
	after, err := json.Marshal(decoded)
	require.NoError(t, err)

	assert.JSONEqT(t, string(before), string(after))
}

// TestGob_EveryOptionalNumberIsCarried walks the types that hold optional numbers and checks
// each of their pointer fields survives a gob round-trip when it points at zero.
//
// It reads the fields by reflection, so a new bound added to one of these types is covered
// without touching this test - and fails it until the encoder in gob.go carries it too.
func TestGob_EveryOptionalNumberIsCarried(t *testing.T) {
	for _, subject := range []struct {
		name  string
		value any
	}{
		{"Schema", &Schema{}},
		{"Parameter", &Parameter{}},
		{"Header", &Header{}},
		{"Items", &Items{}},
	} {
		t.Run(subject.name, func(t *testing.T) {
			fields := optionalNumbers(reflect.ValueOf(subject.value).Elem())
			require.NotEmpty(t, fields, "expected optional numbers on %s", subject.name)

			for _, field := range fields {
				t.Run(field, func(t *testing.T) {
					value := reflect.New(reflect.TypeOf(subject.value).Elem())
					setZeroPointer(t, value.Elem(), field)

					var buf bytes.Buffer
					require.NoError(t, gob.NewEncoder(&buf).EncodeValue(value))

					back := reflect.New(reflect.TypeOf(subject.value).Elem())
					require.NoError(t, gob.NewDecoder(&buf).DecodeValue(back))

					got := fieldByName(back.Elem(), field)
					require.FalseT(t, got.IsNil(), "%s.%s was dropped by gob", subject.name, field)
					assert.EqualT(t, float64(0), got.Elem().Convert(reflect.TypeFor[float64]()).Float())
				})
			}
		})
	}
}

// optionalNumbers returns the names of the *float64 and *int64 fields of a struct, embedded
// fields included.
func optionalNumbers(v reflect.Value) []string {
	var names []string
	var walk func(reflect.Value)
	walk = func(sv reflect.Value) {
		st := sv.Type()
		for i := range st.NumField() {
			f := st.Field(i)
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				walk(sv.Field(i))

				continue
			}
			if f.Type.Kind() != reflect.Pointer {
				continue
			}
			switch f.Type.Elem().Kind() {
			case reflect.Float64, reflect.Int64:
				names = append(names, f.Name)
			default:
			}
		}
	}
	walk(v)

	return names
}

func fieldByName(v reflect.Value, name string) reflect.Value {
	return v.FieldByName(name)
}

func setZeroPointer(t testing.TB, v reflect.Value, name string) {
	t.Helper()

	field := v.FieldByName(name)
	require.TrueT(t, field.IsValid(), "no field %q", name)
	field.Set(reflect.New(field.Type().Elem()))
}
