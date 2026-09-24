# Perses dashboard-as-code (GitOps example)

Author Prometheus dashboards in Go, validate PromQL at build time, generate `PersesDashboard` CRs, and deploy with Argo CD. This demo focuses on the [Prometheus plugin](https://perses.dev/plugins/docs/prometheus/); Perses also supports other datasources. The YAML under `manifests/dashboards/` is the **delivery artifact** source of truth is `dashboards/`.

1. **Author** in Go (`dashboards/`) with the [Perses Go SDK](https://perses.dev/perses/docs/dac/go/) and [promql-builder](https://github.com/perses/promql-builder) import reusable panels from [community-mixins](https://github.com/perses/community-mixins) and extend them for your environment
2. **Validate** structural PromQL checks at build time (`promqlbuilder.Validate`)
3. **Render** `PersesDashboard` CRs (`make render-dashboards`)
4. **Commit** manifests under `manifests/dashboards/`
5. **Deploy** with Argo CD (or `kubectl apply`) → [perses-operator](https://github.com/perses/perses-operator) reconciles CRs to Perses

## Quick start

```sh
# After changing dashboards/ (Go 1.26+)
make validate-dashboards
make render-dashboards

# kind + cert-manager, operator, minimal kube-prometheus, Perses (perses-dev)
make setup-prerequisites

# Argo CD → sync manifests/dashboards
make setup-argocd

# Tear down stack (or delete the kind cluster)
make cleanup
```

Non-interactive:

```sh
YES=true CLUSTER_NAME=perses-demo make setup-prerequisites
YES=true REPO_URL=https://github.com/<you>/perses-gitops-workflow.git make setup-argocd
YES=true CLUSTER_NAME=perses-demo DELETE_KIND_CLUSTER=true make cleanup
```

## Deploy

```sh
# Direct apply (no Argo CD)
kubectl apply -f manifests/dashboards/

# Or GitOps
make setup-argocd
```

Perses UI: `kubectl -n perses-dev port-forward svc/perses-sample 8080:8080` → [http://localhost:8080](http://localhost:8080)

The Argo CD setup here is for **demo purposes**, for production, follow the [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/). See [`deploy/argocd/README.md`](deploy/argocd/README.md) for UI, polling details, and troubleshooting.

![Argo CD Applications list — perses-dashboards Healthy / Synced](docs/img/argocd-app.png)

![Argo CD synced dashboard resources](docs/img/synced-dashboard.png)

![Argo CD synced new dashboard resource](docs/img/synced-new-dashboard.png)

**Node Exporter / Overview** — community-mixins CPU/Memory (kube-prometheus `job="node-exporter"` matchers) + custom Filesystem panel:

![Node Exporter / Overview dashboard in Perses](docs/img/perses-node-exporter-dashboard-demo.png)

## CI

`make render-dashboards` (runs `validate-dashboards` first) → fail if validation fails or `manifests/dashboards/` drifts.

## Optional: semantic validation with metrics-usage

For Day-2 auditing of metric names and labels against a live Prometheus, see [`docs/metrics-usage.md`](docs/metrics-usage.md). Use a **CronJob** for scheduled drift checks or a one-off **Job** for local/CI validation.

## Related

- [perses-operator](https://github.com/perses/perses-operator)
- [perses-operator-examples](https://github.com/slashpai/perses-operator-examples)
- [promql-builder](https://github.com/perses/promql-builder)
- [community-mixins](https://github.com/perses/community-mixins)
- [metrics-usage](https://github.com/perses/metrics-usage)

## License

MIT — see [LICENSE](LICENSE).
