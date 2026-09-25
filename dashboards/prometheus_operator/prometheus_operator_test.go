package prometheus_operator

import (
	"strings"
	"testing"

	"github.com/slashpai/perses-gitops-workflow/dashboards/build"

	operatorv2 "github.com/perses/perses-operator/api/v1alpha2"
	k8syaml "sigs.k8s.io/yaml"
)

func TestBuildPrometheusOperator(t *testing.T) {
	builder, err := build.ValidateBuilder(BuildPrometheusOperator("perses-dev", "prometheus-datasource"))
	if err != nil {
		t.Fatalf("BuildPrometheusOperator: %v", err)
	}

	if builder.Dashboard.Metadata.Name != "prometheus-operator-health" {
		t.Fatalf("unexpected name: %s", builder.Dashboard.Metadata.Name)
	}
	if builder.Dashboard.Metadata.Project != "perses-dev" {
		t.Fatalf("unexpected project: %s", builder.Dashboard.Metadata.Project)
	}
	if len(builder.Dashboard.Spec.Panels) == 0 {
		t.Fatal("expected at least one panel")
	}
}

func TestBuildPrometheusOperatorCR(t *testing.T) {
	builder, err := build.ValidateBuilder(BuildPrometheusOperator("perses-dev", "prometheus-datasource"))
	if err != nil {
		t.Fatalf("BuildPrometheusOperator: %v", err)
	}

	cr := build.ToPersesDashboard(builder)
	pd, ok := cr.(*operatorv2.PersesDashboard)
	if !ok {
		t.Fatalf("expected *operatorv2.PersesDashboard, got %T", cr)
	}
	if pd.APIVersion != "perses.dev/v1alpha2" {
		t.Fatalf("unexpected apiVersion: %s", pd.APIVersion)
	}
	if pd.Namespace != "perses-dev" {
		t.Fatalf("unexpected namespace: %s", pd.Namespace)
	}
	if pd.Spec.Config.Display == nil || pd.Spec.Config.Display.Name != "Prometheus Operator / Health" {
		t.Fatalf("spec.config.display.name = %v, want Prometheus Operator / Health", pd.Spec.Config.Display)
	}

	yamlOutput, err := k8syaml.Marshal(cr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	output := string(yamlOutput)
	if !strings.Contains(output, "apiVersion: perses.dev/v1alpha2") {
		t.Errorf("yaml missing v1alpha2 apiVersion:\n%s", output)
	}
	if !strings.Contains(output, "config:") {
		t.Errorf("yaml missing spec.config wrapper:\n%s", output)
	}
	if !strings.Contains(output, "prometheus_operator_reconcile_operations_total") {
		t.Errorf("yaml missing prometheus-operator PromQL:\n%s", output)
	}
}
