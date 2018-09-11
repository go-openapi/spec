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

package spec_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mimics what the go-openapi/load does
var (
	rex        = regexp.MustCompile(`"\$ref":\s*"(.+)"`)
	testLoader func(string) (json.RawMessage, error)
)

func init() {
	testLoader = func(path string) (json.RawMessage, error) {
		if swag.YAMLMatcher(path) {
			return swag.YAMLDoc(path)
		}
		data, err := swag.LoadFromFileOrHTTP(path)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(data), nil
	}
}
func loadOrFail(t *testing.T, path string) *spec.Swagger {
	raw, err := testLoader(path)
	require.NoErrorf(t, err, "can't load fixture %s: %v", path, err)
	swspec := new(spec.Swagger)
	err = json.Unmarshal(raw, swspec)
	require.NoError(t, err)
	return swspec
}

// Test unitary fixture for dev and bug fixing
func Test_Issue1429(t *testing.T) {
	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()
	spec.PathLoader = testLoader
	path := filepath.Join("fixtures", "bugs", "1429", "swagger.yaml")

	// load and full expand
	sp := loadOrFail(t, path)
	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)

	// assert well expanded
	require.Truef(t, (sp.Paths != nil && sp.Paths.Paths != nil), "expected paths to be available in fixture")

	assertPaths1429(t, sp)

	for _, def := range sp.Definitions {
		assert.Equal(t, "", def.Ref.String())
	}

	// reload and SkipSchemas: true
	sp = loadOrFail(t, path)
	err = spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true})
	require.NoError(t, err)

	// assert well resolved
	require.Truef(t, (sp.Paths != nil && sp.Paths.Paths != nil), "expected paths to be available in fixture")

	assertPaths1429SkipSchema(t, sp)

	for _, def := range sp.Definitions {
		assert.Contains(t, def.Ref.String(), "responses.yaml#/")
	}
}

func assertPaths1429(t testing.TB, sp *spec.Swagger) {
	for _, pi := range sp.Paths.Paths {
		for _, param := range pi.Get.Parameters {
			require.NotNilf(t, param.Schema, "expected param schema not to be nil")
			// all param fixtures are body param with schema
			// all $ref expanded
			assert.Equal(t, "", param.Schema.Ref.String())
		}

		for code, response := range pi.Get.Responses.StatusCodeResponses {
			// all response fixtures are with StatusCodeResponses, but 200
			if code == 200 {
				assert.Nilf(t, response.Schema, "expected response schema to be nil")
				continue
			}
			require.NotNilf(t, response.Schema, "expected response schema not to be nil")
			assert.Equal(t, "", response.Schema.Ref.String())
		}
	}
}

func assertPaths1429SkipSchema(t testing.TB, sp *spec.Swagger) {
	for _, pi := range sp.Paths.Paths {
		for _, param := range pi.Get.Parameters {
			require.NotNilf(t, param.Schema, "expected param schema not to be nil")

			// all param fixtures are body param with schema
			switch param.Name {
			case "plainRequest":
				// this one is expanded
				assert.Equal(t, "", param.Schema.Ref.String())
				continue
			case "nestedBody":
				// this one is local
				assert.Truef(t, strings.HasPrefix(param.Schema.Ref.String(), "#/definitions/"),
					"expected rooted definitions $ref, got: %s", param.Schema.Ref.String())
				continue
			case "remoteRequest":
				assert.Contains(t, param.Schema.Ref.String(), "remote/remote.yaml#/")
				continue
			}
			assert.Contains(t, param.Schema.Ref.String(), "responses.yaml#/")

		}

		for code, response := range pi.Get.Responses.StatusCodeResponses {
			// all response fixtures are with StatusCodeResponses, but 200
			switch code {
			case 200:
				assert.Nilf(t, response.Schema, "expected response schema to be nil")
				continue
			case 204:
				assert.Contains(t, response.Schema.Ref.String(), "remote/remote.yaml#/")
				continue
			case 404:
				assert.Equal(t, "", response.Schema.Ref.String())
				continue
			}
			assert.Containsf(t, response.Schema.Ref.String(), "responses.yaml#/", "expected remote ref at resp. %d", code)
		}
	}
}

func Test_MoreLocalExpansion(t *testing.T) {
	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()
	spec.PathLoader = testLoader
	path := filepath.Join("fixtures", "local_expansion", "spec2.yaml")

	// load and full expand
	sp := loadOrFail(t, path)
	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)

	// asserts all $ref expanded
	jazon, _ := json.MarshalIndent(sp, "", " ")
	assert.NotContains(t, jazon, `"$ref"`)
}

