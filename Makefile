COVERAGE_THRESHOLD := 81
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
	go test -timeout 10m -v ./... -count=1 -race -covermode=atomic -cover -coverprofile=cover.out
	@echo "Remove not critical packages for coverage threshold..."
	@grep -v 'test_utils/' cover.out > cover.out.tmp && mv cover.out.tmp cover.out
	@echo "Calculating total coverage..."
	@COVERAGE_TOTAL=$$(go tool cover -func=cover.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	rm cover.out; \
	echo "Threshold:                $(COVERAGE_THRESHOLD)%"; \
	echo "Current test coverage is: $${COVERAGE_TOTAL}%"; \
	awk 'BEGIN {exit !('"$${COVERAGE_TOTAL}"' >= '$(COVERAGE_THRESHOLD)')}' || { \
		echo "Test coverage is lower than threshold"; \
		exit 1; \
	}

bench: ## Run benchmarks w/o tests
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem -run=^$

fuzz: ## Run fuzz tests (default 30s per test, override with FUZZ_TIME=10s)
	@echo "Running fuzz tests ($(FUZZ_TIME) per test)..."
	@echo "Fuzzing FuzzHashing..."
	@go test -run='^$$' -fuzz=FuzzHashing -fuzztime=$(FUZZ_TIME) ./pkg/hash/
	@echo "Fuzzing FuzzMultiAddressBuilder..."
	@go test -run='^$$' -fuzz=FuzzMultiAddressBuilder -fuzztime=$(FUZZ_TIME) ./pkg/net/
	@echo "Fuzzing FuzzIPClassification..."
	@go test -run='^$$' -fuzz=FuzzIPClassification -fuzztime=$(FUZZ_TIME) ./pkg/net/
	@echo "Fuzzing FuzzSortAddresses..."
	@go test -run='^$$' -fuzz=FuzzSortAddresses -fuzztime=$(FUZZ_TIME) ./pkg/net/
	@echo "Fuzzing FuzzMathConversions..."
	@go test -run='^$$' -fuzz=FuzzMathConversions -fuzztime=$(FUZZ_TIME) ./pkg/math/
	@echo "Fuzzing FuzzSafeAddUint64Ptr..."
	@go test -run='^$$' -fuzz=FuzzSafeAddUint64Ptr -fuzztime=$(FUZZ_TIME) ./pkg/math/
	@echo "Fuzzing FuzzConfigLoad..."
	@go test -run='^$$' -fuzz=FuzzConfigLoad -fuzztime=$(FUZZ_TIME) ./pkg/config/
	@echo "Fuzz tests completed."

lint: ## Run golangci-lint
	@echo "Running linter..."
	@go tool golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "Running govulncheck..."
	@go tool govulncheck ./...

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

release: ## Create a release with GoReleaser (requires tag)
	@echo "Running goreleaser..."
	@go tool goreleaser release --clean

.PHONY: coverage coverhtml lint vulcheck check ci all help tidy fmt vet release bench test fuzz docs docs-check
.DEFAULT_GOAL := help
