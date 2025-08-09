// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{"x-license": "custom term"}}}

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
