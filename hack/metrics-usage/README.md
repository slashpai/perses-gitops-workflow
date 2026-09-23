# hack/metrics-usage

Kubernetes manifests for running [metrics-usage](https://github.com/perses/metrics-usage) as an ephemeral audit Job or CronJob.

For full details on how this repo adapts metrics-usage for GitOps validation (and how it differs from stock metrics-usage), see **[docs/metrics-usage.md](../../docs/metrics-usage.md)**.

## Files

| File | Purpose |
| --- | --- |
| `configmap.yaml` | Collector configuration (Prometheus + Perses endpoints) |
| `audit-job.yaml` | On-demand audit — metrics-usage container + audit sidecar |
| `cronjob.yaml` | Scheduled daily audit (production Day-2) |

## Quick reference

```sh
make setup-metrics-usage   # apply ConfigMap + daily CronJob
make check-metrics         # one-off audit Job (metric existence)
make check-labels          # one-off Job + label validation via hack/checklabels
```
