// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

// remoteFuzzStub is what every remote $ref resolves to while fuzzing.
//
// It answers any path, so the loader never touches the filesystem or the network, and it holds a
// $ref back out to another document: since every path yields this same stub, that $ref chains
// forever unless cycle detection and the expansion budget stop it.
//
// "unmapped" carries a $ref under propertyNames, a keyword this model does not map: inlining
// that subtree must rebase the $ref on the stub, not copy it into the root.
// fuzzBase is where the document being expanded is taken to live.
const fuzzBase = "file:///fuzz/spec.json"

const remoteFuzzStub = `{"definitions":{` +
	`"remote":{"type":"string"},` +
	`"cycle":{"$ref":"other.json#/definitions/cycle"},` +
	`"deep":{"properties":{"a":{"$ref":"#/definitions/remote"},"b":{"$ref":"other.json#/definitions/deep"}}},` +
	`"unmapped":{"propertyNames":{"$ref":"#/definitions/remote"}}` +
	`}}`

// FuzzExpandSpec expands arbitrary documents.
//
// Expansion is the part of this package that walks attacker-shaped input recursively, follows
// $refs across documents and rewrites the tree as it goes. It is also where a hostile document
// costs the most: see ErrExpandTooManyNodes and the confinement notes on ExpandOptions.
//
// The properties are that expansion terminates, that what it produces is still a document -
// an expansion that succeeds but leaves behind something we can no longer read is a silent
// corruption of the caller's spec - and that the local pointers it leaves behind still name
// something. That last one is what found the ExtraProps defect fixed by rebaseExtraRefs.
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

		// what the input claims before anything is expanded: an input that already
		// dangles cannot say whether expansion broke a pointer
		original, err := json.Marshal(doc)
		if err != nil {
			return
		}
		soundBefore := len(unresolvedLocalRefs(&doc, string(original))) == 0

		opts := &ExpandOptions{
			RelativeBase: fuzzBase,
			PathLoader: func(path string) (json.RawMessage, error) {
				if normalizeBase(path) == normalizeBase(fuzzBase) {
					// a $ref spelled "." names the document being expanded: a real loader
					// hands back the bytes it was read from, not another document
					return original, nil
				}

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

		if !soundBefore {
			return // a document whose own pointers were already dangling proves nothing
		}

		require.Emptyf(t, unresolvedLocalRefs(&reread, string(expanded)),
			"expansion left a local $ref pointing at nothing.\ninput: %s\nexpanded: %s", data, expanded)
	})
}

// unresolvedLocalRefs returns the local $ref of jazon that name nothing in doc.
//
// Only "#/..." pointers are looked at: a $ref into another document is the loader's business,
// and the loader is a stub here. The document is walked with the same jsonpointer call the
// resolver makes, so "has no key" here is "has no key" there.
func unresolvedLocalRefs(doc any, jazon string) []string {
	var unresolved []string
	seen := make(map[string]struct{})

	for _, matched := range rex.FindAllStringSubmatch(jazon, -1) {
		refString := matched[1]
		if !strings.HasPrefix(refString, "#/") {
			continue
		}
		if _, already := seen[refString]; already {
			continue
		}
		seen[refString] = struct{}{}

		ref, err := NewRef(refString)
		if err != nil {
			continue // not a reference this package claims to understand
		}

		if _, _, err := ref.GetPointer().Get(doc); err != nil {
			unresolved = append(unresolved, refString)
		}
	}

	sort.Strings(unresolved)

	return unresolved
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
		`{"definitions":{"A":{"$ref":"other.json#/definitions/unmapped"}}}`,
		`{"paths":{"/x":{"$ref":"other.json#/definitions/deep"}}}`,
		`{"paths":{"/x":{"get":{"parameters":[{"$ref":"#/parameters/p"}],` +
			`"responses":{"200":{"$ref":"#/responses/r"}}}}},` +
			`"parameters":{"p":{"name":"p","in":"body","schema":{"$ref":"#/definitions/A"}}},` +
			`"responses":{"r":{"description":"d","schema":{"$ref":"other.json#/definitions/remote"}}},` +
			`"definitions":{"A":{"type":"object"}}}`,
		`{"id":"https://example.com/root","definitions":{"A":{"id":"sub","$ref":"#/definitions/A"}}}`,
	}
}
