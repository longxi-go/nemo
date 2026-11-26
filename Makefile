.DEFAULT_GOAL := help

GO ?= go
UV ?= uv
GOLANGCI_LINT ?= golangci-lint
TAG ?=
REPOSITORY ?= longxi-go/nemo
WORKSPACE ?=
export GO GOLANGCI_LINT
ifeq ($(strip $(WORKSPACE)),)
override export GOWORK := off
override export NEMO_TEST_WORKSPACE :=
else
# Keep directory-based discovery so isolated fixture modules stay independent.
override export GOWORK :=
override export NEMO_TEST_WORKSPACE := $(abspath $(WORKSPACE))
endif
override export GOTOOLCHAIN := local

.PHONY: help setup format check-format test-format test race vet lint verify-mod check generate check-generated build

.PHONY: docs-setup docs-serve docs-build compatibility test-scripts validate-release preflight-release
.PHONY: license-headers check-license-headers verify-environment

help:
	@printf '%s\n' 'Usage: make <target>' '' \
	  '  help              Show this target list (default)' \
	  '  check WORKSPACE=../go.work  Verify explicitly selected local dependencies' \
	  '  license-headers   Refresh owned Go source license headers' \
	  '  check-license-headers  Check source license headers without writes' \
	  '  docs-setup        Install locked documentation dependencies (Python 3.12)' \
	  '  docs-serve        Preview bilingual documentation locally' \
	  '  docs-build        Strictly build both documentation languages' \
	  '  setup             Verify shared golayout tooling' \
	  '  verify-environment Reject hidden Go module/workspace overrides' \
	  '  format            Rewrite owned Go files with shared golayout' \
	  '  check-format      Check shared Go layout without writes' \
	  '  test-format       Test shared layout, ownership and failures' \
	  '  test              Run all tests with statement coverage' \
	  '  race              Run shuffled race tests' \
	  '  vet               Run go vet' \
	  '  lint              Run golangci-lint (must be installed)' \
	  '  verify-mod        Check module dependency tidiness' \
	  '  generate          Regenerate typed configuration example' \
	  '  check-generated   Verify generated output freshness' \
	  '  build             Build nemoctl into .tmp/bin' \
	  '  test-scripts      Test repository release/formatter adapters' \
	  '  compatibility     Run module, vet and test gates on the selected Go version' \
	  '  validate-release  Validate TAG=vX.Y.Z and maintained notes' \
	  '  preflight-release Check module release and GitHub Pages prerequisites' \
	  '  check             Run all required local quality gates'

verify-environment:
	@python3 scripts/go_environment.py
	@if [ -n "$(WORKSPACE)" ]; then \
	  test "$$($(GO) env GOWORK)" = "$(abspath $(WORKSPACE))" || \
	  { printf '%s\n' 'WORKSPACE must match the go.work discovered from this checkout' >&2; exit 2; }; \
	fi

setup format check-format test-format license-headers check-license-headers test race vet lint verify-mod generate check-generated build validate-release: verify-environment

setup:
	@python3 scripts/golayout.py setup
license-headers:
	@python3 scripts/golayout.py headers-write
check-license-headers:
	@python3 scripts/golayout.py headers-check
format:
	@python3 scripts/golayout.py write
check-format:
	@python3 scripts/golayout.py check
test-format:
	@python3 scripts/golayout.py test

test:
	$(GO) test -cover ./...
race:
	$(GO) test -race -shuffle=on ./...
vet:
	$(GO) vet ./...
lint:
	$(GOLANGCI_LINT) run --timeout=10m
verify-mod:
	$(GO) mod tidy -diff
generate:
	$(GO) run ./cmd/nemoctl generate --roots ./examples/config/...
check-generated:
	$(GO) run ./cmd/nemoctl check --roots ./examples/config/...
compatibility: verify-mod vet test

test-scripts:
	python3 -m unittest discover -s scripts/tests -v

validate-release:
	python3 scripts/release.py --tag "$(TAG)"

preflight-release: validate-release verify-mod
	python3 scripts/release.py --tag "$(TAG)" --check-github --repository "$(REPOSITORY)"

check: check-license-headers check-format test-format verify-mod vet test race lint check-generated test-scripts
build:
	$(GO) build -o .tmp/bin/nemoctl ./cmd/nemoctl

docs-setup:
	cd docs/design && $(UV) sync --locked
docs-serve: docs-setup
	cd docs/design && $(UV) run --locked mkdocs serve
docs-build: docs-setup
	cd docs/design && $(UV) run --locked mkdocs build --strict
	cd docs/design && $(UV) run --locked python -m unittest discover -s tests -v

.DEFAULT:
	@printf 'Unknown target: %s. Available targets follow.\n' "$@" >&2
	@$(MAKE) --no-print-directory help
	@exit 2
