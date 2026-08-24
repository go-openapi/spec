// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// TestExpand_IsReproducible expands the same document repeatedly and requires the same bytes
// every time.
//
// Expansion inlines the first branch that reaches a cycle and leaves a $ref on the others, so
// the order siblings are visited in decides which node holds the $ref. Ranging a map straight
// gave that decision to Go's map iteration: fixture-957.json produced 40 different documents in
// 40 runs. See go-openapi/spec#93.
func TestExpand_IsReproducible(t *testing.T) {
	const runs = 10

	for _, fixture := range []struct{ name, path string }{
		{"cycles sharing a node", filepath.Join("testdata", "expansion", "shared-node-cycles.json")},
		{"several entries into remote cycles", filepath.Join("testdata", "expansion", "multi-entry-cycles", "root.json")},
		{"a cycle in the root", filepath.Join("testdata", "expansion", "circularSpec.json")},
		{"issue 957", filepath.Join("testdata", "bugs", "957", "fixture-957.json")},
		{"bitbucket", filepath.Join("testdata", "more_circulars", "bitbucket.json")},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(fixture.path)
			require.NoError(t, err)

			first := expandOnce(t, data, fixture.path)
			for range runs - 1 {
				assert.EqualT(t, first, expandOnce(t, data, fixture.path),
					"expanding %s twice gave two different documents", fixture.path)
			}
		})
	}
}

func expandOnce(t *testing.T, data []byte, basePath string) string {
	t.Helper()

	doc := new(Swagger)
	require.NoError(t, json.Unmarshal(data, doc))
	require.NoError(t, ExpandSpec(doc, &ExpandOptions{RelativeBase: basePath}))

	expanded, err := json.Marshal(doc)
	require.NoError(t, err)

	return string(expanded)
}
