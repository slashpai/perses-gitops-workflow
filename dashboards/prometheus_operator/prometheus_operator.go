// Package prometheus_operator builds a "Prometheus Operator Health" dashboard
// from scratch using metrics emitted by kube-prometheus-stack.
package prometheus_operator

import (
	"github.com/slashpai/perses-gitops-workflow/dashboards/build"
	gitpromql "github.com/slashpai/perses-gitops-workflow/dashboards/promql"

	commonSdk "github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeSeriesPanel "github.com/perses/plugins/timeserieschart/sdk/go"
)

var (
	opsPerSecondUnit = string(commonSdk.DecimalUnit)
	percentUnit      = string(commonSdk.PercentDecimalUnit)
	secondsUnit      = string(commonSdk.SecondsUnit)
)

// BuildPrometheusOperator builds an operational dashboard for prometheus-operator.
// Demonstrates building a dashboard from scratch for an application that has no
// pre-existing community-mixins panels.
func BuildPrometheusOperator(project, datasource string) (dashboard.Builder, error) {
	return dashboard.New("prometheus-operator-health",
		dashboard.ProjectName(project),
		dashboard.Name("Prometheus Operator / Health"),
		dashboard.AddVariable("job",
			listvariable.List(
				labelvalues.PrometheusLabelValues("job",
					labelvalues.Matchers(`prometheus_operator_reconcile_operations_total`),
					build.VariableDatasource(datasource),
				),
				listvariable.DisplayName("job"),
				listvariable.AllowAllValue(true),
			),
		),
		dashboard.AddVariable("namespace",
			listvariable.List(
				labelvalues.PrometheusLabelValues("namespace",
					labelvalues.Matchers(`prometheus_operator_reconcile_operations_total{job=~"$job"}`),
					build.VariableDatasource(datasource),
				),
				listvariable.DisplayName("namespace"),
				listvariable.AllowAllValue(true),
			),
		),

		// --- Reconciliation ---
		dashboard.AddPanelGroup("Reconciliation",
			panelgroup.PanelsPerLine(3),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Reconcile Rate",
				panel.Description("Rate of reconcile operations by controller."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ReconcileRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("{{controller}}"),
					),
				),
			),
			panelgroup.AddPanel("Reconcile Error Ratio",
				panel.Description("Fraction of reconcile operations that failed."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &percentUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ReconcileErrorRatio().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("{{controller}}"),
					),
				),
			),
			panelgroup.AddPanel("Reconcile Duration (p99 / p50)",
				panel.Description("Latency histogram quantiles for reconcile operations."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &secondsUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ReconcileDurationP99().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("p99 {{controller}}"),
					),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ReconcileDurationP50().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("p50 {{controller}}"),
					),
				),
			),
		),

		// --- Triggers ---
		dashboard.AddPanelGroup("Triggers",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Triggered Events",
				panel.Description("Rate of Kubernetes events triggering reconciliation, by resource and action."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.TriggeredRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("{{triggered_by}} / {{action}}"),
					),
				),
			),
		),

		// --- API Operations ---
		dashboard.AddPanelGroup("API Operations",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("List Operations",
				panel.Description("Rate of Kubernetes API list operations (total vs failed)."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ListOperationsRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("total"),
					),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.ListOperationsFailedRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("failed"),
					),
				),
			),
			panelgroup.AddPanel("Watch Operations",
				panel.Description("Rate of Kubernetes API watch operations (total vs failed)."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.WatchOperationsRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("total"),
					),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.WatchOperationsFailedRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("failed"),
					),
				),
			),
		),

		// --- Workqueue ---
		dashboard.AddPanelGroup("Workqueue",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Workqueue Adds",
				panel.Description("Rate of workqueue adds by name and controller."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.WorkqueueAddsRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("{{name}} / {{controller}}"),
					),
				),
			),
		),

		// --- Status ---
		dashboard.AddPanelGroup("Status",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(6),
			panelgroup.AddPanel("Controller Ready",
				panel.Description("1 when the controller is ready to reconcile, 0 otherwise."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.OperatorReady().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("ready"),
					),
				),
			),
			panelgroup.AddPanel("Status Update Errors",
				panel.Description("Rate of errors when updating status subresources."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &opsPerSecondUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.StatusUpdateErrorRate().Pretty(0),
						build.QueryDatasource(datasource),
						query.SeriesNameFormat("errors/s"),
					),
				),
			),
		),
	)
}
