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

| Environment | Manual Approval | URL | Namespace |
|---|---|---|---|
| `dev` | No | `https://dev.oscar-crm.cc` | `oscar-dev` |
| `staging` | Yes | `https://staging.oscar-crm.cc` | `oscar-staging` |
| `production` | Yes | `https://oscar-crm.cc` | `oscar-production` |

All three environments share one OKE cluster with namespace-based isolation. Each has its own `KUBECONFIG_<ENV>` secret in GitHub (all three contain the same cluster kubeconfig for now).

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

### Cluster Details (mx-queretaro-1, June 2026)

| Resource | Value |
|---|---|
| Region | `mx-queretaro-1` (Mexico Central Querétaro) |
| Kubernetes | v1.35.2 |
| VCN | `oscar-vcn` (10.0.0.0/16) |
| Worker subnet | `oscar-workers-subnet` (10.0.1.0/24, public, with IGW route) |
| Endpoint subnet | `oscar-endpoint-subnet` (10.0.32.0/24) |
| Node shape | `VM.Standard2.1` (1 OCPU, 15 GB RAM, Intel Skylake) |
| Node count | 1 |
| Node public IP | `159.54.137.54` |
| Node pool image | `Oracle-Linux-8.10-2026.04.30-3-OKE-1.35.2-1462` |
| Worker subnet OCID | `ocid1.subnet.oc1.mx-queretaro-1.aaaaaaaa66pizbng6pbiizde3varx257objaspnxoyxgffhlbogaix6tqpfq` |
| Cluster OCID | `ocid1.cluster.oc1.mx-queretaro-1.aaaaaaaaggeuom6sy26ehogcbrlk3jsgllij63h5oyccstzplcvld6nsjk6a` |

### Prerequisites
- OCI CLI installed (`C:\oci\Scripts\oci.exe` on Windows, or via `pip install oci-cli`)
- OCI config at `~/.oci/config` with valid credentials
- `kubectl` and `helm` installed
- `oci` must be in PATH or referenced by full path in `~/.kube/config`

### 1. Create OKE cluster
```bash
# Create VCN
oci network vcn create --cidr-block 10.0.0.0/16 --display-name oscar-vcn \
  --compartment-id <compartment-ocid>

# Create Internet Gateway + Route Table
oci network internet-gateway create --vcn-id <vcn-ocid> --is-enabled true \
  --compartment-id <compartment-ocid>
oci network route-table create --vcn-id <vcn-ocid> --route-rules \
  '[{"cidrBlock":"0.0.0.0/0","networkEntityId":"<igw-ocid>"}]' \
  --compartment-id <compartment-ocid>

# Create subnets
oci network subnet create --vcn-id <vcn-ocid> --cidr-block 10.0.1.0/24 \
  --route-table-id <rt-ocid> --compartment-id <compartment-ocid>
oci network subnet create --vcn-id <vcn-ocid> --cidr-block 10.0.32.0/24 \
  --route-table-id <rt-ocid> --compartment-id <compartment-ocid>

# Create OKE cluster
oci ce cluster create --name oscar-cluster --compartment-id <compartment-ocid> \
  --vcn-id <vcn-ocid> --kubernetes-version v1.35.2 \
  --endpoint-subnet-id <endpoint-subnet-ocid> \
  --endpoint-public-ip true

# Create node pool (use an x86 OKE image from Oracle docs)
oci ce node-pool create --cluster-id <cluster-ocid> \
  --compartment-id <compartment-ocid> --name oscar-nodepool \
  --node-shape VM.Standard2.1 --size 1 \
  --node-source-details '{"sourceType":"IMAGE","imageId":"<x86-oke-image-ocid>"}' \
  --placement-configs '[{"availabilityDomain":"<ad-name>","subnetId":"<worker-subnet-ocid>"}]' \
  --kubernetes-version v1.35.2
```

