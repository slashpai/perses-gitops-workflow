package build

import (
	"fmt"

	gitpromql "github.com/slashpai/perses-gitops-workflow/dashboards/promql"

	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeSeriesPanel "github.com/perses/plugins/timeserieschart/sdk/go"
)

// GoOverview builds a minimal dashboard-as-code example for GitOps.
// PromQL is composed with promql-builder (see dashboards/promql) — not hand-written YAML strings.
func GoOverview(project, datasource string) (dashboard.Builder, error) {
	return dashboard.New("go-overview",
		dashboard.ProjectName(project),
		dashboard.Name("Go / Overview"),
		dashboard.AddVariable("job",
			listvariable.List(
				labelvalues.PrometheusLabelValues("job",
					labelvalues.Matchers("go_goroutines"),
					variableDatasource(datasource),
				),
				listvariable.DisplayName("job"),
			),
		),
		dashboard.AddPanelGroup("Runtime",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("CPU usage",
				panel.Description("CPU seconds rate — query built with promql-builder at compile time."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
					timeSeriesPanel.WithVisual(timeSeriesPanel.Visual{
						Display:   timeSeriesPanel.LineDisplay,
						LineWidth: 0.25,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.CPUUsageRate().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{job}}"),
					),
				),
			),
		),
	)
}

func variableDatasource(name string) labelvalues.Option {
	if name == "" {
		return func(*labelvalues.Builder) error { return nil }
	}
	return labelvalues.Datasource(name)
}

func queryDatasource(name string) query.Option {
	if name == "" {
		return func(*query.Builder) error { return nil }
	}
	return query.Datasource(name)
}

// ValidateBuilder ensures the dashboard builder completed without SDK errors.
func ValidateBuilder(builder dashboard.Builder, err error) (dashboard.Builder, error) {
	if err != nil {
		return dashboard.Builder{}, fmt.Errorf("build dashboard: %w", err)
	}
	return builder, nil
}
