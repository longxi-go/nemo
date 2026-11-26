# Development

## Repository tools and quality gates

Development requires Git, Bash, Make, Python 3.12 and the Go toolchain selected by
[the quality workflow](https://github.com/longxi-go/nemo/blob/main/.github/workflows/quality.yml).
`go.mod` declares the minimum Go version. Install the workflow's golangci-lint
version and ensure it is on `PATH`, or set `GOLANGCI_LINT` to its executable.

From the repository root:

```sh
make setup
make license-headers format
make generate
GOWORK=off make check
```

`setup` resolves the pinned golayout tooling without installing a formatter binary
or editing module files. `license-headers` applies `codequality/HEADER` to owned
Go files. `format` applies the shared golayout policy; upstream gofmt or gofumpt
alone does not enforce the complete structural layout.

`check` is the read-only gate for source notices, formatting, formatter integration,
module tidiness, vet, coverage tests, shuffled race tests, golangci-lint and generated
output freshness. It may download tool dependencies. Use `make help` for individual
targets. After generator changes, regenerate the example and check its output
with `make check-generated`.

`make build` writes the development CLI to `.tmp/bin/nemoctl`. Use the public
installation command from [quick start](quick-start.md) when working outside the
checkout.

## Source and fixture conventions

Public APIs belong in `pkg`, implementation in `internal`, and thin executables
in `cmd`. Each owned package includes package documentation, runnable examples
and focused tests. Runtime binding must stay free of reflection and unsafe.

Formatting selects owned eligible Go files and complete parseable
`testdata/**/*.go.tmpl` fixtures. Ordinary testdata, caches, vendor trees and
symlinks are excluded. Header refresh additionally excludes testdata and follows
the copyright owner declared in `codequality/HEADER`.

Generators call golayout's shared formatting API before type validation and output.
The target module owns any optional header template; do not copy nemo's notice
into consumer-generated files. Prefer substantial test programs as readable
per-scenario fixtures, retaining exact malformed syntax where a negative test
requires it.

## Multi-module workspace {#local-workspace}

For joint development, clone the projects being changed into sibling directories.
From their parent, for example:

```sh
go work init ./golayout ./gosr ./gosr/tools ./nemo
```

Use `go work use` to extend an existing workspace. Include `gosr/tools` separately
because it has its own `go.mod`; include other consumers only when needed.
Keep `go.work` and `go.work.sum` local, and keep production modules free of
sibling-directory replacements.

From the Nemo checkout, `make check WORKSPACE=../go.work` verifies the explicitly
selected workspace and applies its module selections to temporary test consumers.
Without `WORKSPACE`, checks use the published dependency graph.
Checks reject process or persistent `GOFLAGS` overrides.

## Documentation

The bilingual site uses Python 3.12 and uv:

```sh
make docs-setup
make docs-serve
make docs-build
```

Preview English at `http://127.0.0.1:8000/nemo/` and Chinese at
`http://127.0.0.1:8000/nemo/zh/`. Sources live in `docs/design/docs`; add matching
`page.md` and `page.zh.md` files and update both navigation trees. Review commands,
API semantics and constraints against source in both languages. Mermaid diagrams
also require browser rendering review; a successful Markdown build does not
validate diagram syntax or readability.

See the [documentation maintenance guide](https://github.com/longxi-go/nemo/blob/main/docs/design/README.md)
for site files and the [maintainer release guide](https://github.com/longxi-go/nemo/blob/main/docs/maintainers/releases.md)
for publication procedures.
