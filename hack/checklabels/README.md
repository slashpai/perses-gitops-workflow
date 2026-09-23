# checklabels — PromQL label validation against metrics-usage

Custom Go tool that validates label matchers in rendered `PersesDashboard`
YAMLs against label data collected by
[metrics-usage](https://github.com/perses/metrics-usage).

**This capability does not exist in metrics-usage itself.**  metrics-usage
collects which labels exist on each metric (via the labels collector and
`/api/v1/metrics`), but it does not parse dashboard PromQL to verify that
label matchers reference real labels.  This tool fills that gap.

## What it does

1. Reads `PersesDashboard` YAML files from `manifests/dashboards/`
2. Extracts PromQL expressions from panel query plugins
3. Parses each expression and extracts label selectors (using the Prometheus
   PromQL parser)
4. Fetches known metrics and their labels from the metrics-usage
   `/api/v1/metrics` endpoint
5. Reports any label matcher that references a label not present on the
   metric in Prometheus

## Usage

Typically called via `make check-labels`, which:
- Deploys the metrics-usage audit Job
- Port-forwards to the metrics-usage API
- Runs this tool against the rendered manifests

```sh
# Standalone (requires a running metrics-usage instance)
cd hack/checklabels
go run . \
  --manifests ../../manifests/dashboards \
  --metrics-usage-url http://localhost:8080
```

## Module

This is a standalone Go module (`hack/checklabels/go.mod`) separate from
the main `dashboards/` module.  It depends only on
`github.com/prometheus/prometheus` (PromQL parser) and `gopkg.in/yaml.v3`
— it does **not** import the dashboards module or metrics-usage code.
