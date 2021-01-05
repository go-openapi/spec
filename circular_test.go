package spec

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandCircular_Issue3(t *testing.T) {
	jazon := expandThisOrDieTrying(t, "fixtures/expansion/overflow.json")
	require.NotEmpty(t, jazon)

	// all $ref are in the root document
	assertRefInJSON(t, jazon, "#/definitions/")
}

func TestExpandCircular_RefExpansion(t *testing.T) {
	carsDoc, err := jsonDoc("fixtures/expansion/circularRefs.json")
	require.NoError(t, err)

	basePath, _ := absPath("fixtures/expansion/circularRefs.json")

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(carsDoc, spec))

	resolver := defaultSchemaLoader(spec, &ExpandOptions{RelativeBase: basePath}, nil, nil)

	schema := spec.Definitions["car"]

	assert.NotPanics(t, func() {
		_, err := expandSchema(schema, []string{"#/definitions/car"}, resolver, basePath, "/definitions/car")
		require.NoError(t, err)
	}, "Calling expand schema with circular refs, should not panic!")
}

func TestExpandCircular_Minimal(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "expansion", "circular-minimal.json")
	jazon := expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)

	assert.NotContainsf(t, jazon, "circular-minimal.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)

	/*
		At the moment, the result of expanding circular references is not stable (issue #93),
		when several cycles have intersections:
		the spec structure is randomly walked through and mutating as expansion is carried out.
		detected cycles in $ref are not necessarily the shortest matches.

		This may result in different, functionally correct expanded specs (e.g. with same validations)
	*/
	assertRefInJSON(t, jazon, "#/definitions/node") // NOTE: we are not sure which node definition is used
}

func TestExpandCircular_Spec2Expansion(t *testing.T) {
	// assert stripped $ref in result

	fixturePath := "fixtures/expansion/circularSpec2.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	assert.NotEmpty(t, jazon)

	assert.NotContainsf(t, jazon, "circularSpec.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)

	assertRefInJSON(t, jazon, "#/definitions/")
}

func TestExpandCircular_MoreCircular(t *testing.T) {
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
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "#/responses/itemResponse/schema")

	fixturePath = "fixtures/more_circulars/spec2.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "#/responses/itemResponse/schema")

	fixturePath = "fixtures/more_circulars/spec3.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "#/definitions/myItems")

	fixturePath = "fixtures/more_circulars/spec4.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "#/parameters/itemParameter/schema")
}

func TestExpandCircular_Issue957(t *testing.T) {
	fixturePath := "fixtures/bugs/957/fixture-957.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)

	require.NotContainsf(t, jazon, "fixture-957.json#/",
		"expected %s to be expanded with stripped circular $ref", fixturePath)

	assertRefInJSON(t, jazon, "#/definitions/")
}

func TestExpandCircular_Bitbucket(t *testing.T) {
	// Additional testcase for circular $ref (from bitbucket api)

	fixturePath := "fixtures/more_circulars/bitbucket.json"
	jazon := expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)

	assertRefInJSON(t, jazon, "#/definitions/")
}

func TestExpandCircular_ResponseWithRoot(t *testing.T) {
	rootDoc := new(Swagger)
	b, err := ioutil.ReadFile("fixtures/more_circulars/resp.json")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, rootDoc))

	path := rootDoc.Paths.Paths["/api/v1/getx"]
	resp := path.Post.Responses.StatusCodeResponses[200]

	thisCache := cacheOrDefault(nil)

	// during the first response expand, refs are getting expanded,
	// so the following expands cannot properly resolve them w/o the document.
	// this happens in validator.Validate() when different validators try to expand the same mutable response.
	require.NoError(t, ExpandResponseWithRoot(&resp, rootDoc, thisCache))

	bbb, _ := json.MarshalIndent(resp, "", " ")
	assertRefInJSON(t, string(bbb), "#/definitions/MyObj")

	// do it again
	require.NoError(t, ExpandResponseWithRoot(&resp, rootDoc, thisCache))
}

func TestExpandCircular_Issue415(t *testing.T) {
	jazon := expandThisOrDieTrying(t, "fixtures/expansion/clickmeter.json")
	require.NotEmpty(t, jazon)

	assertRefInJSON(t, jazon, "#/definitions/")
}

func TestExpandCircular_SpecExpansion(t *testing.T) {
	jazon := expandThisOrDieTrying(t, "fixtures/expansion/circularSpec.json")
	require.NotEmpty(t, jazon)

	assertRefInJSON(t, jazon, "#/definitions/Book")
}

