# optimum-common

This library will serve as the **shared SDK** for Optimum projects.

## High-level structure

- `optimum-common/` contains a Go module that acts as a shared SDK for other
  Optimum services, consolidating logging, configuration, utilities, and version
  helpers used across projects.

## Standards followed

1. Tests employ `require` package, must be table driven where possible
2. Comments in tests follow Given/When/Then structure when possible
3. Comments are capitalized, in full sentences ending with a period.
