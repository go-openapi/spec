// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
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

func TestJSONLookupPaths(t *testing.T) {
	p := Paths{
		VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
		Paths: map[string]PathItem{
			"/pets": {PathItemProps: PathItemProps{
				Get: &Operation{OperationProps: OperationProps{ID: "listPets"}},
			}},
		},
	}

	t.Run("lookup should find a path", func(t *testing.T) {
		res, err := p.JSONLookup("/pets")
		require.NoError(t, err)

		pi, ok := res.(*PathItem)
		require.TrueT(t, ok)
		require.NotNil(t, pi.Get)
		assert.EqualT(t, "listPets", pi.Get.ID)
	})

	t.Run("lookup should find an extension", func(t *testing.T) {
		res, err := p.JSONLookup("x-framework")
		require.NoError(t, err)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, "go-swagger", *ext)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := p.JSONLookup("/unknown")
		require.ErrorIs(t, err, ErrSpec)
		require.Nil(t, res)
	})
}
