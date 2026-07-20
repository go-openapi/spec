// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// TestExpand_ConfinedLoader validates end to end that a WithRoot-confined loader, injected
// through PathLoaderWithOptions, safely expands an untrusted spec:
//   - a legitimate $ref that resolves within the root is expanded (this requires the loader to
//     accept the absolute paths spec normalizes references to);
//   - a "file://" $ref and a "../" traversal $ref that point outside the root are blocked, and
//     no byte of the out-of-root file leaks into the result.
//
// It exercises the go-openapi/swag/loading WithRoot behavior from the consumer side, with an
// absolute RelativeBase (the realistic case).
func TestExpand_ConfinedLoader(t *testing.T) {
	const secretMarker = "TOP_SECRET"

	root := t.TempDir()
	outside := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "child.json"),
		[]byte(`{"definitions":{"Thing":{"type":"string","title":"IN_ROOT"}}}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.json"),
		[]byte(`{"leaked":"`+secretMarker+`"}`), 0o600))

	secretAbs := filepath.Join(outside, "secret.json")
	// a relative traversal, from the spec's base dir (root), that reaches the outside secret
	traversal, err := filepath.Rel(root, secretAbs)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(traversal, ".."), "sanity: traversal must escape the root")

	raw := `{
		"swagger":"2.0","info":{"title":"x","version":"1"},"paths":{},
		"definitions":{
			"Local":     {"$ref":"child.json#/definitions/Thing"},
			"SecretFile":{"$ref":"file://` + filepath.ToSlash(secretAbs) + `"},
			"Traversal": {"$ref":"` + filepath.ToSlash(traversal) + `"}
		}
	}`
	var sw Swagger
	require.NoError(t, json.Unmarshal([]byte(raw), &sw))

	confined := func(pth string, _ ...loading.Option) (json.RawMessage, error) {
		b, err := loading.LoadFromFileOrHTTP(pth, loading.WithRoot(root))
		return json.RawMessage(b), err
	}

	// absolute base, as a real consumer would pass
	err = ExpandSpec(&sw, &ExpandOptions{
		RelativeBase:          filepath.Join(root, "api.json"),
		PathLoaderWithOptions: confined,
		ContinueOnError:       true, // do not abort on the blocked refs; expand what is legitimate
	})
	require.NoError(t, err)

	dump := func(name string) string {
		out, err := json.Marshal(sw.Definitions[name])
		require.NoError(t, err)
		return string(out)
	}

	// 1) the legitimate in-root ref resolved (the WithRoot fix: absolute in-root paths are accepted)
	local := dump("Local")
	assert.StringContainsT(t, local, "IN_ROOT")
	assert.StringNotContainsT(t, local, `"$ref"`)

	// 2) the escaping refs were blocked: they remain unexpanded and leak nothing
	assert.StringContainsT(t, dump("SecretFile"), `"$ref"`)
	assert.StringContainsT(t, dump("Traversal"), `"$ref"`)

	// 3) the secret never appears anywhere in the expanded document
	whole, err := json.Marshal(&sw)
	require.NoError(t, err)
	assert.StringNotContainsT(t, string(whole), secretMarker)
}
