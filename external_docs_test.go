// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"
)

func TestIntegrationExternalDocs(t *testing.T) {
	var extDocs = ExternalDocumentation{Description: "the name", URL: "the url"}
	const extDocsYAML = "description: the name\nurl: the url\n"
	const extDocsJSON = `{"description":"the name","url":"the url"}`
	assertSerializeJSON(t, extDocs, extDocsJSON)
	assertSerializeYAML(t, extDocs, extDocsYAML)
	assertParsesJSON(t, extDocsJSON, extDocs)
	assertParsesYAML(t, extDocsYAML, extDocs)
}
