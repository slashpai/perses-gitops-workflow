# Argo CD Application

> **Note:** This setup is intended for **demo purposes only**. For production deployments, follow the [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/).

Template Application that syncs `manifests/dashboards/` into `perses-dev`.

```sh
make setup-argocd
```

Prompts for `repoURL`, validates it, checks the remote revision exists, installs Argo CD if needed, waits for repo-server readiness, applies this Application, and waits for sync. The committed YAML keeps a `<your-user>` placeholder; the script substitutes your URL at apply time.

## How Argo CD discovers changes

Argo CD **polls** the Git remote every **~3 minutes** (`timeout.reconciliation: 180s` by default). When the HEAD commit on `main` changes, Argo CD diffs all manifests under the configured `path` (`manifests/dashboards/`). Any new, modified, or deleted file triggers a sync because `syncPolicy.automated` is enabled.

## UI

```sh
kubectl -n argocd port-forward svc/argocd-server 8443:443
# → https://localhost:8443 (accept the self-signed cert)
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d; echo
```

## Force a sync

Trigger an immediate sync without waiting for the next poll:

```sh
kubectl -n argocd patch application perses-dashboards --type merge \
  -p '{"operation":{"initiatedBy":{"username":"admin"},"sync":{"revision":"HEAD"}}}'
```

## Troubleshooting

### DNS timeouts (`lookup … i/o timeout`) on kind + Podman

Intermittent ClusterIP/DNS issues are common with kind on Podman (especially with multiple clusters). Prefer one cluster; recreate with `make cleanup` / `make setup-prerequisites` if CoreDNS stays broken.
