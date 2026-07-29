package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/label"
	"github.com/perses/promql-builder/matrix"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Queries for node-exporter dashboard panels.

func withNodeMatchers(base parser.Expr, extra ...*labels.Matcher) parser.Expr {
	matchers := append([]*labels.Matcher{NodeJob, InstanceVarV2}, extra...)
	return SetLabelMatchersV2(base, matchers)
}

// CPUUsageIdleRatio is CPU busy ratio (1 - avg idle rate) per instance.
func CPUUsageIdleRatio(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sub(
		&parser.NumberLiteral{Val: 1},
		promqlbuilder.Avg(
			promqlbuilder.Rate(
				matrix.New(
					vector.New(
						vector.WithMetricName("node_cpu_seconds_total"),
						vector.WithLabelMatchers(
							label.New("mode").Equal("idle"),
						),
					),
					matrix.WithRangeAsVariable("$__rate_interval"),
				),
			),
		).By("instance"),
	)
	return withNodeMatchers(base, labelMatchers...)
}

// MemoryUsageRatio is (1 - MemAvailable / MemTotal).
func MemoryUsageRatio(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sub(
		&parser.NumberLiteral{Val: 1},
		promqlbuilder.Div(
			vector.New(vector.WithMetricName("node_memory_MemAvailable_bytes")),
			vector.New(vector.WithMetricName("node_memory_MemTotal_bytes")),
		),
	)
	return withNodeMatchers(base, labelMatchers...)
}

// NetworkReceiveBytes is receive throughput excluding loopback.
func NetworkReceiveBytes(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Rate(
		matrix.New(
			vector.New(
				vector.WithMetricName("node_network_receive_bytes_total"),
				vector.WithLabelMatchers(
					label.New("device").NotEqual("lo"),
				),
			),
			matrix.WithRangeAsVariable("$__rate_interval"),
		),
	)
	return withNodeMatchers(base, labelMatchers...)
}

// NetworkTransmitBytes is transmit throughput excluding loopback.
func NetworkTransmitBytes(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Rate(
		matrix.New(
			vector.New(
				vector.WithMetricName("node_network_transmit_bytes_total"),
				vector.WithLabelMatchers(
					label.New("device").NotEqual("lo"),
				),
			),
			matrix.WithRangeAsVariable("$__rate_interval"),
		),
	)
	return withNodeMatchers(base, labelMatchers...)
}

// Load1 is the 1-minute load average.
func Load1(labelMatchers ...*labels.Matcher) parser.Expr {
	base := vector.New(vector.WithMetricName("node_load1"))
	return withNodeMatchers(base, labelMatchers...)
}

// FilesystemUsedRatio is (size - avail) / size for node_filesystem_* metrics.
//
// DEMO BUG: the Div is missing its RHS on purpose so query construction
// (SetLabelMatchersV2 / promqlbuilder.Validate) fails in CI
// (make validate-dashboards).
func FilesystemUsedRatio(labelMatchers ...*labels.Matcher) parser.Expr {
	used := promqlbuilder.Sub(
		vector.New(
			vector.WithMetricName("node_filesystem_size_bytes"),
			vector.WithLabelMatchers(
				label.New("fstype").NotEqual(""),
				label.New("mountpoint").NotEqual(""),
			),
		),
		vector.New(
			vector.WithMetricName("node_filesystem_avail_bytes"),
			vector.WithLabelMatchers(
				label.New("fstype").NotEqual(""),
				label.New("mountpoint").NotEqual(""),
			),
		),
	)
	base := &parser.BinaryExpr{
		Op:  parser.DIV,
		LHS: used,
		// RHS intentionally omitted for the CI demo.
	}
	return withNodeMatchers(base, labelMatchers...)
}
