// Check dashboard PromQL label matchers against metrics-usage known labels.
//
// Parses rendered PersesDashboard YAMLs, extracts PromQL expressions,
// and cross-references label matchers against the labels collected by
// metrics-usage from a live Prometheus instance.
//
// Usage:
//
//	go run ./cmd/checklabels \
//	  --manifests ../../manifests/dashboards \
//	  --metrics-usage-url http://localhost:8080
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/yaml.v3"
)

type metricsResponse map[string]struct {
	Labels []string `json:"labels"`
}

type finding struct {
	Dashboard string
	Panel     string
	Metric    string
	Label     string
	Query     string
}

func main() {
	manifestDir := flag.String("manifests", "../../manifests/dashboards", "directory containing PersesDashboard YAMLs")
	metricsURL := flag.String("metrics-usage-url", "http://localhost:8080", "metrics-usage base URL")
	flag.Parse()

	knownLabels, err := fetchKnownLabels(*metricsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: fetch metrics-usage: %v\n", err)
		os.Exit(1)
	}
	if len(knownLabels) == 0 {
		fmt.Fprintf(os.Stderr, "warn: metrics-usage returned 0 metrics — collectors may still be running\n")
	}

	files, err := filepath.Glob(filepath.Join(*manifestDir, "*.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: glob: %v\n", err)
		os.Exit(1)
	}

	var findings []finding
	for _, f := range files {
		ff, err := checkFile(f, knownLabels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: %s: %v\n", f, err)
			continue
		}
		findings = append(findings, ff...)
	}

	if len(findings) == 0 {
		fmt.Println("OK: all label matchers reference known labels")
		os.Exit(0)
	}

	fmt.Printf("FOUND %d label mismatch(es):\n\n", len(findings))
	for _, f := range findings {
		fmt.Printf("  dashboard: %s\n", f.Dashboard)
		fmt.Printf("  panel:     %s\n", f.Panel)
		fmt.Printf("  metric:    %s\n", f.Metric)
		fmt.Printf("  label:     %q not found in Prometheus\n", f.Label)
		fmt.Printf("  query:     %s\n\n", truncate(f.Query, 120))
	}
	os.Exit(1)
}

func fetchKnownLabels(baseURL string) (map[string]map[string]bool, error) {
	resp, err := http.Get(baseURL + "/api/v1/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var metrics metricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool, len(metrics))
	for name, m := range metrics {
		labels := make(map[string]bool, len(m.Labels))
		for _, l := range m.Labels {
			labels[l] = true
		}
		labels["__name__"] = true
		result[name] = labels
	}
	return result, nil
}

func checkFile(path string, knownLabels map[string]map[string]bool) ([]finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Config struct {
				Panels map[string]panel `yaml:"panels"`
			} `yaml:"config"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind != "PersesDashboard" {
		return nil, nil
	}

	dashName := doc.Metadata.Name
	var findings []finding

	for _, p := range doc.Spec.Config.Panels {
		panelName := p.Spec.Display.Name
		for _, q := range p.Spec.Queries {
			query := extractQuery(q)
			if query == "" {
				continue
			}
			ff := checkQuery(dashName, panelName, query, knownLabels)
			findings = append(findings, ff...)
		}
	}
	return findings, nil
}

type panel struct {
	Kind string `yaml:"kind"`
	Spec struct {
		Display struct {
			Name string `yaml:"name"`
		} `yaml:"display"`
		Queries []yaml.Node `yaml:"queries"`
	} `yaml:"spec"`
}

func extractQuery(node yaml.Node) string {
	var q struct {
		Spec struct {
			Plugin struct {
				Spec struct {
					Query string `yaml:"query"`
				} `yaml:"spec"`
			} `yaml:"plugin"`
		} `yaml:"spec"`
	}
	if err := node.Decode(&q); err != nil {
		return ""
	}
	return strings.TrimSpace(q.Spec.Plugin.Spec.Query)
}

func checkQuery(dashboard, panelName, query string, knownLabels map[string]map[string]bool) []finding {
	cleaned := replaceVariables(query)

	p := parser.NewParser(parser.Options{})
	expr, err := p.ParseExpr(cleaned)
	if err != nil {
		return nil
	}

	selectors := parser.ExtractSelectors(expr)

	var findings []finding
	seen := make(map[string]bool)

	for _, matchers := range selectors {
		var metricName string
		for _, m := range matchers {
			if m.Name == "__name__" {
				metricName = m.Value
				break
			}
		}
		if metricName == "" {
			continue
		}

		metricLabels, known := knownLabels[metricName]
		if !known {
			continue
		}

		for _, m := range matchers {
			if m.Name == "__name__" {
				continue
			}
			key := metricName + "/" + m.Name
			if seen[key] {
				continue
			}
			seen[key] = true

			if !metricLabels[m.Name] {
				findings = append(findings, finding{
					Dashboard: dashboard,
					Panel:     panelName,
					Metric:    metricName,
					Label:     m.Name,
					Query:     query,
				})
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Label < findings[j].Label
	})
	return findings
}

func replaceVariables(query string) string {
	result := query
	result = strings.ReplaceAll(result, "$__rate_interval", "5m")
	result = strings.ReplaceAll(result, "$__interval", "5m")
	result = strings.ReplaceAll(result, "$__range", "1h")

	var b strings.Builder
	i := 0
	for i < len(result) {
		if result[i] == '$' && i+1 < len(result) && result[i+1] != '_' {
			if result[i+1] == '{' {
				end := strings.Index(result[i:], "}")
				if end > 0 {
					i += end + 1
					b.WriteString(".*")
					continue
				}
			} else {
				j := i + 1
				for j < len(result) && isAlphaNumUnderscore(result[j]) {
					j++
				}
				if j > i+1 {
					b.WriteString(".*")
					i = j
					continue
				}
			}
		}
		b.WriteByte(result[i])
		i++
	}
	return b.String()
}

func isAlphaNumUnderscore(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
