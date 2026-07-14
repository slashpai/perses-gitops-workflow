package build

import (
	"strings"
	"testing"

	operatorv2 "github.com/perses/perses-operator/api/v1alpha2"
	k8syaml "sigs.k8s.io/yaml"
)

func TestGoOverviewBuilds(t *testing.T) {
	builder, err := ValidateBuilder(GoOverview("monitoring", "prometheus"))
	if err != nil {
		t.Fatalf("GoOverview: %v", err)
	}

	if builder.Dashboard.Metadata.Name != "go-overview" {
		t.Fatalf("unexpected name: %s", builder.Dashboard.Metadata.Name)
	}
	if builder.Dashboard.Metadata.Project != "monitoring" {
		t.Fatalf("unexpected project: %s", builder.Dashboard.Metadata.Project)
	}
	if len(builder.Dashboard.Spec.Panels) == 0 {
		t.Fatal("expected at least one panel")
	}
}

func TestToPersesDashboardCR_v1alpha2(t *testing.T) {
	builder, err := ValidateBuilder(GoOverview("monitoring", "prometheus"))
	if err != nil {
		t.Fatalf("GoOverview: %v", err)
	}

	cr := ToPersesDashboard(builder)
	pd, ok := cr.(*operatorv2.PersesDashboard)
	if !ok {
		t.Fatalf("expected *operatorv2.PersesDashboard, got %T", cr)
	}
	if pd.APIVersion != "perses.dev/v1alpha2" {
		t.Fatalf("unexpected apiVersion: %s", pd.APIVersion)
	}
	if pd.Namespace != "monitoring" {
		t.Fatalf("unexpected namespace: %s", pd.Namespace)
	}
	if pd.Spec.Config.Display == nil || pd.Spec.Config.Display.Name != "Go / Overview" {
		t.Fatalf("spec.config.display.name = %v, want Go / Overview", pd.Spec.Config.Display)
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
	if strings.Contains(output, "perses.dev/v1alpha1") {
		t.Errorf("yaml should not contain v1alpha1:\n%s", output)
	}
}
