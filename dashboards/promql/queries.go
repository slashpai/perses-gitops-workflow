package promql

import (
	communityPanels "github.com/perses/community-mixins/pkg/panels/node_exporter"
	mixinpromql "github.com/perses/community-mixins/pkg/promql"
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/label"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// Queries for custom node-exporter panels (extend community-mixins).

func nodeJobMatcher() *labels.Matcher {
	return &labels.Matcher{
		Name:  "job",
		Type:  labels.MatchEqual,
		Value: communityPanels.GetNodeExporterLabelValue(),
	}
}

func withNodeMatchers(base parser.Expr, extra ...*labels.Matcher) parser.Expr {
	matchers := append([]*labels.Matcher{nodeJobMatcher(), mixinpromql.InstanceVarV2}, extra...)
	return mixinpromql.SetLabelMatchersV2(base, matchers)
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
