# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in optimum-common, please report it
responsibly. **Do not open a public GitHub issue.**

Send an email to **<security@getoptimum.io>** with:

- A description of the vulnerability.
- Steps to reproduce the issue, or a proof-of-concept if possible.
- The version(s) or commit(s) affected.
- Any potential impact you have identified.

## What to Expect

- **Acknowledgment** within 48 hours of your report.
- **Status update** within 5 business days with an initial assessment and
  expected timeline for a fix.
- A coordinated disclosure once a fix is available. We will credit reporters
  unless they prefer to remain anonymous.

## Scope

### In scope

- Code in this repository (`github.com/getoptimum/optimum-common`).
- Dependencies shipped as part of the compiled module, where the vulnerability
  is reachable through optimum-common's public API.

### Out of scope

- Vulnerabilities in third-party dependencies that are not reachable through
  this module's API (please report those upstream).
- Issues in other Optimum repositories (report those to the respective project).
- Social engineering, phishing, or physical attacks.
- Denial-of-service attacks that require privileged network access.

## Preferred Languages

Reports may be submitted in English.

## Acknowledgments

We appreciate the security research community's efforts in helping keep
Optimum projects safe. Thank you for practicing responsible disclosure.
