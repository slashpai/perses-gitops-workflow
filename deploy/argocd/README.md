# Example Argo CD Application

Syncs `manifests/dashboards/` into namespace `perses-dev`.

```sh
make setup-prerequisites   # cluster + minimal Prometheus stack
# Edit repoURL in application.yaml to your fork, then:
kubectl apply -f deploy/argocd/application.yaml
kubectl get persesdashboard -n perses-dev
```
