# metrics-usage — optional semantic validation for dashboards

Deploys [metrics-usage](https://github.com/perses/metrics-usage) alongside Perses to audit whether dashboard PromQL references real metrics and labels in Prometheus. Useful as a **Day-2 operations tool** or **CI check** against a staging Prometheus — not a deploy-time gate.

## How it works

1. **metrics-usage** runs as a Deployment, collecting metrics, labels, and dashboard references from Prometheus and Perses (collectors refresh daily)
2. Its `/api/v1/pending_usages` endpoint exposes metrics referenced in dashboards that were **not found** by the metric collector
3. Run `make check-metrics` or `make check-labels` locally to verify your dashboards

## Deploy

```sh
make setup-metrics-usage   # apply manifests, wait for pod ready
make check-metrics         # metric names exist in Prometheus
make check-labels          # label matchers reference real labels
```

Or manually:

```sh
kubectl apply -f deploy/metrics-usage/deployment.yaml
kubectl -n perses-dev wait --for=condition=available deploy/metrics-usage --timeout=120s
```

## Where to use

| Context | How |
| --- | --- |
| **Local development** | `make check-metrics` / `make check-labels` after authoring new panels |
| **CI pipeline** | Run against a staging Prometheus to catch metric drift |
| **Day-2 operations** | Periodically audit metric inventory vs. dashboard references |

A reference PreSync Job is included at `presync-check.yaml` in this directory for teams that want to experiment with deploy-time gating, but this is **not recommended for production** — coupling deploy availability to a metrics audit service adds fragility.

## Files

| File | Purpose |
| --- | --- |
| `deployment.yaml` | ConfigMap + Deployment + Service for metrics-usage |
| `presync-check.yaml` | Reference Argo CD PreSync Job (not in the sync path) |

## Configuration

All collectors use `period: 1d` — metrics and labels are stable day-to-day. Collectors also run once on startup so data is available immediately. The labels collector uses `concurrency: 5` to speed up per-metric label queries.

## Scaling notes

- For the demo kind cluster (~500 metrics), checks complete in seconds
- For larger Prometheus instances, the long-running Deployment keeps data warm; the check targets read cached results with no Prometheus API calls
