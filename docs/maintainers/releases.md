# Releases

[English](releases.md) | [简体中文](releases.zh_CN.md)

```mermaid
flowchart LR
    A[Review source and notes] --> B[Check dependencies and Pages settings]
    B --> C[Push reviewed main commit]
    C --> D[Quality succeeds]
    D --> E[Push annotated vX.Y.Z tag]
    E --> F[Check tagged source]
    F --> G[Publish GitHub Release]
    G --> H[Deploy released documentation]
    H --> I[Verify public installation and site]
```

## Preparation

Use a stable `vX.Y.Z` tag and maintain its English notes in
`docs/releases/vX.Y.Z.md`, headed `# nemo vX.Y.Z`. The Go directive in `go.mod`
is the language minimum; the tag determines the module version. A v2 release
requires a versioned module path and a corresponding release-policy update.

The module must resolve through published dependencies without workspace or
module replacements. `gosr` and `gosr/tools` have independent tags; the latter's
module version `vX.Y.Z` corresponds to its repository tag `tools/vX.Y.Z`.

Run from the repository root with Go, Python 3.12, Bash, Git, uv,
golangci-lint and an authenticated GitHub CLI:

```sh
make setup
make check docs-build build
make preflight-release TAG=vX.Y.Z
```

The preflight validates notes and module identity, checks stable upstream
Releases and reads GitHub repository settings. Go's module checks independently
verify dependency checksums. The preflight does not change remote settings.

In **Settings → Pages**, select **GitHub Actions**. In **Settings → Environments →
github-pages → Deployment branches and tags**, permit root release tags with a
`v*` **tag** rule. A branch rule named `v*` does not permit tags. Keep any other
reviewed environment protection rules. These settings belong to each repository;
copying workflow files does not copy them.

## Publication and verification

Push the reviewed commit and wait for its Quality workflow to pass before
creating an annotated tag. The Release workflow validates and checks the exact
tagged source, publishes the maintained notes as a stable Release. GitHub selects Latest
using the explicit `make_latest=legacy` policy (creation date and semantic
version); only a release selected as Latest calls Pages. Pages runs serially and
checks Latest again before deployment, refusing a release superseded during its build. Pages checks out that tag and requires its published
stable Release. Branch and pull-request workflows build documentation without
deploying it. A manual Pages deployment also requires the explicit current Latest stable tag.

After the workflow succeeds, verify both documentation languages and test an
independent consumer with `GOWORK=off`, no `replace` directives and a fresh module
cache. Check `go get github.com/longxi-go/nemo@vX.Y.Z`, runtime configuration
loading, generated binding, and `go install github.com/longxi-go/nemo/cmd/nemoctl@vX.Y.Z`.
Run `go mod verify` with checksum verification enabled. A successful source test
or GitHub Release does not establish public module availability.

Never move or reuse a published tag. A failed deployment caused by repository
settings can be retried after correcting those settings. A source correction
requires a new release version.
