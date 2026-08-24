// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

// pin pointing go-swagger/go-swagger#1816 issue with cloning ref's.
func TestCloneRef(t *testing.T) {
	var b bytes.Buffer
	src := MustCreateRef("#/definitions/test")
	require.NoError(t,
		gob.NewEncoder(&b).Encode(&src),
	)

	var dst Ref
	require.NoError(t,
		gob.NewDecoder(&b).Decode(&dst),
	)

	jazon, err := json.Marshal(dst)
	require.NoError(t, err)

	assert.JSONEqT(t, `{"$ref":"#/definitions/test"}`, string(jazon))
}

func TestRef_IsValidURI(t *testing.T) {
	t.Run("empty and fragment-only refs are valid", func(t *testing.T) {
		empty := MustCreateRef("")
		assert.TrueT(t, empty.IsValidURI())

		frag := MustCreateRef("#/definitions/Foo")
		assert.TrueT(t, frag.IsValidURI())
	})

	t.Run("absolute URLs are valid without any network request", func(t *testing.T) {
		// A well-formed absolute URL is a valid URI. IsValidURI must NOT reach out to the network:
		// no timeout to tune, no goroutine to leak, no SSRF against internal addresses.
		// 192.0.2.0/24 is TEST-NET-1 (RFC 5737): guaranteed non-routable, so a real GET would
		// stall on connect. We assert the call returns true promptly to guard against a
		// reintroduced network probe.
		for _, uri := range []string{
			"http://192.0.2.1/schema.json", // unreachable public address
			"http://127.0.0.1:1/internal",  // internal address (SSRF target)
			"https://example.com/openapi.json",
		} {
			ref := MustCreateRef(uri)
			require.TrueT(t, ref.HasFullURL)

			done := make(chan bool, 1)
			go func() { done <- ref.IsValidURI() }()

			select {
			case ok := <-done:
				assert.TrueT(t, ok, "expected %q to be a valid URI", uri)
			case <-time.After(5 * time.Second):
				t.Fatalf("IsValidURI(%q) blocked: it must not perform a network request", uri)
			}
		}
	})

	t.Run("local file references are checked on disk", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "schema.json"), []byte(`{}`), 0o600))
		basePath := filepath.Join(dir, "root.json") // mirrors validate's IsValidURI(specFilePath)

		exists := MustCreateRef("schema.json")
		assert.TrueT(t, exists.IsValidURI(basePath),
			"an existing local file should be a valid URI")

		missing := MustCreateRef("does-not-exist.json")
		assert.FalseT(t, missing.IsValidURI(basePath),
			"a missing local file should be an invalid URI")

		asDir := MustCreateRef(".")
		assert.FalseT(t, asDir.IsValidURI(basePath),
			"a directory should not be a valid file URI")
	})
}

func TestRefInherits(t *testing.T) {
	t.Run("a child ref should inherit its parent's base", func(t *testing.T) {
		parent := MustCreateRef("http://www.example.com/foo/bar/schema.json")
		child := MustCreateRef("other.json#/definitions/Cat")

		inherited, err := parent.Inherits(child)
		require.NoError(t, err)
		require.NotNil(t, inherited)
		assert.EqualT(t, "http://www.example.com/foo/bar/other.json#/definitions/Cat", inherited.String())
	})

	t.Run("an absolute child ref should override its parent", func(t *testing.T) {
		parent := MustCreateRef("http://www.example.com/foo/bar/schema.json")
		child := MustCreateRef("http://www.other.com/other.json")

		inherited, err := parent.Inherits(child)
		require.NoError(t, err)
		assert.EqualT(t, "http://www.other.com/other.json", inherited.String())
	})

	t.Run("a child with no URL should not resolve", func(t *testing.T) {
		parent := MustCreateRef("http://www.example.com/foo/bar/schema.json")
		var child Ref // the zero Ref carries no URL

		inherited, err := parent.Inherits(child)
		require.Error(t, err)
		assert.Nil(t, inherited)
	})
}

func TestNewRefError(t *testing.T) {
	r, err := NewRef("http://[")
	require.Error(t, err)
	assert.Equal(t, Ref{}, r)
}

func TestRefRemoteURI(t *testing.T) {
	t.Run("an empty ref has no remote URI", func(t *testing.T) {
		var r Ref
		assert.EqualT(t, "", r.RemoteURI())
	})

	t.Run("the fragment is stripped from the remote URI", func(t *testing.T) {
		r := MustCreateRef("http://www.example.com/schema.json#/definitions/Cat")
		assert.EqualT(t, "http://www.example.com/schema.json", r.RemoteURI())
	})
}

func TestRefFromMap(t *testing.T) {
	t.Run("a nil map leaves the ref untouched", func(t *testing.T) {
		var r Ref
		require.NoError(t, r.fromMap(nil))
		assert.EqualT(t, "", r.String())
	})

	t.Run("a non-string $ref is ignored", func(t *testing.T) {
		var r Ref
		require.NoError(t, r.fromMap(map[string]any{"$ref": 12}))
		assert.EqualT(t, "", r.String())
	})

	t.Run("an invalid $ref is reported", func(t *testing.T) {
		var r Ref
		require.Error(t, r.fromMap(map[string]any{"$ref": "http://["}))
	})
}

func TestRefUnmarshalJSONError(t *testing.T) {
	var r Ref
	require.Error(t, json.Unmarshal([]byte(`not json`), &r))
}

func TestRefGobDecodeError(t *testing.T) {
	var r Ref
	require.Error(t, r.GobDecode([]byte("not gob")))
}
