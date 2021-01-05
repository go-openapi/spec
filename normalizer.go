// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const windowsOS = "windows"

// normalize absolute path for cache.
// on Windows, drive letters should be converted to lower as scheme in net/url.URL
func normalizeAbsPath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		debugLog("normalize absolute path failed: %s", err)
		return path
	}
	return u.String()
}

// base or refPath could be a file path or a URL
// given a base absolute path and a ref path, return the absolute path of refPath.
//
//   1) if refPath is absolute, return it
//   2) if refPath is relative, join it with basePath keeping the scheme, hosts, and ports if exists
//
// base could be a directory or a full file path.
func normalizePaths(refPath, base string) string {
	refURL, _ := url.Parse(refPath)
	if path.IsAbs(refURL.Path) || filepath.IsAbs(refPath) {
		// refPath is actually absolute
		if refURL.Host != "" {
			return refPath
		}
		parts := strings.Split(refPath, "#")
		result := filepath.FromSlash(parts[0])
		if len(parts) == 2 {
			result += "#" + parts[1]
		}
		return result
	}

	// relative refPath
	baseURL, _ := url.Parse(base)
	if !strings.HasPrefix(refPath, "#") {
		// combining paths
		if baseURL.Host != "" {
			baseURL.Path = path.Join(path.Dir(baseURL.Path), refURL.Path)
		} else { // base is a file
			newBase := fmt.Sprintf("%s#%s", filepath.Join(filepath.Dir(base), filepath.FromSlash(refURL.Path)), refURL.Fragment)
			return newBase
		}

	}
	// copying fragment from ref to base
	baseURL.Fragment = refURL.Fragment
	return baseURL.String()
}

// isRoot is a temporary hack to discern windows file ref for ref.IsRoot().
//
// TODO: a more thorough change is needed to handle windows file refs.
func isRoot(ref Ref) bool {
	if runtime.GOOS != windowsOS {
		return ref.IsRoot()
	}
	return !filepath.IsAbs(ref.String())
}

// isAbs is a temporary hack to discern windows file ref for url IsAbs().
//
// TODO: a more thorough change is needed to handle windows file refs.
func isAbs(u *url.URL) bool {
	if runtime.GOOS != windowsOS {
		return u.IsAbs()
	}
	if len(u.Scheme) <= 1 {
		// drive letter got caught as URI scheme
		return false
	}
	return u.IsAbs()
}

func rebase(str, prefix string) (string, bool) {
	if !strings.HasPrefix(str, prefix) {
		return "", false
	}

	rebased := strings.TrimPrefix(str, prefix)

	if rebased == "" { // exact match: points to its own root document
		return "#", true
	}

	if strings.HasPrefix(rebased, "/") {
		return "#" + rebased, true
	}

	return rebased, true
}

// denormalizePaths returns to simplest notation on file $ref,
// i.e. strips the absolute path and sets a path relative to the base path.
//
// This is currently used when we rewrite ref after a circular ref has been detected
func (r *schemaLoader) denormalizeFileRef(ref Ref, relativeBase string) Ref {
	originalRelativeBase := r.context.basePath
	rootID := r.context.rootID

	str := ref.String()

	debugLog("denormalize: %s, IsRoot: %t,HasFragmentOnly: %t, HasFullURL: %t [relative: %s, original: %s, rootID: %s]",
		str, ref.IsRoot(), ref.HasFragmentOnly, ref.HasFullURL, relativeBase, originalRelativeBase, rootID)

	if str == "" || isRoot(ref) || ref.HasFragmentOnly {
		// nothing to be done
		return ref
	}

	// strip fragments from URI
	relativeBaseURL, _ := url.Parse(relativeBase)
	relativeBaseURL.Fragment = ""

	originalRelativeBaseURL, _ := url.Parse(originalRelativeBase)
	originalRelativeBaseURL.Fragment = ""

	rootIDURL, _ := url.Parse(rootID)
	rootIDURL.Fragment = ""

	if isAbs(relativeBaseURL) {

		// this should work for absolute URI (e.g. http://...): we have an exact match, just trim prefix
		if rebased, ok := rebase(str, originalRelativeBaseURL.String()); ok {
			return MustCreateRef(rebased)
		}

		// when absolute but in some root identified by an ID
		if rebased, ok := rebase(str, rootIDURL.String()); ok {
			return MustCreateRef(rebased)
		}

		// other absolute URL get unchanged (i.e. with a non-empty scheme)
		return ref
	}

	// for relative file URIs:
	// the resulting ref is in the expanded spec: return a local ref
	if rebased, ok := rebase(str, originalRelativeBaseURL.String()); ok {
		return MustCreateRef(rebased)
	}

	// check if we may set a relative path, considering the original base path for this spec.
	// Example:
	//   spec is located at /mypath/spec.json
	//   my normalized ref points to: /mypath/item.json#/target
	//   expected result: item.json#/target
	parts := strings.Split(str, "#")
	relativePath, err := filepath.Rel(filepath.Dir(originalRelativeBaseURL.String()), parts[0])
	if err != nil {
		// there is no common ancestor (e.g. different drives on windows)
		// leaves the ref unchanged
		return ref
	}
	if len(parts) == 2 {
		relativePath += "#" + parts[1]
	}

	return MustCreateRef(relativePath)
}

// relativeBase could be an ABSOLUTE file path or an ABSOLUTE URL
func normalizeFileRef(ref Ref, relativeBase string) Ref {
	// This is important for when the reference is pointing to the root schema
	if ref.String() == "" {
		return MustCreateRef(relativeBase)
	}

	s := normalizePaths(ref.String(), relativeBase)
	return MustCreateRef(s)
}

// absPath returns the absolute path of a file
func absPath(fname string) (string, error) {
	if strings.HasPrefix(fname, "http") {
		return fname, nil
	}
	if filepath.IsAbs(fname) {
		return fname, nil
	}
	wd, err := os.Getwd()
	return normalizeAbsPath(filepath.Join(wd, fname)), err
}
