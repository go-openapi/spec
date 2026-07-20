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
