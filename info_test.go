// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

const infoJSON = `{
	"description": "A sample API that uses a petstore as an example to demonstrate features in ` +
	`the swagger-2.0 specification",
	"title": "Swagger Sample API",
	"termsOfService": "http://helloreverb.com/terms/",
	"contact": {
		"name": "wordnik api team",
		"url": "http://developer.wordnik.com"
	},
	"license": {
		"name": "Creative Commons 4.0 International",
		"url": "http://creativecommons.org/licenses/by/4.0/"
	},
	"version": "1.0.9-abcd",
	"x-framework": "go-swagger"
}`

var testInfo = Info{ //nolint:gochecknoglobals // test fixture
	InfoProps: InfoProps{
		Version: "1.0.9-abcd",
		Title:   "Swagger Sample API",
		Description: "A sample API that uses a petstore as an example to demonstrate features in " +
			"the swagger-2.0 specification",
		TermsOfService: "http://helloreverb.com/terms/",
		Contact:        &ContactInfo{ContactInfoProps: ContactInfoProps{Name: "wordnik api team", URL: "http://developer.wordnik.com"}},
		License: &License{
			LicenseProps: LicenseProps{
				Name: "Creative Commons 4.0 International",
				URL:  "http://creativecommons.org/licenses/by/4.0/",
			},
		},
	},
	VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
}

func TestInfo(t *testing.T) {
	t.Run("should marshal Info", func(t *testing.T) {
		assert.JSONMarshalAsT(t, infoJSON, testInfo)
	})

	t.Run("should unmarshal Info", func(t *testing.T) {
		assert.JSONUnmarshalAsT(t, testInfo, infoJSON)
	})

	t.Run("should GobEncode Info", func(t *testing.T) {
		var src, dst Info
		require.NoError(t, json.Unmarshal([]byte(infoJSON), &src))
		assert.Equal(t, src, testInfo)
		doTestAnyGobEncoding(t, &src, &dst)
	})
}

func TestExtensionsGetBool(t *testing.T) {
	e := Extensions{
		"x-truthy": true,
		"x-falsy":  false,
		"x-string": "not a bool",
	}

	t.Run("should read a bool extension, ignoring key case", func(t *testing.T) {
		v, ok := e.GetBool("X-Truthy")
		require.TrueT(t, ok)
		assert.TrueT(t, v)

		v, ok = e.GetBool("x-falsy")
		require.TrueT(t, ok)
		assert.FalseT(t, v)
	})

	t.Run("should not find a non-bool extension", func(t *testing.T) {
		v, ok := e.GetBool("x-string")
		assert.FalseT(t, ok)
		assert.FalseT(t, v)
	})

	t.Run("should not find an absent extension", func(t *testing.T) {
		v, ok := e.GetBool("x-absent")
		assert.FalseT(t, ok)
		assert.FalseT(t, v)
	})
}

func TestJSONLookupInfo(t *testing.T) {
	t.Run("lookup should find an extension", func(t *testing.T) {
		res, err := testInfo.JSONLookup("x-framework")
		require.NoError(t, err)

		ext, ok := res.(*any)
		require.TrueT(t, ok)
		assert.Equal(t, "go-swagger", *ext)
	})

	t.Run(`lookup should find "title"`, func(t *testing.T) {
		res, err := testInfo.JSONLookup("title")
		require.NoError(t, err)

		title, ok := res.(string)
		require.TrueT(t, ok)
		assert.EqualT(t, "Swagger Sample API", title)
	})

	t.Run(`lookup should fail on "unknown"`, func(t *testing.T) {
		res, err := testInfo.JSONLookup("unknown")
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestExtensionsGetString(t *testing.T) {
	e := Extensions{"x-name": "go-swagger", "x-count": 3}

	v, ok := e.GetString("X-Name")
	require.TrueT(t, ok)
	assert.EqualT(t, "go-swagger", v)

	v, ok = e.GetString("x-count")
	assert.FalseT(t, ok)
	assert.EqualT(t, "", v)

	v, ok = e.GetString("x-absent")
	assert.FalseT(t, ok)
	assert.EqualT(t, "", v)
}

func TestExtensionsGetInt(t *testing.T) {
	e := Extensions{
		"x-from-string":  "12",
		"x-bad-string":   "not a number",
		"x-from-float":   float64(12),
		"x-not-a-number": true,
	}

	t.Run("should parse an int held as a string", func(t *testing.T) {
		v, ok := e.GetInt("X-From-String")
		require.TrueT(t, ok)
		assert.EqualT(t, 12, v)
	})

	t.Run("should read an int held as a float64", func(t *testing.T) {
		v, ok := e.GetInt("x-from-float")
		require.TrueT(t, ok)
		assert.EqualT(t, 12, v)
	})

	t.Run("should not parse a string that is not a number", func(t *testing.T) {
		v, ok := e.GetInt("x-bad-string")
		assert.FalseT(t, ok)
		assert.EqualT(t, -1, v)
	})

	t.Run("should not read a value that is neither string nor float64", func(t *testing.T) {
		v, ok := e.GetInt("x-not-a-number")
		assert.FalseT(t, ok)
		assert.EqualT(t, -1, v)
	})

	t.Run("should not find an absent extension", func(t *testing.T) {
		v, ok := e.GetInt("x-absent")
		assert.FalseT(t, ok)
		assert.EqualT(t, -1, v)
	})
}
