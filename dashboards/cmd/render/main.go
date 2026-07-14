// Render PersesDashboard manifests from dashboard-as-code for GitOps.
//
// Usage:
//
//	go run ./cmd/render \
//	  --project monitoring \
//	  --datasource prometheus \
//	  --output-dir ../../manifests/dashboards
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/slashpai/perses-gitops-workflow/dashboards/build"
	k8syaml "sigs.k8s.io/yaml"
)

func main() {
	project := flag.String("project", "monitoring", "Perses project (CR namespace)")
	datasource := flag.String("datasource", "prometheus", "Prometheus datasource name in Perses")
	outputDir := flag.String("output-dir", "../../manifests/dashboards", "directory for rendered YAML")
	flag.Parse()

	builder, err := build.ValidateBuilder(build.GoOverview(*project, *datasource))
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}

	cr := build.ToPersesDashboard(builder)
	out, err := k8syaml.Marshal(cr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(*outputDir, builder.Dashboard.Metadata.Name+".yaml")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("rendered %s\n", path)
}
