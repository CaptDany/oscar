# CI/CD Strategy — oscar CRM

> **Navigation:** [Documentation Index](README.md) → CI/CD

## Overview

This document defines the CI/CD pipeline for oscar, a multi-tenant CRM (Go 1.24+ / Echo v4 / PostgreSQL 16 / Astro). Every decision maps to Lean (Poppendieck), Continuous Delivery (Humble & Farley), and DORA (Accelerate) principles.

---

## Principles

| Principle | Source | Pipeline Manifestation |
|---|---|---|
| **Fast Feedback** | DORA, CD Ch. 9 | Parallel jobs; path filtering; short PR→deploy cycle |
| **Batch Size Reduction** | Lean (Poppendieck) | Per-commit builds; no batching; small PRs via path filters |
| **Eliminate Waste** | Lean (Poppendieck) | `detect-changes` skips irrelevant jobs; cached Go/npm/Docker layers; single immutable artifact promoted through all envs |
| **Build Quality In** | Lean, CD Ch. 5 | Lint, test, security audit, CodeQL, dependency review — fail fast |
| **Immutable Artifacts** | CD Ch. 9 | One Docker image per commit; never rebuild; same SHA dev→staging→prod |
| **Shift Left on Security** | DevSecOps | govulncheck, CodeQL, dependency-review run on every PR |
| **Trunk-Based Development** | DORA, Accelerate | Short-lived branches off `main`; main always deployable |
| **Progressive Delivery** | Lean, DORA | GitHub Environments with protection rules; manual gates for staging + prod |
| **Provider Agnostic** | Anti-lock-in | kubeconfig secret abstraction — switch clouds by swapping one GitHub Secret per env |

---

## Pipeline Architecture

```
PR / Push → detect-changes → ┌─────────────────┐
                              │  CI (parallel)   │
                              │  lint · test ·    │
                              │  build · audit    │
                              │  docker · codeql  │
                              └────────┬─────────┘
                                       │ (push to main)
                                       ▼
                              ┌─────────────────┐
                              │   CD pipeline    │
                              │  version → build │
                              │  & push image    │
                              │  → dev → staging │
                              │  → production    │
                              └─────────────────┘
```

---

## Workflow Reference

### `.github/workflows/ci.yml` — PR checks
| Job | Trigger | Purpose |
|---|---|---|
| `detect-changes` | Always | Path-based filtering (Lean waste elimination) |
| `lint-backend` | Go files | golangci-lint |
| `test-backend` | Go files | `go test -short -cover` |
| `build-backend` | Go/Docker | Cross-compile Linux binary |
| `build-frontend` | `web/**` | `npm run build` |
| `security-audit` | Go files | `govulncheck` |
| `docker-build-check` | Docker/Go | Build image (no push) |
| `dependency-review` | PR only | Block critical-severity advisories |

### `.github/workflows/cd.yml` — Deploy pipeline
| Job | Gate | Description |
|---|---|---|
| `version` | — | Calculate semver from git tags |
| `build-and-push` | — | Docker build → ghcr.io push with SBOM + attestation |
| `deploy-dev` | Auto | Helm upgrade → dev namespace → smoke test `/health` |
| `deploy-staging` | Manual | Helm upgrade → staging → smoke test |
| `deploy-production` | Manual | Helm upgrade → prod → canary smoke → record deployment |

### `.github/workflows/rollback.yml` — Incident recovery
- Manual trigger: select environment + target tag
- Decodes kubeconfig from `${{ secrets.KUBECONFIG_<ENV> }}`
- Attempts `helm rollback`; falls back to `helm upgrade` with previous tag
- Creates `pipeline-incident` issue for blameless post-mortem

### `.github/workflows/release.yml` — Release management
- Manual trigger: pick bump type (major/minor/patch) + optional pre-release ID
- Generates changelog from conventional commits, creates git tag, publishes GitHub Release

---

## GitHub Environments

| Environment | Manual Approval | URL |
|---|---|---|
| `dev` | No | `https://dev.oscar-crm.cc` |
| `staging` | Yes | `https://staging.oscar-crm.cc` |
| `production` | Yes | `https://oscar-crm.cc` |

---

## Provider Abstraction: Kubeconfig Interface

The pipeline is **provider-agnostic**. Every deploy step uses the same pattern:

```yaml
- name: Configure kubectl
  run: |
    mkdir -p ~/.kube
    echo "${{ secrets.KUBECONFIG_<ENV> }}" | base64 -d > ~/.kube/config
```

Three GitHub Environment secrets must exist:

| Secret | Environment | Value |
|---|---|---|
| `KUBECONFIG_DEV` | `dev` | Base64-encoded kubeconfig for the dev cluster |
| `KUBECONFIG_STAGING` | `staging` | Base64-encoded kubeconfig for the staging cluster |
| `KUBECONFIG_PROD` | `production` | Base64-encoded kubeconfig for the production cluster |

**Switching providers** = replace these three secrets. Zero pipeline edits.

### Kubeconfig Rotation

Kubeconfigs contain cluster certificate authority data and user credentials. Rotate them periodically:

1. **Generate new kubeconfig** — Re-run the provider's command (e.g., `oci ce cluster create-kubeconfig` for OKE). This refreshes the client certificate and CA data.
2. **Base64-encode**:
   ```bash
   base64 -w0 ~/.kube/oscar-<env>-config
   ```
3. **Replace secret** — Go to Settings → Environments → `<env>` → update the `KUBECONFIG_<ENV>` secret.
4. **Verify** — Run the CD workflow manually for that environment — the deploy step decodes the new kubeconfig automatically.

