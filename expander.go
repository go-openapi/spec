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
	"path"
	"strconv"

	"github.com/go-openapi/jsonpointer"
)

var opsPointers = []string{
	// keys to generate json pointers to operations
	"get", "head", "options", "put", "post", "patch", "delete",
}

// ExpandOptions provides options for the spec expander.
type ExpandOptions struct {
	RelativeBase    string
	SkipSchemas     bool
	ContinueOnError bool
	PathLoader      func(string) (json.RawMessage, error) `json:"-"`

	AbsoluteCircularRef bool // AbsoluteCircularRef is now deprecated
}

// ExpandSpec expands the references in a swagger spec
func ExpandSpec(spec *Swagger, options *ExpandOptions) error {
	resolver := defaultSchemaLoader(spec, options, nil, nil)

	// getting the base path of the spec to adjust all subsequent reference resolutions
	specBasePath := ""
	if options != nil && options.RelativeBase != "" {
		specBasePath, _ = absPath(options.RelativeBase)
	}

	if options == nil || !options.SkipSchemas {
		for key, definition := range spec.Definitions {
			parentRefs := make([]string, 0, 10)
			parentRefs = append(parentRefs, fmt.Sprintf("#/definitions/%s", key))
			def, err := expandSchema(definition, parentRefs, resolver, specBasePath, path.Join("/definitions", jsonpointer.Escape(key)))
			if resolver.shouldStopOnError(err) {
				return err
			}
			if def != nil {
				spec.Definitions[key] = *def
			}
		}
	}

	for key := range spec.Parameters {
		parameter := spec.Parameters[key]
		if err := expandParameterOrResponse(&parameter, resolver, specBasePath, path.Join("/parameters", jsonpointer.Escape(key))); resolver.shouldStopOnError(err) {
			return err
		}
		spec.Parameters[key] = parameter
	}

	for key := range spec.Responses {
		response := spec.Responses[key]
		if err := expandParameterOrResponse(&response, resolver, specBasePath, path.Join("/responses", jsonpointer.Escape(key))); resolver.shouldStopOnError(err) {
			return err
		}
		spec.Responses[key] = response
	}

	if spec.Paths != nil {
		for key := range spec.Paths.Paths {
			pth := spec.Paths.Paths[key]
			if err := expandPathItem(&pth, resolver, specBasePath, path.Join("/paths", jsonpointer.Escape(key))); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Paths.Paths[key] = pth
		}
	}

	return nil
}

const rootBase = ".root"

// baseForRoot loads in the cache the root document and produces a fake ".root" base path entry
// for further $ref resolution
//
// Setting the cache is optional and this parameter may safely be left to nil.
func baseForRoot(root interface{}, cache ResolutionCache) string {
	if root == nil {
		return ""
	}

	// cache the root document to resolve $ref's
	base, _ := absPath(rootBase)
	normalizedBase := normalizeAbsPath(base)
	cache.Set(normalizedBase, root)

	return normalizedBase
}

// ExpandSchema expands the refs in the schema object with reference to the root object.
//
// go-openapi/validate uses this function.
//
// Notice that it is impossible to reference a json schema in a different document other than root
// (use ExpandSchemaWithBasePath to resolve external references).
//
// Setting the cache is optional and this parameter may safely be left to nil.
func ExpandSchema(schema *Schema, root interface{}, cache ResolutionCache) error {
	cache = cacheOrDefault(cache)
	if root == nil {
		root = schema
	}

	opts := &ExpandOptions{
		// when a root is specified, cache the root as an in-memory document for $ref retrieval
		RelativeBase:    baseForRoot(root, cache),
		SkipSchemas:     false,
		ContinueOnError: false,
	}

	return ExpandSchemaWithBasePath(schema, cache, opts)
}

// ExpandSchemaWithBasePath expands the refs in the schema object, base path configured through expand options.
//
// Setting the cache is optional and this parameter may safely be left to nil.
func ExpandSchemaWithBasePath(schema *Schema, cache ResolutionCache, opts *ExpandOptions) error {
	if schema == nil {
		return nil
	}

	cache = cacheOrDefault(cache)

	var basePath string
	if opts.RelativeBase != "" {
		basePath, _ = absPath(opts.RelativeBase)
	}

	resolver := defaultSchemaLoader(nil, opts, cache, nil)

	parentRefs := make([]string, 0, 10)
	s, err := expandSchema(*schema, parentRefs, resolver, basePath, "/")
	if err != nil {
		return err
	}

	if s != nil {
		// guard for when continuing on error
		*schema = *s
	}

	return nil
}

