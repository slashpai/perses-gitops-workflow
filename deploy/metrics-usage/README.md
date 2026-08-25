# metrics-usage — optional semantic validation for dashboards

Deploys [metrics-usage](https://github.com/perses/metrics-usage) to audit whether dashboard PromQL references real metrics and labels in Prometheus. Useful as a **Day-2 operations tool** or **CI check** against a staging Prometheus — not a deploy-time gate.

## Workload model

metrics-usage is an HTTP server with periodic collectors. For this optional audit use case, use a **Job** or **CronJob** rather than a long-lived Deployment.

| Pattern | When to use | This repo |
| --- | --- | --- |
| **CI Job / workflow step** | Gate PRs against staging Prometheus | Run `metrics-usage` as a GHA service container or `audit-job.yaml` against a test cluster |
| **CronJob** | Production Day-2 drift detection | `cronjob.yaml` — daily audit, alert on Job failure |
| **On-demand Job** | Local dev after authoring panels | `audit-job.yaml` via `make check-metrics` / `make check-labels` |
| **Deployment** | Central API for many consumers (sidecars, teams) | Not used here — only if you need always-on query API |

## How it works

1. **metrics-usage** container starts and collectors scrape Prometheus + Perses
2. **audit** sidecar waits for `/api/v1/metrics`, then checks `/api/v1/pending_usages`
3. Job exits pass/fail — no long-lived Service required

## Deploy

```sh
make setup-metrics-usage   # apply ConfigMap + daily CronJob
make check-metrics         # one-off audit Job
make check-labels          # one-off Job + Go label matcher check
```

Or manually:

```sh
kubectl apply -f deploy/metrics-usage/configmap.yaml
kubectl apply -f deploy/metrics-usage/cronjob.yaml
kubectl apply -f deploy/metrics-usage/audit-job.yaml
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
| `configmap.yaml` | Collector configuration |
| `audit-job.yaml` | On-demand audit (dev, CI, manual) |
| `cronjob.yaml` | Scheduled daily audit (production Day-2) |

## Configuration

Collectors use `period: 1d` — metrics and labels are stable day-to-day. Collectors also run once on startup so the audit Job can complete without waiting a full day. The labels collector uses `concurrency: 5` to speed up per-metric label queries.

## Production notes

- **CI is the right PR gate** — structural validation (`go test`) in every PR; semantic checks against staging Prometheus in CI
- **CronJob for drift** — catch metric renames or missing series in production over time
- **Central Deployment** — only if multiple teams need a shared metrics-usage API; see [metrics-usage deployment docs](https://github.com/perses/metrics-usage)
