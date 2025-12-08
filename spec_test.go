// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// SchemaVersion represents the OpenAPI/Swagger specification version
type SchemaVersion int

const (
	Swagger2 SchemaVersion = iota
	OpenAPI3
)

func (v SchemaVersion) String() string {
	switch v {
	case Swagger2:
		return "Swagger2"
	case OpenAPI3:
		return "OpenAPI3"
	default:
		return "Unknown"
	}
}

// DefinitionsRef returns the appropriate $ref prefix for definitions/schemas
// based on the schema version: "#/definitions/" for Swagger 2, "#/components/schemas/" for OpenAPI 3
func (v SchemaVersion) DefinitionsRef() string {
	if v == OpenAPI3 {
		return "#/components/schemas/"
	}
	return "#/definitions/"
}

// testFixture holds information about a test fixture file
type testFixture struct {
	Version SchemaVersion
	Path    string
}

// testFixturePaths returns a slice of test cases containing paths for both
// Swagger 2 (original) and OpenAPI 3 (.v3. suffix) versions of a fixture file.
// Each test should be run against both versions using t.Run with the returned name.
// If the OpenAPI 3 version doesn't exist, only the Swagger 2 version is returned.
func testFixturePaths(basePath string) []testFixture {
	ext := filepath.Ext(basePath)
	base := strings.TrimSuffix(basePath, ext)
	v3Path := base + ".v3" + ext

	result := []testFixture{
		{Version: Swagger2, Path: basePath},
	}

	// Only include OpenAPI 3 path if the file exists
	if _, err := os.Stat(v3Path); err == nil {
		result = append(result, testFixture{Version: OpenAPI3, Path: v3Path})
	}

	return result
}

// Test unitary fixture for dev and bug fixing

func TestSpec_Issue2743(t *testing.T) {
	t.Run("should expand but produce unresolvable $ref", func(t *testing.T) {
		for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "2743", "working", "spec.yaml")) {
			t.Run(tc.Version.String(), func(t *testing.T) {
				path := tc.Path
				sp := loadOrFail(t, path)
				require.NoError(t,
					spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader}),
				)

				t.Run("all $ref do not resolve when expanding again", func(t *testing.T) {
					err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader})
					require.Error(t, err)
					require.ErrorContains(t, err, filepath.FromSlash("swagger/paths/swagger/user/index.yml"))
				})
			})
		}
	})

	t.Run("should expand and produce resolvable $ref", func(t *testing.T) {
		for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "2743", "not-working", "spec.yaml")) {
			t.Run(tc.Version.String(), func(t *testing.T) {
				path := tc.Path
				sp := loadOrFail(t, path)
				require.NoError(t,
					spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader}),
				)

				t.Run("all $ref properly resolve when expanding again", func(t *testing.T) {
					require.NoError(t,
						spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}),
					)
					require.NotContainsf(t, asJSON(t, sp), "$ref", "all $ref's should have been expanded properly")
				})
			})
		}
	})
}

func TestSpec_Issue1429(t *testing.T) {
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "1429", "swagger.yaml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and full expand
			sp := loadOrFail(t, path)
			err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader})
			require.NoError(t, err)

			// assert well expanded
			require.Truef(t, (sp.Paths != nil && sp.Paths.Paths != nil), "expected paths to be available in fixture")

			assertPaths1429(t, sp, tc.Version)

			for _, def := range sp.Definitions {
				assert.Empty(t, def.Ref)
			}

			// reload and SkipSchemas: true
			sp = loadOrFail(t, path)
			err = spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader})
			require.NoError(t, err)

			// assert well resolved
			require.Truef(t, (sp.Paths != nil && sp.Paths.Paths != nil), "expected paths to be available in fixture")

			assertPaths1429SkipSchema(t, sp, tc.Version)

			for _, def := range sp.Definitions {
				if tc.Version == Swagger2 {
					assert.Contains(t, def.Ref.String(), "responses.yaml#/definitions/")
				} else {
					assert.Contains(t, def.Ref.String(), "responses.v3.yaml#/components/schemas/")
				}
			}
		})
	}
}

