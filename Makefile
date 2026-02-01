GO := go
COVERAGE_THRESHOLD := 80
COVERPROFILE    := coverage.out
VULNCHECK_VER := v1.1.4
VULNCHECK_RUN := $(GO) run golang.org/x/vuln/cmd/govulncheck@$(VULNCHECK_VER)
FUZZ_TIME ?= 30s
SHELL := /bin/bash

all: check ## run all checks

check: tidy fmt vet test coverage lint vulcheck ## Run tests, coverage gate, lints, vuln check
	@echo "all checks passed"

ci: lint test coverage vulcheck ## Run CI checks (matches GitHub Actions)
	@echo "CI checks passed"

tidy: ## go mod tidy + verify
	@$(GO) mod tidy
	@$(GO) mod verify

fmt: ## go fmt packages
	@$(GO) fmt ./...

vet: ## go vet
	@echo "Running go vet..."
	@$(GO) vet ./...

test: ## Run all tests with race detector and coverage
	@echo "Running tests..."
	@$(GO) test ./... -race -covermode=atomic -coverprofile=$(COVERPROFILE)
	@grep -v 'test_utils/' $(COVERPROFILE) > $(COVERPROFILE).tmp && mv $(COVERPROFILE).tmp $(COVERPROFILE) # exclude test_utils from coverage
bench: ## Run benchmarks w/o tests
	@echo "Running benchmarks..."
	@$(GO) test ./... -bench=. -benchmem -run=^$

fuzz: ## Run fuzz tests (default 30s per test, override with FUZZ_TIME=10s)
	@echo "Running fuzz tests ($(FUZZ_TIME) per test)..."
	@echo "Fuzzing FuzzHashing..."
	@$(GO) test -run='^$$' -fuzz=FuzzHashing -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzMultiaddr..."
	@$(GO) test -run='^$$' -fuzz=FuzzMultiaddr -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzMultiAddressBuilder..."
	@$(GO) test -run='^$$' -fuzz=FuzzMultiAddressBuilder -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzIPClassification..."
	@$(GO) test -run='^$$' -fuzz=FuzzIPClassification -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzSortAddresses..."
	@$(GO) test -run='^$$' -fuzz=FuzzSortAddresses -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzMathConversions..."
	@$(GO) test -run='^$$' -fuzz=FuzzMathConversions -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzSafeAddUint64Ptr..."
	@$(GO) test -run='^$$' -fuzz=FuzzSafeAddUint64Ptr -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzConfigLoad..."
	@$(GO) test -run='^$$' -fuzz=FuzzConfigLoad -fuzztime=$(FUZZ_TIME) ./pkg/config/
	@echo "Fuzz tests completed."

coverage: $(COVERPROFILE) ## Show summary and enforce threshold
	@echo "Coverage summary"
	@$(GO) tool cover -func=$(COVERPROFILE) | tail -1
	@total=$$($(GO) tool cover -func=$(COVERPROFILE) | tail -1 | awk '{print $$3}' | tr -d '%'); \
	 echo "Threshold:            $(COVERAGE_THRESHOLD)%"; \
	 echo "Current test coverage: $$total%"; \
	 awk 'BEGIN{exit !('"$${total:-0}"' >= '$(COVERAGE_THRESHOLD)')}'; \
	 if [ $$? -ne 0 ]; then echo "ERROR: Test coverage is lower than threshold"; exit 1; fi


coverhtml: $(COVERPROFILE) ## Generate HTML coverage report
	@$(GO) tool cover -html=$(COVERPROFILE) -o coverage.html
	@echo "Open coverage.html to view the report"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "Running govulncheck..."
	@$(GO) version
	@$(GO) env GOPRIVATE
	@$(VULNCHECK_RUN) ./...

# NOTE: run before running vulcheck/lint commands locally
tools: ## Install dev/CI tools (govulncheck, golangci-lint, goreleaser)
	@echo "Installing tools (optional locally)"
	@go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
	@go install github.com/goreleaser/goreleaser/v2@v2.12.0
clean: ## Remove generated artifacts
	@rm -f $(COVERPROFILE) coverage.html

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
# guard target
$(COVERPROFILE):
	@echo "No $(COVERPROFILE) found; run 'make test' first." >&2; exit 1

tag-rc: ## Tag new release candidate
	@echo "Calculating next RC tag..."
	@set -euo pipefail; \
	latest=$$(git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+-rc[0-9]+$$' | head -n1 || true); \
	if [ -z "$$latest" ]; then \
		base=$$(git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | head -n1 || echo "v0.0.1"); \
		next_tag="$$base-rc1"; \
	else \
		base=$${latest%-rc*}; \
		rc=$${latest##*-rc}; \
		next_rc=$$((rc+1)); \
		next_tag="$$base-rc$$next_rc"; \
	fi; \
	echo "Tagging $$next_tag"; \
	git tag -a "$$next_tag" -m "Release candidate $$next_tag"; \
	git push origin "$$next_tag"

release: ## Create a release with GoReleaser (requires tag)
	@echo "Running goreleaser..."
	@goreleaser release --clean

.PHONY: coverage coverhtml lint vulcheck tools check ci all help tidy fmt vet clean tag-rc release bench test fuzz
.DEFAULT_GOAL := help