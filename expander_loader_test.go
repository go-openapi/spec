// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// errUnexpectedLoad is returned by a test loader asked to load a path it does not expect.
var errUnexpectedLoad = errors.New("unexpected load")

func TestExpandSchemaWithOptions(t *testing.T) {
	// A schema whose $ref points into an external document is expanded through the injected
	// option-aware loader, exactly as flatten/analysis and validate need for confined expansion.
	const external = `{"definitions":{"Thing":{"type":"string"}}}`

	root := map[string]any{"swagger": "2.0", "definitions": map[string]any{}}
	schema := RefSchema("external.json#/definitions/Thing")

	var loaderCalls int
	err := ExpandSchemaWithOptions(schema, root, nil, &ExpandOptions{
		PathLoaderWithOptions: func(pth string, _ ...loading.Option) (json.RawMessage, error) {
			if strings.Contains(pth, "external.json") {
				loaderCalls++
				return json.RawMessage(external), nil
			}
			return nil, fmt.Errorf("%w: %s", errUnexpectedLoad, pth)
		},
	})
	require.NoError(t, err)
	assert.TrueT(t, loaderCalls > 0, "expected the injected loader to resolve the external $ref")

	out, err := json.Marshal(schema)
	require.NoError(t, err)
	assert.StringContainsT(t, string(out), `"type":"string"`)
	assert.StringNotContainsT(t, string(out), `"$ref"`)

	t.Run("nil options behaves like ExpandSchema (in-memory root, fragment ref)", func(t *testing.T) {
		inMemRoot := map[string]any{
			"definitions": map[string]any{"Local": map[string]any{"type": "integer"}},
		}
		sch := RefSchema("#/definitions/Local")
		require.NoError(t, ExpandSchemaWithOptions(sch, inMemRoot, nil, nil))

		out, err := json.Marshal(sch)
		require.NoError(t, err)
		assert.StringContainsT(t, string(out), `"type":"integer"`)
	})
}

func TestExpandParameterResponseWithOptions(t *testing.T) {
	// Parameter and response $ref pointing into an external document are expanded through the
	// injected option-aware loader — the path go-openapi/validate needs for confined validation.
	const external = `{
		"parameters":{"Foo":{"name":"foo","in":"query","type":"string"}},
		"responses":{"Bar":{"description":"ok"}}
	}`

	var loaderCalls int
	loader := func(pth string, _ ...loading.Option) (json.RawMessage, error) {
		if strings.Contains(pth, "external.json") {
			loaderCalls++
			return json.RawMessage(external), nil
		}
		return nil, fmt.Errorf("%w: %s", errUnexpectedLoad, pth)
	}
	opts := &ExpandOptions{RelativeBase: "spec.json", PathLoaderWithOptions: loader}

	t.Run("parameter", func(t *testing.T) {
		param := new(Parameter)
		param.Ref = MustCreateRef("external.json#/parameters/Foo")
		require.NoError(t, ExpandParameterWithOptions(param, nil, nil, opts))
		assert.EqualT(t, "foo", param.Name)
		assert.EqualT(t, "", param.Ref.String())
	})

	t.Run("response", func(t *testing.T) {
		resp := new(Response)
		resp.Ref = MustCreateRef("external.json#/responses/Bar")
		require.NoError(t, ExpandResponseWithOptions(resp, nil, nil, opts))
		assert.EqualT(t, "ok", resp.Description)
		assert.EqualT(t, "", resp.Ref.String())
	})

	assert.TrueT(t, loaderCalls >= 2, "expected the injected loader to resolve both external $ref")
}

func TestPathLoaderSelection(t *testing.T) {
	t.Run("option-aware loader is used when set", func(t *testing.T) {
		var called string
		ctx := newResolverContext(&ExpandOptions{
			PathLoaderWithOptions: func(string, ...loading.Option) (json.RawMessage, error) {
				called = "withOptions"
				return json.RawMessage(`{}`), nil
			},
		})

		_, err := ctx.loadDoc("x")
		require.NoError(t, err)
		assert.EqualT(t, "withOptions", called)
	})

	t.Run("option-aware loader takes precedence over the plain loader", func(t *testing.T) {
		var called string
		ctx := newResolverContext(&ExpandOptions{
			PathLoader: func(string) (json.RawMessage, error) {
				called = "plain"
				return json.RawMessage(`{}`), nil
			},
			PathLoaderWithOptions: func(string, ...loading.Option) (json.RawMessage, error) {
				called = "withOptions"
				return json.RawMessage(`{}`), nil
			},
		})

		_, err := ctx.loadDoc("x")
		require.NoError(t, err)
		assert.EqualT(t, "withOptions", called)
	})

	t.Run("plain loader is used when only it is set", func(t *testing.T) {
		var called string
		ctx := newResolverContext(&ExpandOptions{
			PathLoader: func(string) (json.RawMessage, error) {
				called = "plain"
				return json.RawMessage(`{}`), nil
			},
		})

		_, err := ctx.loadDoc("x")
		require.NoError(t, err)
		assert.EqualT(t, "plain", called)
	})
}

func TestExpand_PathLoaderWithOptions(t *testing.T) {
	// A cross-file $ref forces a document load: prove it is routed through the option-aware loader.
	const other = `{"definitions":{"Thing":{"type":"string"}}}`

	raw := []byte(`{
		"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{},
		"definitions":{"Ref":{"$ref":"other.json#/definitions/Thing"}}
	}`)
	var sw Swagger
	require.NoError(t, json.Unmarshal(raw, &sw))

	var loaderCalls int
	err := ExpandSpec(&sw, &ExpandOptions{
		RelativeBase: "/base/root.json",
		PathLoaderWithOptions: func(pth string, _ ...loading.Option) (json.RawMessage, error) {
			if strings.Contains(pth, "other.json") {
				loaderCalls++
				return json.RawMessage(other), nil
			}
			return nil, fmt.Errorf("%w: %s", errUnexpectedLoad, pth)
		},
	})
	require.NoError(t, err)
	assert.TrueT(t, loaderCalls > 0, "expected the option-aware loader to be invoked")

	// the cross-file $ref has been expanded in place
	out, err := json.Marshal(sw.Definitions["Ref"])
	require.NoError(t, err)
	assert.StringContainsT(t, string(out), `"type":"string"`)
	assert.StringNotContainsT(t, string(out), `"$ref"`)
}
