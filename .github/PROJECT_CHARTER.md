# Oscar — Project Charter

**PMBOK Process Group:** Initiation (4.1 Develop Project Charter)
**Date:** 2026-06-07
**Author:** PM (via integrated Lean-Agile-PMI framework)

---

## 1. Purpose & Vision

**Vision:** Build a CRM for people that do not care to understand much about tech and that want a visually beautiful product that offers as much power as the competitors.

**Product Name:** Oscar
**Tagline:** *The CRM that feels good.*

## 2. Project Scope

### In Scope (MVP)
- Multi-tenant SaaS CRM backend (Go/Echo/PostgreSQL)
- Beautiful, intuitive frontend (Astro/React/Tailwind)
- Core CRM: Contacts, Companies, Deals, Pipelines, Activities
- Authentication: email/password, Paseto tokens, OAuth (Google, Apple)
- Team management with RBAC
- White-label branding per tenant
- API for integration

### Out of Scope (MVP)
- AI/ML features (Phase 5)
- Mobile native apps (Phase 7)
- Enterprise SSO/SAML (Phase 7)
- Advanced automation engine (Phase 4)

## 3. Key Stakeholders

| Stakeholder | Role | Interest |
|---|---|---|
| @CaptDany (you) | Product Lead & Solo Developer | Delivery, quality, vision |
| End users | Non-technical CRM operators | Usability, beauty, reliability |
| Open-source community | Contributors | Code quality, documentation |

## 4. High-Level Risks

| # | Risk | P (1-5) | I (1-5) | PxI | Mitigation |
|---|---|---|---|---|---|
| R1 | Single-contributor bottleneck | 5 | 3 | 15 | Document patterns; keep scope lean |
| R2 | Redis dependency not wired | 3 | 4 | 12 | Treat as spike in Sprint 1 |
| R3 | Phase 1 scope exceeds deadline | 4 | 5 | 20 | WSJF prioritization; renegotiate if needed |
| R4 | No automated test coverage for API | 3 | 4 | 12 | Add test debt to backlog |
| R5 | Frontend detail views not connected to API | 3 | 4 | 12 | Verify data flow in Sprint 1 |

## 5. MVP Definition

The MVP is delivered when a non-technical user can:
1. Register a tenant account and log in
2. Manage contacts (persons) and companies
3. Track deals through a pipeline (Kanban)
4. Log activities (notes, calls, meetings)
5. Manage team members with roles
6. Customize brand colors/logo
7. Operate without reading documentation

**Lean principle:** "Eliminate waste" — anything not needed for the above is deferred.

## 6. Success Criteria

- Zero manual database operations required to run
- All API endpoints return consistent error format (already implemented)
- Frontend loads < 2s on standard connections
- Tests pass before merge (CI enforced)

## 7. Governance

- **Sprints:** 2-week iterations
- **Backlog:** Prioritized by WSJF (Weighted Shortest Job First)
- **Tracking:** GitHub Projects V2 board with burndown
- **Communication:** Sprint reports via GitHub Discussions
- **Risk review:** At sprint start