func expandItems(target Schema, parentRefs []string, resolver *schemaLoader, basePath, pointer string) (*Schema, error) {
	if target.Items == nil {
		return &target, nil
	}

	// array
	if target.Items.Schema != nil {
		t, err := expandSchema(*target.Items.Schema, parentRefs, resolver, basePath, path.Join(pointer, "items"))
		if err != nil {
			return nil, err
		}
		*target.Items.Schema = *t
	}

	// tuple
	for i := range target.Items.Schemas {
		t, err := expandSchema(target.Items.Schemas[i], parentRefs, resolver, basePath, path.Join(pointer, "items", strconv.Itoa(i)))
		if err != nil {
			return nil, err
		}
		target.Items.Schemas[i] = *t
	}

	return &target, nil
}

func expandSchema(target Schema, parentRefs []string, resolver *schemaLoader, basePath, pointer string) (*Schema, error) {

	// A $ref is encountered: ignore all other keys
	if target.Ref.String() != "" || target.Ref.IsRoot() {
		return expandSchemaRef(target, parentRefs, resolver, basePath, pointer)
	}

	// A schema ID is encountered: rebase and track new parent
	// change the base path of resolution when an ID is encountered
	// otherwise the basePath should inherit the parent's
	if target.ID != "" {
		var parent string
		basePath, parent = resolver.setSchemaID(target, target.ID, basePath, pointer)

		// add normalized ID to the list of parents, in order to detect cycles
		parentRefs = append(parentRefs, parent)

		// remove ID from the expanded spec: IDs are no more required in the expanded spec and
		// remaining nested schema IDs would work against proper $ref resolution on the expanded spec
		// (in the case of circular $ref built on top of ID-based $ref).
		target.ID = ""
	}

	for k := range target.Definitions {
		tt, err := expandSchema(target.Definitions[k], parentRefs, resolver, basePath, path.Join(pointer, "definitions", jsonpointer.Escape(k)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if tt != nil {
			target.Definitions[k] = *tt
		}
	}

	t, err := expandItems(target, parentRefs, resolver, basePath, pointer)
	if resolver.shouldStopOnError(err) {
		return &target, err
	}
	if t != nil {
		target = *t
	}

	for i := range target.AllOf {
		t, err := expandSchema(target.AllOf[i], parentRefs, resolver, basePath, path.Join(pointer, "allOf", strconv.Itoa(i)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AllOf[i] = *t
		}
	}

	for i := range target.AnyOf {
		t, err := expandSchema(target.AnyOf[i], parentRefs, resolver, basePath, path.Join(pointer, "anyOf", strconv.Itoa(i)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AnyOf[i] = *t
		}
	}

	for i := range target.OneOf {
		t, err := expandSchema(target.OneOf[i], parentRefs, resolver, basePath, path.Join(pointer, "oneOf", strconv.Itoa(i)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.OneOf[i] = *t
		}
	}

	if target.Not != nil {
		t, err := expandSchema(*target.Not, parentRefs, resolver, basePath, path.Join(pointer, "not"))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.Not = *t
		}
	}

	for k := range target.Properties {
		t, err := expandSchema(target.Properties[k], parentRefs, resolver, basePath, path.Join(pointer, "properties", jsonpointer.Escape(k)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.Properties[k] = *t
		}
	}

	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		t, err := expandSchema(*target.AdditionalProperties.Schema, parentRefs, resolver, basePath, path.Join(pointer, "additionalProperties"))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalProperties.Schema = *t
		}
	}

	for k := range target.PatternProperties {
		t, err := expandSchema(target.PatternProperties[k], parentRefs, resolver, basePath, path.Join(pointer, "patternProperties", jsonpointer.Escape(k)))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.PatternProperties[k] = *t
		}
	}

	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			t, err := expandSchema(*target.Dependencies[k].Schema, parentRefs, resolver, basePath, path.Join(pointer, "dependencies", jsonpointer.Escape(k)))
			if resolver.shouldStopOnError(err) {
				return &target, err
			}
			if t != nil {
				*target.Dependencies[k].Schema = *t
			}
		}
	}

	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		t, err := expandSchema(*target.AdditionalItems.Schema, parentRefs, resolver, basePath, path.Join(pointer, "additionalItems"))
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalItems.Schema = *t
		}
	}

	return &target, nil
}