> **Note:** OKE images are NOT in your compartment's `oci compute image list` — get the OCID for your region from [Oracle's image documentation](https://docs.oracle.com/en-us/iaas/images/oke-worker-node-oracle-linux-8x/). The x86 image is compatible with Intel shapes (`VM.Standard2.x`, `VM.Standard3.Flex`). ARM shapes (`VM.Standard.A1.Flex`) require the aarch64 OKE image. Free shapes (`VM.Standard.E2.1.Micro`) are often too small for OKE (1 GB RAM).

### 2. Get kubeconfig

Generate and store kubeconfigs for each environment. Since we use namespace-based isolation on a single cluster, you need three base64-encoded copies:

```bash
oci ce cluster create-kubeconfig --cluster-id <cluster-ocid> \
  --file ~/.kube/oscar-config --region mx-queretaro-1

# After generation, update the kubeconfig to use full path to oci:
# Linux/macOS:
sed -i 's|command: oci|command: /usr/local/bin/oci|' ~/.kube/oscar-config
# Windows PowerShell:
(Get-Content ~\.kube\oscar-config) -replace 'command: oci', 'command: C:\oci\Scripts\oci.exe' | Set-Content ~\.kube\oscar-config

# Base64-encode for GitHub secrets
# macOS:
base64 -w0 ~/.kube/oscar-config | pbcopy
# Linux:
base64 -w0 ~/.kube/oscar-config
# Windows PowerShell:
[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes((Get-Content ~\.kube\oscar-config -Raw)))

# Paste into GitHub: Settings → Environments → dev/staging/prod → KUBECONFIG_DEV / _STAGING / _PROD
```

> **Note:** The kubeconfig uses an `exec` credential plugin that calls `oci` to generate tokens. Ensure `oci` is in the runner's PATH or update the kubeconfig `command` to use the full binary path.

### 3. Create ghcr.io image pull secret
OKE free tier does not support OIDC. You need a static pull secret in each namespace:

```bash
for ns in oscar-dev oscar-staging oscar-production; do
  kubectl create secret docker-registry ghcr-pull \
    --docker-server=ghcr.io \
    --docker-username=<your-github-username> \
    --docker-password=<ghcr-pat> \
    -n $ns
done
```

The Helm chart references `imagePullSecrets: [{name: ghcr-pull}]` in `values.yaml`.

### 4. Install nginx-ingress + cert-manager

**Important:** The OCI Cloud Controller Manager is not installed in this OKE cluster (OKE managed control planes may not include it for all shapes/versions). Without the CCM, `type: LoadBalancer` services never get an external IP. The workaround is to use `hostNetwork` mode on the ingress controller:

```bash
# Add repos
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Install cert-manager
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set crds.enabled=true

# Install nginx-ingress with hostNetwork (no LoadBalancer)
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  --set controller.hostNetwork=true \
  --set controller.dnsPolicy=ClusterFirstWithHostNet \
  --set controller.service.type=ClusterIP
```

The ingress controller will bind directly to the node's network ports (80/443). Update the VCN security list to allow inbound TCP 80 and 443 from `0.0.0.0/0`:

```bash
oci network security-list update --security-list-id <security-list-ocid> \
  --ingress-security-rules '[
    {"source":"0.0.0.0/0","protocol":"6","tcp-options":{"destination-port-range":{"min":80,"max":80}}},
    {"source":"0.0.0.0/0","protocol":"6","tcp-options":{"destination-port-range":{"min":443,"max":443}}},
    {"source":"0.0.0.0/0","protocol":"6","tcp-options":{"destination-port-range":{"min":22,"max":22}}},
    {"source":"0.0.0.0/0","protocol":"1","icmp-options":{"type":3,"code":4}},
    {"source":"10.0.0.0/16","protocol":"1","icmp-options":{"type":3}}
  ]' --force
```

> **Note:** On Windows PowerShell, use `--%` to bypass PowerShell's JSON mangling, or write the JSON to a file and use `--from-json file://path/to/file`.

### 5. Create Let's Encrypt ClusterIssuer
```bash
# Run from repository root (e.g. ~/Documents/GitHub/oscar/)
kubectl apply -f deploy/cluster-issuer.yaml
```

This creates two ClusterIssuers:
- `letsencrypt-staging` — for testing (rate limits: 50 certs/week)
- `letsencrypt-prod` — for production (rate limits: 5 certs/week)

The Helm chart ingress annotations already reference `letsencrypt-prod`.

### 6. Configure DNS
Point all subdomains to the node's public IP:

| Domain | Target |
|---|---|
| `oscar-crm.cc` | `159.54.137.54` (production) |
| `dev.oscar-crm.cc` | `159.54.137.54` (dev) |
| `staging.oscar-crm.cc` | `159.54.137.54` (staging) |

Create A records on Cloudflare (or your DNS provider). All three point to the same node IP since we use namespace-based isolation.

---

## Provider Migration Guide

When migrating, the key change is that the target cluster should have a functioning Cloud Controller Manager (CCM) so `type: LoadBalancer` services work correctly. On OKE, the CCM may not be present (workaround: `hostNetwork`). On DOKS, GKE, AKS, or EKS, CCM is built-in and LoadBalancer works out of the box.

### OKE → DigitalOcean (DOKS)

1. Create DOKS cluster:
   ```bash
   doctl kubernetes cluster create oscar --region nyc1 --node-pool "pool=2gb:2"
   doctl kubernetes cluster kubeconfig save oscar
   ```
2. Base64-encode the new kubeconfig and replace the three GitHub Environment secrets.
3. **Optional** — Delete the `imagePullSecrets` block from `values.yaml` if you enable DOKS OIDC.
4. Update `ingress-nginx` to use `type: LoadBalancer` (remove `hostNetwork`).
5. Run `CD` workflow — deploys to the new cluster with zero code changes.

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
| `KUBECONFIG_STAGING` | Env: staging | Same kubeconfig (namespace-based isolation) |
| `KUBECONFIG_PROD` | Env: production | Same kubeconfig (namespace-based isolation) |
| `GITHUB_TOKEN` | Repo (auto) | Used by Actions to push to ghcr.io |

No cloud API keys are stored — only kubeconfigs (which contain short-lived certs).

> **Setup:** All three kubeconfig secrets currently contain the same `oscar-cluster` kubeconfig. Deployments target different namespaces (`oscar-dev`, `oscar-staging`, `oscar-production`) via `helm --namespace <ns>`.
>
> **⚠ Single-cluster risk:** Namespace isolation reduces blast radius for most failure modes (e.g. a bad deploy in `oscar-dev` won't affect `oscar-production`), but a cluster-wide failure (control plane outage, node failure, CVE in kubelet) takes down all three environments simultaneously. To eliminate this shared-fate risk, migrate to separate clusters per environment. This is particularly important when moving past MVP.

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