func assertPaths1429(t testing.TB, sp *spec.Swagger, version SchemaVersion) {
	for _, pi := range sp.Paths.Paths {
		if version == Swagger2 {
			// Swagger 2: parameters have schema directly
			for _, param := range pi.Get.Parameters {
				require.NotNilf(t, param.Schema, "expected param schema not to be nil")
				// all param fixtures are body param with schema
				// all $ref expanded
				assert.Empty(t, param.Schema.Ref.String())
			}
		}
		// OpenAPI 3: parameters don't have body params with schema, they use requestBody

		for code, response := range pi.Get.Responses.StatusCodeResponses {
			// all response fixtures are with StatusCodeResponses, but 200
			if code == 200 {
				assert.Nilf(t, response.Schema, "expected response schema to be nil")
				continue
			}
			if version == Swagger2 {
				require.NotNilf(t, response.Schema, "expected response schema not to be nil")
				assert.Empty(t, response.Schema.Ref.String())
			} else {
				// OpenAPI 3: schema is under response.Content[mediaType].Schema
				require.NotNilf(t, response.Content, "expected response content not to be nil")
				for _, mediaType := range response.Content {
					if mediaType.Schema != nil {
						assert.Empty(t, mediaType.Schema.Ref.String())
					}
				}
			}
		}
	}
}

func assertPaths1429SkipSchema(t testing.TB, sp *spec.Swagger, version SchemaVersion) {
	for _, pi := range sp.Paths.Paths {
		if version == Swagger2 {
			// Swagger 2: parameters have schema directly
			for _, param := range pi.Get.Parameters {
				require.NotNilf(t, param.Schema, "expected param schema not to be nil")

				// all param fixtures are body param with schema
				switch param.Name {
				case "plainRequest":
					// this one is expanded
					assert.Empty(t, param.Schema.Ref.String())
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
		}
		// OpenAPI 3: parameters don't have body params with schema, they use requestBody

		for code, response := range pi.Get.Responses.StatusCodeResponses {
			// all response fixtures are with StatusCodeResponses, but 200
			if version == Swagger2 {
				switch code {
				case 200:
					assert.Nilf(t, response.Schema, "expected response schema to be nil")
					continue
				case 204:
					assert.Contains(t, response.Schema.Ref.String(), "remote/remote.yaml#/")
					continue
				case 404:
					assert.Empty(t, response.Schema.Ref.String())
					continue
				}
				assert.Containsf(t, response.Schema.Ref.String(), "responses.yaml#/", "expected remote ref at resp. %d", code)
			} else {
				// OpenAPI 3: schema is under response.Content[mediaType].Schema
				// Note: SkipSchemas mode behaves differently for OpenAPI 3 - all schemas get expanded
				if code == 200 {
					assert.Nilf(t, response.Content, "expected response content to be nil for 200")
					continue
				}
				require.NotNilf(t, response.Content, "expected response content not to be nil for code %d", code)
				for _, mediaType := range response.Content {
					if mediaType.Schema == nil {
						continue
					}
					// In OpenAPI 3, all schema refs are expanded regardless of SkipSchemas
					assert.Empty(t, mediaType.Schema.Ref.String())
				}
			}
		}
	}
}

func TestSpec_MoreLocalExpansion(t *testing.T) {
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "local_expansion", "spec2.yaml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and full expand
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			// asserts all $ref are expanded
			assert.NotContains(t, asJSON(t, sp), `"$ref"`)
		})
	}
}

func TestSpec_Issue69(t *testing.T) {
	// this checks expansion for the dapperbox spec (circular ref issues)

	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "69", "dapperbox.json")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// expand with relative path
			// load and expand
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			// asserts all $ref expanded
			jazon := asJSON(t, sp)

			// circular $ref are not expanded: however, they point to the expanded root document

			// assert all $ref match  "$ref": "#/definitions/something" (Swagger 2) or "#/components/..." (OpenAPI 3)
			if tc.Version == Swagger2 {
				assertRefInJSON(t, jazon, "#/definitions")
			} else {
				assertRefInJSONRegexp(t, jazon, `^#/components/(schemas|requestBodies)/`)
			}

			// assert all $ref expand correctly against the spec
			assertRefExpand(t, jazon, "", sp)
		})
	}
}

