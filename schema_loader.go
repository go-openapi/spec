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
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"sort"
	"strings"

	"github.com/go-openapi/swag"
)

// PathLoader is a function to use when loading remote refs.
//
// This is a package level default. It may be overridden or bypassed by
// specifying the loader in ExpandOptions.
//
// NOTE: if you are using the go-openapi/loads package, it will override
// this value with its own default (a loader to retrieve YAML documents as
// well as JSON ones).
var PathLoader func(string) (json.RawMessage, error)

func init() {
	PathLoader = func(path string) (json.RawMessage, error) {
		data, err := swag.LoadFromFileOrHTTP(path)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(data), nil
	}
}

type refTracker struct {
	Pointer string // pointer to the referrer of that $ref
}

type refTrackers []refTracker

func (tr refTrackers) Len() int      { return len(tr) }
func (tr refTrackers) Swap(i, j int) { tr[i], tr[j] = tr[j], tr[i] }
func (tr refTrackers) Less(i, j int) bool {
	// prefer to resolve on definitions than other pointers
	defi := strings.HasPrefix(tr[i].Pointer, "/definitions")
	defj := strings.HasPrefix(tr[j].Pointer, "/definitions")

	switch {
	case defi && !defj:
		return true
	case !defi && defj:
		return false
	default:
		// prefer to resolve on shortest path
		partsi := len(strings.Split(tr[i].Pointer, "/"))
		partsj := len(strings.Split(tr[j].Pointer, "/"))
		if partsi == partsj {
			return tr[i].Pointer < tr[j].Pointer
		}
		// lexicographic order on pointers at the same depth
		return partsi < partsj
	}
}

// resolverContext allows to share a context during spec processing.
// At the moment, it just holds the index of circular references found.
type resolverContext struct {
	// circulars holds all visited circular references, to shortcircuit $ref resolution.
	//
	// This structure is privately instantiated and needs not be locked against
	// concurrent access, unless we chose to implement a parallel spec walking.
	circulars map[string]bool
	basePath  string
	loadDoc   func(string) (json.RawMessage, error)
	// allRefs holds all pointers to referrer schemas which hold a $ref.
	// This is used to resolve circular $ref declared outside the root document:
	// in that case a pointer to some referrer in the root document is selected instead.
	allRefs map[string]refTrackers
	// rootID holds the ID of the current root schema. This is used to rebase $ref against ID.
	rootID string
}

func newResolverContext(expandOptions *ExpandOptions) *resolverContext {
	absBase, _ := absPath(expandOptions.RelativeBase)

	// path loader may be overridden from option
	var loader func(string) (json.RawMessage, error)
	if expandOptions.PathLoader == nil {
		loader = PathLoader
	} else {
		loader = expandOptions.PathLoader
	}

	return &resolverContext{
		circulars: make(map[string]bool),
		basePath:  absBase, // keep the root base path in context
		loadDoc:   loader,
		allRefs:   make(map[string]refTrackers, 50),
	}
}

type schemaLoader struct {
	root    interface{}
	options *ExpandOptions
	cache   ResolutionCache
	context *resolverContext
}

func (r *schemaLoader) transitiveResolver(basePath string, ref Ref) *schemaLoader {
	if ref.IsRoot() || ref.HasFragmentOnly {
		return r
	}

	baseRef := MustCreateRef(basePath)
	currentRef := normalizeFileRef(ref, basePath)

	if strings.HasPrefix(currentRef.String(), baseRef.String()) {
		return r
	}

	// set a new root against which to resolve
	rootURL := currentRef.GetURL()
	rootURL.Fragment = ""

	root, _ := r.cache.Get(rootURL.String())

	// shallow copy of resolver options to set a new RelativeBase when
	// traversing multiple documents
	newOptions := r.options
	newOptions.RelativeBase = rootURL.String()
	return defaultSchemaLoader(root, newOptions, r.cache, r.context)
}

