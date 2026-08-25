// Composable dashboard example — imports reusable panels from community-mixins
// and extends them with a custom filesystem panel. Inspired by:
// https://perses.dev/blog/2025/06/10/composable-dashboards----lessons-from-building-perses-community-dashboards/
package nodeexporter

import (
	gitpromql "github.com/slashpai/perses-gitops-workflow/dashboards/promql"

	communityPanels "github.com/perses/community-mixins/pkg/panels/node_exporter"

	commonSdk "github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/perses/go-sdk/panel"
	panelgroup "github.com/perses/perses/go-sdk/panel-group"
	listvariable "github.com/perses/perses/go-sdk/variable/list-variable"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
	timeSeriesPanel "github.com/perses/plugins/timeserieschart/sdk/go"
	"github.com/prometheus/prometheus/model/labels"
)

// BuildComposable builds a dashboard that demonstrates composability:
//   - CPU and Memory panels are imported from community-mixins (reuse)
//   - Filesystem panel is written locally (extend)
//
// This is the pattern recommended by the community-mixins project:
// import panels as Go modules, customise label matchers for your
// environment, and add panels for workload-specific needs.
func BuildComposable(project, datasource string) (dashboard.Builder, error) {
	jobMatcher := &labels.Matcher{Name: "job", Type: labels.MatchEqual, Value: "node-exporter"}
	instanceMatcher := &labels.Matcher{Name: "instance", Type: labels.MatchRegexp, Value: "$instance"}

	return dashboard.New("node-exporter-composable",
		dashboard.ProjectName(project),
		dashboard.Name("Node Exporter / Composable"),
		dashboard.AddVariable("instance",
			listvariable.List(
				labelvalues.PrometheusLabelValues("instance",
					labelvalues.Matchers(`node_uname_info{job="node-exporter",sysname!="Darwin"}`),
					variableDatasource(datasource),
				),
				listvariable.DisplayName("instance"),
				listvariable.AllowAllValue(true),
			),
		),

		// --- Reused from community-mixins ---
		dashboard.AddPanelGroup("CPU (community-mixins)",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(8),
			communityPanels.NodeCPUUsagePercentage(datasource, jobMatcher, instanceMatcher),
			communityPanels.NodeAverage(datasource, jobMatcher, instanceMatcher),
		),
		dashboard.AddPanelGroup("Memory (community-mixins)",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(8),
			communityPanels.NodeMemoryUsageBytes(datasource, jobMatcher, instanceMatcher),
			communityPanels.NodeMemoryUsagePercentage(datasource, jobMatcher, instanceMatcher),
		),

		// --- Custom panel: extend with your own ---
		dashboard.AddPanelGroup("Filesystem (custom)",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Filesystem Used",
				panel.Description("Custom panel — filesystem used ratio from local promql package."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &percentDecimalUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.FilesystemUsedRatio().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}} - {{mountpoint}}"),
					),
				),
			),
		),
	)
}
