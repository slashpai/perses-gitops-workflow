# Example Argo CD Application

Watches `manifests/dashboards/` in this repository and applies `PersesDashboard` resources.

## Prerequisites

- Argo CD installed
- `perses-operator` **v0.4+** CRDs and controller running in the target cluster (v1alpha2 `PersesDashboard`)
- Namespace `monitoring` exists (matches default `--project` in `make render`)

## Apply

```sh
kubectl apply -f application.yaml
```

Edit `spec.source.repoURL` in `application.yaml` to your fork (`https://github.com/<your-user>/perses-gitops-workflow.git`).
