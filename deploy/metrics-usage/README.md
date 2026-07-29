# metrics-usage — semantic validation for dashboards

Deploys [metrics-usage](https://github.com/perses/metrics-usage) alongside Perses to validate that dashboard PromQL references real metrics and labels in Prometheus.

## How it works

1. **metrics-usage** runs as a Deployment, collecting metrics, labels, and dashboard references from Prometheus and Perses (collectors refresh daily)
2. Its `/api/v1/pending_usages` endpoint exposes metrics referenced in dashboards that were **not found** by the metric collector
3. A **PreSync Job** queries this endpoint before Argo CD syncs dashboards — if pending usages exist, the sync is aborted

## Deploy

```sh
make setup-metrics-usage   # apply manifests, wait for pod ready
make check-metrics         # query pending_usages — {} = all clear
```

Or manually:

```sh
kubectl apply -f deploy/metrics-usage/deployment.yaml
kubectl -n perses-dev wait --for=condition=available deploy/metrics-usage --timeout=120s
```

## Enabling the Argo CD PreSync check

The PreSync Job must be in the same path that the Argo CD Application syncs (`manifests/dashboards/`). Copy it there to enable:

```sh
cp deploy/metrics-usage/presync-check.yaml manifests/dashboards/
git add manifests/dashboards/presync-check.yaml && git commit -m "add PreSync metric check"
```

On the next sync, Argo CD discovers the `argocd.argoproj.io/hook: PreSync` annotation and runs the Job before applying dashboards. The Job retries up to 5 times (configurable via `MAX_RETRIES` / `RETRY_INTERVAL` env vars) to handle cases where collectors are still running.

## Files

| File | Purpose |
| --- | --- |
| `deployment.yaml` | ConfigMap + Deployment + Service for metrics-usage |
| `presync-check.yaml` | Argo CD PreSync Job — copy to `manifests/dashboards/` to enable |

## Configuration

All collectors use `period: 1d` — metrics and labels are stable day-to-day. Collectors also run once on startup so data is available immediately. The labels collector uses `concurrency: 5` to speed up per-metric label queries.

## Scaling notes

- For the demo kind cluster (~500 metrics), the PreSync Job completes in seconds
- For larger Prometheus instances, the long-running Deployment keeps data warm; the PreSync check reads cached results with no Prometheus API calls
