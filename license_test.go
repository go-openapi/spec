// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestIntegrationLicense(t *testing.T) {
	const licenseJSON = `{
	"name": "the name",
	"url": "the url",
	"x-license": "custom term"
}`

	testLicense := License{
		LicenseProps:     LicenseProps{Name: "the name", URL: "the url"},
		VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-license": "custom term"}},
	}

	// const licenseYAML = "name: the name\nurl: the url\n"

	t.Run("should marshal license", func(t *testing.T) {
		assert.JSONMarshalAsT(t, licenseJSON, testLicense)
	})

	t.Run("should unmarshal empty license", func(t *testing.T) {
		assert.JSONUnmarshalAsT(t, testLicense, licenseJSON)
	})
}
