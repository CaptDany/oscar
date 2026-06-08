# Oscar — Risk Register

**PMBOK Process Group:** Risk Management (11.2 Identify Risks, 11.3 Perform Qualitative Risk Analysis)
**Last Updated:** 2026-06-07
**Review Cadence:** Start of every sprint

---

## Risk Scoring

- **Probability (P):** 1 (Rare) → 5 (Almost certain)
- **Impact (I):** 1 (Negligible) → 5 (Catastrophic)
- **PxI =** Overall risk score (1–25)

---

## Risk Table

| ID | Risk Description | Category | P | I | PxI | Response Strategy | Mitigation Plan | Contingency Plan | Owner | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| R1 | Single-contributor bottleneck delays delivery | Resource | 5 | 3 | 15 | Mitigate | Document patterns in CONTRIBUTING.md; keep scope lean per sprint; automate CI/CD to reduce manual overhead | If blocked for >1 week, reduce sprint scope by 30% | @CaptDany | Open |
| R2 | Redis dependency is in go.mod but not wired; could block caching features | Technical | 3 | 4 | 12 | Mitigate | Sprint 1 spike to wire Redis; fallocate 2 story points | If spike takes >1 sprint, defer Redis to Phase 2 and use in-memory cache | @CaptDany | Open |
| R3 | Phase 1 scope (3 remaining items) exceeds Aug 2026 deadline | Schedule | 4 | 5 | 20 | Mitigate | WSJF-prioritize remaining items; track burndown weekly | Renegotiate scope: drop lowest-WSJF item to Phase 2 | PM | Open |
| R4 | No automated test coverage for existing API handlers | Quality | 3 | 4 | 12 | Accept (near-term) | Add test-debt story to every sprint until coverage > 60% | If regression occurs, add tests before fix | @CaptDany | Open |
| R5 | Frontend detail views (Contact/Company/Deal) exist as stubs not connected to API | Technical | 3 | 4 | 12 | Mitigate | Sprint 1 verification task: ensure API ↔ frontend data flow works end-to-end | If API missing, file bug and stub with mock data | @CaptDany | Open |
| R6 | OAuth (Google/Apple) provider credentials not configured in CI | Security | 2 | 3 | 6 | Accept | Document required env vars in .env.example | Manual test before release | @CaptDany | Open |
| R7 | PostgreSQL RLS policies may need tuning for complex queries | Technical | 2 | 3 | 6 | Monitor | Add query performance review to PR checklist | Add indexes; log slow queries | @CaptDany | Open |

---

## Risk Burndown

Track total risk score over time:

| Sprint | Total PxI | New Risks | Closed Risks | Notes |
|---|---|---|---|---|
| Sprint 0 | 82 | 7 | 0 | Baseline |
| Sprint 1 | — | — | — | Review at sprint start |

---

## Decision Log

| Date | Decision | Rationale | Impact |
|---|---|---|---|
| 2026-06-07 | Use .github/ files for charter + risk register instead of Wiki | Wiki API unavailable; version-controlled files are traceable (Lean: "see the whole") | Documents live alongside code; PR reviewable |
