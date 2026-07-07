# Contributing to optimum-common

Thank you for your interest in contributing to optimum-common. This document
describes the process for submitting changes and the standards we follow.

## Filing Issues

- **Bug reports**: Include the Go version, OS, a minimal reproduction, and the
  expected vs. actual behavior.
- **Feature requests**: Describe the use case and why the feature belongs in the
  shared SDK rather than a downstream project.

## Pull Request Process

1. **Fork** the repository and create a branch from `main`.
2. **Write tests** for your changes. See the code standards below.
3. **Run checks** locally before pushing:

   ```bash
   make fmt        # Format code
   make vet        # Static analysis
   make lint       # golangci-lint
   make test       # Tests with race detector and coverage
   make fuzz       # Fuzz tests (default 30s per target)
   make vulcheck   # Vulnerability scanner
   ```

   You can also run all checks at once with `make check`.

4. **Open a PR** against `main`. Keep PRs focused -- one logical change per PR.
5. Ensure CI passes. The CI pipeline runs `make ci` which covers linting, tests,
   coverage, and vulnerability checks.
6. A maintainer will review your PR. Address feedback with additional commits
   rather than force-pushing, so the review history is preserved.

## Code Standards

- **Tests**: Use table-driven tests with the `require` package.
- **Test comments**: Follow the Given/When/Then structure where possible.
- **Comments**: Capitalized, full sentences, ending with a period.
- **Formatting**: Run `make fmt` before committing.
- **Linting**: Code must pass `make lint` with no warnings.

## Makefile Targets

| Target     | Description                                |
|------------|--------------------------------------------|
| `test`     | Run tests with race detector and coverage  |
| `lint`     | Run golangci-lint                          |
| `fuzz`     | Run fuzz tests                             |
| `vulcheck` | Run govulncheck for known vulnerabilities  |
| `bench`    | Run benchmarks                             |
| `fmt`      | Format Go source files                     |
| `vet`      | Run go vet                                 |
| `tidy`     | Run go mod tidy and verify                 |
| `check`    | Run all of the above                       |

## Commit Sign-Off (DCO)

This project uses the [Developer Certificate of Origin](https://developercertificate.org/)
(DCO). By contributing, you certify that you have the right to submit the work
under the project's license.

Sign off each commit by adding a `Signed-off-by` line:

```bash
git commit -s -m "feat: add helper for X"
```

This appends a line like:

```text
Signed-off-by: Your Name <your.email@example.com>
```

All commits in a PR must carry a valid sign-off.

## Questions

If you are unsure whether a change fits the project, open an issue to discuss it
before writing code. This saves time for everyone.
