// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// TestNormalizer_Canonicalization pins the rules the normalizer and jsonreference both apply,
// on the edge cases where they used to part company.
//
// Two canonicalizers meet on every $ref. The normalizer builds a URI out of a $ref and a base;
// jsonreference re-renders it whenever the string becomes a Ref, which happens as soon as the
// document is unmarshalled. When the two disagreed, one document could be cached under two
// spellings of itself, and a $ref written back by denormalizeRef did not normalize to what it
// came from. normalizeURI and normalizeBase now end in canonicalString, so both sides answer
// with jsonreference's spelling.
func TestNormalizer_Canonicalization(t *testing.T) {
	t.Parallel()

	const base = "https://example.com/base/spec.json"

	tests := []struct {
		name     string
		rule     string
		refPath  string
		expected string
	}{
		{
			name:     "host case",
			rule:     "the host is lower-cased, the path is not",
			refPath:  "https://EXAMPLE.com/OTHER.json",
			expected: "https://example.com/OTHER.json",
		},
		{
			name:     "default port",
			rule:     ":443 comes off an https URL",
			refPath:  "https://example.com:443/other.json",
			expected: "https://example.com/other.json",
		},
		{
			name:     "non-default port",
			rule:     "any other port stays",
			refPath:  "https://example.com:8443/other.json",
			expected: "https://example.com:8443/other.json",
		},
		{
			name:     "IPv6 literal",
			rule:     "a bracketed host is lower-cased and loses its default port",
			refPath:  "https://[2001:DB8::1]:443/other.json",
			expected: "https://[2001:db8::1]/other.json",
		},
		{ //nolint:gosec // test URLs carrying userinfo, not credentials
			name:     "userinfo",
			rule:     "the colon in userinfo is not a port",
			refPath:  "https://user:pw@EXAMPLE.com:443/other.json",
			expected: "https://user:pw@example.com/other.json",
		},
		// Since go1.26, url.Parse rejects a colon outside a bracketed IPv6 host on an http or https URL
		// (GODEBUG urlstrictcolons=1, the default from a go.mod declaring go 1.26 or later). Both $refs below
		// used to parse - the first as the host ":a" on port 443, the second as "0:443" on port 443 - and both
		// now fail. normalizeURI logs a warning, repairs the $ref to the empty URI and resolves it against the
		// base, so the base itself comes back.
		{
			name:     "degenerate authority, stray colon",
			rule:     "a colon in the host is rejected, and an unresolvable $ref falls back to the base",
			refPath:  "https://:a:443/other.json",
			expected: base,
		},
		{
			name:     "degenerate authority, port twice",
			rule:     "same for a host spelling a port of its own",
			refPath:  "https://0:443:443/other.json",
			expected: base,
		},
		{
			name:     "duplicate slashes, relative",
			rule:     "a run of slashes collapses to one, however long",
			refPath:  "a//b///c////d.json",
			expected: "https://example.com/base/a/b/c/d.json",
		},
		{
			name:     "duplicate slashes, absolute",
			rule:     "same, on a $ref that carries its own host",
			refPath:  "https://example.com/a//b///c////d.json",
			expected: "https://example.com/a/b/c/d.json",
		},
		{
			name:     "unreserved character, escaped",
			rule:     "%41 is decoded: RFC 3986 §6.2.2.2 allows decoding an unreserved character",
			refPath:  "https://example.com/%41.json",
			expected: "https://example.com/A.json",
		},
		{
			name:     "reserved character, escaped",
			rule:     "%2F is decoded too, which changes what the URI denotes - see the note below",
			refPath:  "https://example.com/a%2Fb.json",
			expected: "https://example.com/a/b.json",
		},
		{
			name:     "fragment, unescaped",
			rule:     "a quote in a JSON pointer is escaped on the way out",
			refPath:  "other.json#/definitions/it's",
			expected: "https://example.com/base/other.json#/definitions/it%27s",
		},
		{
			name:     "fragment, escaped",
			rule:     "and the escaped spelling is left as it is, so the two agree",
			refPath:  "other.json#/definitions/it%27s",
			expected: "https://example.com/base/other.json#/definitions/it%27s",
		},
		{
			name:     "fragment, brackets",
			rule:     "same for a definition named o[k]",
			refPath:  "other.json#/definitions/o[k]",
			expected: "https://example.com/base/other.json#/definitions/o%5Bk%5D",
		},
		{
			name:     "fragment, space",
			rule:     "and for one named \"a b\"",
			refPath:  "other.json#/definitions/a b",
			expected: "https://example.com/base/other.json#/definitions/a%20b",
		},
	}

	canonicalBase := normalizeBase(base)
	require.EqualT(t, base, canonicalBase, "the base of these cases is already canonical")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			normalized := normalizeURI(test.refPath, canonicalBase)
			assert.EqualTf(t, test.expected, normalized, "%s: %s", test.name, test.rule)

			// the rule that matters: what the normalizer returns is what jsonreference would
			// have made of it, so turning it into a Ref changes nothing
			ref := MustCreateRef(normalized)
			assert.EqualTf(t, normalized, ref.String(),
				"%q is not a fixpoint of jsonreference's canonicalization", normalized)

			assert.EqualTf(t, normalized, normalizeURI(normalized, canonicalBase),
				"normalizing %q again changed it", normalized)
		})
	}
}