> **Note:** Kubeconfigs are short-lived by default on most managed K8s providers (typically 1–3 years for OKE). The pipeline will fail with an auth error when the cert expires; follow the steps above to refresh.

---

## OKE (Oracle Kubernetes Engine) — Bootstrap Runbook

### Prerequisites
- OCI CLI installed (`brew install oci` or `curl -L -O https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.sh`)
- OCI config at `~/.oci/config` with valid credentials
- `kubectl` and `helm` installed

### 1. Create OKE cluster (free tier)
```bash
# Create VCN and networking (or use existing)
oci network vcn create \
  --cidr-block 10.0.0.0/16 \
  --display-name oscar-vcn \
  --compartment-id <compartment-ocid>

# Create OKE cluster (free: VM.Standard.E2.1.Micro shape × 2)
oci ce cluster create \
  --name oscar \
  --compartment-id <compartment-ocid> \
  --vcn-id <vcn-ocid> \
  --kubernetes-version v1.28.5 \
  --node-image-id ocid1.image.oc1..<latest-olk> \
  --node-shape VM.Standard.E2.1.Micro \
  --node-count 2 \
  --wait-for-state SUCCEEDED
```

### 2. Get kubeconfig for each environment
**Dev:**
```bash
oci ce cluster create-kubeconfig \
  --cluster-id <cluster-ocid> \
  --file ~/.kube/oscar-dev-config \
  --region <region>
base64 -w0 ~/.kube/oscar-dev-config | pbcopy
# Paste into GitHub: Settings → Environments → dev → Secrets → KUBECONFIG_DEV
```

**Staging & Production:** Repeat with separate clusters or namespaces.

### 3. Create ghcr.io image pull secret
OKE free tier does not support OIDC. You need a static pull secret:

```bash
kubectl create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<your-github-username> \
  --docker-password=<ghcr-token> \
  -n oscar-dev
kubectl create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<your-github-username> \
  --docker-password=<ghcr-token> \
  -n oscar-staging
kubectl create secret docker-registry ghcr-pull \
  --docker-server=ghcr.io \
  --docker-username=<your-github-username> \
  --docker-password=<ghcr-token> \
  -n oscar-production
```

The Helm chart already references `imagePullSecrets: [{name: ghcr-pull}]` in `values.yaml`.

### 4. Install nginx-ingress + cert-manager
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace

helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true
```

### 5. Configure DNS
Point `oscar-crm.cc`, `dev.oscar-crm.cc`, `staging.oscar-crm.cc` to the nginx-ingress LoadBalancer external IP:
```bash
kubectl get svc -n ingress-nginx nginx-ingress-ingress-nginx-controller \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

---

## Provider Migration Guide

### OKE → DigitalOcean (DOKS)

1. Create DOKS cluster:
   ```bash
   doctl kubernetes cluster create oscar --region nyc1 --node-pool "pool=2gb:2"
   doctl kubernetes cluster kubeconfig save oscar
   ```
2. Base64-encode the new kubeconfig and replace the three GitHub Environment secrets.
3. **Optional** — Delete the `imagePullSecrets` block from `values.yaml` if you enable DOKS OIDC.
4. Run `CD` workflow — deploys to the new cluster with zero code changes.

### OKE → GKE / AKS / EKS

Same process: create cluster, export kubeconfig, base64-encode, replace secrets, optionally switch to OIDC and remove `ghcr-pull`:

```bash
# GKE with OIDC:
gcloud container clusters create oscar --region us-east1 \
  --workload-pool=<project>.svc.id.goog

# AKS with OIDC:
az aks create --resource-group oscar --name oscar --enable-oidc-issuer
```

---

## Secrets Required

| Secret | Scope | Source |
|---|---|---|
| `KUBECONFIG_DEV` | Env: dev | Base64 kubeconfig from `oci ce cluster create-kubeconfig` |
| `KUBECONFIG_STAGING` | Env: staging | Same, for staging cluster |
| `KUBECONFIG_PROD` | Env: production | Same, for production cluster |
| `GITHUB_TOKEN` | Repo (auto) | Used by Actions to push to ghcr.io |

No cloud API keys are stored — only kubeconfigs (which contain short-lived certs).

---

## DORA Metrics

| Metric | Method | Target (Elite) | Source |
|---|---|---|---|
| **Deployment Frequency** | Count `deploy-production` runs/week | Multiple/day | Accelerate Ch. 3 |
| **Lead Time** | First commit → prod deploy | < 1 hour | Accelerate Ch. 3 |
| **MTTR** | `pipeline-incident` created → rollback success | < 1 hour | Accelerate Ch. 3 |
| **Change Failure Rate** | Rollbacks / total prod deploys | < 15% | Accelerate Ch. 3 |

---

## Security Posture

- **CodeQL** — PR + weekly; Go + JavaScript (security-and-quality queries)
- **Dependency review** — Blocks PRs with critical-severity vulnerabilities
- **Dependabot** — Weekly PRs for Go, npm, Docker, Actions
- **SLSA / Provenance** — Docker images built with `provenance: true`, `sbom: true`, `actions/attest-build-provenance`
- **OpenSSF Scorecard** — Weekly scan; results → GitHub Security tab
- **govulncheck** — PR-level Go vulnerability scan (shift-left)

---

## References

- Humble, J., & Farley, D. (2010). *Continuous Delivery*. Addison-Wesley.
- Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate*. IT Revolution Press.
- Poppendieck, M., & Poppendieck, T. (2003). *Lean Software Development*. Addison-Wesley.
- OCI CLI: https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm
- GitHub Actions: https://docs.github.com/en/actions
