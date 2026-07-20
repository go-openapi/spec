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