// TestNormalizer_CanonicalBase pins the same rules on the base, which normalizeBase canonicalizes
// too, and covers the rendering defect that produced an unparseable base.
func TestNormalizer_CanonicalBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rule     string
		base     string
		expected string
	}{
		{
			name:     "host case and default port",
			rule:     "a base is canonicalized like a $ref, so two spellings key one document",
			base:     "https://EXAMPLE.com:443/base/spec.json",
			expected: "https://example.com/base/spec.json",
		},
		{
			name:     "duplicate slashes",
			rule:     "path.Clean already collapsed these; the rule is pinned here too",
			base:     "https://example.com/base//sub///spec.json",
			expected: "https://example.com/base/sub/spec.json",
		},
		{
			name:     "fragment dropped",
			rule:     "a fragment in the base names nothing",
			base:     "https://example.com/base/spec.json#/definitions/x",
			expected: "https://example.com/base/spec.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			canonical := normalizeBase(test.base)
			assert.EqualTf(t, test.expected, canonical, "%s: %s", test.name, test.rule)

			ref := MustCreateRef(canonical)
			assert.EqualTf(t, canonical, ref.String(),
				"%q is not a fixpoint of jsonreference's canonicalization", canonical)

			assert.EqualTf(t, canonical, normalizeBase(canonical),
				"normalizing %q again changed it", canonical)
		})
	}
}

// TestNormalizer_EscapedPathRendering covers the rendering defect that produced an unparseable
// base: normalizeBase("%2F") returned "file://%2F".
//
// url.Parse reads "%2F" as the path "/" spelled "%2F", and url.URL.String renders RawPath
// whenever it still decodes to Path - here a spelling with no leading slash, so what is left
// reads as an authority. forgetSpelling drops RawPath once Path has been rewritten.
//
// The expectation is written against normalizeBase("/") rather than a literal, because a
// relative base is anchored to the working directory: "file:///" on unix, "file:///c:" or
// whichever drive the tests run from on windows.
func TestNormalizer_EscapedPathRendering(t *testing.T) {
	t.Parallel()

	const escapedSlash = "%2F"

	canonical := normalizeBase(escapedSlash)
	assert.EqualT(t, normalizeBase("/"), canonical, "the escaped spelling of the root must normalize like the plain one")
	assert.NotEqualT(t, "file://"+escapedSlash, canonical, "the escaped path was rendered as an authority")

	_, err := parseURL(canonical)
	require.NoErrorf(t, err, "normalizing %q yielded %q, which does not parse", escapedSlash, canonical)

	ref := MustCreateRef(canonical)
	assert.EqualTf(t, canonical, ref.String(),
		"%q is not a fixpoint of jsonreference's canonicalization", canonical)
	assert.EqualTf(t, canonical, normalizeBase(canonical), "normalizing %q again changed it", canonical)
}

// TestNormalizer_EscapedSlashIsDecoded records a rule that is not ours to change here.
//
// jsonreference clears url.URL.RawPath, so "%2F" comes back as a separator. RFC 3986 §6.2.2.2
// allows decoding an unreserved character and forbids decoding a reserved one, so this is a
// change of meaning rather than of spelling: the GitLab API requires the project path
// percent-encoded, and the decoded URL names a different endpoint.
//
// It is pinned rather than fixed because a $ref loses the escape anyway - the document model
// turns every $ref into a Ref, and jsonreference decodes it there. Making the normalizer keep
// the escape would only hide that. The fix belongs upstream.
func TestNormalizer_EscapedSlashIsDecoded(t *testing.T) {
	t.Parallel()

	const gitlab = "https://gitlab.com/api/v4/projects/mygroup%2Fmyproject/repository/files/swagger.json/raw"
	const decoded = "https://gitlab.com/api/v4/projects/mygroup/myproject/repository/files/swagger.json/raw"

	var ref Ref
	require.NoError(t, ref.UnmarshalJSON([]byte(`{"$ref":"`+gitlab+`"}`)))
	assert.EqualT(t, decoded, ref.String(), "the document model decodes the escape on its own")

	assert.EqualT(t, decoded, normalizeBase(gitlab))
	assert.EqualT(t, decoded, normalizeURI(gitlab, normalizeBase("https://example.com/spec.json")))
}
