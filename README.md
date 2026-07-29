# Perses dashboard-as-code (GitOps example)

Author dashboards in Go, validate PromQL before render, generate `PersesDashboard` CRs, then deploy with kubectl or Argo CD. The YAML under `manifests/dashboards/` is the **delivery artifact** — source of truth is `dashboards/`.

1. **Author** in Go (`dashboards/`) with the [Perses Go SDK](https://github.com/perses/perses) and [promql-builder](https://github.com/perses/promql-builder)
2. **Validate** — `promqlbuilder.Validate` / `equalRegexp` catch bad PromQL on the Go path
3. **Render** `PersesDashboard` CRs (`make render-dashboards`)
4. **Commit** manifests under `manifests/dashboards/`
5. **Sync** with Argo CD (or `kubectl apply`) → [perses-operator](https://github.com/perses/perses-operator) reconciles

## Quick start

```sh
# After changing dashboards/ (Go 1.26+)
make validate-dashboards
make render-dashboards

# kind + cert-manager, operator, minimal kube-prometheus, Perses (perses-dev)
make setup-prerequisites

# Argo CD → sync manifests/dashboards (push main first; prompts for fork repoURL)
make setup-argocd

# (Optional) Deploy metrics-usage for semantic validation
make setup-metrics-usage
make check-metrics

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

# Perses UI
kubectl -n perses-dev port-forward svc/perses-sample 8080:8080
# → http://localhost:8080

# Argo CD UI
kubectl -n argocd port-forward svc/argocd-server 8443:443
# → https://localhost:8443 (user: admin)
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d; echo
```

![Argo CD Applications list — perses-dashboards Healthy / Synced](docs/img/argocd-app.png)

Argo CD resource tree after the initial Nodes dashboard sync:

![Argo CD synced node-exporter-nodes PersesDashboard](docs/img/synced-dashboard.png)

After merging a new dashboard (Filesystem), both CRs sync:

![Argo CD synced Nodes and Filesystem PersesDashboards](docs/img/synced-new-dashboard.png)

**Node Exporter / Nodes** (CPU, load, memory, network).

![Node Exporter / Nodes dashboard in Perses](docs/img/perses-node-exporter-dashboard-demo.png)

**Node Exporter / Filesystem** (used disk space ratio).

![Node Exporter / Filesystem dashboard in Perses](docs/img/perses-node-exporter-fs-dashboard-demo.png)

## CI

`make validate-dashboards` → `make render-dashboards` → fail if `manifests/dashboards/` drifts.

## Semantic validation with metrics-usage

[metrics-usage](https://github.com/perses/metrics-usage) cross-references dashboard PromQL against a live Prometheus — catching missing metrics and label typos that structural validation cannot.

```sh
make setup-metrics-usage   # deploy metrics-usage (after setup-prerequisites)
make check-metrics         # query pending_usages — {} = all clear
```

Collectors run once on startup, then refresh daily (`period: 1d`). The `check-metrics` target queries the `/api/v1/pending_usages` endpoint — any metrics referenced in dashboards but not found in Prometheus will be listed.

To gate Argo CD syncs on metric validation, `make setup-argocd` prompts you to enable the PreSync check. It deploys metrics-usage if needed and copies the PreSync Job into `manifests/dashboards/`.

Non-interactive:

```sh
ENABLE_PRESYNC_CHECK=true make setup-argocd
```

Or manually: `cp deploy/metrics-usage/presync-check.yaml manifests/dashboards/` and commit. Argo CD discovers the `argocd.argoproj.io/hook: PreSync` annotation and runs the Job before every sync. See [`deploy/metrics-usage/README.md`](deploy/metrics-usage/README.md).

## Troubleshooting

### DNS timeouts (`lookup … i/o timeout`) on kind + Podman

Intermittent ClusterIP/DNS issues are common with kind on Podman (especially with multiple clusters). Prefer one cluster; recreate with `make cleanup` / `make setup-prerequisites` if CoreDNS stays broken.

## Related

- [perses-operator](https://github.com/perses/perses-operator)
- [perses-operator-examples](https://github.com/slashpai/perses-operator-examples)
- [promql-builder](https://github.com/perses/promql-builder)
- [community-mixins](https://github.com/perses/community-mixins)
- [metrics-usage](https://github.com/perses/metrics-usage)

## License

MIT — see [LICENSE](LICENSE).
