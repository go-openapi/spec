# Maintainer's guide

## Repo structure

Single go module.

> **NOTE**
>
> Some `go-openapi` repos are mono-repos with multiple modules,
> with adapted CI workflows.

## Repo configuration

* default branch: master
* protected branches: master
* branch protection rules:
  * require pull requests and approval
  * required status checks: 
    - DCO (simple email sign-off)
    - Lint
    - tests completed
* auto-merge enabled (used for dependabot updates)

## Continuous Integration

### Code Quality checks

* meta-linter: golangci-lint
* linter config: [`.golangci.yml`](../.golangci.yml) (see our [posture](./STYLE.md) on linters)

* Code quality assessment: [CodeFactor](https://www.codefactor.io/dashboard)
* Code quality badges
  * go report card: <https://goreportcard.com/>
  * CodeFactor: <https://goreportcard.com/>

> **NOTES**
>
> codefactor inherits roles from github. There is no need to create a dedicated account.
>
> The codefactor app is installed at the organization level (`github.com/go-openapi`).
>
> There is no special token to setup in github for CI usage.

### Testing

* Test reports
  * Uploaded to codecov: <https://app.codecov.io/analytics/gh/go-openapi>
* Test coverage reports
  * Uploaded to codecov: <https://app.codecov.io/gh/go-openapi>

* Fuzz testing
  * Fuzz tests are handled separately by CI and may reuse a cached version of the fuzzing corpus.
    At this moment, cache may not be shared between feature branches or feature branch and master.
    The minimized corpus produced on failure is uploaded as an artifact and should be added manually
    to `testdata/fuzz/...`.

Coverage threshold status is informative and not blocking.
This is because the thresholds are difficult to tune and codecov oftentimes reports false negatives
or may fail to upload coverage.

All tests use our fork of `stretchr/testify`: `github.com/go-openapi/testify`.
This allows for minimal test dependencies.

> **NOTES**
>
> codecov inherits roles from github. There is no need to create a dedicated account.
> However, there is only 1 maintainer allowed to be the admin of the organization on codecov
> with their free plan.
>
> The codecov app is installed at the organization level (`github.com/go-openapi`).
>
> There is no special token to setup in github for CI usage.
> A organization-level token used to upload coverage and test reports is managed at codecov:
> no setup is required on github.

### Automated updates

* dependabot
  * configuration: [`dependabot.yaml`](../.github/dependabot.yaml)

  Principle:

  * codecov applies updates and security patches to the github-actions and golang ecosystems.
  * all updates from "trusted" dependencies (github actions, golang.org packages, go-openapi packages
    are auto-merged if they successfully pass CI.

* go version udpates

  Principle:

  * we support the 2 latest minor versions of the go compiler (`stable`, `oldstable`)
  * `go.mod` should be updated (manually) whenever there is a new go minor release
    (e.g. every 6 months).

* contributors
  * a [`CONTRIBUTORS.md`](../CONTRIBUTORS.md) file is updated weekly, with all-time contributors to the repository
  * the `github-actions[bot]` posts a pull request to do that automatically
  * at this moment, this pull request is not auto-approved/auto-merged (bot cannot approve its own PRs)

### Vulnerability scanners

There are 3 complementary scanners - obviously, there is some overlap, but each has a different focus.

* github `CodeQL`
* `trivy` <https://trivy.dev/docs/latest/getting-started>
* `govulnscan` <https://go.dev/blog/govulncheck>

None of these tools require an additional account or token.

Github CodeQL configuration is set to "Advanced", so we may collect a CI status for this check (e.g. for badges).

Scanners run on every commit to master and at least once a week.

Reports are centralized in github security reports for code scanning tools.

## Releases

The release process is minimalist:

* push a semver tag (i.e v{major}.{minor}.{patch}) to the master branch.
* the CI handles this to generate a github release with release notes

* release notes generator: git-cliff <https://git-cliff.org/docs/>
* configuration: [`cliff.toml`](../.cliff.toml)

Tags are preferably PGP-signed.

The tag message introduces the release notes (e.g. a summary of this release).

The release notes generator does not assume that commits are necessarily "conventional commits".

## Other files

Standard documentation:

* [`CONTRIBUTING.md`](../.github/CONTRIBUTING.md) guidelines
* [`DCO.md`](../.github/DCO.md) terms for first-time contributors to read
* [`CODE_OF_CONDUCT.md`](../CODE_OF_CONDUCT.md)
* [`SECURIY.md`](../SECURITY.md) policy: how to report vulnerabilities privately
* [`LICENSE`](../LICENSE) terms
<!--
* [`NOTICE`](../NOTICE) on supplementary license terms (original authors, copied code etc)
-->

Reference documentation (released):

* [godoc](https://pkg.go.dev/github.com/go-openapi/spec)

## TODOs & other ideas

A few things remain ahead to ease a bit a maintainer's job:

* [x] reuse CI workflows (e.g. in `github.com/go-openapi/workflows`)
* [x] reusable actions with custom tools pinned  (e.g. in `github.com/go-openapi/gh-actions`)
* open-source license checks
* [x] auto-merge for CONTRIBUTORS.md (requires a github app to produce tokens)
* [ ] more automated code renovation / relinting work (possibly built with CLAUDE) (ongoing)
* organization-level documentation web site
* ...
