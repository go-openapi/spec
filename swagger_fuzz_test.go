// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

// FuzzSwaggerRoundTrip feeds arbitrary JSON documents to the spec model.
//
// A document is the first thing this package sees of an untrusted input, and it lands on some
// two dozen hand-written UnmarshalJSON methods - several of them for union types that switch on
// the shape of the value rather than on a discriminator. This target covers them all at once.
//
// The property is that a document we accepted can be written back and read again unchanged:
// marshalling after unmarshalling is where a union type that guessed wrong shows up.
func FuzzSwaggerRoundTrip(f *testing.F) {
	for _, seed := range swaggerSeeds() {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var doc Swagger
		if err := json.Unmarshal(data, &doc); err != nil {
			return // not a document we claim to accept
		}

		once, err := json.Marshal(doc)
		require.NoErrorf(t, err, "an accepted document did not marshal back: %s", data)

		var reread Swagger
		require.NoErrorf(t, json.Unmarshal(once, &reread),
			"a marshalled document no longer parses: %s", once)

		twice, err := json.Marshal(reread)
		require.NoErrorf(t, err, "a re-read document did not marshal back: %s", once)
		require.EqualTf(t, string(once), string(twice),
			"reading a document back changed it, from %s", data)
	})
}

// swaggerSeeds are the documents the fuzzer starts from.
//
// They aim at the union types rather than at coverage of the specification: the members that
// accept either an object or a boolean, either one value or an array, either a $ref or a body.
func swaggerSeeds() []string {
	return []string{
		`{"swagger":"2.0","info":{"title":"t","version":"1"},"paths":{}}`,
		`{"paths":{"/x":{"get":{"responses":{"200":{"description":"ok","schema":{"$ref":"#/definitions/A"}}}}}},` +
			`"definitions":{"A":{"type":"object","properties":{"a":{"type":"string"}}}}}`,
		`{"paths":{"/x":{"$ref":"a.json#/paths/~1y"},"/y":{"parameters":[{"$ref":"#/parameters/p"}]}}}`,
		`{"definitions":{"A":{"additionalProperties":true,"items":[{"type":"string"}],"enum":[1,"a",null,{}]}}}`,
		`{"definitions":{"A":{"additionalProperties":{"type":"string"},"items":{"type":"integer"},` +
			`"type":["string","null"],"required":["a"]}}}`,
		`{"definitions":{"A":{"allOf":[{"$ref":"#/definitions/B"}],"not":{"type":"null"},"x-go-name":"A"}}}`,
		`{"parameters":{"p":{"name":"p","in":"query","type":"array","items":{"type":"string"},"collectionFormat":"csv"}}}`,
		`{"responses":{"r":{"description":"d","headers":{"h":{"type":"array","items":{"type":"number"}}},` +
			`"examples":{"application/json":{"a":1}}}}}`,
		`{"securityDefinitions":{"k":{"type":"apiKey","name":"n","in":"header"},` +
			`"o":{"type":"oauth2","flow":"implicit","authorizationUrl":"https://x/a","scopes":{"s":"d"}}}}`,
		`{"security":[{"k":[]},{"o":["s"]}],"tags":[{"name":"t","externalDocs":{"url":"https://x"}}],"x-ext":{"a":[1,2]}}`,
		`{"definitions":{"A":{"xml":{"name":"a","attribute":true,"wrapped":false},"default":null,"example":[]}}}`,
		`{"$schema":"http://json-schema.org/draft-04/schema#","id":"https://example.com/schema","definitions":{}}`,
	}
}
