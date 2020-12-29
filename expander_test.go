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
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	crossFileRefFixture = "fixtures/expansion/crossFileRef.json"
	withoutSchemaID     = "removed"
	withSchemaID        = "schema"
	pathItemsFixture    = "fixtures/expansion/pathItems.json"
	extraRefFixture     = "fixtures/expansion/extraRef.json"
)

var (
	// PetStoreJSONMessage json raw message for Petstore20
	PetStoreJSONMessage = json.RawMessage([]byte(PetStore20))
	specs               = filepath.Join("fixtures", "specs")
	rex                 = regexp.MustCompile(`"\$ref":\s*"(.+)"`)
)

func TestExpandsKnownRef(t *testing.T) {
	schema := RefProperty("http://json-schema.org/draft-04/schema#")
	if assert.NoError(t, ExpandSchema(schema, nil, nil)) {
		assert.Equal(t, "Core schema meta-schema", schema.Description)
	}
}

func TestExpandResponseSchema(t *testing.T) {
	fp := "./fixtures/local_expansion/spec.json"
	b, err := jsonDoc(fp)
	if assert.NoError(t, err) {
		var spec Swagger
		if err := json.Unmarshal(b, &spec); assert.NoError(t, err) {
			err := ExpandSpec(&spec, &ExpandOptions{RelativeBase: fp})
			if assert.NoError(t, err) {
				sch := spec.Paths.Paths["/item"].Get.Responses.StatusCodeResponses[200].Schema
				if assert.NotNil(t, sch) {
					assert.Empty(t, sch.Ref.String())
					assert.Contains(t, sch.Type, "object")
					assert.Len(t, sch.Properties, 2)
				}
			}
		}
	}
}

func TestSpecExpansion(t *testing.T) {
	spec := new(Swagger)

	require.NoError(t, ExpandSpec(spec, nil))

	specDoc, err := jsonDoc("fixtures/expansion/all-the-things.json")
	require.NoError(t, err)

	specPath, _ := absPath("fixtures/expansion/all-the-things.json")
	opts := &ExpandOptions{
		RelativeBase: specPath,
	}

	spec = new(Swagger)
	require.NoError(t, json.Unmarshal(specDoc, spec))

	pet := spec.Definitions["pet"]
	errorModel := spec.Definitions["errorModel"]
	petResponse := spec.Responses["petResponse"]
	petResponse.Schema = &pet
	stringResponse := spec.Responses["stringResponse"]
	tagParam := spec.Parameters["tag"]
	idParam := spec.Parameters["idParam"]

	require.NoError(t, ExpandSpec(spec, opts))

	assert.Equal(t, tagParam, spec.Parameters["query"])
	assert.Equal(t, petResponse, spec.Responses["petResponse"])
	assert.Equal(t, petResponse, spec.Responses["anotherPet"])
	assert.Equal(t, pet, *spec.Responses["petResponse"].Schema)
	assert.Equal(t, stringResponse, *spec.Paths.Paths["/"].Get.Responses.Default)
	assert.Equal(t, petResponse, spec.Paths.Paths["/"].Get.Responses.StatusCodeResponses[200])
	assert.Equal(t, pet, *spec.Paths.Paths["/pets"].Get.Responses.StatusCodeResponses[200].Schema.Items.Schema)
	assert.Equal(t, errorModel, *spec.Paths.Paths["/pets"].Get.Responses.Default.Schema)
	assert.Equal(t, pet, spec.Definitions["petInput"].AllOf[0])
	assert.Equal(t, spec.Definitions["petInput"], *spec.Paths.Paths["/pets"].Post.Parameters[0].Schema)
	assert.Equal(t, petResponse, spec.Paths.Paths["/pets"].Post.Responses.StatusCodeResponses[200])
	assert.Equal(t, errorModel, *spec.Paths.Paths["/pets"].Post.Responses.Default.Schema)
	pi := spec.Paths.Paths["/pets/{id}"]
	assert.Equal(t, idParam, pi.Get.Parameters[0])
	assert.Equal(t, petResponse, pi.Get.Responses.StatusCodeResponses[200])
	assert.Equal(t, errorModel, *pi.Get.Responses.Default.Schema)
	assert.Equal(t, idParam, pi.Delete.Parameters[0])
	assert.Equal(t, errorModel, *pi.Delete.Responses.Default.Schema)
}

