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
)

var responses = Responses{
	VendorExtensible: VendorExtensible{
		Extensions: map[string]interface{}{
			"x-go-name": "PutDogExists",
		},
	},
	ResponsesProps: ResponsesProps{
		StatusCodeResponses: map[int]Response{
			200: {
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
	var actual Responses
	if assert.NoError(t, json.Unmarshal([]byte(responsesJSON), &actual)) {
		assert.EqualValues(t, actual, responses)
	}

	assertParsesJSON(t, responsesJSON, responses)
}

func TestJSONLookupResponses(t *testing.T) {
	resp200, ok := responses.StatusCodeResponses[200]
	if !assert.True(t, ok) {
		t.FailNow()
		return
	}

	res, err := resp200.JSONLookup("$ref")
	if !assert.NoError(t, err) {
		t.FailNow()
		return
	}
	if assert.IsType(t, &Ref{}, res) {
		ref := res.(*Ref)
		assert.EqualValues(t, MustCreateRef("Dog"), *ref)
	}

	var def string
	res, err = resp200.JSONLookup("description")
	if !assert.NoError(t, err) || !assert.NotNil(t, res) || !assert.IsType(t, def, res) {
		t.FailNow()
		return
	}
	def = res.(string)
	assert.Equal(t, "Dog exists", def)

	var x *interface{}
	res, err = responses.JSONLookup("x-go-name")
	if !assert.NoError(t, err) || !assert.NotNil(t, res) || !assert.IsType(t, x, res) {
		t.FailNow()
		return
	}

	x = res.(*interface{})
	assert.EqualValues(t, "PutDogExists", *x)

	res, err = responses.JSONLookup("unknown")
	if !assert.Error(t, err) || !assert.Nil(t, res) {
		t.FailNow()
		return
	}
}

func TestResponsesBuild(t *testing.T) {
	resp := NewResponse().
		WithDescription("some response").
		WithSchema(new(Schema).Typed("object", "")).
		AddHeader("x-header", ResponseHeader().Typed("string", "")).
		AddExample("application/json", `{"key":"value"}`)
	jazon, _ := json.MarshalIndent(resp, "", " ")
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
