// Copyright 2015 go-swagger maintainers
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

var spec = Swagger{
	SwaggerProps: SwaggerProps{
		ID:          "http://localhost:3849/api-docs",
		Swagger:     "2.0",
		Consumes:    []string{"application/json", "application/x-yaml"},
		Produces:    []string{"application/json"},
		Schemes:     []string{"http", "https"},
		Info:        &testInfo,
		Host:        "some.api.out.there",
		BasePath:    "/",
		Paths:       &paths,
		Definitions: map[string]Schema{"Category": {SchemaProps: SchemaProps{Type: []string{"string"}}}},
		Parameters: map[string]Parameter{
			"categoryParam": {ParamProps: ParamProps{Name: "category", In: "query"}, SimpleSchema: SimpleSchema{Type: "string"}},
		},
		Responses: map[string]Response{
			"EmptyAnswer": {
				ResponseProps: ResponseProps{
					Description: "no data to return for this operation",
				},
			},
		},
		SecurityDefinitions: map[string]*SecurityScheme{
			"internalApiKey": APIKeyAuth("api_key", "header"),
		},
		Security: []map[string][]string{
			{"internalApiKey": {}},
		},
		Tags:         []Tag{NewTag("pets", "", nil)},
		ExternalDocs: &ExternalDocumentation{Description: "the name", URL: "the url"},
	},
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{
		"x-some-extension": "vendor",
		"x-schemes":        []interface{}{"unix", "amqp"},
	}},
}

const specJSON = `{
	"id": "http://localhost:3849/api-docs",
	"consumes": ["application/json", "application/x-yaml"],
	"produces": ["application/json"],
	"schemes": ["http", "https"],
	"swagger": "2.0",
	"info": {
		"contact": {
			"name": "wordnik api team",
			"url": "http://developer.wordnik.com"
		},
		"description": "A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0` +
	` specification",
		"license": {
			"name": "Creative Commons 4.0 International",
			"url": "http://creativecommons.org/licenses/by/4.0/"
		},
		"termsOfService": "http://helloreverb.com/terms/",
		"title": "Swagger Sample API",
		"version": "1.0.9-abcd",
		"x-framework": "go-swagger"
	},
	"host": "some.api.out.there",
	"basePath": "/",
	"paths": {"x-framework":"go-swagger","/":{"$ref":"cats"}},
	"definitions": { "Category": { "type": "string"} },
	"parameters": {
		"categoryParam": {
			"name": "category",
			"in": "query",
			"type": "string"
		}
	},
	"responses": { "EmptyAnswer": { "description": "no data to return for this operation" } },
	"securityDefinitions": {
		"internalApiKey": {
			"type": "apiKey",
			"in": "header",
			"name": "api_key"
		}
	},
	"security": [{"internalApiKey":[]}],
	"tags": [{"name":"pets"}],
	"externalDocs": {"description":"the name","url":"the url"},
	"x-some-extension": "vendor",
	"x-schemes": ["unix","amqp"]
}`

// func verifySpecSerialize(specJSON []byte, spec Swagger) {
// 	expected := map[string]interface{}{}
// 	json.Unmarshal(specJSON, &expected)
// 	b, err := json.MarshalIndent(spec, "", "  ")
// 	So(err, ShouldBeNil)
// 	var actual map[string]interface{}
// 	err = json.Unmarshal(b, &actual)
// 	So(err, ShouldBeNil)
// 	compareSpecMaps(actual, expected)
// }