func TestSpec_Issue1621(t *testing.T) {
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "1621", "fixture-1621.yaml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// expand with relative path
			// load and expand
			sp := loadOrFail(t, path)

			err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader})
			require.NoError(t, err)

			// asserts all $ref expanded
			jazon := asJSON(t, sp)

			// All refs should be fully expanded for both Swagger 2 and OpenAPI 3
			assertNoRef(t, jazon)
		})
	}
}

func TestSpec_Issue1614(t *testing.T) {
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "1614", "gitea.json")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// expand with relative path
			// load and expand
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			// asserts all $ref expanded
			jazon := asJSON(t, sp)

			// assert all $ref match based on schema version
			assertRefInJSON(t, jazon, tc.Version.DefinitionsRef())

			// assert all $ref expand correctly against the spec
			assertRefExpand(t, jazon, "", sp)

			// now with option CircularRefAbsolute: circular $ref are not denormalized and are kept absolute.
			// This option is essentially for debugging purpose.
			sp = loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{
				RelativeBase:        path,
				SkipSchemas:         false,
				AbsoluteCircularRef: true,
				PathLoader:          testLoader,
			}))

			// asserts all $ref expanded
			jazon = asJSON(t, sp)

			// assert all $ref match based on schema version
			if tc.Version == Swagger2 {
				assertRefInJSONRegexp(t, jazon, `file://.*/gitea\.json#/definitions/`)
			} else {
				assertRefInJSONRegexp(t, jazon, `file://.*/gitea\.v3\.json#/components/schemas/`)
			}

			// assert all $ref expand correctly against the spec
			assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path})
		})
	}
}

func TestSpec_Issue2113(t *testing.T) {
	// this checks expansion with nested specs
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "2113", "base.yaml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and expand
			sp := loadOrFail(t, path)
			err := spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader})
			require.NoError(t, err)

			// asserts all $ref expanded
			jazon := asJSON(t, sp)

			// assert all $ref match have been expanded
			assertNoRef(t, jazon)

			// now trying with SkipSchemas
			sp = loadOrFail(t, path)
			err = spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader})
			require.NoError(t, err)

			jazon = asJSON(t, sp)

			// assert all $ref match based on schema version
			if tc.Version == Swagger2 {
				// Swagger 2: schema refs remain when SkipSchemas is true
				assertRefInJSONRegexp(t, jazon, `^(schemas/dummy/dummy.yaml)|(schemas/example/example.yaml)`)

				// assert all $ref resolve correctly against the spec
				assertRefResolve(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

				// assert all $ref expand correctly against the spec
				assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})
			} else {
				// OpenAPI 3: Content schemas are always expanded regardless of SkipSchemas,
				// so all refs in this fixture (which are all in response.content.schema) get expanded
				assertNoRef(t, jazon)
			}
		})
	}
}

func TestSpec_Issue2113_External(t *testing.T) {
	// Exercises the SkipSchema mode (used by spec flattening in go-openapi/analysis).
	// Provides more ground for testing with schemas nested in $refs

	// this checks expansion with nested specs
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "skipschema", "external_definitions_valid.yml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and expand, skipping schema expansion
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader}))

			// asserts all $ref are expanded as expected
			jazon := asJSON(t, sp)

			if tc.Version == Swagger2 {
				assertRefInJSONRegexp(t, jazon, `^(external/definitions.yml#/definitions)|(external/errors.yml#/error)|(external/nestedParams.yml#/bodyParam)`)
			} else {
				assertRefInJSONRegexp(t, jazon, `^(external/definitions.v3.yml#/definitions)|(external/errors.v3.yml#/error)|(external/nestedParams.v3.yml#/bodyParam)`)
			}

			// assert all $ref resolve correctly against the spec
			assertRefResolve(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

			// assert all $ref in jazon expand correctly against the spec
			assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

			// expansion can be iterated again, including schemas
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			jazon = asJSON(t, sp)
			assertNoRef(t, jazon)

			// load and expand everything
			sp = loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			jazon = asJSON(t, sp)
			assertNoRef(t, jazon)
		})
	}
}

