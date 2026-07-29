// Package promql builds PromQL with promql-builder — queries are validated at
// build time via promqlbuilder.Validate.
package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/label"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// InstanceVarV2 is the Perses instance template variable matcher (node-exporter).
var InstanceVarV2 = label.New("instance").EqualRegexp("$instance")

// NodeJob is the default kube-prometheus-stack node-exporter job label
// (prometheus-node-exporter.podLabels.jobLabel = node-exporter).
var NodeJob = label.New("job").Equal("node-exporter")

// SetLabelMatchersV2 applies label matchers to every vector selector, then
// validates the resulting expression (including Perses dashboard variables).
func SetLabelMatchersV2(query parser.Expr, matchers []*labels.Matcher) parser.Expr {
	copy := promqlbuilder.DeepCopyExpr(query)
	for _, l := range matchers {
		copy = labelsSetPromQLV2(copy, l.Type, l.Name, l.Value)
	}
	if err := promqlbuilder.Validate(copy); err != nil {
		panic(err)
	}
	return copy
}

func labelsSetPromQLV2(query parser.Expr, matchType labels.MatchType, name, value string) parser.Expr {
	if name == "" || value == "" {
		return query
	}

	promqlbuilder.Inspect(query, func(node parser.Node, _ []parser.Node) error {
		if n, ok := node.(*parser.VectorSelector); ok {
			var found bool
			for i, l := range n.LabelMatchers {
				if l.Name == name {
					n.LabelMatchers[i].Type = matchType
					n.LabelMatchers[i].Value = value
					found = true
				}
			}
			if !found {
				n.LabelMatchers = append(n.LabelMatchers, &labels.Matcher{
					Type:  matchType,
					Name:  name,
					Value: value,
				})
			}
		}
		return nil
	})

	return query
}
