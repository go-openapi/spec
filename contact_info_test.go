// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

const contactInfoJSON = `{
	"name": "wordnik api team",
	"url": "http://developer.wordnik.com",
	"email": "some@mailayada.dkdkd",
	"x-teams": "test team"
}`

var contactInfo = ContactInfo{ContactInfoProps: ContactInfoProps{ //nolint:gochecknoglobals // test fixture
	Name:  "wordnik api team",
	URL:   "http://developer.wordnik.com",
	Email: "some@mailayada.dkdkd",
}, VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-teams": "test team"}}}

func TestIntegrationContactInfo(t *testing.T) {
	assert.JSONMarshalAsT(t, contactInfoJSON, contactInfo)
	assert.JSONUnmarshalAsT(t, contactInfo, contactInfoJSON)
}
