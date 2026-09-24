package nodeexporter

import (
	"fmt"

	commonSdk "github.com/perses/perses/go-sdk/common"
	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
)

var percentDecimalUnit = string(commonSdk.PercentDecimalUnit)

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
