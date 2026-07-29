// Render PersesDashboard manifests from dashboard-as-code for GitOps.
//
// Usage:
//
//	go run ./cmd/render \
//	  --project perses-dev \
//	  --datasource prometheus-datasource \
//	  --output-dir ../../manifests/dashboards
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/slashpai/perses-gitops-workflow/dashboards/build"
	nodeexporter "github.com/slashpai/perses-gitops-workflow/dashboards/node_exporter"

	"github.com/perses/perses/go-sdk/dashboard"
	k8syaml "sigs.k8s.io/yaml"
)

func main() {
	project := flag.String("project", "perses-dev", "Perses project (CR namespace)")
	datasource := flag.String("datasource", "prometheus-datasource", "Prometheus datasource name in Perses")
	outputDir := flag.String("output-dir", "../../manifests/dashboards", "directory for rendered YAML")
	flag.Parse()

	dashboards := []struct {
		name  string
		build func(string, string) (dashboard.Builder, error)
	}{
		{"node-exporter-nodes", nodeexporter.BuildNodes},
		{"node-exporter-filesystem", nodeexporter.BuildFilesystem},
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	for _, d := range dashboards {
		builder, err := d.build(*project, *datasource)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render %s: %v\n", d.name, err)
			os.Exit(1)
		}

		cr := build.ToPersesDashboard(builder)
		out, err := k8syaml.Marshal(cr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal %s: %v\n", d.name, err)
			os.Exit(1)
		}

		path := filepath.Join(*outputDir, builder.Dashboard.Metadata.Name+".yaml")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("rendered %s\n", path)
	}
}
