// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationLicense(t *testing.T) {
	const licenseJSON = `{
	"name": "the name",
	"url": "the url",
	"x-license": "custom term"
}`

	var testLicense = License{
		LicenseProps:     LicenseProps{Name: "the name", URL: "the url"},
		VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-license": "custom term"}}}

	// const licenseYAML = "name: the name\nurl: the url\n"

	t.Run("should marshal license", func(t *testing.T) {
		b, err := json.MarshalIndent(testLicense, "", "\t")
		require.NoError(t, err)
		assert.JSONEq(t, licenseJSON, string(b))
	})

	t.Run("should unmarshal empty license", func(t *testing.T) {
		actual := License{}
		err := json.Unmarshal([]byte(licenseJSON), &actual)
		require.NoError(t, err)
		assert.Equal(t, testLicense, actual)
	})
}
