// Package promql builds PromQL with promql-builder — queries are structurally valid at compile time.
//
// SetLabelMatchersV2 is adapted from github.com/perses/community-mixins/pkg/promql/matchers.go.
package promql

import (
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// JobVarV2 is the standard Perses job template variable matcher.
var JobVarV2 = &labels.Matcher{
	Name:  "job",
	Value: "$job",
	Type:  labels.MatchRegexp,
}

// SetLabelMatchersV2 applies label matchers to every vector selector in the expression.
func SetLabelMatchersV2(query parser.Expr, matchers []*labels.Matcher) parser.Expr {
	copy := promqlbuilder.DeepCopyExpr(query)
	for _, l := range matchers {
		copy = labelsSetPromQLV2(copy, l.Type, l.Name, l.Value)
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