func (r *schemaLoader) updateBasePath(transitive *schemaLoader, basePath string) string {
	if transitive != r {
		if transitive.options != nil && transitive.options.RelativeBase != "" {
			basePath, _ = absPath(transitive.options.RelativeBase)
		}
	}
	return basePath
}

func (r *schemaLoader) resolveRef(ref *Ref, target interface{}, basePath, pointer string) error {
	tgt := reflect.ValueOf(target)
	if tgt.Kind() != reflect.Ptr {
		return ErrResolveRefNeedsAPointer
	}

	if ref == nil || ref.GetURL() == nil {
		return nil
	}

	var (
		res  interface{}
		data interface{}
		err  error
	)

	// resolve against the root if it isn't nil: if ref is pointing at the root, or has a fragment only which means
	// it is pointing somewhere in the root.
	root := r.root

	if (ref.IsRoot() || ref.HasFragmentOnly) && root == nil {
		if basePath != "" {
			// this resolver already has a root set up
			baseRef := MustCreateRef(basePath)
			root, _ = r.load(baseRef.GetURL())
		} else {
			// resolve the root against the base path
			baseRef := MustCreateRef(r.options.RelativeBase)
			root, _ = r.load(baseRef.GetURL())
		}
	}

	baseRef := normalizeFileRef(*ref, basePath)

	// track the current referrer to this $ref (for future simplification of a cyclical ref)
	r.context.allRefs[baseRef.String()] = append(r.context.allRefs[baseRef.String()], refTracker{Pointer: pointer})

	if (ref.IsRoot() || ref.HasFragmentOnly) && root != nil {
		data = root
	} else {
		data, err = r.load(baseRef.GetURL())
		if err != nil {
			return err
		}
	}

	res = data
	if ref.String() != "" || ref.IsRoot() {
		res, _, err = ref.GetPointer().Get(data)
		if err != nil {
			return err
		}
	}

	return swag.DynamicJSONToStruct(res, target)
}

// load a document from an URL or from cache
func (r *schemaLoader) load(refURL *url.URL) (interface{}, error) {
	toFetch := *refURL
	toFetch.Fragment = ""

	var err error
	pth := toFetch.String()
	if pth == rootBase {
		pth, err = absPath(rootBase)
		if err != nil {
			return nil, err
		}
	}
	normalized := normalizeAbsPath(pth)

	data, fromCache := r.cache.Get(normalized)
	if !fromCache {
		b, err := r.context.loadDoc(normalized)
		if err != nil {
			return nil, fmt.Errorf("%s [normalized: %s]: %w", pth, normalized, err)
		}

		var doc interface{}
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, err
		}

		r.cache.Set(normalized, doc)

		return doc, nil
	}

	return data, nil
}

// isCircular detects cycles in sequences of $ref.
//
// It relies on a private context (which needs not be locked).
func (r *schemaLoader) isCircular(ref Ref, basePath string, parentRefs ...string) (foundCycle bool) {
	normalizedRef := normalizePaths(ref.String(), basePath)
	if _, ok := r.context.circulars[normalizedRef]; ok {
		// circular $ref has been already detected in another explored cycle
		foundCycle = true
		return
	}
	foundCycle = swag.ContainsStringsCI(parentRefs, normalizedRef) // TODO(fred): normalize windows url and remove CI equality
	if foundCycle {
		r.context.circulars[normalizedRef] = true
	}
	return
}

func (r *schemaLoader) resolveCircularRef(ref Ref, basePath string) (result Ref) {
	// ref and basePath must be normalized

	result = r.denormalizeFileRef(ref, basePath)

	if result.RemoteURI() == "" {
		// circularity is already captured in the root document: simply rebase the ref
		return
	}

	// case of circularity detected while walking through remote documents:
	// find an earlier referrer to this resource and replace the $ref by a json pointer to it
	referrers, ok := r.context.allRefs[ref.String()]
	if !ok {
		// guard against dev errors: a circular ref has necessarily been referred to at least once
		panic(ErrInternalRef)
	}
	sort.Sort(referrers) // pick the preferred referrer known at the time the circular is found

	return MustCreateRef("#" + referrers[0].Pointer)
}

