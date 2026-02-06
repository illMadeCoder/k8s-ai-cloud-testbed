package crossplane

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// createVCluster creates a virtual cluster using vcluster
func (m *ClusterManager) createVCluster(ctx context.Context, name string, spec experimentsv1alpha1.ClusterSpec) (*unstructured.Unstructured, error) {
	// Use vcluster CRD
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1alpha1",
		Kind:    "VCluster",
	})

	cluster.SetName(name)
	cluster.SetNamespace("crossplane-system")

	// Set labels for experiment tracking
	cluster.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": "experiment-operator",
		"experiments.illm.io/cluster":  name,
	})

	// Build the vcluster spec
	clusterSpec := map[string]interface{}{
		"helmRelease": map[string]interface{}{
			"chart": map[string]interface{}{
				"repo":    "https://charts.loft.sh",
				"name":    "vcluster",
				"version": "0.15.0",
			},
			"values": map[string]interface{}{
				"syncer": map[string]interface{}{
					"extraArgs": []interface{}{
						"--tls-san=" + name + ".vcluster.svc",
					},
				},
				"storage": map[string]interface{}{
					"size": "5Gi",
				},
			},
		},
		"kubernetesVersion": "1.28",
	}

	if err := unstructured.SetNestedMap(cluster.Object, clusterSpec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set vcluster spec: %w", err)
	}

	return cluster, nil
}
