# Repository scripts

[English](README.md) | [简体中文](README.zh_CN.md)

Run the Make targets from the repository root with Python 3.12 and a Go toolchain
compatible with `go.mod`. Formatting also requires Git and Bash; GitHub preflight
checks require an authenticated GitHub CLI.

| Script | Responsibility | Targets |
| --- | --- | --- |
| `golayout.py` | Resolve the declared golayout module and delegate source selection, formatting and headers to its shared scripts | `setup`, `format`, `check-format`, `test-format`, `license-headers`, `check-license-headers` |
| `release.py` | Validate stable tags, module identity, dependency versions and notes; optionally inspect Pages settings and upstream Releases | `validate-release`, `preflight-release` |
| `go_environment.py` | Reject process/persistent Go flags that can substitute a different module graph | `verify-environment` |
| `go_module.py` | Parse a module file with `go mod edit -json` without resolving dependencies | Imported helper |

`GO` selects the Go executable. The formatter language is read from this module's
Go directive. `make check` disables workspaces and automatic toolchain switching;
public dependency versions come from `go.mod` and `go.sum`. Nonempty `GOFLAGS`,
including flags in persistent Go configuration, are rejected by verification gates.
Configured proxy and checksum settings are retained.

`make test-scripts` runs Python unit tests. Module metadata tests use real Go.
A real Go regression also verifies that missing checksums fail without changing
`go.mod` or `go.sum`. Tests replace selected subprocess/API boundaries with controlled responses to
exercise malformed data, missing scripts, denied tag policies and unpublished
Releases without network requests. The formatter adapter tests check delegation;
`make test-format` exercises real golayout formatting and source selection.
The GitHub preflight tests do not prove that live repository settings are correct;
run `make preflight-release TAG=vX.Y.Z` for the live check.

See the [release guide](../docs/maintainers/releases.md) for publication order and
public installation verification.
