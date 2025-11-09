// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var paths = Paths{
	VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
	Paths: map[string]PathItem{
		"/": {
			Refable: Refable{Ref: MustCreateRef("cats")},
		},
	},
}

const pathsJSON = `{"x-framework":"go-swagger","/":{"$ref":"cats"}}`

func TestIntegrationPaths(t *testing.T) {
	var actual Paths
	require.NoError(t, json.Unmarshal([]byte(pathsJSON), &actual))
	assert.Equal(t, actual, paths)

	assertParsesJSON(t, pathsJSON, paths)

}