func Test_Issue69(t *testing.T) {
	// this checks expansion for the dapperbox spec (circular ref issues)

	path := filepath.Join("fixtures", "bugs", "69", "dapperbox.json")

	// expand with relative path
	// load and expand
	sp := loadOrFail(t, path)
	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)

	// asserts all $ref expanded
	jazon, _ := json.MarshalIndent(sp, "", " ")

	// assert all $ref match  "$ref": "#/definitions/something"
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	if assert.NotNil(t, m) {
		for _, matched := range m {
			subMatch := matched[1]
			assert.True(t, strings.HasPrefix(subMatch, "#/definitions/"),
				"expected $ref to be inlined, got: %s", matched[0])
		}
	}
}

func Test_Issue1621(t *testing.T) {
	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()
	spec.PathLoader = testLoader
	path := filepath.Join("fixtures", "bugs", "1621", "fixture-1621.yaml")

	// expand with relative path
	// load and expand
	sp := loadOrFail(t, path)

	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)

	// asserts all $ref expanded
	jazon, _ := json.MarshalIndent(sp, "", " ")
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	assert.Nil(t, m)
}

func Test_Issue1614(t *testing.T) {

	path := filepath.Join("fixtures", "bugs", "1614", "gitea.json")

	// expand with relative path
	// load and expand
	sp := loadOrFail(t, path)
	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)

	// asserts all $ref expanded
	jazon, _ := json.MarshalIndent(sp, "", " ")

	// assert all $ref maches  "$ref": "#/definitions/something"
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	if assert.NotNil(t, m) {
		for _, matched := range m {
			subMatch := matched[1]
			assert.True(t, strings.HasPrefix(subMatch, "#/definitions/"),
				"expected $ref to be inlined, got: %s", matched[0])
		}
	}

	// now with option CircularRefAbsolute
	sp = loadOrFail(t, path)
	err = spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, AbsoluteCircularRef: true})
	require.NoError(t, err)

	// asserts all $ref expanded
	jazon, _ = json.MarshalIndent(sp, "", " ")

	// assert all $ref maches  "$ref": "{file path}#/definitions/something"
	refPath, _ := os.Getwd()
	refPath = filepath.Join(refPath, path)
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	if assert.NotNil(t, m) {
		for _, matched := range m {
			subMatch := matched[1]
			assert.True(t, strings.HasPrefix(subMatch, refPath+"#/definitions/"),
				"expected $ref to be inlined, got: %s", matched[0])
		}
	}
}

func Test_Issue2113(t *testing.T) {
	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()
	spec.PathLoader = testLoader
	// this checks expansion with nested specs
	path := filepath.Join("fixtures", "bugs", "2113", "base.yaml")

	// load and expand
	sp := loadOrFail(t, path)
	err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false})
	require.NoError(t, err)
	// asserts all $ref expanded
	jazon, _ := json.MarshalIndent(sp, "", " ")

	// assert all $ref match have been expanded
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	assert.Emptyf(t, m, "expected all $ref to be expanded")

	// now trying with SkipSchemas
	sp = loadOrFail(t, path)
	err = spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true})
	require.NoError(t, err)

	jazon, _ = json.MarshalIndent(sp, "", " ")
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotEmpty(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		switch {
		case strings.Contains(subMatch, "dummy"):
			assert.True(t, strings.HasPrefix(subMatch, "schemas/dummy/dummy.yaml"),
				"expected $ref to be rebased to new relative base, got: %s", matched[0])
		case strings.Contains(subMatch, "example"):
			assert.True(t, strings.HasPrefix(subMatch, "schemas/example/example.yaml"),
				"expected $ref to be rebased to new relative base, got: %s", matched[0])
		default:
			t.Fail()
			t.Logf("unexpected $ref after skip-schemas expansion: %s", subMatch)
		}
	}
}

func Test_Issue2113_External(t *testing.T) {
	// Exercises the SkipSchema mode from spec flattening in go-openapi/analysis
	// Provides more ground for testing with schemas nested in $refs

	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()

	spec.PathLoader = testLoader
	// this checks expansion with nested specs
	path := filepath.Join("fixtures", "skipschema", "external_definitions_valid.yml")

	// load and expand, skipping schema expansion
	sp := loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true}))

	// asserts all $ref are expanded as expected
	jazon, _ := json.MarshalIndent(sp, "", " ")

	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotEmpty(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		require.Truef(t,
			strings.HasPrefix(subMatch, "external/definitions.yml#/definitions") ||
				strings.HasPrefix(subMatch, "external/errors.yml#/error") ||
				strings.HasPrefix(subMatch, "external/nestedParams.yml#/bodyParam"),
			"$ref %q did not match expectation", subMatch,
		)
	}

	// load and expand everything
	sp = loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false}))

	jazon, _ = json.MarshalIndent(sp, "", " ")
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	require.Empty(t, m)
}

