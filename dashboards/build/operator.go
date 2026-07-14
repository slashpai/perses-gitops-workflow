package build

import (
	operatorv2 "github.com/perses/perses-operator/api/v1alpha2"
	"github.com/perses/perses/go-sdk/dashboard"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ToPersesDashboard converts a dashboard builder into a PersesDashboard CR for GitOps.
// Emits perses.dev/v1alpha2 with dashboard content under spec.config (perses-operator v0.4+).
func ToPersesDashboard(builder dashboard.Builder) runtime.Object {
	return &operatorv2.PersesDashboard{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersesDashboard",
			APIVersion: "perses.dev/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Dashboard.Metadata.Name,
			Namespace: builder.Dashboard.Metadata.Project,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "perses-dashboard",
				"app.kubernetes.io/instance":   builder.Dashboard.Metadata.Name,
				"app.kubernetes.io/part-of":    "perses-operator",
				"app.kubernetes.io/component":  "dashboard",
				"app.kubernetes.io/managed-by": "perses-gitops-workflow",
			},
		},
		Spec: operatorv2.PersesDashboardSpec{
			Config: operatorv2.Dashboard{
				DashboardSpec: builder.Dashboard.Spec,
			},
		},
	}
}
