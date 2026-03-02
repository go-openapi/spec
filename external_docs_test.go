// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	_ "github.com/go-openapi/testify/enable/yaml/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestIntegrationExternalDocs(t *testing.T) {
	extDocs := ExternalDocumentation{Description: "the name", URL: "the url"}
	const extDocsYAML = "description: the name\nurl: the url\n"
	const extDocsJSON = `{"description":"the name","url":"the url"}`
	assert.JSONMarshalAsT(t, extDocsJSON, extDocs)
	assert.YAMLMarshalAsT(t, extDocsYAML, extDocs)
	assert.JSONUnmarshalAsT(t, extDocs, extDocsJSON)
	assert.YAMLUnmarshalAsT(t, extDocs, extDocsYAML)
}
