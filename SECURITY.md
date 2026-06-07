# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Oscar, please report it by
opening a GitHub Security Advisory at:

https://github.com/CaptDany/oscar/security/advisories/new

Do **not** report security vulnerabilities via public GitHub issues.

## What to Include

- A clear description of the vulnerability
- Steps to reproduce (proof of concept preferred)
- Impact assessment (what an attacker could achieve)
- Any suggested mitigation if known

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Triage and assessment**: Within 5 business days
- **Fix deployed**: Depends on severity, typically 7-14 days for critical issues

## Scope

The following are in scope:

- The Go backend (API server, database interactions, auth)
- The Astro frontend (XSS, CSRF, dependency vulnerabilities)
- Infrastructure configuration (CI/CD, Docker)

The following are out of scope:

- Issues in dependencies that have already been disclosed upstream
- Social engineering attacks against project maintainers

## Supported Versions

Only the latest commit on the `main` branch is supported.
There are no LTS releases at this time.
