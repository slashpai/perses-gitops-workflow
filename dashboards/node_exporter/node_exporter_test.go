package nodeexporter

import (
	"strings"
	"testing"

	"github.com/slashpai/perses-gitops-workflow/dashboards/build"

	k8syaml "sigs.k8s.io/yaml"
)

func TestBuildOverview(t *testing.T) {
	builder, err := ValidateBuilder(BuildOverview("perses-dev", "prometheus-datasource"))
	if err != nil {
		t.Fatalf("BuildOverview: %v", err)
	}
	if builder.Dashboard.Metadata.Name != "node-exporter-overview" {
		t.Fatalf("unexpected name: %s", builder.Dashboard.Metadata.Name)
	}
	if builder.Dashboard.Metadata.Project != "perses-dev" {
		t.Fatalf("unexpected project: %s", builder.Dashboard.Metadata.Project)
	}
	if len(builder.Dashboard.Spec.Panels) == 0 {
		t.Fatal("expected at least one panel")
	}

	cr := build.ToPersesDashboard(builder)
	yamlOutput, err := k8syaml.Marshal(cr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	output := string(yamlOutput)
	if !strings.Contains(output, "apiVersion: perses.dev/v1alpha2") {
		t.Errorf("yaml missing v1alpha2 apiVersion:\n%s", output)
	}
	if !strings.Contains(output, "Filesystem (custom)") {
		t.Errorf("yaml missing custom Filesystem panel group:\n%s", output)
	}
	if !strings.Contains(output, "node_filesystem_size_bytes") {
		t.Errorf("yaml missing filesystem PromQL from extend panel:\n%s", output)
	}
	if !strings.Contains(output, "job=\"node-exporter\"") {
		t.Errorf("yaml missing kube-prometheus node-exporter job matcher:\n%s", output)
	}
}
