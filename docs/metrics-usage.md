# Semantic validation with metrics-usage

This repo adapts [metrics-usage](https://github.com/perses/metrics-usage)
for **GitOps dashboard validation** — auditing whether PromQL in rendered
`PersesDashboard` CRs references real metrics and labels in Prometheus.
It is an optional Day-2 / CI tool, not a deploy-time gate.

## How this differs from stock metrics-usage

metrics-usage is designed as a **long-running HTTP server** with periodic
collectors. It exposes usage data via REST but has no built-in pass/fail
audit mode. This repo adapts it for a validation use case that
metrics-usage does not natively support:

| Capability | Stock metrics-usage | Adaptation in this repo |
| --- | --- | --- |
| **Deployment model** | Long-running Deployment or sidecar; collectors run on a timer | Ephemeral **Job / CronJob** — starts metrics-usage, waits for one collector cycle, runs checks, then exits |
| **Metric existence audit** | `/api/v1/pending_usages` passively lists dashboard metrics not yet matched to Prometheus series | Custom **`audit` sidecar** (curl container) polls the API, treats non-empty `pending_usages` as a **failure**, and exits non-zero — turning the endpoint into a CI gate |
| **Label validation** | Labels collector enriches `/api/v1/metrics` with per-metric label names; no query-level checking | Custom **[`hack/checklabels`](../hack/checklabels/)** Go tool parses rendered `PersesDashboard` YAMLs, extracts PromQL selectors, and cross-references label matchers against the labels reported by metrics-usage — **this capability does not exist in metrics-usage itself** |

In short: metrics-usage provides the **data** (which metrics exist, which
labels they carry, which dashboards reference them). This repo wraps that
data in an **ephemeral audit workflow** with pass/fail semantics for CI and
Day-2 drift detection.

## Workload model

| Pattern | When to use | This repo |
| --- | --- | --- |
| **CI Job / workflow step** | Gate PRs against staging Prometheus | Run `metrics-usage` as a GHA service container or `audit-job.yaml` against a test cluster |
| **CronJob** | Production Day-2 drift detection | `cronjob.yaml` — daily audit, alert on Job failure |
| **On-demand Job** | Local dev after authoring panels | `audit-job.yaml` via `make check-metrics` / `make check-labels` |
| **Deployment** | Central API for many consumers (sidecars, teams) | Not used here — only if you need always-on query API |

## How it works

1. **metrics-usage** container starts; metric, labels, and Perses collectors run their first cycle
2. **audit** sidecar polls `/api/v1/metrics` until collectors have populated data
3. Sidecar checks `/api/v1/pending_usages` — non-empty means dashboards reference metrics Prometheus doesn't have
4. Job exits pass/fail — no long-lived Service required
5. *(optional)* `make check-labels` port-forwards to the still-running API and runs `hack/checklabels` for label-level validation

## Deploy

```sh
make setup-metrics-usage   # apply ConfigMap + daily CronJob
make check-metrics         # one-off audit Job (metric existence only)
make check-labels          # one-off Job + Go label matcher check
```

Or manually:

```sh
kubectl apply -f hack/metrics-usage/configmap.yaml
kubectl apply -f hack/metrics-usage/cronjob.yaml
kubectl apply -f hack/metrics-usage/audit-job.yaml
kubectl -n perses-dev wait --for=condition=complete job/metrics-usage-audit --timeout=600s
kubectl -n perses-dev logs job/metrics-usage-audit -c audit
```

## Where to use

| Context | How |
| --- | --- |
| **Local development** | `make check-metrics` / `make check-labels` after authoring new panels |
| **CI pipeline** | Run `audit-job.yaml` against staging, or start metrics-usage in the workflow |
| **Day-2 operations** | `cronjob.yaml` runs daily; alert on failed Jobs |

## Files

| File | Purpose |
| --- | --- |
| [`hack/metrics-usage/configmap.yaml`](../hack/metrics-usage/configmap.yaml) | Collector configuration (Prometheus + Perses endpoints) |
| [`hack/metrics-usage/audit-job.yaml`](../hack/metrics-usage/audit-job.yaml) | On-demand audit (dev, CI, manual) — metrics-usage container + audit sidecar |
| [`hack/metrics-usage/cronjob.yaml`](../hack/metrics-usage/cronjob.yaml) | Scheduled daily audit (production Day-2) |
| [`hack/checklabels/`](../hack/checklabels/) | Go tool for label-level validation (separate go.mod) |

## Configuration

Collectors use `period: 1d` — metrics and labels are stable day-to-day.
Collectors also run once on startup so the audit Job can complete without
waiting a full day. The labels collector uses `concurrency: 5` to speed up
per-metric label queries.

## Production notes

- **CI is the right PR gate** — structural validation (`go test`) in every PR; semantic checks against staging Prometheus in CI
- **CronJob for drift** — catch metric renames or missing series in production over time
- **Central Deployment** — only if multiple teams need a shared metrics-usage API; see [metrics-usage deployment docs](https://github.com/perses/metrics-usage)
- **Upstream feature gap** — if metrics-usage adds a native audit/CI mode in the future, the custom sidecar and checklabels tool here can be retired
