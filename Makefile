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

# --- Fuzzing ---
# Note: no fuzzing in CI
# Usage:
#   make fuzz FUZZ=FuzzNewVerifierFromDomain PKG=./auth FUZZTIME=15s
#   make fuzz-list
#   make fuzz-all FUZZTIME=10s           # sequentially fuzz every Fuzz* found

FUZZ        ?=
PKG         ?= ./...
FUZZTIME    ?= 15s

fuzz: ## Run a single fuzz target (set FUZZ=FuzzName, optionally PKG=./pkg)
	@if [ -z "$(FUZZ)" ]; then \
		echo "❗ Set FUZZ=FuzzName (e.g., FuzzNewVerifierFromDomain) [PKG=./pkg]"; exit 2; \
	fi
	@echo "🐛 fuzzing $(FUZZ) in $(PKG) for $(FUZZTIME)..."
	@$(GO) test -run=^$$ -fuzz=$(FUZZ) -fuzztime=$(FUZZTIME) $(PKG)

fuzz-list: ## List discovered fuzz targets (Fuzz* functions) with their packages
	@grep -R --include='*_test.go' -n 'func[[:space:]]\+Fuzz[[:alnum:]_]*\s*\(' . \
	  | sed -E 's|^(.*)/([^/]+):[0-9]+:.*func[[:space:]]+(Fuzz[[:alnum:]_]+).*|\3\t./\2|' \
	  | sort -u

fuzz-all: ## Run all fuzz targets sequentially (short fuzz time recommended)
	@targets=$$(grep -R --include='*_test.go' -n 'func[[:space:]]\+Fuzz[[:alnum:]_]*\s*\(' . \
	  | sed -E 's|^(.*)/([^/]+):[0-9]+:.*func[[:space:]]+(Fuzz[[:alnum:]_]+).*|\3 ./\2|' \
	  | sort -u); \
	if [ -z "$$targets" ]; then echo "No fuzz targets found."; exit 0; fi; \
	echo "🐛 Running all fuzz targets for $(FUZZTIME) each..."; \
	status=0; \
	while read -r name pkg; do \
	  echo "——→ fuzz $$name in $$pkg"; \
	  $(GO) test -run=^$$ -fuzz=$$name -fuzztime=$(FUZZTIME) $$pkg || status=$$?; \
	done <<EOF \
	$$targets \
EOF ; exit $$status

# -------- Deps / Phony --------
$(COVERPROFILE):
	@echo "No $(COVERPROFILE) found; run 'make test' first." >&2; exit 1

.PHONY: all test coverage coverhtml lint vet vulcheck tools tidy fmt help clean
