# Argo CD Application

Template Application that syncs `manifests/dashboards/` into `perses-dev`.

```sh
make setup-argocd
```

Prompts for `repoURL`, validates it, checks the remote revision exists, installs Argo CD if needed, waits for repo-server readiness, applies this Application, then waits for sync. The committed YAML keeps a `<your-user>` placeholder; the script substitutes your URL at apply time.
