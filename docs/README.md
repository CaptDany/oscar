# Oscar Documentation

Welcome to the Oscar documentation. This is the canonical index for all project documentation.

---

## Getting Started

| Document | Location | Audience |
|---|---|---|
| [Project Overview](../README.md) | Root README | Everyone |
| [Quick Start Guide](../README.md#quick-start) | Root README | Developers |
| [Frontend README](../web/README.md) | `web/README.md` | Frontend developers |

## Technical Reference

| Document | Location | Contents |
|---|---|---|
| [Architecture](../README.md#architecture) | Root README | Layered architecture, repository pattern, error conventions |
| [API Reference](../README.md#api-endpoints) | Root README | All REST endpoints by module |
| [Database Schema](../README.md#database-schema) | Root README | Key tables, RLS policies |
| [CI/CD Pipeline](CI-CD.md) | `docs/CI-CD.md` | Pipeline architecture, DORA metrics, deployment runbooks |

## Project Management

| Document | Location | PMBOK Reference |
|---|---|---|
| [Project Charter](../.github/PROJECT_CHARTER.md) | `.github/PROJECT_CHARTER.md` | 4.1 Develop Project Charter |
| [Risk Register](../.github/RISK_REGISTER.md) | `.github/RISK_REGISTER.md` | 11.2 Identify Risks |
| [Phase 1 Retrospective](../.github/RETROSPECTIVE_PHASE1.md) | `.github/RETROSPECTIVE_PHASE1.md` | 4.4 Monitor & Control |
| [Project Board](https://github.com/users/CaptDany/projects/2) | GitHub Projects V2 | 4.3 Direct & Manage |
| [Sprint Reports](https://github.com/CaptDany/oscar/discussions) | GitHub Discussions | 10.1 Plan Communications |

## Community & Governance

| Document | Location |
|---|---|
| [Contributing Guide](../CONTRIBUTING.md) | Root |
| [Code of Conduct](../CODE_OF_CONDUCT.md) | Root |
| [Security Policy](../SECURITY.md) | Root |
| [Pull Request Template](../.github/PULL_REQUEST_TEMPLATE.md) | `.github/` |
| [Issue Templates](../.github/ISSUE_TEMPLATE/) | `.github/ISSUE_TEMPLATE/` |

## Issue Types

| Template | When to Use |
|---|---|
| [User Story](../.github/ISSUE_TEMPLATE/user_story.md) | Feature from end-user perspective with WSJF |
| [Task](../.github/ISSUE_TEMPLATE/task.md) | Technical or process work |
| [Spike](../.github/ISSUE_TEMPLATE/spike.md) | Research, exploration, proof-of-concept |
| [Epic](../.github/ISSUE_TEMPLATE/epic.md) | Large body spanning multiple issues |
| [Bug Report](../.github/ISSUE_TEMPLATE/bug_report.md) | Something isn't working |
| [Feature Request](../.github/ISSUE_TEMPLATE/feature_request.md) | Idea or enhancement |

## GitHub Wiki

> **Note:** The GitHub Wiki requires a one-time initialization. Visit
> [https://github.com/CaptDany/oscar/wiki](https://github.com/CaptDany/oscar/wiki)
> in your browser to create it. After that, wiki content can be pushed via git or
> the GitHub Wiki API. The following pages are staged for the wiki:
>
> - **Home** — Overview and quick links (this document)
> - **Project Charter** — Vision, scope, risks, MVP
> - **Risk Register** — Probability/Impact risk matrix
> - **API Reference** — All REST endpoints

## Document Map

```
.github/                         # Governance (PM artifacts, templates)
├── PROJECT_CHARTER.md           # Project charter (PMBOK 4.1)
├── RISK_REGISTER.md             # Risk register (PMBOK 11.2)
├── RETROSPECTIVE_PHASE1.md      # Phase 1 retrospective
├── PULL_REQUEST_TEMPLATE.md     # PR template
└── ISSUE_TEMPLATE/              # Issue templates (6)

docs/                            # Technical documentation
├── README.md                    # ← You are here
└── CI-CD.md                     # CI/CD strategy + runbooks

web/README.md                    # Frontend docs

README.md                        # Project landing page
CONTRIBUTING.md                  # Contribution guide
CODE_OF_CONDUCT.md               # Code of conduct
SECURITY.md                      # Security policy
```

---

*This documentation structure follows the Lean "see the whole" principle — all project knowledge in one navigable index.*
