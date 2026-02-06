package crossplane

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

var gkeClusterGVK = schema.GroupVersionKind{
	Group:   "illm.io",
	Version: "v1alpha1",
	Kind:    "GKECluster",
}

const (
	claimNamespace       = "experiments"
	connectionSecretNS   = "crossplane-system"
	connectionSecretFmt  = "%s-cluster-conn"
	managedByLabel       = "app.kubernetes.io/managed-by"
	managedByValue       = "experiment-operator"
	experimentClusterLbl = "experiments.illm.io/cluster"
)

// ClusterManager manages Crossplane cluster resources
type ClusterManager struct {
	client.Client
}

// NewClusterManager creates a new ClusterManager
func NewClusterManager(c client.Client) *ClusterManager {
	return &ClusterManager{Client: c}
}

// CreateCluster creates a Crossplane GKECluster claim based on the spec
func (m *ClusterManager) CreateCluster(ctx context.Context, experimentName string, target experimentsv1alpha1.Target) (string, error) {
	log := log.FromContext(ctx)

	clusterName := fmt.Sprintf("%s-%s", experimentName, target.Name)

	if target.Cluster.Type == "hub" {
		log.Info("Using existing hub cluster", "target", target.Name)
		return "hub", nil
	}

	claim := buildGKEClusterClaim(clusterName, target.Cluster)

	if err := m.Create(ctx, claim); err != nil {
		return "", fmt.Errorf("failed to create GKECluster claim: %w", err)
	}

	log.Info("Created GKECluster claim", "name", clusterName, "namespace", claimNamespace)
	return clusterName, nil
}

// buildGKEClusterClaim constructs a GKECluster claim from ClusterSpec
func buildGKEClusterClaim(name string, spec experimentsv1alpha1.ClusterSpec) *unstructured.Unstructured {
	claim := &unstructured.Unstructured{}
	claim.SetGroupVersionKind(gkeClusterGVK)
	claim.SetName(name)
	claim.SetNamespace(claimNamespace)
	claim.SetLabels(map[string]string{
		managedByLabel:       managedByValue,
		experimentClusterLbl: name,
	})

	// Apply defaults
	zone := spec.Zone
	if zone == "" {
		zone = "us-central1-a"
	}
	nodeCount := spec.NodeCount
	if nodeCount == 0 {
		nodeCount = 1
	}
	machineType := spec.MachineType
	if machineType == "" {
		machineType = "e2-medium"
	}
	diskSizeGb := spec.DiskSizeGb
	if diskSizeGb == 0 {
		diskSizeGb = 50
	}

	claimSpec := map[string]interface{}{
		"zone":        zone,
		"nodeCount":   int64(nodeCount),
		"machineType": machineType,
		"diskSizeGb":  int64(diskSizeGb),
		"preemptible": spec.Preemptible,
	}

	// Safe to ignore error — values are all primitives
	_ = unstructured.SetNestedMap(claim.Object, claimSpec, "spec")

	return claim
}

// IsClusterReady checks if a cluster is ready
func (m *ClusterManager) IsClusterReady(ctx context.Context, clusterName string, clusterType string) (bool, error) {
	if clusterType == "hub" {
		return true, nil
	}

	claim := &unstructured.Unstructured{}
	claim.SetGroupVersionKind(gkeClusterGVK)

	if err := m.Get(ctx, client.ObjectKey{
		Name:      clusterName,
		Namespace: claimNamespace,
	}, claim); err != nil {
		return false, fmt.Errorf("failed to get GKECluster claim: %w", err)
	}

	conditions, found, err := unstructured.NestedSlice(claim.Object, "status", "conditions")
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

// GetClusterKubeconfig retrieves the kubeconfig for a cluster.
// The composition writes the connection secret to crossplane-system/{xr-name}-cluster-conn.
// We look up the XR name from the claim's spec.resourceRef.name.
func (m *ClusterManager) GetClusterKubeconfig(ctx context.Context, clusterName string, clusterType string) ([]byte, error) {
	if clusterType == "hub" {
		return nil, fmt.Errorf("hub cluster does not have a separate kubeconfig")
	}

	// Read the claim to get the XR name
	claim := &unstructured.Unstructured{}
	claim.SetGroupVersionKind(gkeClusterGVK)
	if err := m.Get(ctx, client.ObjectKey{
		Name:      clusterName,
		Namespace: claimNamespace,
	}, claim); err != nil {
		return nil, fmt.Errorf("failed to get GKECluster claim: %w", err)
	}

	xrName, found, err := unstructured.NestedString(claim.Object, "spec", "resourceRef", "name")
	if err != nil || !found || xrName == "" {
		return nil, fmt.Errorf("claim has no resourceRef.name yet (XR not bound)")
	}

	// Composition writes to crossplane-system/{xr-name}-cluster-conn
	secretName := fmt.Sprintf(connectionSecretFmt, xrName)

	secret := &unstructured.Unstructured{}
	secret.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})

	if err := m.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: connectionSecretNS,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get connection secret %s: %w", secretName, err)
	}

	data, found, err := unstructured.NestedMap(secret.Object, "data")
	if err != nil || !found {
		return nil, fmt.Errorf("connection secret has no data")
	}

	kubeconfigB64, ok := data["kubeconfig"].(string)
	if !ok {
		return nil, fmt.Errorf("kubeconfig not found in connection secret")
	}

	// Secret data from unstructured API is base64-encoded
	kubeconfig, err := base64.StdEncoding.DecodeString(kubeconfigB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode kubeconfig: %w", err)
	}

	return kubeconfig, nil
}

// DeleteCluster deletes a cluster resource
func (m *ClusterManager) DeleteCluster(ctx context.Context, clusterName string, clusterType string) error {
	log := log.FromContext(ctx)

	if clusterType == "hub" {
		log.Info("Skipping deletion of hub cluster")
		return nil
	}

	claim := &unstructured.Unstructured{}
	claim.SetGroupVersionKind(gkeClusterGVK)
	claim.SetName(clusterName)
	claim.SetNamespace(claimNamespace)

	if err := m.Delete(ctx, claim); err != nil {
		return fmt.Errorf("failed to delete GKECluster claim: %w", err)
	}

	log.Info("Deleted GKECluster claim", "name", clusterName)
	return nil
}

// GetClusterEndpoint returns the external endpoint for a cluster
func (m *ClusterManager) GetClusterEndpoint(ctx context.Context, clusterName string, clusterType string) (string, error) {
	if clusterType == "hub" {
		return "", nil
	}

	claim := &unstructured.Unstructured{}
	claim.SetGroupVersionKind(gkeClusterGVK)

	if err := m.Get(ctx, client.ObjectKey{
		Name:      clusterName,
		Namespace: claimNamespace,
	}, claim); err != nil {
		return "", fmt.Errorf("failed to get GKECluster claim: %w", err)
	}

	endpoint, found, err := unstructured.NestedString(claim.Object, "status", "endpoint")
	if err != nil || !found {
		return "", nil // Not ready yet
	}
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