func expandSchemaRef(target Schema, parentRefs []string, resolver *schemaLoader, basePath, pointer string) (*Schema, error) {
	// if a Ref is found, all sibling fields are skipped
	// Ref also changes the resolution scope of children expandSchema

	// here the resolution scope is changed because a $ref was encountered
	normalizedRef := normalizeFileRef(target.Ref, basePath)
	normalizedBasePath := normalizedRef.RemoteURI()

	if resolver.isCircular(normalizedRef, basePath, parentRefs...) {
		// this means there is a cycle in the recursion tree: return the Ref
		target.Ref = resolver.resolveCircularRef(normalizedRef, normalizedBasePath)
		return &target, nil
	}

	var t *Schema
	err := resolver.Resolve(&target.Ref, &t, basePath, pointer)
	if resolver.shouldStopOnError(err) {
		return nil, err
	}

	if t == nil {
		// guard for when continuing on error
		return &target, nil
	}

	parentRefs = append(parentRefs, normalizedRef.String())
	transitiveResolver := resolver.transitiveResolver(basePath, target.Ref)

	basePath = resolver.updateBasePath(transitiveResolver, normalizedBasePath)

	return expandSchema(*t, parentRefs, transitiveResolver, basePath, pointer)
}

func expandPathItem(pathItem *PathItem, resolver *schemaLoader, basePath, pointer string) error {
	if pathItem == nil {
		return nil
	}

	parentRefs := make([]string, 0, 10)
	if err := resolver.deref(pathItem, parentRefs, basePath, pointer); resolver.shouldStopOnError(err) {
		return err
	}

	if pathItem.Ref.String() != "" || pathItem.Ref.IsRoot() {
		transitiveResolver := resolver.transitiveResolver(basePath, pathItem.Ref)
		basePath = transitiveResolver.updateBasePath(resolver, basePath)
		resolver = transitiveResolver
	}

	pathItem.Ref = Ref{}
	for i := range pathItem.Parameters {
		if err := expandParameterOrResponse(&(pathItem.Parameters[i]), resolver, basePath, path.Join(pointer, "parameters", strconv.Itoa(i))); resolver.shouldStopOnError(err) {
			return err
		}
	}

	ops := []*Operation{
		pathItem.Get,
		pathItem.Head,
		pathItem.Options,
		pathItem.Put,
		pathItem.Post,
		pathItem.Patch,
		pathItem.Delete,
	}
	for i, op := range ops {
		if err := expandOperation(op, resolver, basePath, path.Join(pointer, opsPointers[i])); resolver.shouldStopOnError(err) {
			return err
		}
	}

	return nil
}

func expandOperation(op *Operation, resolver *schemaLoader, basePath, pointer string) error {
	if op == nil {
		return nil
	}

	for i := range op.Parameters {
		param := op.Parameters[i]
		if err := expandParameterOrResponse(&param, resolver, basePath, path.Join(pointer, "parameters", strconv.Itoa(i))); resolver.shouldStopOnError(err) {
			return err
		}
		op.Parameters[i] = param
	}

	if op.Responses == nil {
		return nil
	}

	responses := op.Responses
	if err := expandParameterOrResponse(responses.Default, resolver, basePath, path.Join(pointer, "default")); resolver.shouldStopOnError(err) {
		return err
	}

	for code := range responses.StatusCodeResponses {
		response := responses.StatusCodeResponses[code]
		if err := expandParameterOrResponse(&response, resolver, basePath, path.Join(pointer, strconv.Itoa(code))); resolver.shouldStopOnError(err) {
			return err
		}
		responses.StatusCodeResponses[code] = response
	}

	return nil
}

// ExpandResponseWithRoot expands a response based on a root document, not a fetchable document
//
// Notice that it is impossible to reference a json schema in a different document other than root
// (use ExpandResponse to resolve external references).
//
// Setting the cache is optional and this parameter may safely be left to nil.
func ExpandResponseWithRoot(response *Response, root interface{}, cache ResolutionCache) error {
	cache = cacheOrDefault(cache)
	opts := &ExpandOptions{
		RelativeBase:    baseForRoot(root, cache),
		SkipSchemas:     false,
		ContinueOnError: false,
	}
	resolver := defaultSchemaLoader(root, opts, cache, nil)

	return expandParameterOrResponse(response, resolver, opts.RelativeBase, "/")
}

