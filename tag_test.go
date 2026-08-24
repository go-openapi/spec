// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestJSONLookupTag(t *testing.T) {
	tag := NewTag("pet", "everything about pets", &ExternalDocumentation{URL: "http://readthedocs.org/swagger"})
	tag.AddExtension("x-framework", "go-swagger")

	t.Run("lookup should find an extension", func(t *testing.T) {
		res, err := tag.JSONLookup("x-framework")
		require.NoError(t, err)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, "go-swagger", *ext)
	})

	t.Run(`lookup should find "name"`, func(t *testing.T) {
		res, err := tag.JSONLookup("name")
		require.NoError(t, err)

		name, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "pet", name)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := tag.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}
