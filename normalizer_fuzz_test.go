// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

// FuzzNormalizer explores the $ref normalizer with arbitrary ($ref, base) pairs.
//
// The normalizer is the only place in this package where MustCreateRef is handed a URI
// that we built rather than one the caller wrote: a $ref that survives url.Parse but not
// the round-trip through url.URL.String() would panic while merely reading a document.
//
// Beyond that, the properties asserted here are the invariants the expander relies on:
// normalizing yields a canonical URI, normalizing twice changes nothing, and
// denormalizing a canonical $ref yields a shorthand that normalizes back to it.
func FuzzNormalizer(f *testing.F) {
	for _, seed := range normalizerSeeds() {
		f.Add(seed.refPath, seed.base)
	}

	f.Fuzz(func(t *testing.T, refPath, base string) {
		canonicalBase := normalizeBase(base)
		requireCanonicalURI(t, canonicalBase)
		require.EqualTf(t, canonicalBase, normalizeBase(canonicalBase),
			"normalizeBase is not idempotent on base %q", base)

		normalized := normalizeURI(refPath, canonicalBase)
		requireCanonicalURI(t, normalized)
		require.EqualTf(t, normalized, normalizeURI(normalized, canonicalBase),
			"normalizeURI is not idempotent on $ref %q against base %q", refPath, canonicalBase)

		if respelled(normalized) {
			return
		}

		ref := MustCreateRef(normalized)
		denormalized := denormalizeRef(&ref, canonicalBase, "")
		require.EqualTf(t, normalized, normalizeURI(denormalized.String(), canonicalBase),
			"denormalizing %q against base %q yielded %q, which does not normalize back",
			normalized, canonicalBase, denormalized.String())
	})
}

// respelled reports whether jsonreference renders a URI otherwise than the normalizer does.
//
// Turning a URI into a Ref lower-cases the host, drops a default port and re-escapes path and
// fragment in their canonical form, whereas the normalizer keeps the spelling it was handed.
// Where the two disagree, denormalizing cannot be asked to round-trip.
func respelled(in string) bool {
	ref := MustCreateRef(in)

	return ref.String() != in
}

// requireCanonicalURI asserts the postcondition shared by normalizeBase and normalizeURI:
// whatever comes out is a parseable, absolute URI, safe to use as a document cache key.
func requireCanonicalURI(t *testing.T, in string) {
	t.Helper()

	u, err := parseURL(in)
	require.NoErrorf(t, err, "the normalizer produced an unparseable URI: %q", in)
	require.NotEmptyf(t, u.Scheme, "the normalizer produced a URI with no scheme: %q", in)
}

// normalizerSeeds are the ($ref, base) pairs the fuzzer starts from.
//
// They cover the shapes the normalizer branches on: empty and fragment-only $refs,
// relative and absolute paths, the file and http schemes, query components, percent-escaping,
// the windows spellings (drive letters and UNC shares) and URIs that only repairURI can salvage.
func normalizerSeeds() []struct{ refPath, base string } {
	return []struct{ refPath, base string }{
		{"", ""},
		{"", "/base/path.json"},
		{"#", "file:///base/path.json"},
		{"#/definitions/Pet", "http://www.example.com/base/path/swagger.json"},
		{"#/dir/definitions/Pet", "https://www.example.com:8443/base/path.json?raw=true"},
		{"another/base/path.json#/definitions/Pet", "/base/path.json"},
		{"./another/base/path.json#/definitions/Pet", "file:///base/path.json"},
		{"../other/another.json#/definitions/X", "file:///base/path.json"},
		{"/another///base//path.json/", "/base/path.json"},
		{"http://www.anotherexample.com/swagger.json#/definitions/Pet", "http://www.example.com/base/path/swagger.json"},
		{"https://origin.com/another/file.json?raw=true", "https://www.example.com:8443/base/path.json?raw=true"},
		{"Resources [x86].yaml#/definitions/Pets", "https://example.com/base (x86)/Spec.json"},
		{"Resources.yaml#/definitions (x86)/Pets", "https://example.com/base/Spec.json"},
		{"\x7f\x9a", "file:///base/path.json"},
		{"gs://bucket/folder/file.json", "smb://host"},
		{`C:\another\Base\path.json#/definitions/Pet`, `file:///c:/base/path.json`},
		{`another\base\path.json#/definitions/Pet`, `file:///c:/base/path.json`},
		{`\\host\share\another\base\path.json#/definitions/Pet`, `\\host\share@1234\Folder\File.json`},
		{`file://E:\Base\sub\File.json`, `file:///.\Base\sub\File.json`},
		{`file:/..`, `file:///.`},
	}
}
