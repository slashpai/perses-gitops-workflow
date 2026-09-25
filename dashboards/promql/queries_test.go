package promql

import (
	"strings"
	"testing"

	communityPanels "github.com/perses/community-mixins/pkg/panels/node_exporter"
	mixinpromql "github.com/perses/community-mixins/pkg/promql"
	"github.com/perses/promql-builder/label"
	"github.com/prometheus/prometheus/promql/parser"
)

func TestFilesystemUsedRatioProducesExpectedPromQL(t *testing.T) {
	communityPanels.SetNodeExporterLabelValue("node-exporter")
	t.Cleanup(func() { communityPanels.SetNodeExporterLabelValue("node") })

	expr := FilesystemUsedRatio(
		label.New("cluster").Equal("$cluster"),
	)
	query := expr.Pretty(0)

	for _, want := range []string{
		`node_filesystem_size_bytes`,
		`node_filesystem_avail_bytes`,
		`job="node-exporter"`,
		`instance=~"$instance"`,
		`cluster="$cluster"`,
		`fstype!=""`,
		`mountpoint!=""`,
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q:\n%s", want, query)
		}
	}
}

func TestFilesystemUsedRatioPassesValidate(t *testing.T) {
	communityPanels.SetNodeExporterLabelValue("node-exporter")
	t.Cleanup(func() { communityPanels.SetNodeExporterLabelValue("node") })

	expr := FilesystemUsedRatio()
	if expr == nil {
		t.Fatal("expected non-nil expression")
	}
}

func TestReconcileRateProducesExpectedPromQL(t *testing.T) {
	expr := ReconcileRate(
		label.New("cluster").Equal("$cluster"),
	)
	query := expr.Pretty(0)

	for _, want := range []string{
		`prometheus_operator_reconcile_operations_total`,
		`job=~"$job"`,
		`namespace=~"$namespace"`,
		`cluster="$cluster"`,
		`by (controller, namespace)`,
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q:\n%s", want, query)
		}
	}
}

func TestReconcileErrorRatioProducesExpectedPromQL(t *testing.T) {
	expr := ReconcileErrorRatio()
	query := expr.Pretty(0)

	for _, want := range []string{
		`prometheus_operator_reconcile_errors_total`,
		`prometheus_operator_reconcile_operations_total`,
		`job=~"$job"`,
		`namespace=~"$namespace"`,
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q:\n%s", want, query)
		}
	}
}

func TestReconcileRatePassesValidate(t *testing.T) {
	if ReconcileRate() == nil {
		t.Fatal("expected non-nil expression")
	}
}

func TestInvalidExpressionPanicsOnValidate(t *testing.T) {
	expr := &parser.BinaryExpr{
		Op:  parser.ADD,
		LHS: FilesystemUsedRatio(),
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected SetLabelMatchersV2 to panic on invalid expression")
		}
	}()
	mixinpromql.SetLabelMatchersV2(expr, nil)
}

func TestInvalidRawStringFailsParse(t *testing.T) {
	if _, err := parser.NewParser(parser.Options{}).ParseExpr(`rate(up[)`); err == nil {
		t.Fatal("expected invalid PromQL string to fail parse")
	}
}

func TestEqualRegexpPanicsOnInvalidPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected EqualRegexp to panic on invalid regex")
		}
	}()
	_ = label.New("instance").EqualRegexp("[invalid")
}
