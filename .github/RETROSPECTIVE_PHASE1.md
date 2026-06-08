# Phase 1 Academic Retrospective

**Date:** 2026-06-07
**Phase:** Phase 1 — Foundation Gaps (Core CRM Completion)
**Sprint:** Pre-Sprint 0 (prior to Sprint 1 establishment)

> **Navigation:** [Documentation Index](../docs/README.md) → Phase 1 Retrospective
> **See also:** [Risk Register](RISK_REGISTER.md), [Sprint 1 Board](https://github.com/users/CaptDany/projects/2)

---

## 1. What Happened

Phase 1 was planned as the first deliverable of Oscar, targeting core CRM feature completion. Of 10 planned issues, **7 are closed** and **3 remain open**. The remaining items are:

| # | Issue | WSJF Priority | Status |
|---|---|---|---|
| #10 | Wire Redis for Caching | 1 (3.33) | No code started |
| #8 | API Key Management Backend | 2 (3.0) | Frontend done, backend missing |
| #11 | File Attachments for Records | 3 (2.2) | No code started |

Phase 1's deadline (Aug 2026) is still achievable if the remaining items are completed within Sprint 1 (June 8–21).

---

## 2. PMI Process Group Analysis

### Planning (PMBOK 5.1 — Plan Scope Management)
- **Good:** The README roadmap provided a clear phased scope with linked issues
- **Gap:** No sprint-level decomposition existed — issues were in a phase milestone without prioritization or timeboxing
- **Corrective Action:** WSJF scores now provide sprint-level priority; Sprint 1 (2-week) milestone replaces flat phase bucket

### Executing (PMBOK 4.3 — Direct and Manage Project Work)
- **Good:** 7 issues closed without a formal PM system — the developer worked autonomously
- **Gap:** No visibility into WIP limits, cycle time, or bottlenecks
- **Corrective Action:** Project V2 board with Status field enables WIP visualization; burndown tracks remaining effort

### Monitoring & Controlling (PMBOK 4.4 — Monitor and Control Project Work)
- **Gap:** No burndown chart, no variance analysis — phase completion could not be tracked
- **Corrective Action:** Project V2 insights provide burndown; weekly sprint reports via Discussions

---

## 3. Lean Waste Analysis

| Waste | Manifestation | Corrective Action |
|---|---|---|
| **Waiting** | Redis was in `go.mod` and `config.go` since project inception but never wired — dependency was "available" but not acted upon. The project waited for a dependency that was already decided on. | Treat Redis as a Sprint 1 spike (Risk R2 mitigation). Use JIT principle: when a dependency is decided, execute it immediately. |
| **Partially Done Work** | API Key Management has a complete frontend UI but no backend — the frontend was built before the API existed. This work delivers zero value until the backend is complete. | "Stop starting, start finishing" (Kanban principle). Complete the backend for one feature before starting the frontend for another. |
| **Task Switching** | 10 issues were started simultaneously under Phase 1. | WSJF prioritization in Sprint 1 enforces single-piece flow. Do Redis first, then API Keys, then File Attachments. |
| **Defects** | Not observed in Phase 1 | CI workflow already prevents merge with failing tests |

---

## 4. Agile Principles Assessment

| Agile Principle | Score (1-5) | Notes |
|---|---|---|
| #1: Highest priority is customer satisfaction | 4 | CRM core features are being built; visual design is strong |
| #3: Deliver working software frequently | 3 | Delivered in batches (phase-level), not incremental sprints |
| #7: Working software is the measure of progress | 3 | Frontend stubs exist but don't count as working software |
| #9: Technical excellence and good design | 5 | Codebase is well-structured (repository pattern, layered architecture) |
| #10: Simplicity — maximize work not done | 2 | 10 items in Phase 1 includes items that could wait |

**Average:** 3.4 / 5 — Room for improvement, primarily around delivery cadence and scope management.

---

## 5. Recommendations for Sprint 1

1. **Start with Redis (#10) — WSJF 3.33** — Risk reduction first. Wiring a pre-decided dependency eliminates "waiting" waste.
2. **Follow with API Keys (#8) — WSJF 3.0** — Complete the backend to deliver value from the already-built frontend. Eliminates "partially done work" waste.
3. **File Attachments (#11) — WSJF 2.2** — If this exceeds Sprint 1, renegotiate scope to Sprint 2 (Phase 1 deadline is Aug — two sprints remain).

---

## 6. Metric Baseline

| Metric | Phase 1 (Pre-PM) | Target (Sprint 1) |
|---|---|---|
| Cycle Time per issue | Unknown | < 5 days |
| Throughput | ~7 issues / ~10 weeks (0.7/week) | 3 issues / 2 weeks (1.5/week) |
| WIP Limit | Unlimited | 1 (single-piece flow) |
| Scope Change | None tracked | WSJF review at sprint start |

---

*This retrospective applies the Lean "see the whole" principle — understanding Phase 1's delays requires looking at the entire workflow, not individual tasks.*
