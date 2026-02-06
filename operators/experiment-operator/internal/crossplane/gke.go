package crossplane

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// createGKECluster creates a GKE cluster using Crossplane/Upbound GCP provider
func (m *ClusterManager) createGKECluster(ctx context.Context, name string, spec experimentsv1alpha1.ClusterSpec) (*unstructured.Unstructured, error) {
	// Use Upbound GCP provider for GKE
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "container.gcp.upbound.io",
		Version: "v1beta1",
		Kind:    "Cluster",
	})

	cluster.SetName(name)
	cluster.SetNamespace("crossplane-system")

	// Set labels for experiment tracking
	cluster.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": "experiment-operator",
		"experiments.illm.io/cluster":  name,
	})

	// Default values
	zone := spec.Zone
	if zone == "" {
		zone = "us-central1-a"
	}

	nodeCount := spec.NodeCount
	if nodeCount == 0 {
		nodeCount = 2
	}

	machineType := spec.MachineType
	if machineType == "" {
		machineType = "e2-medium"
	}

	// Build the cluster spec
	clusterSpec := map[string]interface{}{
		"forProvider": map[string]interface{}{
			"location": zone,
			"initialNodeCount": nodeCount,
			"nodeConfig": []interface{}{
				map[string]interface{}{
					"machineType": machineType,
					"preemptible": spec.Preemptible,
					"oauthScopes": []interface{}{
						"https://www.googleapis.com/auth/cloud-platform",
					},
					"diskSizeGb": 50,
					"diskType":   "pd-standard",
					"metadata": map[string]interface{}{
						"disable-legacy-endpoints": "true",
					},
					"shieldedInstanceConfig": []interface{}{
						map[string]interface{}{
							"enableSecureBoot":          true,
							"enableIntegrityMonitoring": true,
						},
					},
				},
			},
			"networkingMode": "VPC_NATIVE",
			"ipAllocationPolicy": []interface{}{
				map[string]interface{}{
					"clusterIpv4CidrBlock":  "",
					"servicesIpv4CidrBlock": "",
				},
			},
			"addonsConfig": []interface{}{
				map[string]interface{}{
					"httpLoadBalancing": []interface{}{
						map[string]interface{}{
							"disabled": false,
						},
					},
					"horizontalPodAutoscaling": []interface{}{
						map[string]interface{}{
							"disabled": false,
						},
					},
				},
			},
			"releaseChannel": []interface{}{
				map[string]interface{}{
					"channel": "REGULAR",
				},
			},
			"workloadIdentityConfig": []interface{}{
				map[string]interface{}{
					"workloadPool": "PROJECT_ID.svc.id.goog",
				},
			},
			"binaryAuthorization": []interface{}{
				map[string]interface{}{
					"evaluationMode": "PROJECT_SINGLETON_POLICY_ENFORCE",
				},
			},
		},
		"providerConfigRef": map[string]interface{}{
			"name": "default",
		},
		"writeConnectionSecretToRef": map[string]interface{}{
			"name":      fmt.Sprintf("%s-kubeconfig", name),
			"namespace": "crossplane-system",
		},
	}

	if err := unstructured.SetNestedMap(cluster.Object, clusterSpec, "spec"); err != nil {
		return nil, fmt.Errorf("failed to set cluster spec: %w", err)
	}

	return cluster, nil
}
