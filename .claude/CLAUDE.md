# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go types modeling the [Swagger 2.0 / OpenAPI 2.0](https://swagger.io/specification/v2/)
specification. Every object in the spec --- `Swagger`, `Info`, `PathItem`, `Operation`,
`Parameter`, `Schema`, `Response`, `Header`, `SecurityScheme`, etc. --- has a corresponding
Go struct with JSON serialization (`encoding/json`) that round-trips through the spec's
JSON representation.

This package is the **foundational data model** for the
[go-swagger](https://github.com/go-swagger/go-swagger) ecosystem. Higher-level packages
(`analysis`, `loads`, `validate`, `runtime`) consume these types to load, analyze, validate,
and serve Swagger specifications. Because it sits at the bottom of the dependency graph,
changes here ripple through the entire ecosystem.

Key capabilities beyond plain structs:

- **`$ref` resolution** --- the `Ref` type wraps JSON Reference pointers; the `expander`
  resolves `$ref` nodes (local, remote, circular) into fully expanded specs.
- **Schema composition** --- `Schema` supports `allOf`, `additionalProperties`,
  `additionalItems`, and JSON Schema validations (`minimum`, `pattern`, `enum`, etc.).
- **URL normalization** --- cross-platform path/URL normalization for `$ref` targets.
- **Embedded spec** --- a copy of the Swagger 2.0 JSON Schema is embedded via `go:embed`
  for offline use.

See [docs/MAINTAINERS.md](../docs/MAINTAINERS.md) for CI/CD, release process, and repo structure details.

### Package layout (single package)

| File | Contents |
|------|----------|
| `swagger.go` | Root `Swagger` type (top-level spec object) |
| `info.go` | `Info`, `ContactInfo`, `LicenseInfo` |
| `paths.go` | `Paths` (map of path patterns to `PathItem`) |
| `path_item.go` | `PathItem` (GET/PUT/POST/DELETE/... operations per path) |
| `operation.go` | `Operation` (single API operation) |
| `parameter.go` | `Parameter` (query, header, path, body, formData) |
| `header.go` | `Header` |
| `response.go`, `responses.go` | `Response`, `Responses` |
| `schema.go` | `Schema` (JSON Schema subset used by Swagger) |
| `security_scheme.go` | `SecurityScheme` |
| `items.go` | `Items` (non-body parameter schema) |
| `ref.go` | `Ref` type, JSON Reference (`$ref`) handling |
| `expander.go` | `$ref` expansion / resolution engine |
| `normalizer.go` | URL/path normalization (platform-specific variants) |
| `cache.go` | Resolution cache for expanded specs |
| `validations.go` | Common validation properties shared across types |
| `properties.go` | `SchemaProperties` ordered map |
| `embed.go` | Embedded Swagger 2.0 JSON Schema (`go:embed`) |
| `spec.go` | `MustLoadSwagger20Schema()` loader |
| `external_docs.go` | `ExternalDocumentation` |
| `tag.go` | `Tag` |
| `xml_object.go` | `XMLObject` |
| `debug.go` | Debug logging helpers |

### Key API

- `Swagger` --- root specification object; deserialize with `json.Unmarshal`
- `Schema` --- JSON Schema with Swagger extensions; supports `allOf`, `$ref`, validations
- `Ref` / `MustCreateRef(uri)` --- JSON Reference wrapper
- `ExpandSpec(spec, opts)` --- resolve all `$ref` nodes in a specification
- `ExpandSchema(schema, root, cache)` --- resolve `$ref` nodes in a single schema
- `ResolveRef(root, ref)` / `ResolveParameter` / `ResolveResponse` --- targeted resolution

### Dependencies

- `github.com/go-openapi/jsonpointer` --- JSON Pointer (RFC 6901) navigation
- `github.com/go-openapi/jsonreference` --- JSON Reference parsing
- `github.com/go-openapi/swag` --- JSON/YAML utilities, name mangling
- `github.com/go-openapi/testify/v2` --- test-only assertions (zero-dep testify fork)

### Notable historical design decisions

- **Mixin of spec types and `$ref`** --- many types embed both their data fields and a `Ref`
  field. When `$ref` is present, the data fields are ignored per the Swagger specification.
  This is modeled by custom `MarshalJSON`/`UnmarshalJSON` on each type.
- **`VendorExtensible`** --- most types embed `VendorExtensible` to capture `x-` extension
  fields as `map[string]any`.
- **`SchemaProperties` as ordered slice** --- schema properties are stored as a slice of
  key-value pairs (not a map) to preserve declaration order during round-trip serialization.
- **Platform-specific normalization** --- Windows path handling differs from Unix; separate
  `normalizer_windows.go` / `normalizer_nonwindows.go` files handle this.