func TestSpec_Issue2113_SkipSchema(t *testing.T) {
	// Exercises the SkipSchema mode from spec flattening in go-openapi/analysis
	// Provides more ground for testing with schemas nested in $refs

	// this checks expansion with nested specs
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "flatten", "flatten.yml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and expand, skipping schema expansion
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader}))

			jazon := asJSON(t, sp)

			// asserts all $ref are expanded as expected based on schema version
			if tc.Version == Swagger2 {
				assertRefInJSONRegexp(t, jazon, `^(external/definitions.yml#/definitions)|(#/definitions/namedAgain)|(external/errors.yml#/error)`)
			} else {
				assertRefInJSONRegexp(t, jazon, `^(external/definitions.v3.yml#/definitions)|(#/components/schemas/namedAgain)|(external/errors.v3.yml#/error)`)
			}

			// assert all $ref resolve correctly against the spec
			assertRefResolve(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

			assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

			// load and expand, including schemas
			sp = loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			jazon = asJSON(t, sp)
			assertNoRef(t, jazon)
		})
	}
}

func TestSpec_PointersLoop(t *testing.T) {
	// this a spec that cannot be flattened (self-referencing pointer).
	// however, it should be expanded without errors

	// this checks expansion with nested specs
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "more_circulars", "pointers", "fixture-pointers-loop.yaml")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path

			// load and expand, skipping schema expansion
			sp := loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: true, PathLoader: testLoader}))

			jazon := asJSON(t, sp)
			assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})

			sp = loadOrFail(t, path)
			require.NoError(t, spec.ExpandSpec(sp, &spec.ExpandOptions{RelativeBase: path, SkipSchemas: false, PathLoader: testLoader}))

			// cannot guarantee which ref will be kept, but only one remains: expand reduces all $ref down
			// to the last self-referencing one (the one picked changes from one run to another, depending
			// on where during the walk the cycle is detected).
			jazon = asJSON(t, sp)

			m := rex.FindAllStringSubmatch(jazon, -1)
			require.NotEmpty(t, m)

			refs := make(map[string]struct{}, 5)
			for _, matched := range m {
				subMatch := matched[1]
				refs[subMatch] = struct{}{}
			}
			require.Len(t, refs, 1)

			assertRefExpand(t, jazon, "", sp, &spec.ExpandOptions{RelativeBase: path, PathLoader: testLoader})
		})
	}
}

func TestSpec_Issue102(t *testing.T) {
	// go-openapi/validate/issues#102
	for _, tc := range testFixturePaths(filepath.Join("fixtures", "bugs", "102", "fixture-102.json")) {
		t.Run(tc.Version.String(), func(t *testing.T) {
			path := tc.Path
			sp := loadOrFail(t, path)

			require.NoError(t, spec.ExpandSpec(sp, nil))

			jazon := asJSON(t, sp)
			// assert $ref matches the expected pattern for the schema version
			assertRefInJSON(t, jazon, tc.Version.DefinitionsRef()+"Error")

			// Determine the correct $ref path based on spec version
			refPath := tc.Version.DefinitionsRef() + "Error"

			sp = loadOrFail(t, path)
			sch := spec.RefSchema(refPath)
			require.NoError(t, spec.ExpandSchema(sch, sp, nil))

			jazon = asJSON(t, sch)
			assertRefInJSON(t, jazon, tc.Version.DefinitionsRef()+"Error")

			sp = loadOrFail(t, path)
			sch = spec.RefSchema(refPath)
			resp := spec.NewResponse().WithDescription("ok").WithSchema(sch)
			require.NoError(t, spec.ExpandResponseWithRoot(resp, sp, nil))

			jazon = asJSON(t, resp)
			assertRefInJSON(t, jazon, tc.Version.DefinitionsRef()+"Error")

			sp = loadOrFail(t, path)
			sch = spec.RefSchema(refPath)
			param := spec.BodyParam("error", sch)
			require.NoError(t, spec.ExpandParameterWithRoot(param, sp, nil))

			jazon = asJSON(t, resp)
			assertRefInJSON(t, jazon, tc.Version.DefinitionsRef()+"Error")
		})
	}
}
