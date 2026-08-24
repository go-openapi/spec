// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

//nolint:gochecknoglobals // it's okay to have embedded test fixtures as globals
var (
	specJSON        []byte
	minimalJSONSpec []byte
	miniJSONSpec    []byte
)

func init() { //nolint:gochecknoinits // it's okay to load embedded fixtures in init().
	// load embedded fixtures

	var err error
	specJSON, err = fixtureAssets.ReadFile("testdata/specs/spec.json")
	if err != nil {
		panic(fmt.Sprintf("could not find fixture: %v", err))
	}

	minimalJSONSpec, err = fixtureAssets.ReadFile("testdata/specs/minimal_spec.json")
	if err != nil {
		panic(fmt.Sprintf("could not find fixture: %v", err))
	}

	miniJSONSpec, err = fixtureAssets.ReadFile("testdata/specs/mini_spec.json")
	if err != nil {
		panic(fmt.Sprintf("could not find fixture: %v", err))
	}
}

var spec = Swagger{ //nolint:gochecknoglobals // test fixture
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
	VendorExtensible: VendorExtensible{Extensions: map[string]any{
		"x-some-extension": "vendor",
		"x-schemes":        []any{"unix", "amqp"},
	}},
}

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
			if typeutil.IsZero(expectedValue) && typeutils.IsZero(reflect.ValueOf(actual)) {
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
			if typeutils.IsZero(expectedValue) && typeutils.IsZero(reflect.ValueOf(actual)) {
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

func assertSpecs(t testing.TB, actual, expected Swagger) {
	t.Helper()
	expected.Swagger = "2.0"
	assert.Equal(t, expected, actual)
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
	expected := make(map[string]any)
	_ = json.Unmarshal(specJSON, &expected)
	b, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err)
	var actual map[string]any
	require.NoError(t, json.Unmarshal(b, &actual))
	assert.Equal(t, expected, actual)
}

func TestSwaggerSpec_Deserialize(t *testing.T) {
	var actual Swagger
	require.NoError(t, json.Unmarshal(specJSON, &actual))
	assert.Equal(t, actual, spec)
}

func TestVendorExtensionStringSlice(t *testing.T) {
	var actual Swagger
	require.NoError(t, json.Unmarshal(specJSON, &actual))
	schemes, ok := actual.Extensions.GetStringSlice("x-schemes")
	require.TrueT(t, ok)
	assert.Equal(t, []string{"unix", "amqp"}, schemes)

	notSlice, ok := actual.Extensions.GetStringSlice("x-some-extension")
	assert.Nil(t, notSlice)
	assert.FalseT(t, ok)

	actual.AddExtension("x-another-ext", 100)
	notString, ok := actual.Extensions.GetStringSlice("x-another-ext")
	assert.Nil(t, notString)
	assert.FalseT(t, ok)

	actual.AddExtension("x-another-slice-ext", []any{100, 100})
	notStringSlice, ok := actual.Extensions.GetStringSlice("x-another-slice-ext")
	assert.Nil(t, notStringSlice)
	assert.FalseT(t, ok)

	_, ok = actual.Extensions.GetStringSlice("x-notfound-ext")
	assert.FalseT(t, ok)
}

func TestOptionalSwaggerProps_Serialize(t *testing.T) {
	var minimalSpec Swagger
	err := json.Unmarshal(miniJSONSpec, &minimalSpec)
	require.NoError(t, err)
	bytes, err := json.Marshal(&minimalSpec)
	require.NoError(t, err)

	var ms map[string]any
	require.NoError(t, json.Unmarshal(bytes, &ms))

	assert.MapNotContainsT(t, ms, "consumes")
	assert.MapNotContainsT(t, ms, "produces")
	assert.MapNotContainsT(t, ms, "schemes")
	assert.MapNotContainsT(t, ms, "host")
	assert.MapNotContainsT(t, ms, "basePath")
	assert.MapNotContainsT(t, ms, "definitions")
	assert.MapNotContainsT(t, ms, "parameters")
	assert.MapNotContainsT(t, ms, "responses")
	assert.MapNotContainsT(t, ms, "securityDefinitions")
	assert.MapNotContainsT(t, ms, "security")
	assert.MapNotContainsT(t, ms, "tags")
	assert.MapNotContainsT(t, ms, "externalDocs")
}

func TestSecurityRequirements(t *testing.T) {
	var minimalSpec Swagger
	require.NoError(t, json.Unmarshal(minimalJSONSpec, &minimalSpec))

	sec := minimalSpec.Paths.Paths["/"].Get.Security
	require.Len(t, sec, 3)
	assert.MapContainsT(t, sec[0], "basic")
	assert.MapContainsT(t, sec[0], "apiKey")
	assert.NotNil(t, sec[1])
	assert.Empty(t, sec[1])
	assert.MapContainsT(t, sec[2], "queryKey")
}

func TestSwaggerGobEncoding(t *testing.T) {
	doTestSwaggerGobEncoding(t, specJSON)

	doTestSwaggerGobEncoding(t, minimalJSONSpec)
}

func doTestSwaggerGobEncoding(t *testing.T, fixture []byte) {
	t.Helper()

	var src, dst Swagger
	require.NoError(t, json.Unmarshal(fixture, &src))

	doTestAnyGobEncoding(t, &src, &dst)
}

func TestJSONLookupSchemaOrBool(t *testing.T) {
	s := SchemaOrBool{Allows: true, Schema: Int32Property()}

	t.Run(`lookup should find "allows"`, func(t *testing.T) {
		res, err := s.JSONLookup("allows")
		require.NoError(t, err)

		allows, ok := res.(bool)
		require.TrueT(t, ok)
		assert.TrueT(t, allows)
	})

	t.Run(`lookup should delegate to the schema`, func(t *testing.T) {
		res, err := s.JSONLookup("format")
		require.NoError(t, err)

		format, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "int32", format)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := s.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestJSONLookupSchemaOrStringArray(t *testing.T) {
	s := SchemaOrStringArray{Schema: StrFmtProperty("uuid")}

	res, err := s.JSONLookup("format")
	require.NoError(t, err)

	format, ok := res.(string)
	require.TrueT(t, ok)
	assert.EqualT(t, "uuid", format)

	res, err = s.JSONLookup("unknown")
	require.Error(t, err)
	require.Nil(t, res)
}

func TestJSONLookupSchemaOrArray(t *testing.T) {
	t.Run("a numeric token should index the schemas", func(t *testing.T) {
		s := SchemaOrArray{Schemas: []Schema{*StringProperty(), *Int64Property()}}

		res, err := s.JSONLookup("1")
		require.NoError(t, err)

		sch, ok := res.(Schema)
		require.TrueT(t, ok)
		assert.EqualT(t, "int64", sch.Format)

		res, err = s.JSONLookup("42")
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("a named token should address the single schema", func(t *testing.T) {
		s := SchemaOrArray{Schema: Int32Property()}

		res, err := s.JSONLookup("format")
		require.NoError(t, err)

		format, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "int32", format)

		res, err = s.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestSchemaOrArrayLenAndContainsType(t *testing.T) {
	t.Run("a single schema counts as one", func(t *testing.T) {
		s := SchemaOrArray{Schema: StringProperty()}
		assert.EqualT(t, 1, s.Len())
		assert.TrueT(t, s.ContainsType("string"))
		assert.FalseT(t, s.ContainsType("integer"))
	})

	t.Run("a schema list counts its members", func(t *testing.T) {
		s := SchemaOrArray{Schemas: []Schema{*StringProperty(), *Int64Property()}}
		assert.EqualT(t, 2, s.Len())
		// ContainsType only looks at the single schema
		assert.FalseT(t, s.ContainsType("string"))
	})

	t.Run("an untyped single schema contains no type", func(t *testing.T) {
		s := SchemaOrArray{Schema: new(Schema)}
		assert.FalseT(t, s.ContainsType("string"))
	})

	t.Run("the zero value is empty", func(t *testing.T) {
		var s SchemaOrArray
		assert.EqualT(t, 0, s.Len())
		assert.FalseT(t, s.ContainsType("string"))
	})
}

func TestStringOrArrayContains(t *testing.T) {
	s := StringOrArray{"string", "null"}
	assert.TrueT(t, s.Contains("null"))
	assert.FalseT(t, s.Contains("integer"))
}

func TestJSONLookupSwagger(t *testing.T) {
	var sp Swagger
	require.NoError(t, json.Unmarshal(specJSON, &sp))

	t.Run("lookup should find an extension", func(t *testing.T) {
		res, err := sp.JSONLookup("x-schemes")
		require.NoError(t, err)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, []any{"unix", "amqp"}, *ext)
	})

	t.Run(`lookup should find "swagger"`, func(t *testing.T) {
		res, err := sp.JSONLookup("swagger")
		require.NoError(t, err)

		version, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "2.0", version)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := sp.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestSchemaOrStringArrayMarshalJSON(t *testing.T) {
	t.Run("a property list should marshal as an array", func(t *testing.T) {
		s := SchemaOrStringArray{Property: []string{"a", "b"}}
		assert.JSONMarshalAsT(t, `["a","b"]`, s)
	})

	t.Run("the zero value should marshal as null", func(t *testing.T) {
		var s SchemaOrStringArray
		assert.JSONMarshalAsT(t, `null`, s)
	})
}

func TestStringOrArrayUnmarshalJSON(t *testing.T) {
	t.Run("should unmarshal an array", func(t *testing.T) {
		var s StringOrArray
		require.NoError(t, json.Unmarshal([]byte(`["string","null"]`), &s))
		assert.Equal(t, StringOrArray{"string", "null"}, s)
	})

	t.Run("should unmarshal a single string", func(t *testing.T) {
		var s StringOrArray
		require.NoError(t, json.Unmarshal([]byte(`"string"`), &s))
		assert.Equal(t, StringOrArray{"string"}, s)
	})

	t.Run("should leave null untouched", func(t *testing.T) {
		var s StringOrArray
		require.NoError(t, json.Unmarshal([]byte(`null`), &s))
		assert.Nil(t, s)
	})

	t.Run("should reject anything else", func(t *testing.T) {
		var s StringOrArray
		require.ErrorIs(t, json.Unmarshal([]byte(`12`), &s), ErrSpec)
	})

	t.Run("should report a malformed array", func(t *testing.T) {
		var s StringOrArray
		require.Error(t, json.Unmarshal([]byte(`[12]`), &s))
	})
}
