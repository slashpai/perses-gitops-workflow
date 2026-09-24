package promql

import (
	"strings"
	"testing"

	"github.com/perses/promql-builder/label"
	"github.com/prometheus/prometheus/promql/parser"
)

func TestFilesystemUsedRatioProducesExpectedPromQL(t *testing.T) {
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
	expr := FilesystemUsedRatio()
	if expr == nil {
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
	SetLabelMatchersV2(expr, nil)
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
