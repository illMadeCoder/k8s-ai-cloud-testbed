package argocd

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// Client provides ArgoCD integration
type Client struct {
	client.Client
	AppManager *ApplicationManager
}

// NewClient creates a new ArgoCD client
func NewClient(c client.Client) *Client {
	return &Client{
		Client:     c,
		AppManager: NewApplicationManager(c),
	}
}

// RegisterClusterAndCreateApps registers a cluster and creates apps for all components
func (c *Client) RegisterClusterAndCreateApps(ctx context.Context, experimentName string, target experimentsv1alpha1.Target, clusterName string, kubeconfig []byte, server string) error {
	// Register cluster with ArgoCD
	if err := RegisterCluster(ctx, c.Client, clusterName, kubeconfig, server); err != nil {
		return err
	}

	// Create ArgoCD Application for this target
	if err := c.AppManager.CreateApplication(ctx, experimentName, target, server); err != nil {
		return err
	}

	return nil
}

// DeleteClusterAndApps unregisters a cluster and deletes all its applications
func (c *Client) DeleteClusterAndApps(ctx context.Context, experimentName string, targets []experimentsv1alpha1.Target, clusterNames []string) error {
	// Delete all applications
	for _, target := range targets {
		if err := c.AppManager.DeleteApplication(ctx, experimentName, target.Name); err != nil {
			// Log but continue with other deletions
			continue
		}
	}

	// Unregister clusters
	for _, clusterName := range clusterNames {
		if err := UnregisterCluster(ctx, c.Client, clusterName); err != nil {
			// Log but continue
			continue
		}
	}

	return nil
}
