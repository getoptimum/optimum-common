GO := go
COVERAGE_THRESHOLD := 70
COVERPROFILE    := coverage.out

all: check ## run all checks

check: tidy fmt vet test coverage lint vulcheck ## Run tests, coverage gate, lints, vuln check
	@echo "✅ all checks passed"

tidy: ## go mod tidy + verify
	@$(GO) mod tidy
	@$(GO) mod verify

fmt: ## go fmt packages
	@$(GO) fmt ./...

vet: ## go vet
	@echo "🔍 go vet..."
	@$(GO) vet ./...

test: ## Run all tests with race detector and coverage
	@echo "🧪 running tests..."
	@$(GO) test ./... -race -covermode=atomic -coverprofile=$(COVERPROFILE)

coverage: $(COVERPROFILE) ## Show summary and enforce threshold
	@echo "📈 coverage summary"
	@$(GO) tool cover -func=$(COVERPROFILE) | tail -1
	@total=$$($(GO) tool cover -func=$(COVERPROFILE) | tail -1 | awk '{print $$3}' | tr -d '%'); \
	 echo "Threshold:            $(COVERAGE_THRESHOLD)%"; \
	 echo "Current test coverage: $$total%"; \
	 awk 'BEGIN{exit !('"$${total:-0}"' >= '$(COVERAGE_THRESHOLD)')}'; \
	 if [ $$? -ne 0 ]; then echo "❌ Test coverage is lower than threshold"; exit 1; fi


coverhtml: $(COVERPROFILE) ## Generate HTML coverage report
	@$(GO) tool cover -html=$(COVERPROFILE) -o coverage.html
	@echo "📊 open coverage.html to view the report"

lint: ## Run golangci-lint
	@echo "🧹 linting..."
	@golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "🔍 govulncheck..."
	@$(GO) version
	@$(GO) env GOPRIVATE
	@$(GOPATH)/bin/govulncheck ./... || govulncheck ./...

# NOTE: run before running vulcheck/lint commands locally
tools: ## Install dev/CI tools (govulncheck, golangci-lint)
	@echo "🔧 installing tools (optional locally)"
	@go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.1

clean: ## Remove generated artifacts
	@rm -f $(COVERPROFILE) coverage.html

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

$(COVERPROFILE):
	@echo "No $(COVERPROFILE) found; run 'make test' first." >&2; exit 1

.PHONY: test coverage lint vulcheck tools check all help
.DEFAULT_GOAL := help