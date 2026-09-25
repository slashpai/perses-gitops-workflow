// Package promql holds custom PromQL helpers for this repo.
// Shared matchers / SetLabelMatchersV2 come from community-mixins/pkg/promql.
package promql

import "github.com/perses/promql-builder/label"

// NamespaceVar is namespace=~"$namespace" so AllowAllValue (.*) works.
// community-mixins NamespaceVarV2 uses exact (=), which breaks All.
var NamespaceVar = label.New("namespace").EqualRegexp("$namespace")