/*
	// assertEquivalent is currently unused
	func assertEquivalent(t testing.TB, actual, expected interface{}) bool {
		if actual == nil || expected == nil || reflect.DeepEqual(actual, expected) {
			return true
		}

		actualType := reflect.TypeOf(actual)
		expectedType := reflect.TypeOf(expected)
		if reflect.TypeOf(actual).ConvertibleTo(expectedType) {
			expectedValue := reflect.ValueOf(expected)
			if swag.IsZero(expectedValue) && swag.IsZero(reflect.ValueOf(actual)) {
				return true
			}

			// Attempt comparison after type conversion
			if reflect.DeepEqual(actual, expectedValue.Convert(actualType).Interface()) {
				return true
			}
		}

		// Last ditch effort
		if fmt.Sprintf("%#v", expected) == fmt.Sprintf("%#v", actual) {
			return true
		}
		errFmt := "Expected: '%[1]T(%[1]#v)'\nActual:   '%[2]T(%[2]#v)'\n(Should be equivalent)!"
		return assert.Fail(t, errFmt, expected, actual)
	}

	// ShouldBeEquivalentTo is currently unused
	func ShouldBeEquivalentTo(actual interface{}, expecteds ...interface{}) string {
		expected := expecteds[0]
		if actual == nil || expected == nil {
			return ""
		}

		if reflect.DeepEqual(expected, actual) {
			return ""
		}

		actualType := reflect.TypeOf(actual)
		expectedType := reflect.TypeOf(expected)
		if reflect.TypeOf(actual).ConvertibleTo(expectedType) {
			expectedValue := reflect.ValueOf(expected)
			if swag.IsZero(expectedValue) && swag.IsZero(reflect.ValueOf(actual)) {
				return ""
			}

			// Attempt comparison after type conversion
			if reflect.DeepEqual(actual, expectedValue.Convert(actualType).Interface()) {
				return ""
			}
		}

		// Last ditch effort
		if fmt.Sprintf("%#v", expected) == fmt.Sprintf("%#v", actual) {
			return ""
		}
		errFmt := "Expected: '%[1]T(%[1]#v)'\nActual:   '%[2]T(%[2]#v)'\n(Should be equivalent)!"
		return fmt.Sprintf(errFmt, expected, actual)
	}

	// assertSpecMaps is currently unused
	func assertSpecMaps(t testing.TB, actual, expected map[string]interface{}) bool {
		res := true
		if id, ok := expected["id"]; ok {
			res = assert.Equal(t, id, actual["id"])
		}
		res = res && assert.Equal(t, expected["consumes"], actual["consumes"])
		res = res && assert.Equal(t, expected["produces"], actual["produces"])
		res = res && assert.Equal(t, expected["schemes"], actual["schemes"])
		res = res && assert.Equal(t, expected["swagger"], actual["swagger"])
		res = res && assert.Equal(t, expected["info"], actual["info"])
		res = res && assert.Equal(t, expected["host"], actual["host"])
		res = res && assert.Equal(t, expected["basePath"], actual["basePath"])
		res = res && assert.Equal(t, expected["paths"], actual["paths"])
		res = res && assert.Equal(t, expected["definitions"], actual["definitions"])
		res = res && assert.Equal(t, expected["responses"], actual["responses"])
		res = res && assert.Equal(t, expected["securityDefinitions"], actual["securityDefinitions"])
		res = res && assert.Equal(t, expected["tags"], actual["tags"])
		res = res && assert.Equal(t, expected["externalDocs"], actual["externalDocs"])
		res = res && assert.Equal(t, expected["x-some-extension"], actual["x-some-extension"])
		res = res && assert.Equal(t, expected["x-schemes"], actual["x-schemes"])

		return res
	}
*/

func assertSpecs(t testing.TB, actual, expected Swagger) bool {
	expected.Swagger = "2.0"
	return assert.Equal(t, expected, actual)
}

/*
// assertSpecJSON is currently unused
func assertSpecJSON(t testing.TB, specJSON []byte) bool {
	var expected map[string]interface{}
	if !assert.NoError(t, json.Unmarshal(specJSON, &expected)) {
		return false
	}

	obj := Swagger{}
	if !assert.NoError(t, json.Unmarshal(specJSON, &obj)) {
		return false
	}

	cb, err := json.MarshalIndent(obj, "", "  ")
	if assert.NoError(t, err) {
		return false
	}
	var actual map[string]interface{}
	if !assert.NoError(t, json.Unmarshal(cb, &actual)) {
		return false
	}
	return assertSpecMaps(t, expected, actual )
}
*/

func TestSwaggerSpec_Serialize(t *testing.T) {
	expected := make(map[string]interface{})
	_ = json.Unmarshal([]byte(specJSON), &expected)
	b, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err)
	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &actual))
	assert.EqualValues(t, expected, actual)
}

func TestSwaggerSpec_Deserialize(t *testing.T) {
	var actual Swagger
	require.NoError(t, json.Unmarshal([]byte(specJSON), &actual))
	assert.EqualValues(t, actual, spec)
}

