// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestSerialization_AuthSerialization(t *testing.T) {
	assert.JSONMarshalAsT(t, `{"type":"basic"}`, BasicAuth())

	assert.JSONMarshalAsT(t, `{"type":"apiKey","name":"api-key","in":"header"}`, APIKeyAuth("api-key", "header"))

	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"implicit","authorizationUrl":"http://foo.com/authorization"}`,
		OAuth2Implicit("http://foo.com/authorization"))

	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"password","tokenUrl":"http://foo.com/token"}`,
		OAuth2Password("http://foo.com/token"))

	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"application","tokenUrl":"http://foo.com/token"}`,
		OAuth2Application("http://foo.com/token"))

	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"accessCode","authorizationUrl":"http://foo.com/authorization",`+
			`"tokenUrl":"http://foo.com/token"}`,
		OAuth2AccessToken("http://foo.com/authorization", "http://foo.com/token"))

	auth1 := OAuth2Implicit("http://foo.com/authorization")
	auth1.AddScope("email", "read your email")
	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"implicit","authorizationUrl":"http://foo.com/authorization",`+
			`"scopes":{"email":"read your email"}}`,
		auth1)

	auth2 := OAuth2Password("http://foo.com/authorization")
	auth2.AddScope("email", "read your email")
	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"password","tokenUrl":"http://foo.com/authorization",`+
			`"scopes":{"email":"read your email"}}`,
		auth2)

	auth3 := OAuth2Application("http://foo.com/token")
	auth3.AddScope("email", "read your email")
	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"application","tokenUrl":"http://foo.com/token","scopes":{"email":"read your email"}}`,
		auth3)

	auth4 := OAuth2AccessToken("http://foo.com/authorization", "http://foo.com/token")
	auth4.AddScope("email", "read your email")
	assert.JSONMarshalAsT(t,
		`{"type":"oauth2","flow":"accessCode","authorizationUrl":"http://foo.com/authorization",`+
			`"tokenUrl":"http://foo.com/token","scopes":{"email":"read your email"}}`,
		auth4)
}

func TestSerialization_AuthDeserialization(t *testing.T) {
	assert.JSONUnmarshalAsT(t, BasicAuth(), `{"type":"basic"}`)

	assert.JSONUnmarshalAsT(t,
		APIKeyAuth("api-key", "header"),
		`{"in":"header","name":"api-key","type":"apiKey"}`)

	assert.JSONUnmarshalAsT(t,
		OAuth2Implicit("http://foo.com/authorization"),
		`{"authorizationUrl":"http://foo.com/authorization","flow":"implicit","type":"oauth2"}`)

	assert.JSONUnmarshalAsT(t,
		OAuth2Password("http://foo.com/token"),
		`{"flow":"password","tokenUrl":"http://foo.com/token","type":"oauth2"}`)

	assert.JSONUnmarshalAsT(t,
		OAuth2Application("http://foo.com/token"),
		`{"flow":"application","tokenUrl":"http://foo.com/token","type":"oauth2"}`)

	assert.JSONUnmarshalAsT(t,
		OAuth2AccessToken("http://foo.com/authorization", "http://foo.com/token"),
		`{"authorizationUrl":"http://foo.com/authorization","flow":"accessCode","tokenUrl":"http://foo.com/token",`+
			`"type":"oauth2"}`)

	auth1 := OAuth2Implicit("http://foo.com/authorization")
	auth1.AddScope("email", "read your email")
	assert.JSONUnmarshalAsT(t, auth1,
		`{"authorizationUrl":"http://foo.com/authorization","flow":"implicit","scopes":{"email":"read your email"},`+
			`"type":"oauth2"}`)

	auth2 := OAuth2Password("http://foo.com/token")
	auth2.AddScope("email", "read your email")
	assert.JSONUnmarshalAsT(t, auth2,
		`{"flow":"password","scopes":{"email":"read your email"},"tokenUrl":"http://foo.com/token","type":"oauth2"}`)

	auth3 := OAuth2Application("http://foo.com/token")
	auth3.AddScope("email", "read your email")
	assert.JSONUnmarshalAsT(t, auth3,
		`{"flow":"application","scopes":{"email":"read your email"},"tokenUrl":"http://foo.com/token","type":"oauth2"}`)

	auth4 := OAuth2AccessToken("http://foo.com/authorization", "http://foo.com/token")
	auth4.AddScope("email", "read your email")
	assert.JSONUnmarshalAsT(t, auth4,
		`{"authorizationUrl":"http://foo.com/authorization","flow":"accessCode","scopes":{"email":"read your email"},`+
			`"tokenUrl":"http://foo.com/token","type":"oauth2"}`)
}
