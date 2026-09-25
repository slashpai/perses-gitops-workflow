package promql

import (
	mixinpromql "github.com/perses/community-mixins/pkg/promql"
	promqlbuilder "github.com/perses/promql-builder"
	"github.com/perses/promql-builder/matrix"
	"github.com/perses/promql-builder/vector"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func withOperatorMatchers(base parser.Expr, extra ...*labels.Matcher) parser.Expr {
	matchers := append([]*labels.Matcher{mixinpromql.JobVarV2, NamespaceVar}, extra...)
	return mixinpromql.SetLabelMatchersV2(base, matchers)
}

// ReconcileRate is rate(prometheus_operator_reconcile_operations_total) by controller.
func ReconcileRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_reconcile_operations_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	).By("controller", "namespace")
	return withOperatorMatchers(base, labelMatchers...)
}

// ReconcileErrorRatio is reconcile errors / total operations by controller.
func ReconcileErrorRatio(labelMatchers ...*labels.Matcher) parser.Expr {
	errors := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_reconcile_errors_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	).By("controller", "namespace")

	total := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_reconcile_operations_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	).By("controller", "namespace")

	base := promqlbuilder.Div(errors, total)
	return withOperatorMatchers(base, labelMatchers...)
}

// ReconcileDurationP99 is histogram_quantile(0.99, reconcile_duration_seconds) by controller.
func ReconcileDurationP99(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.HistogramQuantile(
		0.99,
		promqlbuilder.Sum(
			promqlbuilder.Rate(
				matrix.New(
					vector.New(vector.WithMetricName("prometheus_operator_reconcile_duration_seconds_bucket")),
					matrix.WithRangeAsVariable("$__rate_interval"),
				),
			),
		).By("controller", "namespace", "le"),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// ReconcileDurationP50 is histogram_quantile(0.5, reconcile_duration_seconds) by controller.
func ReconcileDurationP50(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.HistogramQuantile(
		0.5,
		promqlbuilder.Sum(
			promqlbuilder.Rate(
				matrix.New(
					vector.New(vector.WithMetricName("prometheus_operator_reconcile_duration_seconds_bucket")),
					matrix.WithRangeAsVariable("$__rate_interval"),
				),
			),
		).By("controller", "namespace", "le"),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// TriggeredRate is rate(prometheus_operator_triggered_total) by triggered_by, action.
func TriggeredRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_triggered_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	).By("triggered_by", "action")
	return withOperatorMatchers(base, labelMatchers...)
}

// ListOperationsRate is rate(prometheus_operator_list_operations_total).
func ListOperationsRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_list_operations_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// ListOperationsFailedRate is rate(prometheus_operator_list_operations_failed_total).
func ListOperationsFailedRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_list_operations_failed_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// WatchOperationsRate is rate(prometheus_operator_watch_operations_total).
func WatchOperationsRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_watch_operations_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// WatchOperationsFailedRate is rate(prometheus_operator_watch_operations_failed_total).
func WatchOperationsFailedRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_watch_operations_failed_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	)
	return withOperatorMatchers(base, labelMatchers...)
}

// OperatorReady is prometheus_operator_ready gauge.
func OperatorReady(labelMatchers ...*labels.Matcher) parser.Expr {
	base := vector.New(vector.WithMetricName("prometheus_operator_ready"))
	return withOperatorMatchers(base, labelMatchers...)
}

// StatusUpdateErrorRate is rate(prometheus_operator_status_update_errors_total).
func StatusUpdateErrorRate(labelMatchers ...*labels.Matcher) parser.Expr {
	base := promqlbuilder.Sum(
		promqlbuilder.Rate(
			matrix.New(
				vector.New(vector.WithMetricName("prometheus_operator_status_update_errors_total")),
				matrix.WithRangeAsVariable("$__rate_interval"),
			),
		),
	)
	return withOperatorMatchers(base, labelMatchers...)
}
