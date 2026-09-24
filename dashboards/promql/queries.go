package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/label"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Queries for custom node-exporter panels (extend community-mixins).

func withNodeMatchers(base parser.Expr, extra ...*labels.Matcher) parser.Expr {
	matchers := append([]*labels.Matcher{NodeJob, InstanceVarV2}, extra...)
	return SetLabelMatchersV2(base, matchers)
}

// FilesystemUsedRatio is (size - avail) / size for node_filesystem_* metrics.
func FilesystemUsedRatio(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Div(
		promqlbuilder.Sub(
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
		),
		vector.New(
			vector.WithMetricName("node_filesystem_size_bytes"),
			vector.WithLabelMatchers(
				label.New("fstype").NotEqual(""),
				label.New("mountpoint").NotEqual(""),
			),
		),
	)
	return withNodeMatchers(base, labelMatchers...)
}
