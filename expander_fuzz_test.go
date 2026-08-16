// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

// remoteFuzzStub is what every remote $ref resolves to while fuzzing.
//
// It answers any path, so the loader never touches the filesystem or the network, and it holds a
// $ref back out to another document: since every path yields this same stub, that $ref chains
// forever unless cycle detection and the expansion budget stop it.
const remoteFuzzStub = `{"definitions":{` +
	`"remote":{"type":"string"},` +
	`"cycle":{"$ref":"other.json#/definitions/cycle"},` +
	`"deep":{"properties":{"a":{"$ref":"#/definitions/remote"},"b":{"$ref":"other.json#/definitions/deep"}}}` +
	`}}`

// FuzzExpandSpec expands arbitrary documents.
//
// Expansion is the part of this package that walks attacker-shaped input recursively, follows
// $refs across documents and rewrites the tree as it goes. It is also where a hostile document
// costs the most: see ErrExpandTooManyNodes and the confinement notes on ExpandOptions.
//
// The properties are that expansion terminates, and that what it produces is still a document -
// an expansion that succeeds but leaves behind something we can no longer read is a silent
// corruption of the caller's spec.
//
// The document loader is stubbed, so nothing here reads a file or opens a socket.
func FuzzExpandSpec(f *testing.F) {
	for _, seed := range expanderSeeds() {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var doc Swagger
		if err := json.Unmarshal(data, &doc); err != nil {
			return // not a document we claim to accept
		}

		opts := &ExpandOptions{
			RelativeBase: "file:///fuzz/spec.json",
			PathLoader: func(string) (json.RawMessage, error) {
				return json.RawMessage(remoteFuzzStub), nil
			},
			MaxExpansionNodes: 500,
		}

		if err := ExpandSpec(&doc, opts); err != nil {
			return // refusing a document is a legitimate outcome
		}

		expanded, err := json.Marshal(doc)
		require.NoErrorf(t, err, "an expanded document did not marshal back: %s", data)

		var reread Swagger
		require.NoErrorf(t, json.Unmarshal(expanded, &reread),
			"an expanded document no longer parses: %s", expanded)
	})
}

// expanderSeeds are the documents the fuzzer starts from.
//
// They aim at the shapes expansion has to walk: $refs in every position that accepts one,
// $refs into another document, and the cycles - direct, mutual and through a remote document -
// that expansion has to notice rather than follow.
func expanderSeeds() []string {
	return []string{
		`{"definitions":{"A":{"$ref":"#/definitions/B"},"B":{"type":"string"}}}`,
		`{"definitions":{"A":{"$ref":"#/definitions/A"}}}`,
		`{"definitions":{"A":{"$ref":"#/definitions/B"},"B":{"$ref":"#/definitions/A"}}}`,
		`{"definitions":{"A":{"$ref":"other.json#/definitions/remote"}}}`,
		`{"definitions":{"A":{"$ref":"other.json#/definitions/cycle"}}}`,
		`{"definitions":{"A":{"properties":{"a":{"$ref":"#/definitions/A"}},"items":{"$ref":"#/definitions/A"}}}}`,
		`{"definitions":{"A":{"allOf":[{"$ref":"#/definitions/A"}],"additionalProperties":{"$ref":"#/definitions/A"}}}}`,
		`{"paths":{"/x":{"$ref":"other.json#/definitions/deep"}}}`,
		`{"paths":{"/x":{"get":{"parameters":[{"$ref":"#/parameters/p"}],` +
			`"responses":{"200":{"$ref":"#/responses/r"}}}}},` +
			`"parameters":{"p":{"name":"p","in":"body","schema":{"$ref":"#/definitions/A"}}},` +
			`"responses":{"r":{"description":"d","schema":{"$ref":"other.json#/definitions/remote"}}},` +
			`"definitions":{"A":{"type":"object"}}}`,
		`{"id":"https://example.com/root","definitions":{"A":{"id":"sub","$ref":"#/definitions/A"}}}`,
	}
}
