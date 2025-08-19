GO := go # so can subst. with tinygo

test:
	@echo "🧪 running tests..."
	@$(GO) test ./... -race -covermode=atomic -coverprofile=coverage.out

coverage:
	@echo "📈 coverage summary"
	@$(GO) tool cover -func=coverage.out | tail -n1 || true

lint:
	@echo "🧹 linting..."
	@golangci-lint run --timeout=7m

vulcheck:
	@echo "🔍 govulncheck..."
	@$(GO) version
	@$(GO) env GOPRIVATE
	@$(GOPATH)/bin/govulncheck ./... || govulncheck ./...
# NOTE: run before running vulcheck/lint commands
tools:
	#for keeping dev and CI in sync.
	@echo "🔧 installing tools (optional locally)"
	@go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7

check: test coverage lint vulcheck
	@echo "✅ all checks passed"
all : tools check
.PHONY: test coverage lint vulcheck tools
