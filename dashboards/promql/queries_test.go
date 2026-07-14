package promql

import (
	"strings"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func TestCPUUsageRateProducesExpectedPromQL(t *testing.T) {
	expr := CPUUsageRate(
		&labels.Matcher{Name: "namespace", Value: "$namespace", Type: labels.MatchEqual},
	)
	query := expr.Pretty(0)

	want := `rate(process_cpu_seconds_total{job=~"$job",namespace="$namespace"}[$__rate_interval])`
	if query != want {
		t.Fatalf("got %q, want %q", query, want)
	}
}

func TestCPUUsageRateParseableAfterSubstitution(t *testing.T) {
	query := CPUUsageRate().Pretty(0)
	substituted := strings.ReplaceAll(query, "$__rate_interval", "5m")
	substituted = strings.ReplaceAll(substituted, `job=~"$job"`, `job="demo"`)

	if _, err := parser.NewParser(parser.Options{}).ParseExpr(substituted); err != nil {
		t.Fatalf("expected parseable PromQL after substitution: %v\n%s", err, substituted)
	}
}

func TestInvalidRawStringFailsParse(t *testing.T) {
	if _, err := parser.NewParser(parser.Options{}).ParseExpr(`rate(up[)`); err == nil {
		t.Fatal("expected invalid PromQL string to fail parse")
	}
}
