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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var response = Response{
	Refable: Refable{Ref: MustCreateRef("Dog")},
	VendorExtensible: VendorExtensible{
		Extensions: map[string]interface{}{
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
	var actual Response
	require.NoError(t, json.Unmarshal([]byte(responseJSON), &actual))
	assert.EqualValues(t, actual, response)

	assertParsesJSON(t, responseJSON, response)
}

func TestJSONLookupResponse(t *testing.T) {
	res, err := response.JSONLookup("$ref")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, &Ref{}, res)

	var ok bool
	ref, ok := res.(*Ref)
	require.True(t, ok)
	assert.EqualValues(t, MustCreateRef("Dog"), *ref)

	var def string
	res, err = response.JSONLookup("description")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, def, res)

	def, ok = res.(string)
	require.True(t, ok)
	assert.Equal(t, "Dog exists", def)

	var x *interface{}
	res, err = response.JSONLookup("x-go-name")
	require.NoError(t, err)
	require.NotNil(t, res)
	require.IsType(t, x, res)

	x, ok = res.(*interface{})
	require.True(t, ok)
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
	jazon, err := json.MarshalIndent(resp, "", " ")
	require.NoError(t, err)

	assert.JSONEq(t, `{
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
			 }`, string(jazon))
}