func Test_Issue2113_SkipSchema(t *testing.T) {
	// Exercises the SkipSchema mode from spec flattening in go-openapi/analysis
	// Provides more ground for testing with schemas nested in $refs

	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()

	spec.PathLoader = testLoader
	// this checks expansion with nested specs
	path := filepath.Join("fixtures", "flatten", "flatten.yml")

	// load and expand, skipping schema expansion
	sp := loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true}))

	// asserts all $ref are expanded as expected
	jazon, err := json.MarshalIndent(sp, "", " ")
	require.NoError(t, err)

	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotEmpty(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		require.Truef(t,
			strings.HasPrefix(subMatch, "external/definitions.yml#/") ||
				strings.HasPrefix(subMatch, "#/definitions/namedAgain") ||
				strings.HasPrefix(subMatch, "external/errors.yml#/error"),
			"$ref %q did not match expectation", subMatch,
		)
	}

	sp = loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false}))

	jazon, _ = json.MarshalIndent(sp, "", " ")
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	require.Empty(t, m)
}

func Test_PointersLoop(t *testing.T) {
	// this a spec that cannot be flattened (self-referencing pointer).
	// however, it should be expanded without errors

	prevPathLoader := spec.PathLoader
	defer func() {
		spec.PathLoader = prevPathLoader
	}()

	spec.PathLoader = testLoader
	// this checks expansion with nested specs
	path := filepath.Join("fixtures", "more_circulars", "pointers", "fixture-pointers-loop.yaml")

	// load and expand, skipping schema expansion
	sp := loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true}))

	sp = loadOrFail(t, path)
	require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false}))

	// cannot guarantee which ref will be kept, but only one remains: expand reduces all $ref down
	// to the last self-referencing one (the one picked changes from one run to another, depending
	// on where during the walk the cycle is detected).
	jazon, _ := json.MarshalIndent(sp, "", " ")
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotEmpty(t, m)

	refs := make(map[string]struct{}, 5)
	for _, matched := range m {
		subMatch := matched[1]
		refs[subMatch] = struct{}{}
	}

	require.Len(t, refs, 1)
}

func Test_Issue102(t *testing.T) {
	// go-openapi/validate/issues#102
	path := filepath.Join("fixtures", "bugs", "102", "fixture-102.json")
	sp := loadOrFail(t, path)

	require.NoError(t, spec.ExpandSpec(sp, nil))

	jazon, err := json.MarshalIndent(sp, " ", "")
	require.NoError(t, err)

	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.NotEmpty(t, m)

	for _, matched := range m {
		subMatch := matched[1]
		assert.Equal(t, "#/definitions/Error", subMatch)
	}

	sp = loadOrFail(t, path)
	sch := spec.RefSchema("#/definitions/Error")
	require.NoError(t, spec.ExpandSchema(sch, sp, nil))

	jazon, err = json.MarshalIndent(sch, " ", "")
	require.NoError(t, err)

	m = rex.FindAllStringSubmatch(string(jazon), -1)
	for _, matched := range m {
		subMatch := matched[1]
		assert.Equal(t, "#/definitions/Error", subMatch)
	}

	sp = loadOrFail(t, path)
	sch = spec.RefSchema("#/definitions/Error")
	resp := spec.NewResponse().WithDescription("ok").WithSchema(sch)
	require.NoError(t, spec.ExpandResponseWithRoot(resp, sp, nil))

	jazon, err = json.MarshalIndent(resp, " ", "")
	require.NoError(t, err)

	m = rex.FindAllStringSubmatch(string(jazon), -1)
	for _, matched := range m {
		subMatch := matched[1]
		assert.Equal(t, "#/definitions/Error", subMatch)
	}

	sp = loadOrFail(t, path)
	sch = spec.RefSchema("#/definitions/Error")
	param := spec.BodyParam("error", sch)
	require.NoError(t, spec.ExpandParameterWithRoot(param, sp, nil))

	jazon, err = json.MarshalIndent(resp, " ", "")
	require.NoError(t, err)

	m = rex.FindAllStringSubmatch(string(jazon), -1)
	for _, matched := range m {
		subMatch := matched[1]
		assert.Equal(t, "#/definitions/Error", subMatch)
	}
}
