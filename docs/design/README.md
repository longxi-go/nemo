# Documentation site

[English](README.md) | [简体中文](README.zh.md)

The nemo manual uses MkDocs Material with English and Simplified Chinese pages.
Building it requires Python 3.12 and uv; dependencies are locked in `uv.lock`.

## Build and preview

Run from the repository root:

```sh
make docs-setup
make docs-serve
make docs-build
```

`docs-serve` previews English at `http://127.0.0.1:8000/nemo/` and Chinese at
`http://127.0.0.1:8000/nemo/zh/`. `docs-build` runs the strict build and site checks.
To rerun built-site checks directly from this directory:

```sh
uv run --locked python -m unittest discover -s tests
```

Mermaid diagrams render in the browser during preview. Review their rendering
and bilingual labels in addition to running the static checks.

## Source files

| Path within `docs/design` | Purpose |
| --- | --- |
| `mkdocs.yml` | Theme, navigation, repository and language configuration |
| `docs/` | English pages and matching `*.zh.md` translations |
| `uv.lock` | Locked documentation dependencies |
| `tests/` | Built-site regression checks |

Pair page files by basename and add each page to both navigation trees. Generated
`site/` and `.venv/` directories stay untracked. Set `DOCS_SITE_URL` to a URL ending
in `/` to override the base URL. Keep the native Material repository component
linked to `longxi-go/nemo` for repository metadata.

Publication procedures belong in the [maintainer release guide](../maintainers/releases.md).
