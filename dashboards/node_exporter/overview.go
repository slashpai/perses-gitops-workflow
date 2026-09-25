// Node Exporter overview — imports reusable panels from community-mixins
// and extends them with a custom filesystem panel.

package nodeexporter

import (
	"fmt"

	gitpromql "github.com/slashpai/perses-gitops-workflow/dashboards/promql"

	communityPanels "github.com/perses/community-mixins/pkg/panels/node_exporter"
	mixinpromql "github.com/perses/community-mixins/pkg/promql"

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

// BuildOverview builds the Node Exporter / Overview dashboard:
//   - CPU and Memory panels are imported from community-mixins (reuse)
//   - Filesystem panel is written locally (extend)
//
// Job label follows community-mixins library usage:
// https://github.com/perses/community-mixins#library-usage
func BuildOverview(project, datasource string) (dashboard.Builder, error) {
	// kube-prometheus-stack scrapes node-exporter as job="node-exporter"
	// (mixins default is "node").
	communityPanels.SetNodeExporterLabelValue("node-exporter")
	jobValue := communityPanels.GetNodeExporterLabelValue()
	jobMatcher := &labels.Matcher{Name: "job", Type: labels.MatchEqual, Value: jobValue}
	instanceMatcher := mixinpromql.InstanceVarV2

	return dashboard.New("node-exporter-overview",
		dashboard.ProjectName(project),
		dashboard.Name("Node Exporter / Overview"),
		dashboard.AddVariable("instance",
			listvariable.List(
				labelvalues.PrometheusLabelValues("instance",
					labelvalues.Matchers(fmt.Sprintf(`node_uname_info{job="%s",sysname!="Darwin"}`, jobValue)),
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
