# Argo CD Application

Template Application that syncs `manifests/dashboards/` into `perses-dev`.

```sh
make setup-argocd
```

Prompts for `repoURL`, validates it, checks the remote revision exists, installs Argo CD if needed, waits for repo-server readiness, applies this Application, waits for sync, and deploys `metrics-usage` for PreSync validation. The committed YAML keeps a `<your-user>` placeholder; the script substitutes your URL at apply time.

## How Argo CD discovers changes

Argo CD **polls** the Git remote every **~3 minutes** (`timeout.reconciliation: 180s` by default). When the HEAD commit on `main` changes, Argo CD diffs all manifests under the configured `path` (`manifests/dashboards/`). Any new, modified, or deleted file triggers a sync because `syncPolicy.automated` is enabled.

## PreSync hook

The hook is activated by the **presence** of `presync-check.yaml` in the sync path — no separate enablement is needed. Argo CD reads the `argocd.argoproj.io/hook: PreSync` annotation on the Job and automatically runs it **before** applying dashboards. If the Job fails (e.g. unresolved metrics), the sync is blocked.

See [`../metrics-usage/README.md`](../metrics-usage/README.md) for collector configuration and scaling notes.

## UI

```sh
kubectl -n argocd port-forward svc/argocd-server 8443:443
# → https://localhost:8443 (accept the self-signed cert)
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d; echo
```

## Troubleshooting

### DNS timeouts (`lookup … i/o timeout`) on kind + Podman

Intermittent ClusterIP/DNS issues are common with kind on Podman (especially with multiple clusters). Prefer one cluster; recreate with `make cleanup` / `make setup-prerequisites` if CoreDNS stays broken.
