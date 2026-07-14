# Generated PersesDashboard manifests

This directory contains **rendered** Kubernetes custom resources. Do not edit by hand — run from repo root:

```sh
make render
```

Then commit the updated YAML. GitOps controllers (Argo CD, Flux) should sync from this path.

Each file is a `perses.dev/v1alpha2` `PersesDashboard` (dashboard body under `spec.config`) reconciled by [perses-operator](https://github.com/perses/perses-operator) v0.4+.
