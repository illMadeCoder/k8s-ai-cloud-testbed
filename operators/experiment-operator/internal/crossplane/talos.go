package crossplane

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// createTalosCluster creates a Talos cluster using Cluster API
func (m *ClusterManager) createTalosCluster(ctx context.Context, name string, spec experimentsv1alpha1.ClusterSpec) (*unstructured.Unstructured, error) {
	// Use Cluster API for Talos
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "Cluster",
	})

	cluster.SetName(name)
	cluster.SetNamespace("crossplane-system")

	// Set labels for experiment tracking
	cluster.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": "experiment-operator",
		"experiments.illm.io/cluster":  name,
		"cluster.x-k8s.io/provider":    "talos",
	})

	// Default values
	nodeCount := spec.NodeCount
	if nodeCount == 0 {
		nodeCount = 3 // Talos usually needs at least 3 nodes for HA
	}

	// Build the cluster spec
	clusterSpec := map[string]interface{}{
		"clusterNetwork": map[string]interface{}{
			"pods": map[string]interface{}{
				"cidrBlocks": []interface{}{"10.244.0.0/16"},
			},
			"services": map[string]interface{}{
				"cidrBlocks": []interface{}{"10.96.0.0/12"},
			},
		},
		"infrastructureRef": map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha3",
			"kind":       "TalosControlPlane",
			"name":       name,
		},
		"controlPlaneRef": map[string]interface{}{
			"apiVersion": "controlplane.cluster.x-k8s.io/v1alpha3",
			"kind":       "TalosControlPlane",
			"name":       name,
		},
	}

	if err := unstructured.SetNestedMap(cluster.Object, clusterSpec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set cluster spec: %w", err)
	}

	// TODO: Also create TalosControlPlane and supporting resources
	// For now, this is a simplified version that assumes the Talos provider
	// will reconcile the cluster based on this Cluster resource

	return cluster, nil
}
