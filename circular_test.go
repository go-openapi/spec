package spec

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandCircular_Issue3(t *testing.T) {
	jazon := expandThisOrDieTrying(t, "fixtures/expansion/overflow.json")
	require.NotEmpty(t, jazon)

	// TODO: assert $ref
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
		_, err := expandSchema(schema, []string{"#/definitions/car"}, resolver, basePath)
		require.NoError(t, err)
	}, "Calling expand schema with circular refs, should not panic!")
}

func TestExpandCircular_Spec2Expansion(t *testing.T) {
	// TODO: assert repeatable results (see commented section below)

	fixturePath := filepath.Join("fixtures", "expansion", "circular-minimal.json")
	jazon := expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)

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

		This may result in different, functionally correct expanded specs (e.g. with same validations)

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
	assertRefInJSON(t, jazon, "item.json#/item")

	fixturePath = "fixtures/more_circulars/spec2.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "item2.json#/item")

	fixturePath = "fixtures/more_circulars/spec3.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "item.json#/item")

	fixturePath = "fixtures/more_circulars/spec4.json"
	jazon = expandThisOrDieTrying(t, fixturePath)
	require.NotEmpty(t, jazon)
	assertRefInJSON(t, jazon, "item4.json#/item")
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
	// but we can assert that there is only one
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotEmpty(t, m)

	refs := make(map[string]struct{}, 5)
	for _, matched := range m {
		subMatch := matched[1]
		refs[subMatch] = struct{}{}
	}

	// TODO(fred): the expansion is incorrect (it was already, with an undetected empty $ref)
	// require.Len(t, refs, 1)
}
