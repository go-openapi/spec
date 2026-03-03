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

var responses = Responses{ //nolint:gochecknoglobals // test fixture
	VendorExtensible: VendorExtensible{
		Extensions: map[string]any{
			"x-go-name": "PutDogExists",
		},
	},
	ResponsesProps: ResponsesProps{
		StatusCodeResponses: map[int]Response{
			200: {
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
			},
		},
	},
}

const responsesJSON = `{
	"x-go-name": "PutDogExists",
	"200": {
		"$ref": "Dog",
		"x-go-name": "PutDogExists",
		"description": "Dog exists",
		"schema": {
			"type": "string"
		}
	}
}`

func TestIntegrationResponses(t *testing.T) {
	assert.JSONUnmarshalAsT(t, responses, responsesJSON)
}

func TestJSONLookupResponses(t *testing.T) {
	resp200, ok := responses.StatusCodeResponses[200]
	require.TrueT(t, ok)

	res, err := resp200.JSONLookup("$ref")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, &Ref{}, res)

	ref, ok := res.(*Ref)
	require.TrueT(t, ok)
	assert.Equal(t, MustCreateRef("Dog"), *ref)

	var def string
	res, err = resp200.JSONLookup("description")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, def, res)

	def, ok = res.(string)
	require.TrueT(t, ok)
	assert.EqualT(t, "Dog exists", def)

	var x *any
	res, err = responses.JSONLookup("x-go-name")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, x, res)

	x, ok = res.(*any)
	require.TrueT(t, ok)
	assert.EqualValues(t, "PutDogExists", *x)

	res, err = responses.JSONLookup("unknown")
	require.Error(t, err)
	require.Nil(t, res)
}

func TestResponsesBuild(t *testing.T) {
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
