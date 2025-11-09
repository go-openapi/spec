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

var testInfo = Info{
	InfoProps: InfoProps{
		Version: "1.0.9-abcd",
		Title:   "Swagger Sample API",
		Description: "A sample API that uses a petstore as an example to demonstrate features in " +
			"the swagger-2.0 specification",
		TermsOfService: "http://helloreverb.com/terms/",
		Contact:        &ContactInfo{ContactInfoProps: ContactInfoProps{Name: "wordnik api team", URL: "http://developer.wordnik.com"}},
		License: &License{LicenseProps: LicenseProps{
			Name: "Creative Commons 4.0 International",
			URL:  "http://creativecommons.org/licenses/by/4.0/",
		},
		},
	},
	VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-framework": "go-swagger"}},
}

func TestInfo(t *testing.T) {
	t.Run("should marshal Info", func(t *testing.T) {
		b, err := json.MarshalIndent(testInfo, "", "\t")
		require.NoError(t, err)
		assert.JSONEq(t, infoJSON, string(b))
	})

	t.Run("should unmarshal Info", func(t *testing.T) {
		actual := Info{}
		require.NoError(t, json.Unmarshal([]byte(infoJSON), &actual))
		assert.Equal(t, testInfo, actual)
	})

	t.Run("should GobEncode Info", func(t *testing.T) {
		var src, dst Info
		require.NoError(t, json.Unmarshal([]byte(infoJSON), &src))
		assert.Equal(t, src, testInfo)
		doTestAnyGobEncoding(t, &src, &dst)
	})
}
