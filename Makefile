GO := go # so can subst. with tinygo

test: ## Run all tests with race detector and coverage
	@echo "🧪 running tests..."
	@$(GO) test ./... -race -covermode=atomic -coverprofile=coverage.out

coverage: ## Show coverage summary
	@echo "📈 coverage summary"
	@$(GO) tool cover -func=coverage.out | tail -n1 || true

lint: ## Run golangci-lint
	@echo "🧹 linting..."
	@golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "🔍 govulncheck..."
	@$(GO) version
	@$(GO) env GOPRIVATE
	@$(GOPATH)/bin/govulncheck ./... || govulncheck ./...

# NOTE: run before running vulcheck/lint commands
tools: ## Install dev/CI tools (govulncheck, golangci-lint)
	# for keeping dev and CI in sync.
	@echo "🔧 installing tools (optional locally)"
	@go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7

check: test coverage lint vulcheck ## Run tests, coverage, lint, and vuln check
	@echo "✅ all checks passed"

all: tools check ## Install tools and run all checks

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test coverage lint vulcheck tools check all help
