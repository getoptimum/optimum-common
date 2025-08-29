GO                ?= go
GOPATH            ?= $(shell $(GO) env GOPATH)

PKGS_ALL          := ./...
PKGS_FOR_TESTS    := $(shell $(GO) list ./... | grep -v '/examples')
COVERPROFILE      ?= coverage.out
COVERMODE         ?= atomic

# tool versions
VULNCHECK_VER ?= v1.1.4
GOLANGCI_VER  ?= v2.1.1

.DEFAULT_GOAL := help

# -------- Commands --------

all: tidy fmt vet test lint coverage vulcheck

test: ## Run all tests with race detector and coverage
	@echo "🧪 running tests..."
	@$(GO) test $(PKGS_FOR_TESTS) -race -covermode=$(COVERMODE) -coverprofile=$(COVERPROFILE)

coverage: ## Show coverage summary
	@echo "📈 coverage summary"
	@$(GO) tool cover -func=$(COVERPROFILE) | tail -1

coverhtml: $(COVERPROFILE)
	@$(GO) tool cover -html=$(COVERPROFILE) -o coverage.html
	@echo "📊 open coverage.html to view the report"

lint: ## Run golangci-lint
	@echo "🧹 linting..."
	@$(GOPATH)/bin/golangci-lint run --timeout=7m

vulcheck: ## Run govulncheck for vulnerabilities
	@echo "🔍 govulncheck..."
	@$(GO) version
	@$(GO) env GOPRIVATE
	@$(GOPATH)/bin/govulncheck ./... || govulncheck ./...

tools: ## (Optional) install dev tools locally
	@echo "🔧 installing tools locally"
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@$(VULNCHECK_VER)
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VER)

tidy: ## go mod tidy + verify
	@$(GO) mod tidy
	@$(GO) mod verify

fmt: ## go fmt packages
	@$(GO) fmt $(PKGS_ALL)

vet:
	@echo "🔍 go vet..."
	@$(GO) vet $(PKGS_ALL)

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

clean: ## Remove generated artifacts
	@rm -f $(COVERPROFILE) coverage.html

# -------- Deps / Phony --------
$(COVERPROFILE):
	@echo "No $(COVERPROFILE) found; run 'make test' first." >&2; exit 1

.PHONY: all test coverage coverhtml lint vet vulcheck tools tidy fmt help clean
