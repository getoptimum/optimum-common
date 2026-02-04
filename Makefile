COVERAGE_THRESHOLD := 80
COVERPROFILE := coverage.out
FUZZ_TIME ?= 30s
SHELL := /bin/bash

all: check ## run all checks

check: tidy fmt vet test coverage lint vulcheck ## Run tests, coverage gate, lints, vuln check
	@echo "all checks passed"

ci: lint test coverage vulcheck ## Run CI checks (matches GitHub Actions)
	@echo "CI checks passed"

tidy: ## go mod tidy + verify
	@go mod tidy
	@go mod verify

fmt: ## go fmt packages
	@go fmt ./...

vet: ## go vet
	@echo "Running go vet..."
	@go vet ./...

test: ## Run all tests with race detector and coverage
	@echo "Running tests..."
	@go test ./... -race -covermode=atomic -coverprofile=$(COVERPROFILE)
	@grep -v 'test_utils/' $(COVERPROFILE) > $(COVERPROFILE).tmp && mv $(COVERPROFILE).tmp $(COVERPROFILE)

bench: ## Run benchmarks w/o tests
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem -run=^$

fuzz: ## Run fuzz tests (default 30s per test, override with FUZZ_TIME=10s)
	@echo "Running fuzz tests ($(FUZZ_TIME) per test)..."
	@echo "Fuzzing FuzzHashing..."
	@go test -run='^$$' -fuzz=FuzzHashing -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzMultiAddressBuilder..."
	@go test -run='^$$' -fuzz=FuzzMultiAddressBuilder -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzIPClassification..."
	@go test -run='^$$' -fuzz=FuzzIPClassification -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzSortAddresses..."
	@go test -run='^$$' -fuzz=FuzzSortAddresses -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzMathConversions..."
	@go test -run='^$$' -fuzz=FuzzMathConversions -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzSafeAddUint64Ptr..."
	@go test -run='^$$' -fuzz=FuzzSafeAddUint64Ptr -fuzztime=$(FUZZ_TIME) ./pkg/utils/
	@echo "Fuzzing FuzzConfigLoad..."
	@go test -run='^$$' -fuzz=FuzzConfigLoad -fuzztime=$(FUZZ_TIME) ./pkg/config/
	@echo "Fuzz tests completed."

coverage: $(COVERPROFILE) ## Show summary and enforce threshold
	@echo "Coverage summary"
	@go tool cover -func=$(COVERPROFILE) | tail -1
	@total=$$(go tool cover -func=$(COVERPROFILE) | tail -1 | awk '{print $$3}' | tr -d '%'); \
	 echo "Threshold:            $(COVERAGE_THRESHOLD)%"; \
	 echo "Current test coverage: $$total%"; \
	 awk 'BEGIN{exit !('"$${total:-0}"' >= '$(COVERAGE_THRESHOLD)')}'; \
	 if [ $$? -ne 0 ]; then echo "ERROR: Test coverage is lower than threshold"; exit 1; fi

coverhtml: $(COVERPROFILE) ## Generate HTML coverage report
	@go tool cover -html=$(COVERPROFILE) -o coverage.html
	@echo "Open coverage.html to view the report"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@go tool golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "Running govulncheck..."
	@go version
	@go env GOPRIVATE
	@set -euo pipefail; \
	 IGNORED_VULNS=$$(awk '/- id:/ {print $$3}' govulncheck.yaml | paste -sd '|' - || echo "^$$"); \
	 if [ -z "$$IGNORED_VULNS" ] || [ "$$IGNORED_VULNS" = "^$$" ]; then \
	   echo "No ignored vulnerabilities configured in govulncheck.yaml"; \
	   IGNORED_VULNS="^$$"; \
	 else \
	   echo "Ignored vulnerabilities from govulncheck.yaml: $$IGNORED_VULNS"; \
	 fi; \
	 echo "Running govulncheck..."; \
	 set +e; \
	 go tool govulncheck -show verbose ./... 2>&1 | tee /tmp/govulncheck-output.txt; \
	 govulncheck_exit=$$?; \
	 set -e; \
	 found_vulns=$$(grep -o 'GO-[0-9]\{4\}-[0-9]\+' /tmp/govulncheck-output.txt | sort -u || true); \
	 echo "Found vulnerabilities: $$found_vulns"; \
	 echo "Govulncheck exit code: $$govulncheck_exit"; \
	 if [ $$govulncheck_exit -ne 0 ] && [ -z "$$found_vulns" ]; then \
	   echo "ERROR: govulncheck failed with exit code $$govulncheck_exit (no vulnerabilities detected in output)"; \
	   exit $$govulncheck_exit; \
	 fi; \
	 if [ -z "$$found_vulns" ]; then \
	   echo "No vulnerabilities found"; \
	   exit 0; \
	 fi; \
	 non_ignored=$$(echo "$$found_vulns" | grep -Ev "$$IGNORED_VULNS" || true); \
	 echo "Non-ignored vulnerabilities: $$non_ignored"; \
	 if [ -z "$$non_ignored" ]; then \
	   echo "Only accepted vulnerabilities found: $$found_vulns"; \
	   echo "   (Documented in govulncheck.yaml as accepted risks)"; \
	   exit 0; \
	 else \
	   echo "ERROR: Found non-accepted vulnerabilities: $$non_ignored"; \
	   echo "   Please fix or document these in govulncheck.yaml"; \
	   exit 1; \
	 fi

clean: ## Remove generated artifacts
	@rm -f $(COVERPROFILE) coverage.html

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

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
	@go tool goreleaser release --clean

# guard target
$(COVERPROFILE):
	@echo "No $(COVERPROFILE) found; run 'make test' first." >&2; exit 1

docs: ## Generate documentation from Go code comments
	@echo "Generating documentation..."
	@go run scripts/generate-docs.go

docs-check: docs ## Check if documentation is up-to-date
	@git diff --exit-code docs/ || (echo "Documentation is out of date. Run 'make docs' to regenerate." && exit 1)

.PHONY: coverage coverhtml lint vulcheck check ci all help tidy fmt vet clean tag-rc release bench test fuzz docs docs-check
.DEFAULT_GOAL := help
