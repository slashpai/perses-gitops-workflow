package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/label"
	"github.com/perses/promql-builder/matrix"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Queries mirror community-mixins node-exporter panels
// (pkg/panels/node_exporter), simplified for this GitOps demo.

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
