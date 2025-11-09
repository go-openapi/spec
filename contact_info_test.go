// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

const contactInfoJSON = `{
	"name": "wordnik api team",
	"url": "http://developer.wordnik.com",
	"email": "some@mailayada.dkdkd",
	"x-teams": "test team"
}`

var contactInfo = ContactInfo{ContactInfoProps: ContactInfoProps{
	Name:  "wordnik api team",
	URL:   "http://developer.wordnik.com",
	Email: "some@mailayada.dkdkd",
}, VendorExtensible: VendorExtensible{Extensions: map[string]any{"x-teams": "test team"}}}

func TestIntegrationContactInfo(t *testing.T) {
	b, err := json.MarshalIndent(contactInfo, "", "\t")
	require.NoError(t, err)
	assert.JSONEq(t, contactInfoJSON, string(b))

	actual := ContactInfo{}
	err = json.Unmarshal([]byte(contactInfoJSON), &actual)
	require.NoError(t, err)
	assert.Equal(t, contactInfo, actual)
}