func TestVendorExtensionStringSlice(t *testing.T) {
	var actual Swagger
	require.NoError(t, json.Unmarshal([]byte(specJSON), &actual))
	schemes, ok := actual.Extensions.GetStringSlice("x-schemes")
	require.True(t, ok)
	assert.EqualValues(t, []string{"unix", "amqp"}, schemes)

	notSlice, ok := actual.Extensions.GetStringSlice("x-some-extension")
	assert.Nil(t, notSlice)
	assert.False(t, ok)

	actual.AddExtension("x-another-ext", 100)
	notString, ok := actual.Extensions.GetStringSlice("x-another-ext")
	assert.Nil(t, notString)
	assert.False(t, ok)

	actual.AddExtension("x-another-slice-ext", []interface{}{100, 100})
	notStringSlice, ok := actual.Extensions.GetStringSlice("x-another-slice-ext")
	assert.Nil(t, notStringSlice)
	assert.False(t, ok)

	_, ok = actual.Extensions.GetStringSlice("x-notfound-ext")
	assert.False(t, ok)
}

func TestOptionalSwaggerProps_Serialize(t *testing.T) {
	minimalJSONSpec := []byte(`{
	"swagger": "2.0",
	"info": {
		"version": "0.0.0",
		"title": "Simple API"
	},
	"paths": {
		"/": {
			"get": {
				"responses": {
					"200": {
						"description": "OK"
					}
				}
			}
		}
	}
}`)

	var minimalSpec Swagger
	err := json.Unmarshal(minimalJSONSpec, &minimalSpec)
	require.NoError(t, err)
	bytes, err := json.Marshal(&minimalSpec)
	require.NoError(t, err)

	var ms map[string]interface{}
	require.NoError(t, json.Unmarshal(bytes, &ms))

	assert.NotContains(t, ms, "consumes")
	assert.NotContains(t, ms, "produces")
	assert.NotContains(t, ms, "schemes")
	assert.NotContains(t, ms, "host")
	assert.NotContains(t, ms, "basePath")
	assert.NotContains(t, ms, "definitions")
	assert.NotContains(t, ms, "parameters")
	assert.NotContains(t, ms, "responses")
	assert.NotContains(t, ms, "securityDefinitions")
	assert.NotContains(t, ms, "security")
	assert.NotContains(t, ms, "tags")
	assert.NotContains(t, ms, "externalDocs")
}

var minimalJSONSpec = []byte(`{
		"swagger": "2.0",
		"info": {
			"version": "0.0.0",
			"title": "Simple API"
		},
		"securityDefinitions": {
			"basic": {
				"type": "basic"
			},
			"apiKey": {
				"type": "apiKey",
				"in": "header",
				"name": "X-API-KEY"
			},
			"queryKey": {
				"type": "apiKey",
				"in": "query",
				"name": "api_key"
			}
		},
		"paths": {
			"/": {
				"get": {
					"security": [
						{
							"apiKey": [],
							"basic": []
						},
						{},
						{
							"queryKey": [],
							"basic": []
						}
					],
					"responses": {
						"200": {
							"description": "OK"
						}
					}
				}
			}
		}
	}`)

func TestSecurityRequirements(t *testing.T) {
	var minimalSpec Swagger
	require.NoError(t, json.Unmarshal(minimalJSONSpec, &minimalSpec))

	sec := minimalSpec.Paths.Paths["/"].Get.Security
	require.Len(t, sec, 3)
	assert.Contains(t, sec[0], "basic")
	assert.Contains(t, sec[0], "apiKey")
	assert.NotNil(t, sec[1])
	assert.Empty(t, sec[1])
	assert.Contains(t, sec[2], "queryKey")
}

func TestSwaggerGobEncoding(t *testing.T) {
	doTestSwaggerGobEncoding(t, specJSON)

	doTestSwaggerGobEncoding(t, string(minimalJSONSpec))
}

func doTestSwaggerGobEncoding(t *testing.T, fixture string) {
	var src, dst Swagger
	require.NoError(t, json.Unmarshal([]byte(fixture), &src))

	doTestAnyGobEncoding(t, &src, &dst)
}
