## Description
Briefly describe the changes and the motivation behind them.

## Related Issues
Closes #(issue-number)

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would break existing functionality)
- [ ] Chore (maintenance, cleanup, dependency updates)
- [ ] Documentation (docs, README, comments)

## CI/CD Automation Checks
<!-- These run automatically — don't merge until all pass. -->
- [ ] **Lint** — `golangci-lint` (Go) — no new violations
- [ ] **Tests** — `go test -short ./...` — all pass
- [ ] **Build** — Go binary + Astro frontend — compile cleanly
- [ ] **Security** — `govulncheck` — no new vulnerabilities
- [ ] **Docker check** — Container image builds
- [ ] **Dependency review** — No critical advisories
- [ ] **CodeQL** — Static analysis passes

## Testing Notes
Describe how the changes were tested and any manual testing steps.

## Checklist
- [ ] My code follows the project's code style
- [ ] I have added tests that prove my fix/feature works
- [ ] All existing tests pass locally
- [ ] I have updated the documentation if needed
