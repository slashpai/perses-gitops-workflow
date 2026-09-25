package build

import (
	"fmt"

	"github.com/perses/perses/go-sdk/dashboard"
	"github.com/perses/plugins/prometheus/sdk/go/query"
	labelvalues "github.com/perses/plugins/prometheus/sdk/go/variable/label-values"
)

// VariableDatasource returns a label-values option that sets the datasource.
// Returns a no-op if name is empty.
func VariableDatasource(name string) labelvalues.Option {
	if name == "" {
		return func(*labelvalues.Builder) error { return nil }
	}
	return labelvalues.Datasource(name)
}

// QueryDatasource returns a query option that sets the datasource.
// Returns a no-op if name is empty.
func QueryDatasource(name string) query.Option {
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
