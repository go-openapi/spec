// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

var paths = Paths{ //nolint:gochecknoglobals // test fixture
	VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
	Paths: map[string]PathItem{
		"/": {
			Refable: Refable{Ref: MustCreateRef("cats")},
		},
	},
}

const pathsJSON = `{"x-framework":"go-swagger","/":{"$ref":"cats"}}`

func TestIntegrationPaths(t *testing.T) {
	assert.JSONUnmarshalAsT(t, paths, pathsJSON)
}