func TestExpandCircular_RemoteCircularID(t *testing.T) {
	go func() {
		err := http.ListenAndServe("localhost:1234", http.FileServer(http.Dir("fixtures/more_circulars/remote")))
		if err != nil {
			panic(err.Error())
		}
	}()
	time.Sleep(100 * time.Millisecond)

	t.Run("CircularID", func(t *testing.T) {
		fixturePath := "http://localhost:1234/tree"
		jazon := expandThisSchemaOrDieTrying(t, fixturePath)

		// all $ref are now in the single root
		assertRefInJSONRegexp(t, jazon, "(^#/definitions/node$)|(^#?$)") // root $ref should be '#' or ""

		sch := new(Schema)
		require.NoError(t, json.Unmarshal([]byte(jazon), sch))

		// expand already expanded: this is not an idempotent operation: circular $ref
		// are expanded again until a (deeper) cycle is detected
		require.NoError(t, ExpandSchema(sch, nil, nil))

		// expand already expanded
		require.NoError(t, ExpandSchema(sch, nil, nil))

		// Empty base path fails:
		require.Error(t, ExpandSchemaWithBasePath(sch, nil, &ExpandOptions{}))
	})

	t.Run("withID", func(t *testing.T) {
		// This test exhibits a broken feature when using nested schema ID
		const fixturePath = "fixtures/more_circulars/with-id.json"
		jazon := expandThisOrDieTrying(t, fixturePath)

		// TODO(fred): the $ref expanded as: "$ref": "" is incorrect.
		assertRefInJSONRegexp(t, jazon, "(^#/definitions/)|(^#?$)")

		// cannot guarantee that the circular will always hook on the same $ref
		// but we can assert that thre is only one
		//
		// TODO(fred): the expansion is incorrect (it was already, with an undetected empty $ref)
		// At the moment there is one single non-empty $ref (which is correct)
		// and one empty $ref (which is invalid)
		nonEmptyRef := regexp.MustCompile(`"\$ref":\s*"(.+)"`)
		m := nonEmptyRef.FindAllStringSubmatch(jazon, -1)
		require.NotEmpty(t, m)

		refs := make(map[string]struct{}, 2)
		for _, matched := range m {
			subMatch := matched[1]
			refs[subMatch] = struct{}{}
		}

		require.Len(t, refs, 1)
	})
}

func TestSortRefTracker(t *testing.T) {
	tracked := refTrackers{
		refTracker{Pointer: "/c/d/e"},
		refTracker{Pointer: "/definitions/x"},
		refTracker{Pointer: "/a/b/c/d"},
		refTracker{Pointer: "/b"},
		refTracker{Pointer: "/z"},
		refTracker{Pointer: "/definitions/a"},
	}
	sort.Sort(tracked)
	require.EqualValues(t, refTrackers{
		refTracker{Pointer: "/definitions/a"},
		refTracker{Pointer: "/definitions/x"},
		refTracker{Pointer: "/b"},
		refTracker{Pointer: "/z"},
		refTracker{Pointer: "/c/d/e"},
		refTracker{Pointer: "/a/b/c/d"},
	}, tracked)
}

func TestRemoteExpandAzure(t *testing.T) {
	// local copy of : https://raw.githubusercontent.com/Azure/azure-rest-api-specs/master/specification/network/resource-manager/Microsoft.Network/stable/2020-04-01/publicIpAddress.json
	server := httptest.NewServer(http.FileServer(http.Dir("fixtures/azure")))
	defer server.Close()

	jazon := expandThisOrDieTrying(t, server.URL+"/publicIpAddress.json")

	assertRefInJSONRegexp(t, jazon, `^(#/definitions/)|(#/paths/.+/get/default/schema/properties/error)|(\./examples/)`)
}

func TestDocRef(t *testing.T) {
	doc := []byte(`{
        "description": "root pointer ref",
        "schema": {
            "properties": {
                "foo": {"$ref": "#"}
            },
            "additionalProperties": false
					}
				}`)
	var schema Schema

	require.NoError(t, json.Unmarshal(doc, &schema))

	// expand from root
	require.NoError(t, ExpandSchema(&schema, &schema, nil))

	jazon, err := json.MarshalIndent(schema, "", " ")
	require.NoError(t, err)

	assertRefInJSONRegexp(t, string(jazon), `(^#$)|(^$)`)

	// expand from self
	require.NoError(t, ExpandSchema(&schema, nil, nil))

	jazon, err = json.MarshalIndent(schema, "", " ")
	require.NoError(t, err)

	assertRefInJSONRegexp(t, string(jazon), `(^#$)|(^$)`)

	// expand from file
	temp, err := ioutil.TempFile(".", "test_doc_ref*.json")
	require.NoError(t, err)

	file := temp.Name()
	defer func() {
		_ = os.Remove(file)
	}()
	_, err = temp.Write(doc)
	require.NoError(t, err)
	require.NoError(t, temp.Close())

	require.NoError(t, ExpandSchemaWithBasePath(&schema, nil, &ExpandOptions{RelativeBase: file}))

	jazon, err = json.MarshalIndent(schema, "", " ")
	require.NoError(t, err)

	assertRefInJSONRegexp(t, string(jazon), `(^#$)|(^$)`)

	ref := RefSchema("#")
	require.NoError(t, ExpandSchema(ref, &schema, nil))
	jazon, err = json.MarshalIndent(ref, "", " ")
	require.NoError(t, err)
	assertRefInJSONRegexp(t, string(jazon), `(^#$)|(^$)`)
}