// ExpandResponse expands a response based on a basepath
//
// All refs inside response will be resolved relative to basePath
func ExpandResponse(response *Response, basePath string) error {
	var specBasePath string
	if basePath != "" {
		specBasePath, _ = absPath(basePath)
	}
	opts := &ExpandOptions{
		RelativeBase: specBasePath,
	}
	resolver := defaultSchemaLoader(nil, opts, nil, nil)

	return expandParameterOrResponse(response, resolver, opts.RelativeBase, "/")
}

// ExpandParameterWithRoot expands a parameter based on a root document, not a fetchable document.
//
// Notice that it is impossible to reference a json schema in a different document other than root
// (use ExpandParameter to resolve external references).
func ExpandParameterWithRoot(parameter *Parameter, root interface{}, cache ResolutionCache) error {
	cache = cacheOrDefault(cache)
	opts := &ExpandOptions{
		RelativeBase:    baseForRoot(root, cache),
		SkipSchemas:     false,
		ContinueOnError: false,
	}
	resolver := defaultSchemaLoader(root, opts, cache, nil)

	return expandParameterOrResponse(parameter, resolver, opts.RelativeBase, "/")
}

// ExpandParameter expands a parameter based on a basepath.
//
// All refs inside parameter will be resolved relative to basePath
func ExpandParameter(parameter *Parameter, basePath string) error {
	var specBasePath string
	if basePath != "" {
		specBasePath, _ = absPath(basePath)
	}
	opts := &ExpandOptions{
		RelativeBase: specBasePath,
	}
	resolver := defaultSchemaLoader(nil, opts, nil, nil)

	return expandParameterOrResponse(parameter, resolver, opts.RelativeBase, "/")
}

func getRefAndSchema(input interface{}) (*Ref, *Schema, error) {
	var (
		ref *Ref
		sch *Schema
	)

	switch refable := input.(type) {
	case *Parameter:
		if refable == nil {
			return nil, nil, nil
		}
		ref = &refable.Ref
		sch = refable.Schema
	case *Response:
		if refable == nil {
			return nil, nil, nil
		}
		ref = &refable.Ref
		sch = refable.Schema
	default:
		return nil, nil, fmt.Errorf("unsupported type: %T: %w", input, ErrExpandUnsupportedType)
	}

	return ref, sch, nil
}

func expandParameterOrResponse(input interface{}, resolver *schemaLoader, basePath, pointer string) error {
	ref, _, err := getRefAndSchema(input)
	if err != nil {
		return err
	}

	if ref == nil {
		// empty parameter or response
		return nil
	}

	parentRefs := make([]string, 0, 10)
	if err = resolver.deref(input, parentRefs, basePath, pointer); resolver.shouldStopOnError(err) {
		return err
	}

	ref, sch, _ := getRefAndSchema(input)
	if ref.String() != "" || ref.IsRoot() {
		transitiveResolver := resolver.transitiveResolver(basePath, *ref)
		basePath = resolver.updateBasePath(transitiveResolver, basePath)
		resolver = transitiveResolver
	}

	if sch == nil {
		// nothing to be expanded
		if ref != nil {
			*ref = Ref{}
		}
		return nil
	}

	pointer = path.Join(pointer, "schema")

	if sch.Ref.String() != "" || sch.Ref.IsRoot() {
		rebasedRef := normalizeFileRef(sch.Ref, basePath)

		switch {
		case resolver.isCircular(rebasedRef, basePath, parentRefs...):
			// this is a circular $ref: stop expansion
			sch.Ref = resolver.resolveCircularRef(rebasedRef, basePath)
		case !resolver.options.SkipSchemas:
			// schema expanded to a $ref in another root
			sch.Ref = rebasedRef
		default:
			// skip schema expansion but rebase $ref to schema
			sch.Ref = resolver.rebaseRef(rebasedRef, basePath)
		}
	}

	if ref != nil {
		*ref = Ref{}
	}

	// expand schema
	if !resolver.options.SkipSchemas {
		s, err := expandSchema(*sch, parentRefs, resolver, basePath, pointer)
		if resolver.shouldStopOnError(err) {
			return err
		}
		if s == nil {
			// guard for when continuing on error
			return nil
		}
		*sch = *s
	}

	return nil
}
