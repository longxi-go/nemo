.DEFAULT_GOAL := help

GO ?= go
GOLANGCI_LINT ?= golangci-lint
override export GOWORK := off
override export GOTOOLCHAIN := local
MODULE := github.com/longxi-go/nemo
VERSION := $(shell GOWORK=off GOTOOLCHAIN=local $(GO) list -m -f '{{.Version}}' $(MODULE))
COMPILER := $(MODULE)/cmd/nemoctl@$(VERSION)

.PHONY: help generate check-generated format check-format test race vet lint verify-mod check run

help:
	@printf '%s\n' 'Usage: make <target>' '' \
	  '  generate          Regenerate adapters with the public matching compiler' \
	  '  check-generated   Check adapters without rewriting source' \
	  '  format            Apply the parent repository shared formatter' \
	  '  check-format      Check shared formatting and source notices' \
	  '  test              Run example tests with statement coverage' \
	  '  race              Run shuffled race tests' \
	  '  vet               Run go vet' \
	  '  lint              Run the repository linter configuration' \
	  '  verify-mod        Check standalone module tidiness and checksums' \
	  '  check             Run all nested-module quality gates' \
	  '  run               Dispatch a real Gin request without opening a listener'

generate:
	$(GO) run $(COMPILER) generate --roots .
check-generated:
	$(GO) run $(COMPILER) check --roots .
format:
	$(MAKE) -C ../.. format
check-format:
	$(MAKE) -C ../.. check-format check-license-headers
test:
	$(GO) test -cover ./...
race:
	$(GO) test -race -shuffle=on ./...
vet:
	$(GO) vet ./...
lint:
	$(GOLANGCI_LINT) run --config ../../.golangci.yml --timeout=10m
verify-mod:
	$(GO) mod tidy -diff
	$(GO) mod verify
check: check-format verify-mod vet test race lint check-generated
run:
	$(GO) run .

.DEFAULT:
	@printf 'Unknown target: %s. Run make help for available targets.\n' "$@" >&2
	@exit 2