func (r *schemaLoader) rebaseRef(ref Ref, basePath string) Ref {
	// rebase $ref to the initial base path for the root document
	return r.denormalizeFileRef(ref, basePath)
}

// Resolve resolves a reference against basePath and stores the result in target.
//
// Resolve is not in charge of following references: it only resolves ref by following its URL.
//
// If the schema the ref is referring to holds nested refs, Resolve doesn't resolve them.
//
// If basePath is an empty string, ref is resolved against the root schema stored in the schemaLoader struct
func (r *schemaLoader) Resolve(ref *Ref, target interface{}, basePath, pointer string) error {
	return r.resolveRef(ref, target, basePath, pointer)
}

func (r *schemaLoader) deref(input interface{}, parentRefs []string, basePath, pointer string) error {
	var ref *Ref
	switch refable := input.(type) {
	// all spec types which support $ref
	case *Schema:
		ref = &refable.Ref
	case *Parameter:
		ref = &refable.Ref
	case *Response:
		ref = &refable.Ref
	case *PathItem:
		ref = &refable.Ref
	default:
		return fmt.Errorf("unsupported type: %T: %w", input, ErrDerefUnsupportedType)
	}

	curRef := ref.String()
	if curRef == "" && !ref.IsRoot() {
		return nil
	}

	normalizedRef := normalizeFileRef(*ref, basePath)
	normalizedBasePath := normalizedRef.RemoteURI()

	if r.isCircular(normalizedRef, basePath, parentRefs...) {
		return nil
	}

	if err := r.resolveRef(ref, input, basePath, pointer); r.shouldStopOnError(err) {
		return err
	}

	if ref.String() == "" && !ref.IsRoot() || ref.String() == curRef {
		// done with rereferencing: no more $ref or self-referencing $ref
		return nil
	}

	// continue dereferencing down the rabbit hole
	parentRefs = append(parentRefs, normalizedRef.String())
	return r.deref(input, parentRefs, normalizedBasePath, pointer)
}

func (r *schemaLoader) shouldStopOnError(err error) bool {
	if err != nil && !r.options.ContinueOnError {
		return true
	}

	if err != nil {
		log.Println(err)
	}

	return false
}

func (r *schemaLoader) setSchemaID(target interface{}, id, basePath, pointer string) (string, string) {
	// sets a schema ID for the current resolved target

	// handling the case when id is a folder
	// remember that basePath has to point to a file
	var refPath string
	if strings.HasSuffix(id, "/") {
		// path.Clean here would not work correctly if there is a scheme (e.g. https://...)
		refPath = fmt.Sprintf("%s%s", id, "placeholder.json")
	} else {
		refPath = id
	}

	if r.context.rootID == "" {
		r.context.rootID = id
	}

	// updates the current base path
	// * important: ID can be a relative path
	// * registers target to be fetchable from the new base proposed by this id
	newBasePath := normalizePaths(refPath, basePath)

	// store found IDs for possible future reuse in $ref
	if _, found := r.cache.Get(newBasePath); !found {
		// keep the original document: do not account for mutated content
		r.cache.Set(newBasePath, target)
	}

	// track the referrer of this ID
	r.context.allRefs[refPath] = append(r.context.allRefs[refPath], refTracker{Pointer: pointer})

	return newBasePath, refPath
}

func defaultSchemaLoader(
	root interface{},
	expandOptions *ExpandOptions,
	cache ResolutionCache,
	context *resolverContext) *schemaLoader {

	if expandOptions == nil {
		expandOptions = &ExpandOptions{}
	}

	if context == nil {
		context = newResolverContext(expandOptions)
	}

	if schema, ok := root.(*Schema); ok && schema.ID != "" {
		// keep the ID of the root schema to rebase cyclical $ref to root
		context.rootID = schema.ID
	}

	return &schemaLoader{
		root:    root,
		options: expandOptions,
		cache:   cacheOrDefault(cache),
		context: context,
	}
}
