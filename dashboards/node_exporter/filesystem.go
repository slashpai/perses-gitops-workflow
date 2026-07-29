package nodeexporter

import (
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

// BuildFilesystem builds a Node Exporter / Filesystem dashboard using
// node_filesystem_* metrics.
func BuildFilesystem(project, datasource string) (dashboard.Builder, error) {
	return dashboard.New("node-exporter-filesystem",
		dashboard.ProjectName(project),
		dashboard.Name("Node Exporter / Filesystem"),
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
		dashboard.AddPanelGroup("Filesystem",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Filesystem Used",
				panel.Description("Used disk space ratio from node_filesystem_size/avail."),
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
