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

### DNS / ClusterIP failures on kind + Podman

Symptoms (often appear together):

- `lookup argocd-repo-server on 10.96.0.10:53: dial udp … i/o timeout`
- `revision main must be resolved`
- `argocd-repo-server` restart loop — liveness `healthz?full=true` times out
- TCP to the `argocd-repo-server` Service or Pod IP fails from other pods

This is a known intermittent **kind + Podman ClusterIP/CNI** issue (worse with multiple kind clusters). Prefer a **single** kind cluster.

**1. Prefer one cluster, then recreate**

```sh
kind get clusters
# delete unused clusters, then:
YES=true CLUSTER_NAME=perses-demo DELETE_KIND_CLUSTER=true make cleanup
YES=true CLUSTER_NAME=perses-demo make setup-prerequisites
YES=true REPO_URL=https://github.com/<you>/perses-gitops-workflow.git make setup-argocd
```

Confirm the remote branch exists before blaming Argo (`revision main must be resolved` is also returned when `main` is not on the remote):

```sh
git ls-remote origin refs/heads/main
```

**2. Soften repo-server probes** (stops restart loops when Redis/DNS is flaky)

```sh
kubectl -n argocd patch deployment argocd-repo-server --type=json -p='[
  {"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/httpGet/path","value":"/healthz"},
  {"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/timeoutSeconds","value":10},
  {"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/failureThreshold","value":10},
  {"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/timeoutSeconds","value":5},
  {"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/failureThreshold","value":10}
]'
```

**3. Bypass ClusterIP** — if recreate is not enough, put repo-server, redis, and application-controller on `hostNetwork` and dial localhost. Intended for **local single-node kind only**, not production.

```sh
for kind_name in deployment/argocd-repo-server deployment/argocd-redis; do
  kubectl -n argocd patch "${kind_name}" --type=strategic -p '{
    "spec":{"template":{"spec":{"hostNetwork":true,"dnsPolicy":"ClusterFirstWithHostNet"}}}
  }'
done
kubectl -n argocd patch statefulset argocd-application-controller --type=strategic -p '{
  "spec":{"template":{"spec":{"hostNetwork":true,"dnsPolicy":"ClusterFirstWithHostNet"}}}
}'

kubectl -n argocd patch cm argocd-cmd-params-cm --type merge -p '{
  "data":{
    "repo.server":"127.0.0.1:8081",
    "redis.server":"127.0.0.1:6379"
  }
}'

kubectl -n argocd rollout restart deployment/argocd-server
kubectl -n argocd rollout status deployment/argocd-repo-server
kubectl -n argocd rollout status deployment/argocd-redis
kubectl -n argocd rollout status statefulset/argocd-application-controller
kubectl -n argocd rollout status deployment/argocd-server

kubectl -n argocd annotate application perses-dashboards \
  argocd.argoproj.io/refresh=hard --overwrite
```

When it works, the Application should be `Synced` / `Healthy` and `kubectl -n perses-dev get persesdashboards` should list the rendered dashboards.

**4. Deploy without Argo CD**

If Argo still cannot sync, apply the rendered manifests directly:

```sh
kubectl apply -f manifests/dashboards/
```
