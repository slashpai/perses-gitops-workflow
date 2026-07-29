package nodeexporter

import (
	"fmt"

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
	percentDecimalUnit  = string(commonSdk.PercentDecimalUnit)
	bytesPerSecondsUnit = string(commonSdk.BytesPerSecondsUnit)
	decimalUnit         = string(commonSdk.DecimalUnit)
)

// BuildNodes builds a Node Exporter / Nodes dashboard for GitOps demos.
// PromQL uses promql-builder + Validate — not hand-written YAML strings.
func BuildNodes(project, datasource string) (dashboard.Builder, error) {
	return dashboard.New("node-exporter-nodes",
		dashboard.ProjectName(project),
		dashboard.Name("Node Exporter / Nodes"),
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
		dashboard.AddPanelGroup("CPU",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("CPU Usage",
				panel.Description("CPU busy ratio (1 − idle) via promql-builder + Validate."),
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
						gitpromql.CPUUsageIdleRatio().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}}"),
					),
				),
			),
			panelgroup.AddPanel("Load (1m)",
				panel.Description("node_load1 — 1-minute load average."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &decimalUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.Load1().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}}"),
					),
				),
			),
		),
		dashboard.AddPanelGroup("Memory",
			panelgroup.PanelsPerLine(1),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Memory Usage",
				panel.Description("1 − MemAvailable/MemTotal — node-exporter memory utilisation."),
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
						gitpromql.MemoryUsageRatio().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}}"),
					),
				),
			),
		),
		dashboard.AddPanelGroup("Network",
			panelgroup.PanelsPerLine(2),
			panelgroup.PanelHeight(8),
			panelgroup.AddPanel("Network Received",
				panel.Description("node_network_receive_bytes_total rate (excluding lo)."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &bytesPerSecondsUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.NetworkReceiveBytes().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}} - {{device}}"),
					),
				),
			),
			panelgroup.AddPanel("Network Transmitted",
				panel.Description("node_network_transmit_bytes_total rate (excluding lo)."),
				timeSeriesPanel.Chart(
					timeSeriesPanel.WithYAxis(timeSeriesPanel.YAxis{
						Format: &commonSdk.Format{Unit: &bytesPerSecondsUnit},
					}),
					timeSeriesPanel.WithLegend(timeSeriesPanel.Legend{
						Position: timeSeriesPanel.BottomPosition,
						Mode:     timeSeriesPanel.ListMode,
					}),
				),
				panel.AddQuery(
					query.PromQL(
						gitpromql.NetworkTransmitBytes().Pretty(0),
						queryDatasource(datasource),
						query.SeriesNameFormat("{{instance}} - {{device}}"),
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
