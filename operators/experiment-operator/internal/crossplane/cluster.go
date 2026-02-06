package crossplane

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// ClusterManager manages Crossplane cluster resources
type ClusterManager struct {
	client.Client
}

// NewClusterManager creates a new ClusterManager
func NewClusterManager(c client.Client) *ClusterManager {
	return &ClusterManager{Client: c}
}

// CreateCluster creates a Crossplane cluster resource based on the spec
func (m *ClusterManager) CreateCluster(ctx context.Context, experimentName string, target experimentsv1alpha1.Target) (string, error) {
	log := log.FromContext(ctx)

	clusterName := fmt.Sprintf("%s-%s", experimentName, target.Name)

	var cluster *unstructured.Unstructured
	var err error

	switch target.Cluster.Type {
	case "gke":
		cluster, err = m.createGKECluster(ctx, clusterName, target.Cluster)
	case "vcluster":
		cluster, err = m.createVCluster(ctx, clusterName, target.Cluster)
	case "hub":
		// Hub cluster already exists, just return the name
		log.Info("Using existing hub cluster", "target", target.Name)
		return "hub", nil
	default:
		return "", fmt.Errorf("unsupported cluster type: %s", target.Cluster.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create cluster spec: %w", err)
	}

	// Create the cluster resource
	if err := m.Create(ctx, cluster); err != nil {
		return "", fmt.Errorf("failed to create cluster resource: %w", err)
	}

	log.Info("Created cluster resource", "name", clusterName, "type", target.Cluster.Type)
	return clusterName, nil
}

// IsClusterReady checks if a cluster is ready
func (m *ClusterManager) IsClusterReady(ctx context.Context, clusterName string, clusterType string) (bool, error) {
	if clusterType == "hub" {
		// Hub cluster is always ready
		return true, nil
	}

	var gvk schema.GroupVersionKind
	switch clusterType {
	case "gke":
		gvk = schema.GroupVersionKind{
			Group:   "container.gcp.upbound.io",
			Version: "v1beta1",
			Kind:    "Cluster",
		}
	case "vcluster":
		gvk = schema.GroupVersionKind{
			Group:   "infrastructure.cluster.x-k8s.io",
			Version: "v1alpha1",
			Kind:    "VCluster",
		}
	default:
		return false, fmt.Errorf("unsupported cluster type: %s", clusterType)
	}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(gvk)

	if err := m.Get(ctx, client.ObjectKey{
		Name:      clusterName,
		Namespace: "crossplane-system",
	}, cluster); err != nil {
		return false, fmt.Errorf("failed to get cluster: %w", err)
	}

	// Check status conditions
	conditions, found, err := unstructured.NestedSlice(cluster.Object, "status", "conditions")
	if err != nil || !found {
		return false, nil
	}

	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condition, "type")
		status, _, _ := unstructured.NestedString(condition, "status")

		if condType == "Ready" && status == "True" {
			return true, nil
		}
	}

	return false, nil
}

// GetClusterKubeconfig retrieves the kubeconfig for a cluster
func (m *ClusterManager) GetClusterKubeconfig(ctx context.Context, clusterName string, clusterType string) ([]byte, error) {
	if clusterType == "hub" {
		return nil, fmt.Errorf("hub cluster does not have a separate kubeconfig")
	}

	// Kubeconfig is typically stored in a Secret named {clusterName}-kubeconfig
	secretName := fmt.Sprintf("%s-kubeconfig", clusterName)

	secret := &unstructured.Unstructured{}
	secret.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})

	if err := m.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: "crossplane-system",
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig secret: %w", err)
	}

	data, found, err := unstructured.NestedMap(secret.Object, "data")
	if err != nil || !found {
		return nil, fmt.Errorf("kubeconfig secret has no data")
	}

	kubeconfigBase64, ok := data["kubeconfig"].(string)
	if !ok {
		return nil, fmt.Errorf("kubeconfig not found in secret")
	}

	// Data in secrets is already base64 encoded in the API, but unstructured
	// gives us the decoded value, so we can return it directly
	return []byte(kubeconfigBase64), nil
}

// DeleteCluster deletes a cluster resource
func (m *ClusterManager) DeleteCluster(ctx context.Context, clusterName string, clusterType string) error {
	log := log.FromContext(ctx)

	if clusterType == "hub" {
		// Don't delete the hub cluster
		log.Info("Skipping deletion of hub cluster")
		return nil
	}

	var gvk schema.GroupVersionKind
	switch clusterType {
	case "gke":
		gvk = schema.GroupVersionKind{
			Group:   "container.gcp.upbound.io",
			Version: "v1beta1",
			Kind:    "Cluster",
		}
	case "vcluster":
		gvk = schema.GroupVersionKind{
			Group:   "infrastructure.cluster.x-k8s.io",
			Version: "v1alpha1",
			Kind:    "VCluster",
		}
	default:
		return fmt.Errorf("unsupported cluster type: %s", clusterType)
	}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(gvk)
	cluster.SetName(clusterName)
	cluster.SetNamespace("crossplane-system")

	if err := m.Delete(ctx, cluster); err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	log.Info("Deleted cluster resource", "name", clusterName, "type", clusterType)
	return nil
}

// GetClusterEndpoint returns the external endpoint for a cluster
func (m *ClusterManager) GetClusterEndpoint(ctx context.Context, clusterName string, clusterType string) (string, error) {
	if clusterType == "hub" {
		return "", nil // Hub cluster endpoint is not exposed
	}

	var gvk schema.GroupVersionKind
	switch clusterType {
	case "gke":
		gvk = schema.GroupVersionKind{
			Group:   "container.gcp.upbound.io",
			Version: "v1beta1",
			Kind:    "Cluster",
		}
	case "vcluster":
		gvk = schema.GroupVersionKind{
			Group:   "infrastructure.cluster.x-k8s.io",
			Version: "v1alpha1",
			Kind:    "VCluster",
		}
	default:
		return "", fmt.Errorf("unsupported cluster type: %s", clusterType)
	}

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(gvk)

	if err := m.Get(ctx, client.ObjectKey{
		Name:      clusterName,
		Namespace: "crossplane-system",
	}, cluster); err != nil {
		return "", fmt.Errorf("failed to get cluster: %w", err)
	}

	// For GKE, endpoint is in status.endpoint
	if clusterType == "gke" {
		endpoint, found, err := unstructured.NestedString(cluster.Object, "status", "endpoint")
		if err != nil || !found {
			return "", nil // Not ready yet
		}
		return endpoint, nil
	}

	// For other cluster types, might be in different locations
	endpoint, _, _ := unstructured.NestedString(cluster.Object, "status", "controlPlaneEndpoint")
	return endpoint, nil
}

// CalculateTTL calculates when a cluster should be deleted based on TTL
func CalculateTTL(creationTime time.Time, ttlDays int) time.Time {
	if ttlDays <= 0 {
		// Default to 1 day
		ttlDays = 1
	}
	return creationTime.Add(time.Duration(ttlDays) * 24 * time.Hour)
}

// ShouldDeleteCluster checks if a cluster has exceeded its TTL
func ShouldDeleteCluster(creationTime time.Time, ttlDays int) bool {
	ttlExpiry := CalculateTTL(creationTime, ttlDays)
	return time.Now().After(ttlExpiry)
}
