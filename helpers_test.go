package spec

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/require"
)

func jsonDoc(path string) (json.RawMessage, error) {
	data, err := swag.LoadFromFileOrHTTP(path)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func docAndOpts(t testing.TB, fixturePath string) ([]byte, *ExpandOptions) {
	doc, err := jsonDoc(fixturePath)
	require.NoError(t, err)

	specPath, _ := absPath(fixturePath)

	return doc, &ExpandOptions{
		RelativeBase: specPath,
	}
}

func expandThisSchemaOrDieTrying(t testing.TB, fixturePath string) string {
	doc, opts := docAndOpts(t, fixturePath)

	sch := new(Schema)
	require.NoError(t, json.Unmarshal(doc, sch))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSchemaWithBasePath(sch, nil, opts))
	}, "Calling expand schema circular refs, should not panic!")

	bbb, err := json.MarshalIndent(sch, "", " ")
	require.NoError(t, err)

	return string(bbb)
}

func expandThisOrDieTrying(t testing.TB, fixturePath string) string {
	doc, opts := docAndOpts(t, fixturePath)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "Calling expand spec with circular refs, should not panic!")

	bbb, err := json.MarshalIndent(spec, "", " ")
	require.NoError(t, err)

	return string(bbb)
}

func resolutionContextServer() *httptest.Server {
	var servedAt string
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/resolution.json" {

			b, _ := ioutil.ReadFile(filepath.Join(specs, "resolution.json"))
			var ctnt map[string]interface{}
			_ = json.Unmarshal(b, &ctnt)
			ctnt["id"] = servedAt

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(200)
			bb, _ := json.Marshal(ctnt)
			_, _ = rw.Write(bb)
			return
		}
		if req.URL.Path == "/resolution2.json" {
			b, _ := ioutil.ReadFile(filepath.Join(specs, "resolution2.json"))
			var ctnt map[string]interface{}
			_ = json.Unmarshal(b, &ctnt)
			ctnt["id"] = servedAt

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(200)
			bb, _ := json.Marshal(ctnt)
			_, _ = rw.Write(bb)
			return
		}

		if req.URL.Path == "/boolProp.json" {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(200)
			b, _ := json.Marshal(map[string]interface{}{
				"type": "boolean",
			})
			_, _ = rw.Write(b)
			return
		}

		if req.URL.Path == "/deeper/stringProp.json" {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(200)
			b, _ := json.Marshal(map[string]interface{}{
				"type": "string",
			})
			_, _ = rw.Write(b)
			return
		}

		if req.URL.Path == "/deeper/arrayProp.json" {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(200)
			b, _ := json.Marshal(map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "file",
				},
			})
			_, _ = rw.Write(b)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	}))
	servedAt = server.URL
	return server
}
