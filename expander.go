// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"encoding/json"
	"fmt"
)

const smallPrealloc = 10

// ExpandOptions provides options for the spec expander.
//
// RelativeBase is the path to the root document. This can be a remote URL or a path to a local file.
//
// If left empty, the root document is assumed to be located in the current working directory:
// all relative $ref's will be resolved from there.
//
// PathLoader injects a document loading method. By default, this resolves to the function provided by the SpecLoader package variable.
type ExpandOptions struct {
	RelativeBase        string                                // the path to the root document to expand. This is a file, not a directory
	SkipSchemas         bool                                  // do not expand schemas, just paths, parameters and responses
	ContinueOnError     bool                                  // continue expanding even after and error is found
	PathLoader          func(string) (json.RawMessage, error) `json:"-"` // the document loading method that takes a path as input and yields a json document
	AbsoluteCircularRef bool                                  // circular $ref remaining after expansion remain absolute URLs
}

func optionsOrDefault(opts *ExpandOptions) *ExpandOptions {
	if opts != nil {
		clone := *opts // shallow clone to avoid internal changes to be propagated to the caller
		if clone.RelativeBase != "" {
			clone.RelativeBase = normalizeBase(clone.RelativeBase)
		}
		// if the relative base is empty, let the schema loader choose a pseudo root document
		return &clone
	}
	return &ExpandOptions{}
}