func TestResponseExpansion(t *testing.T) {
	specDoc, err := jsonDoc("fixtures/expansion/all-the-things.json")
	require.NoError(t, err)

	basePath, err := absPath("fixtures/expansion/all-the-things.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(specDoc, spec))

	resolver, err := defaultSchemaLoader(spec, nil, nil, nil)
	require.NoError(t, err)

	resp := spec.Responses["anotherPet"]
	expected := spec.Responses["petResponse"]
	require.NoError(t, expandParameterOrResponse(&expected, resolver, basePath))

	jazon, err := json.MarshalIndent(expected, "", " ")
	require.NoError(t, err)

	assert.JSONEq(t, `{
         "description": "pet response",
         "schema": {
          "required": [
           "id",
           "name"
          ],
          "properties": {
           "id": {
            "type": "integer",
            "format": "int64"
           },
           "name": {
            "type": "string"
           },
           "tag": {
            "type": "string"
           }
          }
         }
			 }`, string(jazon))

	require.NoError(t, expandParameterOrResponse(&resp, resolver, basePath))
	assert.Equal(t, expected, resp)

	resp2 := spec.Paths.Paths["/"].Get.Responses.Default
	expected = spec.Responses["stringResponse"]

	require.NoError(t, expandParameterOrResponse(resp2, resolver, basePath))
	assert.Equal(t, expected, *resp2)

	// cascading ref
	resp = spec.Paths.Paths["/"].Get.Responses.StatusCodeResponses[200]
	expected = spec.Responses["petResponse"]
	jazon, err = json.MarshalIndent(resp, "", " ")
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"$ref": "#/responses/anotherPet"
  }`, string(jazon))

	require.NoError(t, expandParameterOrResponse(&resp, resolver, basePath))
	assert.Equal(t, expected, resp)
}

// test the exported version of ExpandResponse
func TestExportedResponseExpansion(t *testing.T) {
	specDoc, err := jsonDoc("fixtures/expansion/all-the-things.json")
	require.NoError(t, err)

	basePath, err := absPath("fixtures/expansion/all-the-things.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(specDoc, spec))

	resp := spec.Responses["anotherPet"]
	expected := spec.Responses["petResponse"]
	require.NoError(t, ExpandResponse(&expected, basePath))

	require.NoError(t, ExpandResponse(&resp, basePath))
	assert.Equal(t, expected, resp)

	resp2 := spec.Paths.Paths["/"].Get.Responses.Default
	expected = spec.Responses["stringResponse"]

	require.NoError(t, ExpandResponse(resp2, basePath))
	assert.Equal(t, expected, *resp2)

	resp = spec.Paths.Paths["/"].Get.Responses.StatusCodeResponses[200]
	expected = spec.Responses["petResponse"]

	require.NoError(t, ExpandResponse(&resp, basePath))
	assert.Equal(t, expected, resp)
}

func TestExpandResponseAndParamWithRoot(t *testing.T) {
	specDoc, err := jsonDoc("fixtures/bugs/1614/gitea.json")
	require.NoError(t, err)

	var spec Swagger
	err = json.Unmarshal(specDoc, &spec)
	require.NoError(t, err)

	// check responses with $ref
	resp := spec.Paths.Paths["/admin/users"].Post.Responses.StatusCodeResponses[201]
	require.NoError(t, ExpandResponseWithRoot(&resp, spec, nil))

	jazon, _ := json.MarshalIndent(resp, "", " ")
	m := rex.FindAllStringSubmatch(string(jazon), -1)
	require.Nil(t, m)

	resp = spec.Paths.Paths["/admin/users"].Post.Responses.StatusCodeResponses[403]
	require.NoError(t, ExpandResponseWithRoot(&resp, spec, nil))

	jazon, _ = json.MarshalIndent(resp, "", " ")
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	require.Nil(t, m)

	// check param with $ref
	param := spec.Paths.Paths["/admin/users"].Post.Parameters[0]
	require.NoError(t, ExpandParameterWithRoot(&param, spec, nil))

	jazon, _ = json.MarshalIndent(param, "", " ")
	m = rex.FindAllStringSubmatch(string(jazon), -1)
	require.Nil(t, m)
}

func TestExpandResponseWithRoot_CircularRefs(t *testing.T) {
	rootDoc := new(Swagger)
	b, err := ioutil.ReadFile("fixtures/more_circulars/resp.json")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, rootDoc))

	path := rootDoc.Paths.Paths["/api/v1/getx"]
	resp := path.Post.Responses.StatusCodeResponses[200]

	thisCache := cacheOrDefault(nil)

	// during first response expand, refs are getting expanded,
	// so the following expands cannot properly resolve them w/o the document.
	// this happens in validator.Validate() when different validators try to expand the same mutable response.
	require.NoError(t, ExpandResponseWithRoot(&resp, rootDoc, thisCache))

	require.NoError(t, ExpandResponseWithRoot(&resp, rootDoc, thisCache))
}

func TestIssue3(t *testing.T) {
	spec := new(Swagger)
	specDoc, err := jsonDoc("fixtures/expansion/overflow.json")
	require.NoError(t, err)

	specPath, _ := absPath("fixtures/expansion/overflow.json")
	opts := &ExpandOptions{
		RelativeBase: specPath,
	}

	require.NoError(t, json.Unmarshal(specDoc, spec))

	assert.NotPanics(t, func() {
		err = ExpandSpec(spec, opts)
		assert.NoError(t, err)
	}, "Calling expand spec with circular refs, should not panic!")
}

func TestParameterExpansion(t *testing.T) {
	paramDoc, err := jsonDoc("fixtures/expansion/params.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(paramDoc, spec))

	basePath, err := absPath("fixtures/expansion/params.json")
	require.NoError(t, err)

	resolver, err := defaultSchemaLoader(spec, nil, nil, nil)
	require.NoError(t, err)

	param := spec.Parameters["query"]
	expected := spec.Parameters["tag"]

	require.NoError(t, expandParameterOrResponse(&param, resolver, basePath))

	assert.Equal(t, expected, param)

	param = spec.Paths.Paths["/cars/{id}"].Parameters[0]
	expected = spec.Parameters["id"]

	require.NoError(t, expandParameterOrResponse(&param, resolver, basePath))

	assert.Equal(t, expected, param)
}

func TestExportedParameterExpansion(t *testing.T) {
	paramDoc, err := jsonDoc("fixtures/expansion/params.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(paramDoc, spec))

	basePath, err := absPath("fixtures/expansion/params.json")
	require.NoError(t, err)

	param := spec.Parameters["query"]
	expected := spec.Parameters["tag"]

	require.NoError(t, ExpandParameter(&param, basePath))
	assert.Equal(t, expected, param)

	param = spec.Paths.Paths["/cars/{id}"].Parameters[0]
	expected = spec.Parameters["id"]

	require.NoError(t, ExpandParameter(&param, basePath))
	assert.Equal(t, expected, param)
}

func TestCircularRefsExpansion(t *testing.T) {
	carsDoc, err := jsonDoc("fixtures/expansion/circularRefs.json")
	require.NoError(t, err)

	basePath, _ := absPath("fixtures/expansion/circularRefs.json")

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(carsDoc, spec))

	resolver, err := defaultSchemaLoader(spec, &ExpandOptions{RelativeBase: basePath}, nil, nil)
	require.NoError(t, err)

	schema := spec.Definitions["car"]

	assert.NotPanics(t, func() {
		_, err := expandSchema(schema, []string{"#/definitions/car"}, resolver, basePath)
		require.NoError(t, err)
	}, "Calling expand schema with circular refs, should not panic!")
}

func TestCircularSpec2Expansion(t *testing.T) {
	// TODO: assert repeatable results (see commented section below)

	fixturePath := filepath.Join("fixtures", "expansion", "circular-minimal.json")
	jazon := expandThisOrDieTrying(t, fixturePath)
	assert.NotEmpty(t, jazon)

	// assert stripped $ref in result
	assert.NotContainsf(t, jazon, "circular-minimal.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)

	fixturePath = "fixtures/expansion/circularSpec2.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	assert.NotEmpty(t, jazon)

	assert.NotContainsf(t, jazon, "circularSpec.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)

	/*

		At the moment, the result of expanding circular references is not stable,
		when several cycles have intersections:
		the spec structure is randomly walked through and mutating as expansion is carried out.
		detected cycles in $ref are not necessarily the shortest matches.

		This may result in different, functionally correct expanded spec (e.g. with same validations)

			for i := 0; i < 1; i++ {
				bbb := expandThisOrDieTrying(t, fixturePath)
				t.Log(bbb)
				if !assert.JSONEqf(t, jazon, bbb, "on iteration %d, we should have stable expanded spec", i) {
					t.FailNow()
					return
				}
			}
	*/
}

func Test_MoreCircular(t *testing.T) {
	// Additional testcase for circular $ref (from go-openapi/validate):
	// - $ref with file = current file
	// - circular is located in remote file
	//
	// There are 4 variants to run:
	// - with/without $ref with local file (so its not really remote)
	// - with circular in a schema in  #/responses
	// - with circular in a schema in  #/parameters

	fixturePath := "fixtures/more_circulars/spec.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "item.json#/item"),
			"expected $ref to be relative, got: %s", matched[0])
	}

	fixturePath = "fixtures/more_circulars/spec2.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	m = rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "item2.json#/item"),
			"expected $ref to be relative, got: %s", matched[0])
	}

	fixturePath = "fixtures/more_circulars/spec3.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	m = rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "item.json#/item"),
			"expected $ref to be relative, got: %s", matched[0])
	}

	fixturePath = "fixtures/more_circulars/spec4.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	m = rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "item4.json#/item"),
			"expected $ref to be relative, got: %s", matched[0])
	}
}

func Test_Issue957(t *testing.T) {
	fixturePath := "fixtures/bugs/957/fixture-957.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)

	assert.NotContainsf(t, jazon, "fixture-957.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "#/definitions/"),
			"expected $ref to be inlined, got: %s", matched[0])
	}
}

func Test_Bitbucket(t *testing.T) {
	// Additional testcase for circular $ref (from bitbucket api)

	fixturePath := "fixtures/more_circulars/bitbucket.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "#/definitions/"),
			"expected $ref to be inlined, got: %s", matched[0])
	}
}

func Test_ExpandJSONSchemaDraft4(t *testing.T) {
	fixturePath := filepath.Join("schemas", "jsonschema-draft-04.json")
	jazon := expandThisSchemaOrDieTrying(t, fixturePath)
	// assert all $ref maches  "$ref": "http://json-schema.org/draft-04/something"
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "http://json-schema.org/draft-04/"),
			"expected $ref to be remote, got: %s", matched[0])
	}
}

func Test_ExpandSwaggerSchema(t *testing.T) {
	fixturePath := filepath.Join("schemas", "v2", "schema.json")
	jazon := expandThisSchemaOrDieTrying(t, fixturePath)
	// assert all $ref maches  "$ref": "#/definitions/something"
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	for _, matched := range m {
		// matched := m[0]
		subMatch := matched[1]
		assert.True(t, strings.HasPrefix(subMatch, "#/definitions/"),
			"expected $ref to be inlined, got: %s", matched[0])
	}
}

func TestContinueOnErrorExpansion(t *testing.T) {
	defer log.SetOutput(os.Stdout)
	log.SetOutput(ioutil.Discard)

	missingRefDoc, err := jsonDoc("fixtures/expansion/missingRef.json")
	assert.NoError(t, err)

	specPath, _ := absPath("fixtures/expansion/missingRef.json")

	testCase := struct {
		Input    *Swagger `json:"input"`
		Expected *Swagger `json:"expected"`
	}{}
	require.NoError(t, json.Unmarshal(missingRefDoc, &testCase))

	opts := &ExpandOptions{
		ContinueOnError: true,
		RelativeBase:    specPath,
	}
	require.NoError(t, ExpandSpec(testCase.Input, opts))
	assert.Equal(t, testCase.Input, testCase.Expected, "Should continue expanding spec when a definition can't be found.")

	doc, err := jsonDoc("fixtures/expansion/missingItemRef.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	assert.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "Array of missing refs should not cause a panic, and continue to expand spec.")
}

func TestIssue415(t *testing.T) {
	doc, err := jsonDoc("fixtures/expansion/clickmeter.json")
	require.NoError(t, err)

	specPath, _ := absPath("fixtures/expansion/clickmeter.json")

	opts := &ExpandOptions{
		RelativeBase: specPath,
	}

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	assert.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "Calling expand spec with response schemas that have circular refs, should not panic!")
}

func TestCircularSpecExpansion(t *testing.T) {
	doc, err := jsonDoc("fixtures/expansion/circularSpec.json")
	require.NoError(t, err)

	specPath, err := absPath("fixtures/expansion/circularSpec.json")
	require.NoError(t, err)

	opts := &ExpandOptions{
		RelativeBase: specPath,
	}

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	assert.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "Calling expand spec with circular refs, should not panic!")
}

func TestItemsExpansion(t *testing.T) {
	carsDoc, err := jsonDoc("fixtures/expansion/schemas2.json")
	require.NoError(t, err)

	basePath, err := absPath("fixtures/expansion/schemas2.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(carsDoc, spec))

	resolver, err := defaultSchemaLoader(spec, nil, nil, nil)
	require.NoError(t, err)

	schema := spec.Definitions["car"]
	oldBrand := schema.Properties["brand"]
	assert.NotEmpty(t, oldBrand.Items.Schema.Ref.String())
	assert.NotEqual(t, spec.Definitions["brand"], oldBrand)

	_, err = expandSchema(schema, []string{"#/definitions/car"}, resolver, basePath)
	require.NoError(t, err)

	newBrand := schema.Properties["brand"]
	assert.Empty(t, newBrand.Items.Schema.Ref.String())
	assert.Equal(t, spec.Definitions["brand"], *newBrand.Items.Schema)

	schema = spec.Definitions["truck"]
	require.NotEmpty(t, schema.Items.Schema.Ref.String())

	s, err := expandSchema(schema, []string{"#/definitions/truck"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Items.Schema.Ref.String())
	assert.Equal(t, spec.Definitions["car"], *schema.Items.Schema)

	sch := new(Schema)
	_, err = expandSchema(*sch, []string{""}, resolver, basePath)
	require.NoError(t, err)

	schema = spec.Definitions["batch"]
	s, err = expandSchema(schema, []string{"#/definitions/batch"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Items.Schema.Items.Schema.Ref.String())
	assert.Equal(t, *schema.Items.Schema.Items.Schema, spec.Definitions["brand"])

	schema = spec.Definitions["batch2"]
	s, err = expandSchema(schema, []string{"#/definitions/batch2"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Items.Schemas[0].Items.Schema.Ref.String())
	assert.Empty(t, schema.Items.Schemas[1].Items.Schema.Ref.String())
	assert.Equal(t, *schema.Items.Schemas[0].Items.Schema, spec.Definitions["brand"])
	assert.Equal(t, *schema.Items.Schemas[1].Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["allofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/allofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AllOf[0].Items.Schema.Ref.String())
	assert.Empty(t, schema.AllOf[1].Items.Schema.Ref.String())
	assert.Equal(t, *schema.AllOf[0].Items.Schema, spec.Definitions["brand"])
	assert.Equal(t, *schema.AllOf[1].Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["anyofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/anyofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AnyOf[0].Items.Schema.Ref.String())
	assert.Empty(t, schema.AnyOf[1].Items.Schema.Ref.String())
	assert.Equal(t, *schema.AnyOf[0].Items.Schema, spec.Definitions["brand"])
	assert.Equal(t, *schema.AnyOf[1].Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["oneofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/oneofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.OneOf[0].Items.Schema.Ref.String())
	assert.Empty(t, schema.OneOf[1].Items.Schema.Ref.String())
	assert.Equal(t, *schema.OneOf[0].Items.Schema, spec.Definitions["brand"])
	assert.Equal(t, *schema.OneOf[1].Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["notSomething"]
	s, err = expandSchema(schema, []string{"#/definitions/notSomething"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Not.Items.Schema.Ref.String())
	assert.Equal(t, *schema.Not.Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["withAdditional"]
	s, err = expandSchema(schema, []string{"#/definitions/withAdditional"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AdditionalProperties.Schema.Items.Schema.Ref.String())
	assert.Equal(t, *schema.AdditionalProperties.Schema.Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["withAdditionalItems"]
	s, err = expandSchema(schema, []string{"#/definitions/withAdditionalItems"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AdditionalItems.Schema.Items.Schema.Ref.String())
	assert.Equal(t, *schema.AdditionalItems.Schema.Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["withPattern"]
	s, err = expandSchema(schema, []string{"#/definitions/withPattern"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop := schema.PatternProperties["^x-ab"]
	assert.Empty(t, prop.Items.Schema.Ref.String())
	assert.Equal(t, *prop.Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["deps"]
	s, err = expandSchema(schema, []string{"#/definitions/deps"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop2 := schema.Dependencies["something"]
	assert.Empty(t, prop2.Schema.Items.Schema.Ref.String())
	assert.Equal(t, *prop2.Schema.Items.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["defined"]
	s, err = expandSchema(schema, []string{"#/definitions/defined"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop = schema.Definitions["something"]
	assert.Empty(t, prop.Items.Schema.Ref.String())
	assert.Equal(t, *prop.Items.Schema, spec.Definitions["tag"])
}

func TestSchemaExpansion(t *testing.T) {
	carsDoc, err := jsonDoc("fixtures/expansion/schemas1.json")
	require.NoError(t, err)

	basePath, err := absPath("fixtures/expansion/schemas1.json")
	require.NoError(t, err)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(carsDoc, spec))

	resolver, err := defaultSchemaLoader(spec, nil, nil, nil)
	require.NoError(t, err)

	schema := spec.Definitions["car"]
	oldBrand := schema.Properties["brand"]
	assert.NotEmpty(t, oldBrand.Ref.String())
	assert.NotEqual(t, spec.Definitions["brand"], oldBrand)

	s, err := expandSchema(schema, []string{"#/definitions/car"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s

	newBrand := schema.Properties["brand"]
	assert.Empty(t, newBrand.Ref.String())
	assert.Equal(t, spec.Definitions["brand"], newBrand)

	schema = spec.Definitions["truck"]
	assert.NotEmpty(t, schema.Ref.String())

	s, err = expandSchema(schema, []string{"#/definitions/truck"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Ref.String())
	assert.Equal(t, spec.Definitions["car"], schema)

	sch := new(Schema)
	_, err = expandSchema(*sch, []string{""}, resolver, basePath)
	require.NoError(t, err)

	schema = spec.Definitions["batch"]
	s, err = expandSchema(schema, []string{"#/definitions/batch"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Items.Schema.Ref.String())
	assert.Equal(t, *schema.Items.Schema, spec.Definitions["brand"])

	schema = spec.Definitions["batch2"]
	s, err = expandSchema(schema, []string{"#/definitions/batch2"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Items.Schemas[0].Ref.String())
	assert.Empty(t, schema.Items.Schemas[1].Ref.String())
	assert.Equal(t, schema.Items.Schemas[0], spec.Definitions["brand"])
	assert.Equal(t, schema.Items.Schemas[1], spec.Definitions["tag"])

	schema = spec.Definitions["allofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/allofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AllOf[0].Ref.String())
	assert.Empty(t, schema.AllOf[1].Ref.String())
	assert.Equal(t, schema.AllOf[0], spec.Definitions["brand"])
	assert.Equal(t, schema.AllOf[1], spec.Definitions["tag"])

	schema = spec.Definitions["anyofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/anyofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AnyOf[0].Ref.String())
	assert.Empty(t, schema.AnyOf[1].Ref.String())
	assert.Equal(t, schema.AnyOf[0], spec.Definitions["brand"])
	assert.Equal(t, schema.AnyOf[1], spec.Definitions["tag"])

	schema = spec.Definitions["oneofBoth"]
	s, err = expandSchema(schema, []string{"#/definitions/oneofBoth"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.OneOf[0].Ref.String())
	assert.Empty(t, schema.OneOf[1].Ref.String())
	assert.Equal(t, schema.OneOf[0], spec.Definitions["brand"])
	assert.Equal(t, schema.OneOf[1], spec.Definitions["tag"])

	schema = spec.Definitions["notSomething"]
	s, err = expandSchema(schema, []string{"#/definitions/notSomething"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.Not.Ref.String())
	assert.Equal(t, *schema.Not, spec.Definitions["tag"])

	schema = spec.Definitions["withAdditional"]
	s, err = expandSchema(schema, []string{"#/definitions/withAdditional"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AdditionalProperties.Schema.Ref.String())
	assert.Equal(t, *schema.AdditionalProperties.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["withAdditionalItems"]
	s, err = expandSchema(schema, []string{"#/definitions/withAdditionalItems"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	assert.Empty(t, schema.AdditionalItems.Schema.Ref.String())
	assert.Equal(t, *schema.AdditionalItems.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["withPattern"]
	s, err = expandSchema(schema, []string{"#/definitions/withPattern"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop := schema.PatternProperties["^x-ab"]
	assert.Empty(t, prop.Ref.String())
	assert.Equal(t, prop, spec.Definitions["tag"])

	schema = spec.Definitions["deps"]
	s, err = expandSchema(schema, []string{"#/definitions/deps"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop2 := schema.Dependencies["something"]
	assert.Empty(t, prop2.Schema.Ref.String())
	assert.Equal(t, *prop2.Schema, spec.Definitions["tag"])

	schema = spec.Definitions["defined"]
	s, err = expandSchema(schema, []string{"#/definitions/defined"}, resolver, basePath)
	require.NoError(t, err)
	require.NotNil(t, s)

	schema = *s
	prop = schema.Definitions["something"]
	assert.Empty(t, prop.Ref.String())
	assert.Equal(t, prop, spec.Definitions["tag"])

}

func TestRelativeBaseURI(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("fixtures/remote")))
	defer server.Close()

	spec := new(Swagger)
	// resolver, err := defaultSchemaLoader(spec, nil, nil,nil)
	// assert.NoError(t, err)

	require.NoError(t, ExpandSpec(spec, nil))

	specDoc, err := jsonDoc("fixtures/remote/all-the-things.json")
	require.NoError(t, err)

	opts := &ExpandOptions{
		RelativeBase: server.URL + "/all-the-things.json",
	}

	spec = new(Swagger)
	require.NoError(t, json.Unmarshal(specDoc, spec))

	pet := spec.Definitions["pet"]
	errorModel := spec.Definitions["errorModel"]
	petResponse := spec.Responses["petResponse"]
	petResponse.Schema = &pet
	stringResponse := spec.Responses["stringResponse"]
	tagParam := spec.Parameters["tag"]
	idParam := spec.Parameters["idParam"]

	anotherPet := spec.Responses["anotherPet"]
	anotherPet.Ref = MustCreateRef(server.URL + "/" + anotherPet.Ref.String())
	require.NoError(t, ExpandResponse(&anotherPet, opts.RelativeBase))
	spec.Responses["anotherPet"] = anotherPet

	circularA := spec.Responses["circularA"]
	circularA.Ref = MustCreateRef(server.URL + "/" + circularA.Ref.String())
	require.NoError(t, ExpandResponse(&circularA, opts.RelativeBase))
	spec.Responses["circularA"] = circularA

	require.NoError(t, ExpandSpec(spec, opts))

	assert.Equal(t, tagParam, spec.Parameters["query"])

	assert.Equal(t, petResponse, spec.Responses["petResponse"])
	assert.Equal(t, petResponse, spec.Responses["anotherPet"])
	assert.Equal(t, petResponse, spec.Paths.Paths["/pets"].Post.Responses.StatusCodeResponses[200])
	assert.Equal(t, petResponse, spec.Paths.Paths["/"].Get.Responses.StatusCodeResponses[200])

	assert.Equal(t, pet, *spec.Responses["petResponse"].Schema)
	assert.Equal(t, pet, *spec.Paths.Paths["/pets"].Get.Responses.StatusCodeResponses[200].Schema.Items.Schema)
	assert.Equal(t, pet, spec.Definitions["petInput"].AllOf[0])

	assert.Equal(t, spec.Definitions["petInput"], *spec.Paths.Paths["/pets"].Post.Parameters[0].Schema)

	assert.Equal(t, stringResponse, *spec.Paths.Paths["/"].Get.Responses.Default)

	assert.Equal(t, errorModel, *spec.Paths.Paths["/pets"].Get.Responses.Default.Schema)
	assert.Equal(t, errorModel, *spec.Paths.Paths["/pets"].Post.Responses.Default.Schema)

	pi := spec.Paths.Paths["/pets/{id}"]
	assert.Equal(t, idParam, pi.Get.Parameters[0])
	assert.Equal(t, petResponse, pi.Get.Responses.StatusCodeResponses[200])
	assert.Equal(t, idParam, pi.Delete.Parameters[0])

	assert.Equal(t, errorModel, *pi.Get.Responses.Default.Schema)
	assert.Equal(t, errorModel, *pi.Delete.Responses.Default.Schema)
}

func TestExpandRemoteRef_WithResolutionContext(t *testing.T) {
	server := resolutionContextServer()
	defer server.Close()

	var tgt Schema
	ref, err := NewRef(server.URL + "/resolution.json#/definitions/bool")
	require.NoError(t, err)

	tgt.Ref = ref
	require.NoError(t, ExpandSchema(&tgt, nil, nil))
	assert.Equal(t, StringOrArray([]string{"boolean"}), tgt.Type)
}

func TestExpandRemoteRef_WithNestedResolutionContext(t *testing.T) {
	server := resolutionContextServer()
	defer server.Close()

	var tgt Schema
	ref, err := NewRef(server.URL + "/resolution.json#/items")
	require.NoError(t, err)

	tgt.Ref = ref
	require.NoError(t, ExpandSchema(&tgt, nil, nil))
	assert.Equal(t, StringOrArray([]string{"string"}), tgt.Items.Schema.Type)
}

/* This next test will have to wait until we do full $ID analysis for every subschema on every file that is referenced */
/* For now, TestExpandRemoteRef_WithNestedResolutionContext replaces this next test */
// func TestExpandRemoteRef_WithNestedResolutionContext_WithParentID(t *testing.T) {
// 	server := resolutionContextServer()
// 	defer server.Close()

// 	var tgt Schema
// 	ref, err := NewRef(server.URL + "/resolution.json#/items/items")
// 	require.NoError(t, err)
//
// 	tgt.Ref = ref
// 	require.NoError(t, ExpandSchema(&tgt, nil, nil))
// 	assert.Equal(t, StringOrArray([]string{"string"}), tgt.Type)
// }

func TestExpandRemoteRef_WithNestedResolutionContextWithFragment(t *testing.T) {
	server := resolutionContextServer()
	defer server.Close()

	var tgt Schema
	ref, err := NewRef(server.URL + "/resolution2.json#/items")
	require.NoError(t, err)

	tgt.Ref = ref
	require.NoError(t, ExpandSchema(&tgt, nil, nil))
	assert.Equal(t, StringOrArray([]string{"file"}), tgt.Items.Schema.Type)
}

func TestExpandForTransitiveRefs(t *testing.T) {
	var spec *Swagger
	rawSpec, err := ioutil.ReadFile(filepath.Join(specs, "todos.json"))
	require.NoError(t, err)

	basePath, err := absPath(filepath.Join(specs, "todos.json"))
	require.NoError(t, err)

	opts := &ExpandOptions{
		RelativeBase: basePath,
	}

	require.NoError(t, json.Unmarshal(rawSpec, &spec))

	require.NoError(t, ExpandSpec(spec, opts))
}

func TestExpandSchemaWithRoot(t *testing.T) {
	root := new(Swagger)
	require.NoError(t, json.Unmarshal(PetStoreJSONMessage, root))

	// 1. remove ID from root definition
	origPet := root.Definitions["Pet"]
	newPet := origPet
	newPet.ID = ""
	root.Definitions["Pet"] = newPet
	expandRootWithID(t, root, withoutSchemaID)

	// 2. put back ID in Pet definition
	// nested $ref should fail
	// Debug = true
	root.Definitions["Pet"] = origPet
	expandRootWithID(t, root, withSchemaID)
}

func expandRootWithID(t testing.TB, root *Swagger, testcase string) {
	t.Logf("case: expanding $ref to schema without ID, with nested $ref with %s ID", testcase)
	sch := &Schema{
		SchemaProps: SchemaProps{
			Ref: MustCreateRef("#/definitions/newPet"),
		},
	}
	err := ExpandSchema(sch, root, nil)
	if testcase == withSchemaID {
		require.Errorf(t, err, "expected %s NOT to expand properly because of the ID in the parent schema", sch.Ref.String())
	} else {
		require.NoErrorf(t, err, "expected %s to expand properly", sch.Ref.String())
	}

	if Debug {
		bbb, _ := json.MarshalIndent(sch, "", " ")
		t.Log(string(bbb))
	}

	t.Log("case: expanding $ref to schema without nested $ref")
	sch = &Schema{
		SchemaProps: SchemaProps{
			Ref: MustCreateRef("#/definitions/Category"),
		},
	}
	require.NoErrorf(t, ExpandSchema(sch, root, nil), "expected %s to expand properly", sch.Ref.String())

	if Debug {
		bbb, _ := json.MarshalIndent(sch, "", " ")
		t.Log(string(bbb))
	}
	t.Logf("case: expanding $ref to schema with %s ID and nested $ref", testcase)
	sch = &Schema{
		SchemaProps: SchemaProps{
			Ref: MustCreateRef("#/definitions/Pet"),
		},
	}
	err = ExpandSchema(sch, root, nil)
	if testcase == withSchemaID {
		require.Errorf(t, err, "expected %s NOT to expand properly because of the ID in the parent schema", sch.Ref.String())
	} else {
		require.NoErrorf(t, err, "expected %s to expand properly", sch.Ref.String())
	}

	if Debug {
		bbb, _ := json.MarshalIndent(sch, "", " ")
		t.Log(string(bbb))
	}
}

func TestExpandPathItem(t *testing.T) {
	spec := new(Swagger)
	specDoc, err := jsonDoc(pathItemsFixture)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(specDoc, spec))

	specPath, err := absPath(pathItemsFixture)
	require.NoError(t, err)

	// ExpandSpec use case
	require.NoError(t, ExpandSpec(spec, &ExpandOptions{RelativeBase: specPath}))

	jazon, _ := json.MarshalIndent(spec, "", " ")
	assert.JSONEq(t, `{
         "swagger": "2.0",
         "info": {
          "title": "PathItems refs",
          "version": "1.0"
         },
         "paths": {
          "/todos": {
           "get": {
            "responses": {
             "200": {
              "description": "List Todos",
              "schema": {
               "type": "array",
               "items": {
                "type": "string"
               }
              }
             },
             "404": {
              "description": "error"
             }
            }
           }
          }
         }
			 }`, string(jazon))
}

func TestExpandExtraItems(t *testing.T) {
	spec := new(Swagger)
	specDoc, err := jsonDoc(extraRefFixture)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(specDoc, spec))

	specPath, err := absPath(extraRefFixture)
	require.NoError(t, err)

	// ExpandSpec use case: unsupported $refs are not expanded
	require.NoError(t, ExpandSpec(spec, &ExpandOptions{RelativeBase: specPath}))

	jazon, _ := json.MarshalIndent(spec, "", " ")
	assert.JSONEq(t, `{
         "schemes": [
          "http"
         ],
         "swagger": "2.0",
         "info": {
          "title": "Supported, but non Swagger 20 compliant $ref constructs",
          "version": "2.1.0"
         },
         "host": "item.com",
         "basePath": "/extraRefs",
         "paths": {
          "/employees": {
           "get": {
            "summary": "List Employee Types",
            "operationId": "LIST-Employees",
            "parameters": [
             {
							"description": "unsupported $ref in simple param",
              "type": "array",
              "items": {
               "$ref": "#/definitions/arrayType"
              },
              "name": "myQueryParam",
              "in": "query"
             }
            ],
            "responses": {
             "200": {
							"description": "unsupported $ref in header",
              "schema": {
               "type": "string"
              },
              "headers": {
               "X-header": {
                  "type": "array",
                  "items": {
                    "$ref": "#/definitions/headerType"
                  }
							  }
              }
             }
            }
           }
          }
         },
         "definitions": {
          "arrayType": {
           "type": "integer",
           "format": "int32"
          },
          "headerType": {
           "type": "string",
           "format": "uuid"
          }
         }
			 }`, string(jazon))
}

func Test_CircularID(t *testing.T) {
	go func() {
		err := http.ListenAndServe("localhost:1234", http.FileServer(http.Dir("fixtures/more_circulars/remote")))
		if err != nil {
			panic(err.Error())
		}
	}()
	time.Sleep(100 * time.Millisecond)

	fixturePath := "http://localhost:1234/tree"
	jazon := expandThisSchemaOrDieTrying(t, fixturePath)

	sch := new(Schema)
	require.NoError(t, json.Unmarshal([]byte(jazon), sch))

	require.NotPanics(t, func() {
		assert.NoError(t, ExpandSchemaWithBasePath(sch, nil, &ExpandOptions{}))
	})

	fixturePath = "fixtures/more_circulars/with-id.json"
	jazon = expandThisOrDieTrying(t, fixturePath)

	// cannot guarantee that the circular will always hook on the same $ref
	// but we can assert that thre is only one
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotEmpty(t, m)

	refs := make(map[string]struct{}, 5)
	for _, matched := range m {
		subMatch := matched[1]
		refs[subMatch] = struct{}{}
	}

	require.Len(t, refs, 1)
}

// PetStore20 json doc for swagger 2.0 pet store
const PetStore20 = `{
  "swagger": "2.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "contact": {
      "name": "Wordnik API Team",
      "url": "http://developer.wordnik.com"
    },
    "license": {
      "name": "Creative Commons 4.0 International",
      "url": "http://creativecommons.org/licenses/by/4.0/"
    }
  },
  "host": "petstore.swagger.wordnik.com",
  "basePath": "/api",
  "schemes": [
    "http"
  ],
  "paths": {
    "/pets": {
      "get": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "getAllPets",
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "description": "The status to filter by",
            "type": "string"
          },
          {
            "name": "limit",
            "in": "query",
            "description": "The maximum number of results to return",
            "type": "integer",
						"format": "int64"
          }
        ],
        "summary": "Finds all pets in the system",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Pet"
              }
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "post": {
        "security": [
          {
            "basic": []
          }
        ],
        "tags": [ "Pet Operations" ],
        "operationId": "createPet",
        "summary": "Creates a new pet",
        "consumes": ["application/x-yaml"],
        "produces": ["application/x-yaml"],
        "parameters": [
          {
            "name": "pet",
            "in": "body",
            "description": "The Pet to create",
            "required": true,
            "schema": {
              "$ref": "#/definitions/newPet"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Created Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/pets/{id}": {
      "delete": {
        "security": [
          {
            "apiKey": []
          }
        ],
        "description": "Deletes the Pet by id",
        "operationId": "deletePet",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "ID of pet to delete",
            "required": true,
            "type": "integer",
            "format": "int64"
          }
        ],
        "responses": {
          "204": {
            "description": "pet deleted"
          },
          "default": {
            "description": "unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "get": {
        "tags": [ "Pet Operations" ],
        "operationId": "getPetById",
        "summary": "Finds the pet by id",
        "responses": {
          "200": {
            "description": "Pet response",
            "schema": {
              "$ref": "#/definitions/Pet"
            }
          },
          "default": {
            "description": "Unexpected error",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "description": "ID of pet",
          "required": true,
          "type": "integer",
          "format": "int64"
        }
      ]
    }
  },
  "definitions": {
    "Category": {
      "id": "Category",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Pet": {
      "id": "Pet",
      "properties": {
        "category": {
          "$ref": "#/definitions/Category"
        },
        "id": {
          "description": "unique identifier for the pet",
          "format": "int64",
          "maximum": 100.0,
          "minimum": 0.0,
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "photoUrls": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "status": {
          "description": "pet status in the store",
          "enum": [
            "available",
            "pending",
            "sold"
          ],
          "type": "string"
        },
        "tags": {
          "items": {
            "$ref": "#/definitions/Tag"
          },
          "type": "array"
        }
      },
      "required": [
        "id",
        "name"
      ]
    },
    "newPet": {
      "anyOf": [
        {
          "$ref": "#/definitions/Pet"
        },
        {
          "required": [
            "name"
          ]
        }
      ]
    },
    "Tag": {
      "id": "Tag",
      "properties": {
        "id": {
          "format": "int64",
          "type": "integer"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "Error": {
      "required": [
        "code",
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        }
      }
    }
  },
  "consumes": [
    "application/json",
    "application/xml"
  ],
  "produces": [
    "application/json",
    "application/xml",
    "text/plain",
    "text/html"
  ],
  "securityDefinitions": {
    "basic": {
      "type": "basic"
    },
    "apiKey": {
      "type": "apiKey",
      "in": "header",
      "name": "X-API-KEY"
    }
  }
}
`
