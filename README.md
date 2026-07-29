# Perses GitOps workflow example

End-to-end example of **Kubernetes-native dashboard-as-code** with PromQL validation for GitOps:

1. **Author** dashboards in Go (`dashboards/`) with the [Perses Go SDK](https://github.com/perses/perses) and [promql-builder](https://github.com/perses/promql-builder)
2. **Validate** in CI — `promqlbuilder.Validate` / `EqualRegexp` catch bad PromQL on the Go path
3. **Render** `PersesDashboard` CRs (`make render`)
4. **Commit** manifests under `manifests/dashboards/`
5. **Sync** with Argo CD → [perses-operator](https://github.com/perses/perses-operator) reconciles

Demo dashboard: **Node Exporter / Nodes** (CPU, load, memory, network).  
Defaults: namespace `perses-dev`, datasource `prometheus-datasource` (same as [perses-operator-examples](https://github.com/slashpai/perses-operator-examples)).

![Node Exporter / Nodes dashboard in Perses](docs/img/perses-node-exporter-dashboard-demo.png)

## Layout

```text
dashboards/                      # Go source of truth
manifests/dashboards/            # generated CRs for GitOps
scripts/setup-prerequisites.sh
scripts/cleanup.sh
scripts/kube-prometheus-values.yaml
deploy/argocd/
```

## Quick start

```sh
# Local (Go 1.26+)
make validate
make render

# Cluster (kubectl, helm, kind) — recommended: create a new kind cluster
make setup-prerequisites

# Tear down when done
make cleanup
```

`setup-prerequisites` prompts you to choose a target; **creating a new kind cluster** (`perses-demo` by default) is recommended. It then installs cert-manager, perses-operator, a **minimal** kube-prometheus-stack (Prometheus operator, Prometheus, Alertmanager, node-exporter), plus Perses and the datasource in `perses-dev`.

`cleanup` removes that stack (or deletes the kind cluster entirely).

Non-interactive:

```sh
YES=true CLUSTER_NAME=perses-demo make setup-prerequisites
YES=true CLUSTER_NAME=perses-demo DELETE_KIND_CLUSTER=true make cleanup
```

## Deploy

```sh
kubectl apply -f manifests/dashboards/

# Or Argo CD: set repoURL in deploy/argocd/application.yaml, then:
kubectl apply -f deploy/argocd/application.yaml

# UI
kubectl -n perses-dev port-forward svc/perses-sample 8080:8080
```

## CI

`make validate` → `make render` → fail if `manifests/dashboards/` drifts.

## Note

`dashboards/go.mod` temporarily `replace`s promql-builder with [`slashpai/promql-builder`](https://github.com/slashpai/promql-builder) until [perses/promql-builder#53](https://github.com/perses/promql-builder/pull/53) lands.

## Related

- [perses-operator](https://github.com/perses/perses-operator)
- [perses-operator-examples](https://github.com/slashpai/perses-operator-examples)
- [promql-builder](https://github.com/perses/promql-builder)
- [community-mixins](https://github.com/perses/community-mixins)

## License

MIT — see [LICENSE](LICENSE).
