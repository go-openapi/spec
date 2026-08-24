// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var pathItem = PathItem{ //nolint:gochecknoglobals // test fixture
	Refable: Refable{Ref: MustCreateRef("Dog")},
	VendorExtensible: VendorExtensible{
		Extensions: map[string]any{
			"x-framework": "go-swagger",
		},
	},
	PathItemProps: PathItemProps{
		Get: &Operation{
			OperationProps: OperationProps{Description: "get operation description"},
		},
		Put: &Operation{
			OperationProps: OperationProps{Description: "put operation description"},
		},
		Post: &Operation{
			OperationProps: OperationProps{Description: "post operation description"},
		},
		Delete: &Operation{
			OperationProps: OperationProps{Description: "delete operation description"},
		},
		Options: &Operation{
			OperationProps: OperationProps{Description: "options operation description"},
		},
		Head: &Operation{
			OperationProps: OperationProps{Description: "head operation description"},
		},
		Patch: &Operation{
			OperationProps: OperationProps{Description: "patch operation description"},
		},
		Parameters: []Parameter{
			{
				ParamProps: ParamProps{In: "path"},
			},
		},
	},
}

const pathItemJSON = `{
	"$ref": "Dog",
	"x-framework": "go-swagger",
	"get": { "description": "get operation description" },
	"put": { "description": "put operation description" },
	"post": { "description": "post operation description" },
	"delete": { "description": "delete operation description" },
	"options": { "description": "options operation description" },
	"head": { "description": "head operation description" },
	"patch": { "description": "patch operation description" },
	"parameters": [{"in":"path"}]
}`

func TestIntegrationPathItem(t *testing.T) {
	assert.JSONUnmarshalAsT(t, pathItem, pathItemJSON)
}

func TestJSONLookupPathItem(t *testing.T) {
	t.Run("lookup should find an extension", func(t *testing.T) {
		res, err := pathItem.JSONLookup("x-framework")
		require.NoError(t, err)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, "go-swagger", *ext)
	})

	t.Run(`lookup should find "$ref"`, func(t *testing.T) {
		res, err := pathItem.JSONLookup("$ref")
		require.NoError(t, err)

		ref, ok := res.(*Ref)
		require.TrueT(t, ok)
		assert.Equal(t, MustCreateRef("Dog"), *ref)
	})

	t.Run(`lookup should find "get"`, func(t *testing.T) {
		res, err := pathItem.JSONLookup("get")
		require.NoError(t, err)

		op, ok := res.(*Operation)
		require.TrueT(t, ok)
		assert.EqualT(t, "get operation description", op.Description)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := pathItem.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}