// ExpandSpec expands the references in a swagger spec
func ExpandSpec(spec *Swagger, options *ExpandOptions) error {
	options = optionsOrDefault(options)
	resolver := defaultSchemaLoader(spec, options, nil, nil)

	specBasePath := options.RelativeBase

	// Handle OpenAPI 3.x Components.Schemas
	if !options.SkipSchemas && spec.Components != nil {
		for key, schema := range spec.Components.Schemas {
			parentRefs := make([]string, 0, smallPrealloc)
			parentRefs = append(parentRefs, "#/components/schemas/"+key)

			def, err := expandSchema(schema, parentRefs, resolver, specBasePath)
			if resolver.shouldStopOnError(err) {
				return err
			}
			if def != nil {
				spec.Components.Schemas[key] = *def
			}
		}
	}

	// Handle Swagger 2.0 Definitions
	if !options.SkipSchemas && spec.Definitions != nil {
		for key, definition := range spec.Definitions {
			parentRefs := make([]string, 0, smallPrealloc)
			parentRefs = append(parentRefs, "#/definitions/"+key)

			def, err := expandSchema(definition, parentRefs, resolver, specBasePath)
			if resolver.shouldStopOnError(err) {
				return err
			}
			if def != nil {
				spec.Definitions[key] = *def
			}
		}
	}

	// Handle OpenAPI 3.x Components parameters, responses, and requestBodies
	if spec.Components != nil {
		for key := range spec.Components.Parameters {
			parameter := spec.Components.Parameters[key]
			if err := expandParameterOrResponse(&parameter, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Components.Parameters[key] = parameter
		}

		for key := range spec.Components.Responses {
			response := spec.Components.Responses[key]
			if err := expandParameterOrResponse(&response, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Components.Responses[key] = response
		}

		for key := range spec.Components.RequestBodies {
			requestBody := spec.Components.RequestBodies[key]
			if err := expandRequestBody(&requestBody, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Components.RequestBodies[key] = requestBody
		}
	}

	// Handle Swagger 2.0 top-level Parameters (backward compatibility)
	if spec.Parameters != nil {
		for key := range spec.Parameters {
			parameter := spec.Parameters[key]
			if err := expandParameterOrResponse(&parameter, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Parameters[key] = parameter
		}
	}

	// Handle Swagger 2.0 top-level Responses (backward compatibility)
	if spec.Responses != nil {
		for key := range spec.Responses {
			response := spec.Responses[key]
			if err := expandParameterOrResponse(&response, resolver, specBasePath); resolver.shouldStopOnError(err) {
				return err
			}
			spec.Responses[key] = response
		}
	}

	if spec.Paths != nil {
		for key := range spec.Paths.Paths {
			pth := spec.Paths.Paths[key]
			if err := expandPathItem(&pth, resolver, specBasePath); resolver.shouldStopOnError(err) {
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
func baseForRoot(root any, cache ResolutionCache) string {
	// cache the root document to resolve $ref's
	normalizedBase := normalizeBase(rootBase)

	if root == nil {
		// ensure that we never leave a nil root: always cache the root base pseudo-document
		cachedRoot, found := cache.Get(normalizedBase)
		if found && cachedRoot != nil {
			// the cache is already preloaded with a root
			return normalizedBase
		}

		root = map[string]any{}
	}

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
func ExpandSchema(schema *Schema, root any, cache ResolutionCache) error {
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

	opts = optionsOrDefault(opts)

	resolver := defaultSchemaLoader(nil, opts, cache, nil)

	// Preserve $defs from the original schema, as JSON Schema 2020-12 allows
	// $ref alongside other keywords, but our expander replaces the schema with
	// the referenced one when $ref is present.
	originalDefs := schema.Defs
	originalDefinitions := schema.Definitions

	parentRefs := make([]string, 0, smallPrealloc)
	s, err := expandSchema(*schema, parentRefs, resolver, opts.RelativeBase)
	if err != nil {
		return err
	}
	if s != nil {
		// guard for when continuing on error
		*schema = *s
	}

	// Restore $defs and definitions if they were lost during expansion
	if len(schema.Defs) == 0 && len(originalDefs) > 0 {
		schema.Defs = originalDefs
	}
	if len(schema.Definitions) == 0 && len(originalDefinitions) > 0 {
		schema.Definitions = originalDefinitions
	}

	return nil
}

func expandItems(target Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*Schema, error) {
	if target.Items == nil {
		return &target, nil
	}

	// array
	if target.Items.Schema != nil {
		t, err := expandSchema(*target.Items.Schema, parentRefs, resolver, basePath)
		if err != nil {
			return nil, err
		}
		*target.Items.Schema = *t
	}

	// tuple
	for i := range target.Items.Schemas {
		t, err := expandSchema(target.Items.Schemas[i], parentRefs, resolver, basePath)
		if err != nil {
			return nil, err
		}
		target.Items.Schemas[i] = *t
	}

	return &target, nil
}

func expandSchema(target Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*Schema, error) {
	if target.Ref.String() == "" && target.Ref.IsRoot() {
		newRef := normalizeRef(&target.Ref, basePath)
		target.Ref = *newRef
		return &target, nil
	}

	// change the base path of resolution when an ID is encountered
	// otherwise the basePath should inherit the parent's
	if target.ID != "" {
		basePath, _ = resolver.setSchemaID(target, target.ID, basePath)
	}

	if target.Ref.String() != "" {
		if !resolver.options.SkipSchemas {
			return expandSchemaRef(target, parentRefs, resolver, basePath)
		}

		// when "expand" with SkipSchema, we just rebase the existing $ref without replacing
		// the full schema.
		rebasedRef, err := NewRef(normalizeURI(target.Ref.String(), basePath))
		if err != nil {
			return nil, err
		}
		target.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)

		return &target, nil
	}

	for k := range target.Definitions {
		tt, err := expandSchema(target.Definitions[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if tt != nil {
			target.Definitions[k] = *tt
		}
	}

	// JSON Schema 2020-12 uses $defs instead of definitions
	for k := range target.Defs {
		tt, err := expandSchema(target.Defs[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if tt != nil {
			target.Defs[k] = *tt
		}
	}

	t, err := expandItems(target, parentRefs, resolver, basePath)
	if resolver.shouldStopOnError(err) {
		return &target, err
	}
	if t != nil {
		target = *t
	}

	for i := range target.AllOf {
		t, err := expandSchema(target.AllOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AllOf[i] = *t
		}
	}

	for i := range target.AnyOf {
		t, err := expandSchema(target.AnyOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.AnyOf[i] = *t
		}
	}

	for i := range target.OneOf {
		t, err := expandSchema(target.OneOf[i], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.OneOf[i] = *t
		}
	}

	if target.Not != nil {
		t, err := expandSchema(*target.Not, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.Not = *t
		}
	}

	for k := range target.Properties {
		t, err := expandSchema(target.Properties[k], parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			target.Properties[k] = *t
		}
	}

	if target.AdditionalProperties != nil && target.AdditionalProperties.Schema != nil {
		t, err := expandSchema(*target.AdditionalProperties.Schema, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalProperties.Schema = *t
		}
	}

	for k := range target.PatternProperties {
		if target.PatternProperties[k].Schema != nil {
			t, err := expandSchema(*target.PatternProperties[k].Schema, parentRefs, resolver, basePath)
			if resolver.shouldStopOnError(err) {
				return &target, err
			}
			if t != nil {
				v := target.PatternProperties[k]
				v.Schema = t
				target.PatternProperties[k] = v
			}
		}
	}

	for k := range target.Dependencies {
		if target.Dependencies[k].Schema != nil {
			t, err := expandSchema(*target.Dependencies[k].Schema, parentRefs, resolver, basePath)
			if resolver.shouldStopOnError(err) {
				return &target, err
			}
			if t != nil {
				*target.Dependencies[k].Schema = *t
			}
		}
	}

	if target.AdditionalItems != nil && target.AdditionalItems.Schema != nil {
		t, err := expandSchema(*target.AdditionalItems.Schema, parentRefs, resolver, basePath)
		if resolver.shouldStopOnError(err) {
			return &target, err
		}
		if t != nil {
			*target.AdditionalItems.Schema = *t
		}
	}
	return &target, nil
}

func expandSchemaRef(target Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*Schema, error) {
	// if a Ref is found, all sibling fields are skipped
	// Ref also changes the resolution scope of children expandSchema

	// here the resolution scope is changed because a $ref was encountered
	normalizedRef := normalizeRef(&target.Ref, basePath)
	normalizedBasePath := normalizedRef.RemoteURI()

	if resolver.isCircular(normalizedRef, basePath, parentRefs...) {
		// this means there is a cycle in the recursion tree: return the Ref
		// - circular refs cannot be expanded. We leave them as ref.
		// - denormalization means that a new local file ref is set relative to the original basePath
		debugLog("short circuit circular ref: basePath: %s, normalizedPath: %s, normalized ref: %s",
			basePath, normalizedBasePath, normalizedRef.String())
		if !resolver.options.AbsoluteCircularRef {
			target.Ref = denormalizeRef(normalizedRef, resolver.context.basePath, resolver.context.rootID)
		} else {
			target.Ref = *normalizedRef
		}
		return &target, nil
	}

	var t *Schema
	err := resolver.Resolve(&target.Ref, &t, basePath)
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

	return expandSchema(*t, parentRefs, transitiveResolver, basePath)
}

// expandContentSchema expands a schema from Content (OpenAPI 3.x).
// It handles both schemas with a root $ref and inline schemas with nested refs.
// Note: Content schemas are always expanded regardless of SkipSchemas option,
// because in OpenAPI 3.x the schema structure in responses/parameters is different
// from Swagger 2.0, and partial expansion would leave inconsistent results.
func expandContentSchema(target Schema, parentRefs []string, resolver *schemaLoader, basePath string) (*Schema, error) {
	if target.Ref.String() != "" {
		// Schema has a root $ref - always expand it (ignore SkipSchemas for Content schemas)
		return expandSchemaRef(target, parentRefs, resolver, basePath)
	}
	// Schema is inline (no root $ref) - expand nested refs
	return expandSchema(target, parentRefs, resolver, basePath)
}

func expandPathItem(pathItem *PathItem, resolver *schemaLoader, basePath string) error {
	if pathItem == nil {
		return nil
	}

	parentRefs := make([]string, 0, smallPrealloc)
	if err := resolver.deref(pathItem, parentRefs, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	if pathItem.Ref.String() != "" {
		transitiveResolver := resolver.transitiveResolver(basePath, pathItem.Ref)
		basePath = transitiveResolver.updateBasePath(resolver, basePath)
		resolver = transitiveResolver
	}

	pathItem.Ref = Ref{}
	for i := range pathItem.Parameters {
		if err := expandParameterOrResponse(&(pathItem.Parameters[i]), resolver, basePath); resolver.shouldStopOnError(err) {
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
	for _, op := range ops {
		if err := expandOperation(op, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
	}

	return nil
}

func expandOperation(op *Operation, resolver *schemaLoader, basePath string) error {
	if op == nil {
		return nil
	}

	for i := range op.Parameters {
		param := op.Parameters[i]
		if err := expandParameterOrResponse(&param, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
		op.Parameters[i] = param
	}

	// Handle OpenAPI 3.x RequestBody
	if op.RequestBody != nil {
		if err := expandRequestBody(op.RequestBody, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
	}

	if op.Responses == nil {
		return nil
	}

	responses := op.Responses
	if err := expandParameterOrResponse(responses.Default, resolver, basePath); resolver.shouldStopOnError(err) {
		return err
	}

	for code := range responses.StatusCodeResponses {
		response := responses.StatusCodeResponses[code]
		if err := expandParameterOrResponse(&response, resolver, basePath); resolver.shouldStopOnError(err) {
			return err
		}
		responses.StatusCodeResponses[code] = response
	}

	return nil
}

// expandRequestBody expands a RequestBody and its content schemas
func expandRequestBody(requestBody *RequestBody, resolver *schemaLoader, basePath string) error {
	if requestBody == nil {
		return nil
	}

	parentRefs := make([]string, 0, smallPrealloc)

	// Handle $ref on the RequestBody itself
	if requestBody.Ref.String() != "" {
		if err := resolver.deref(requestBody, parentRefs, basePath); resolver.shouldStopOnError(err) {
			return err
		}

		if requestBody.Ref.String() != "" {
			transitiveResolver := resolver.transitiveResolver(basePath, requestBody.Ref)
			basePath = resolver.updateBasePath(transitiveResolver, basePath)
			resolver = transitiveResolver
		}
		requestBody.Ref = Ref{}
	}

	// Expand schemas in Content
	if requestBody.Content != nil {
		for mediaType, mediaTypeObj := range requestBody.Content {
			if mediaTypeObj.Schema != nil {
				sch, err := expandContentSchema(*mediaTypeObj.Schema, parentRefs, resolver, basePath)
				if resolver.shouldStopOnError(err) {
					return err
				}
				if sch != nil {
					mediaTypeObj.Schema = sch
					requestBody.Content[mediaType] = mediaTypeObj
				}
			}
		}
	}

	return nil
}

// ExpandResponseWithRoot expands a response based on a root document, not a fetchable document
//
// Notice that it is impossible to reference a json schema in a different document other than root
// (use ExpandResponse to resolve external references).
//
// Setting the cache is optional and this parameter may safely be left to nil.
func ExpandResponseWithRoot(response *Response, root any, cache ResolutionCache) error {
	cache = cacheOrDefault(cache)
	opts := &ExpandOptions{
		RelativeBase: baseForRoot(root, cache),
	}
	resolver := defaultSchemaLoader(root, opts, cache, nil)

	return expandParameterOrResponse(response, resolver, opts.RelativeBase)
}

// ExpandResponse expands a response based on a basepath
//
// All refs inside response will be resolved relative to basePath
func ExpandResponse(response *Response, basePath string) error {
	opts := optionsOrDefault(&ExpandOptions{
		RelativeBase: basePath,
	})
	resolver := defaultSchemaLoader(nil, opts, nil, nil)

	return expandParameterOrResponse(response, resolver, opts.RelativeBase)
}

// ExpandParameterWithRoot expands a parameter based on a root document, not a fetchable document.
//
// Notice that it is impossible to reference a json schema in a different document other than root
// (use ExpandParameter to resolve external references).
func ExpandParameterWithRoot(parameter *Parameter, root any, cache ResolutionCache) error {
	cache = cacheOrDefault(cache)

	opts := &ExpandOptions{
		RelativeBase: baseForRoot(root, cache),
	}
	resolver := defaultSchemaLoader(root, opts, cache, nil)

	return expandParameterOrResponse(parameter, resolver, opts.RelativeBase)
}

// ExpandParameter expands a parameter based on a basepath.
// This is the exported version of expandParameter
// all refs inside parameter will be resolved relative to basePath
func ExpandParameter(parameter *Parameter, basePath string) error {
	opts := optionsOrDefault(&ExpandOptions{
		RelativeBase: basePath,
	})
	resolver := defaultSchemaLoader(nil, opts, nil, nil)

	return expandParameterOrResponse(parameter, resolver, opts.RelativeBase)
}

func getRefAndSchema(input any) (*Ref, *Schema, error) {
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
		// For OpenAPI v3, we'll need to handle Content separately
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

func expandParameterOrResponse(input any, resolver *schemaLoader, basePath string) error {
	ref, sch, err := getRefAndSchema(input)
	if err != nil {
		return err
	}

	if ref == nil && sch == nil { // nothing to do
		return nil
	}

	parentRefs := make([]string, 0, smallPrealloc)
	if ref != nil {
		// dereference this $ref
		if err = resolver.deref(input, parentRefs, basePath); resolver.shouldStopOnError(err) {
			return err
		}

		ref, sch, _ = getRefAndSchema(input)
	}

	if ref.String() != "" {
		transitiveResolver := resolver.transitiveResolver(basePath, *ref)
		basePath = resolver.updateBasePath(transitiveResolver, basePath)
		resolver = transitiveResolver
	}

	if sch == nil {
		// For OpenAPI v3, also expand schemas in Content (for both Response and Parameter)
		if resp, ok := input.(*Response); ok && resp != nil && resp.Content != nil {
			for mediaType, mediaTypeObj := range resp.Content {
				if mediaTypeObj.Schema != nil {
					sch, err := expandContentSchema(*mediaTypeObj.Schema, parentRefs, resolver, basePath)
					if resolver.shouldStopOnError(err) {
						return err
					}
					if sch != nil {
						mediaTypeObj.Schema = sch
						resp.Content[mediaType] = mediaTypeObj
					}
				}
			}
		}

		if param, ok := input.(*Parameter); ok && param != nil && param.Content != nil {
			for mediaType, mediaTypeObj := range param.Content {
				if mediaTypeObj.Schema != nil {
					sch, err := expandContentSchema(*mediaTypeObj.Schema, parentRefs, resolver, basePath)
					if resolver.shouldStopOnError(err) {
						return err
					}
					if sch != nil {
						mediaTypeObj.Schema = sch
						param.Content[mediaType] = mediaTypeObj
					}
				}
			}
		}

		// nothing to be expanded
		if ref != nil {
			*ref = Ref{}
		}

		return nil
	}

	if sch.Ref.String() != "" {
		rebasedRef, ern := NewRef(normalizeURI(sch.Ref.String(), basePath))
		if ern != nil {
			return ern
		}

		if resolver.isCircular(&rebasedRef, basePath, parentRefs...) {
			// this is a circular $ref: stop expansion
			if !resolver.options.AbsoluteCircularRef {
				sch.Ref = denormalizeRef(&rebasedRef, resolver.context.basePath, resolver.context.rootID)
			} else {
				sch.Ref = rebasedRef
			}
		}
	}

	// $ref expansion or rebasing is performed by expandSchema below
	if ref != nil {
		*ref = Ref{}
	}

	// expand schema
	// yes, we do it even if options.SkipSchema is true: we have to go down that rabbit hole and rebase nested $ref)
	s, err := expandSchema(*sch, parentRefs, resolver, basePath)
	if resolver.shouldStopOnError(err) {
		return err
	}

	if s != nil { // guard for when continuing on error
		*sch = *s
	}

	// For v3, also expand schemas in Content (for both Response and Parameter)
	if resp, ok := input.(*Response); ok && resp != nil && resp.Content != nil {
		for mediaType, mediaTypeObj := range resp.Content {
			if mediaTypeObj.Schema != nil {
				sch, err := expandContentSchema(*mediaTypeObj.Schema, parentRefs, resolver, basePath)
				if resolver.shouldStopOnError(err) {
					return err
				}
				if sch != nil {
					mediaTypeObj.Schema = sch
					resp.Content[mediaType] = mediaTypeObj
				}
			}
		}
	}

	if param, ok := input.(*Parameter); ok && param != nil && param.Content != nil {
		for mediaType, mediaTypeObj := range param.Content {
			if mediaTypeObj.Schema != nil {
				sch, err := expandContentSchema(*mediaTypeObj.Schema, parentRefs, resolver, basePath)
				if resolver.shouldStopOnError(err) {
					return err
				}
				if sch != nil {
					mediaTypeObj.Schema = sch
					param.Content[mediaType] = mediaTypeObj
				}
			}
		}
	}

	return nil
}
