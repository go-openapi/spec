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
