// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-openapi/swag/loading"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

var rex = regexp.MustCompile(`"\$ref":\s*"(.*?)"`)

// fixtureServer returns an httptest.Server serving the given subdirectory
// from the embedded fixtureAssets FS. This avoids OS-level file serving
// (and the Windows TransmitFile/sendfile code path that has a data race
// in Go 1.26).
func fixtureServer(t testing.TB, dir string) *httptest.Server {
	t.Helper()

	sub, err := fs.Sub(fixtureAssets, filepath.ToSlash(dir))
	require.NoError(t, err)

	server := httptest.NewServer(http.FileServerFS(sub))
	t.Cleanup(server.Close)

	return server
}

// rewritingFixtureServer serves a subdirectory of the embedded fixtureAssets FS,
// replacing every occurrence of placeholder with the address the server listens on.
//
// Fixtures that carry an absolute schema id cannot be served from a random port
// unless the id follows the port. The handler is wired before Start, so the URL
// it captures is the one the listener already holds.
func rewritingFixtureServer(t testing.TB, dir, placeholder string) *httptest.Server {
	t.Helper()

	sub, err := fs.Sub(fixtureAssets, filepath.ToSlash(dir))
	require.NoError(t, err)

	server := httptest.NewUnstartedServer(nil)
	url := "http://" + server.Listener.Addr().String()
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(sub, strings.TrimPrefix(r.URL.Path, "/"))
		if err != nil {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strings.ReplaceAll(string(data), placeholder, url))) //nolint:gosec // serves a fixture from the embedded FS
	})
	server.Start()
	t.Cleanup(server.Close)

	return server
}

// rewriteFixture copies a fixture into the test's temporary directory, replacing
// every occurrence of placeholder with replacement, and returns the copy's path.
func rewriteFixture(t testing.TB, path, placeholder, replacement string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	target := filepath.Join(t.TempDir(), filepath.Base(path))
	//nolint:gosec // writes a fixture into the test's own temporary directory
	require.NoError(t, os.WriteFile(target, []byte(strings.ReplaceAll(string(data), placeholder, replacement)), 0o600))

	return target
}

func jsonDoc(path string) (json.RawMessage, error) {
	data, err := loading.LoadFromFileOrHTTP(path)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func docAndOpts(t testing.TB, fixturePath string) ([]byte, *ExpandOptions) {
	doc, err := jsonDoc(fixturePath)
	require.NoError(t, err)

	return doc, &ExpandOptions{
		RelativeBase: fixturePath,
	}
}

func expandThisSchemaOrDieTrying(t testing.TB, fixturePath string) (string, *Schema) {
	doc, opts := docAndOpts(t, fixturePath)

	sch := new(Schema)
	require.NoError(t, json.Unmarshal(doc, sch))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSchemaWithBasePath(sch, nil, opts))
	}, "calling expand schema circular refs, should not panic!")

	bbb, err := json.MarshalIndent(sch, "", " ")
	require.NoError(t, err)

	return string(bbb), sch
}

func expandThisOrDieTrying(t testing.TB, fixturePath string) (string, *Swagger) {
	doc, opts := docAndOpts(t, fixturePath)

	spec := new(Swagger)
	require.NoError(t, json.Unmarshal(doc, spec))

	require.NotPanics(t, func() {
		require.NoError(t, ExpandSpec(spec, opts))
	}, "calling expand spec with circular refs, should not panic!")

	bbb, err := json.MarshalIndent(spec, "", " ")
	require.NoError(t, err)

	return string(bbb), spec
}

// assertRefInJSONRegexp ensures all $ref in a jazon document have a given prefix.
//
// NOTE: matched $ref might be empty.
func assertRefInJSON(t testing.TB, jazon, prefix string) {
	// assert a match in a references
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)

	for _, matched := range m {
		subMatch := matched[1]
		assert.TrueT(t, strings.HasPrefix(subMatch, prefix),
			"expected $ref to match %q, got: %s", prefix, matched[0])
	}
}

// assertRefInJSONRegexp ensures all $ref in a jazon document match a given regexp
//
// NOTE: matched $ref might be empty.
func assertRefInJSONRegexp(t testing.TB, jazon, match string) {
	// assert a match in a references
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)

	refMatch, err := regexp.Compile(match)
	require.NoError(t, err)

	for _, matched := range m {
		subMatch := matched[1]
		assert.TrueT(t, refMatch.MatchString(subMatch),
			"expected $ref to match %q, got: %s", match, matched[0])
	}
}

// assertNoRef ensures that no $ref is remaining in json doc.
func assertNoRef(t testing.TB, jazon string) {
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.Nil(t, m)
}

// assertRefExpand ensures that all $ref in some json doc expand properly against a root document.
//
// "exclude" is a regexp pattern to ignore certain $ref (e.g. some specs may embed $ref that are not processed, such as extensions).
func assertRefExpand(t *testing.T, jazon, _ string, root any, opts ...*ExpandOptions) {
	assertRefWithFunc(t, jazon, "", func(t *testing.T, match string) {
		ref := RefSchema(match)
		if len(opts) > 0 {
			options := *opts[0]
			require.NoError(t, ExpandSchemaWithBasePath(ref, nil, &options))
		} else {
			require.NoError(t, ExpandSchema(ref, root, nil))
		}
	})
}

// assertRefResolve ensures that all $ref in some json doc resolve properly against a root document.
//
// "exclude" is a regexp pattern to ignore certain $ref (e.g. some specs may embed $ref that are not processed, such as extensions).
func assertRefResolve(t *testing.T, jazon, exclude string, root any, opts ...*ExpandOptions) {
	assertRefWithFunc(t, jazon, exclude, func(t *testing.T, match string) {
		ref := MustCreateRef(match)
		var (
			sch *Schema
			err error
		)
		if len(opts) > 0 {
			options := *opts[0]
			sch, err = ResolveRefWithBase(root, &ref, &options)
		} else {
			sch, err = ResolveRef(root, &ref)
		}

		require.NoErrorf(t, err, `%v: for "$ref": %q`, err, match)
		require.NotNil(t, sch)
	})
}

// assertRefResolve ensures that all $ref in some json doc verify some asserting func.
//
// "exclude" is a regexp pattern to ignore certain $ref (e.g. some specs may embed $ref that are not processed, such as extensions).
func assertRefWithFunc(t *testing.T, jazon, exclude string, asserter func(t *testing.T, match string)) {
	filterRex := regexp.MustCompile(exclude)
	m := rex.FindAllStringSubmatch(jazon, -1)
	require.NotNil(t, m)
	allRefs := make(map[string]struct{}, len(m))
	for _, matched := range m {
		subMatch := matched[1]
		if exclude != "" && filterRex.MatchString(subMatch) {
			continue
		}
		_, ok := allRefs[subMatch]
		if ok {
			continue
		}
		allRefs[subMatch] = struct{}{}

		t.Run(fmt.Sprintf("%s-%s", t.Name(), subMatch), func(t *testing.T) {
			t.Parallel()
			asserter(t, subMatch)
		})
	}
}

func asJSON(t testing.TB, sp any) string {
	bbb, err := json.MarshalIndent(sp, "", " ")
	require.NoError(t, err)

	return string(bbb)
}
