package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/matrix"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// CPUUsageRate returns a rate() query for process_cpu_seconds_total with Perses variables.
func CPUUsageRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Rate(
		matrix.New(
			vector.New(vector.WithMetricName("process_cpu_seconds_total")),
			matrix.WithRangeAsVariable("$__rate_interval"),
		),
	)

	matchers := append([]*labels.Matcher{JobVarV2}, labelMatchers...)
	return SetLabelMatchersV2(base, matchers)
}
