// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

// Copyright 2017 go-swagger maintainers
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
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var response = Response{ //nolint:gochecknoglobals // test fixture
	Refable: Refable{Ref: MustCreateRef("Dog")},
	VendorExtensible: VendorExtensible{
		Extensions: map[string]any{
			"x-go-name": "PutDogExists",
		},
	},
	ResponseProps: ResponseProps{
		Description: "Dog exists",
		Schema:      &Schema{SchemaProps: SchemaProps{Type: []string{"string"}}},
	},
}

const responseJSON = `{
	"$ref": "Dog",
	"x-go-name": "PutDogExists",
	"description": "Dog exists",
	"schema": {
		"type": "string"
	}
}`

func TestIntegrationResponse(t *testing.T) {
	assert.JSONUnmarshalAsT(t, response, responseJSON)
}

func TestJSONLookupResponse(t *testing.T) {
	res, err := response.JSONLookup("$ref")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, &Ref{}, res)

	var ok bool
	ref, ok := res.(*Ref)
	require.TrueT(t, ok)
	assert.Equal(t, MustCreateRef("Dog"), *ref)

	var def string
	res, err = response.JSONLookup("description")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, def, res)

	def, ok = res.(string)
	require.TrueT(t, ok)
	assert.EqualT(t, "Dog exists", def)

	var x *any
	res, err = response.JSONLookup("x-go-name")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, x, res)

	x, ok = res.(*any)
	require.TrueT(t, ok)
	assert.EqualValues(t, "PutDogExists", *x)

	res, err = response.JSONLookup("unknown")
	require.Error(t, err)
	require.Nil(t, res)
}

func TestResponseBuild(t *testing.T) {
	resp := NewResponse().
		WithDescription("some response").
		WithSchema(new(Schema).Typed("object", "")).
		AddHeader("x-header", ResponseHeader().Typed("string", "")).
		AddExample("application/json", `{"key":"value"}`)
	assert.JSONMarshalAsT(t, `{
         "description": "some response",
         "schema": {
          "type": "object"
         },
         "headers": {
          "x-header": {
           "type": "string"
          }
         },
         "examples": {
          "application/json": "{\"key\":\"value\"}"
         }
			 }`, resp)
}

func TestResponseRef(t *testing.T) {
	r := ResponseRef("Dog")
	assert.Equal(t, MustCreateRef("Dog"), r.Ref)
	assert.JSONMarshalAsT(t, `{"$ref":"Dog"}`, r)
}

func TestResponseRemoveHeader(t *testing.T) {
	r := NewResponse().
		AddHeader("X-Rate-Limit", ResponseHeader().Typed("integer", "int32")).
		AddHeader("X-Trace-Id", ResponseHeader().Typed("string", ""))
	require.Len(t, r.Headers, 2)

	r = r.RemoveHeader("X-Trace-Id")
	require.Len(t, r.Headers, 1)
	_, ok := r.Headers["X-Trace-Id"]
	assert.FalseT(t, ok)

	// removing an absent header is a no-op
	r = r.RemoveHeader("X-Absent")
	assert.Len(t, r.Headers, 1)

	// AddHeader with a nil header removes it too
	r = r.AddHeader("X-Rate-Limit", nil)
	assert.Empty(t, r.Headers)
}
