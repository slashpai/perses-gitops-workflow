# Perses GitOps workflow example

End-to-end example of **dashboard-as-code** with PromQL safety for GitOps:

1. **Author** dashboards in Go (`dashboards/`) using [promql-builder](https://github.com/perses/promql-builder)
2. **Validate** in CI — `go test` catches structural PromQL errors before merge
3. **Render** `PersesDashboard` CRs as `perses.dev/v1alpha2` (`make render`)
4. **Commit** manifests under `manifests/dashboards/`
5. **Sync** with Argo CD / Flux → [perses-operator](https://github.com/perses/perses-operator) reconciles

Inspired by the [Perses community dashboards](https://github.com/perses/community-mixins) talk on GitOps and the exploration in [perses-harness](https://github.com/slashpai/perses-harness).

## Why dashboard-as-code for GitOps?

| JSON/YAML in Git | Dashboard-as-code |
| ---------------- | ----------------- |
| `"query": "rate(up[)"` can merge if nobody spots it | `go test` fails — invalid AST cannot compile |
| CUE validates shape only | PromQL structure validated at build time |
| Manual review of query strings | Typed Go + unit tests in CI |

Perses variables (`$job`, `$__rate_interval`) are substituted at query time; CI validates structure by building queries in Go and optionally parsing after substitution (see `dashboards/promql/queries_test.go`).

## Repository layout

```text
dashboards/           # Go module — source of truth
  promql/             # promql-builder queries (compile-time safe)
  build/              # Perses go-sdk dashboard definition
  cmd/render/         # emits PersesDashboard YAML
manifests/
  dashboards/         # generated CRs — commit these for GitOps
deploy/
  argocd/             # example Argo CD Application
.github/workflows/    # CI: test + render
```

## Quick start

```sh
# Prerequisites: Go 1.26+
make validate    # run PromQL + dashboard tests
make render      # write manifests/dashboards/go-overview.yaml
```

### Customize

```sh
make render PROJECT=my-team DATASOURCE=platform-prometheus
```

## GitOps deployment

### Prerequisites

- Kubernetes cluster with [perses-operator](https://github.com/perses/perses-operator) **v0.4+** installed (v1alpha2 `PersesDashboard` API)
- A `Perses` instance and `PersesDatasource` named `prometheus` (or match `--datasource`)
- Argo CD or Flux (optional)

### Apply manually

```sh
kubectl apply -f manifests/dashboards/
```

### Argo CD

Copy and edit `deploy/argocd/application.yaml`, then:

```sh
kubectl apply -f deploy/argocd/application.yaml
```

Argo CD watches `manifests/dashboards/` in this repo and syncs `PersesDashboard` resources.

## CI workflow

On every PR:

1. `go test ./...` — PromQL + dashboard build tests
2. `make render` — regenerate manifests
3. `git diff --exit-code manifests/dashboards/` — fail if committed YAML drift from render

## Relationship to community-mixins

This repo is a **minimal standalone example** of the same dashboard-as-code + GitOps pattern used in [community-mixins](https://github.com/perses/community-mixins):

| | perses-gitops-workflow | community-mixins |
| --- | --- | --- |
| Scope | One demo dashboard (`go-overview`) | Full production dashboard library |
| Output | `manifests/dashboards/*.yaml` for your GitOps repo | `examples/dashboards/operator/` via `make build-dashboards` |
| PromQL helpers | Local `dashboards/promql/` (from community-mixins patterns) | `pkg/promql/` |

For production dashboards, use community-mixins. Use this repo to learn the GitOps render → commit → sync workflow.

## Workflow diagram

```text
  Go dashboards (dashboards/)
           │
           ▼
    ┌──────────────┐
    │  CI: go test │  ← PromQL structural validation
    └──────┬───────┘
           ▼
    ┌──────────────┐
    │ make render  │  ← PersesDashboard YAML
    └──────┬───────┘
           ▼
    ┌──────────────┐
    │  Git commit  │
    └──────┬───────┘
           ▼
    ┌──────────────┐
    │ Argo CD sync │
    └──────┬───────┘
           ▼
    perses-operator → Perses UI
```

## Related

- [community-mixins](https://github.com/perses/community-mixins) — production-scale dashboard-as-code

## License

MIT — see [LICENSE](LICENSE).
