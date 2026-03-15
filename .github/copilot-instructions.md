# Copilot Instructions — spec

## Project Overview

Go types modeling the Swagger 2.0 / OpenAPI 2.0 specification. This package is the
foundational data model for the go-swagger ecosystem — every specification object
(`Swagger`, `Schema`, `Operation`, `Parameter`, etc.) is a Go struct with JSON
round-trip serialization. It also includes a `$ref` expansion engine for resolving
JSON References across local and remote documents.

Single module: `github.com/go-openapi/spec`.

### Package layout (single package)

| File | Contents |
|------|----------|
| `swagger.go` | Root `Swagger` type (top-level spec object) |
| `schema.go` | `Schema` (JSON Schema subset used by Swagger) |
| `operation.go` | `Operation` (single API operation) |
| `parameter.go` | `Parameter` (query, header, path, body, formData) |
| `response.go`, `responses.go` | `Response`, `Responses` |
| `ref.go` | `Ref` type, JSON Reference (`$ref`) handling |
| `expander.go` | `$ref` expansion / resolution engine |
| `normalizer.go` | URL/path normalization (platform-specific variants) |

### Key API

- `Swagger` — root specification object; deserialize with `json.Unmarshal`
- `Schema` — JSON Schema with Swagger extensions; supports `allOf`, `$ref`, validations
- `Ref` / `MustCreateRef(uri)` — JSON Reference wrapper
- `ExpandSpec(spec, opts)` — resolve all `$ref` nodes in a specification
- `ExpandSchema(schema, root, cache)` — resolve `$ref` nodes in a single schema
- `ResolveRef(root, ref)` / `ResolveParameter` / `ResolveResponse` — targeted resolution

### Dependencies

- `github.com/go-openapi/jsonpointer` — JSON Pointer (RFC 6901) navigation
- `github.com/go-openapi/jsonreference` — JSON Reference parsing
- `github.com/go-openapi/swag` — JSON/YAML utilities, name mangling
- `github.com/go-openapi/testify/v2` — test-only assertions (zero-dep testify fork)

## Building & testing

```sh
go test ./...
```

## Conventions

Coding conventions are found beneath `.github/copilot`

### Summary

- All `.go` files must have SPDX license headers (Apache-2.0).
- Commits require DCO sign-off (`git commit -s`).
- Linting: `golangci-lint run` — config in `.golangci.yml` (posture: `default: all` with explicit disables).
- Every `//nolint` directive **must** have an inline comment explaining why.
- Tests: `go test ./...`. CI runs on `{ubuntu, macos, windows} x {stable, oldstable}` with `-race`.
- Test framework: `github.com/go-openapi/testify/v2` (not `stretchr/testify`; `testifylint` does not work).

See `.github/copilot/` (symlinked to `.claude/rules/`) for detailed rules on Go conventions, linting, testing, and contributions.
