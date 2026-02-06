package argocd

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RegisterCluster registers a cluster with ArgoCD by creating a Secret
func RegisterCluster(ctx context.Context, c client.Client, clusterName string, kubeconfig []byte, server string) error {
	log := log.FromContext(ctx)

	// ArgoCD expects cluster secrets in the argocd namespace
	namespace := "argocd"

	// Create the cluster secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("cluster-%s", clusterName),
			Namespace: namespace,
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
				"experiments.illm.io/cluster":    clusterName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"name":   clusterName,
			"server": server,
			"config": string(kubeconfig),
		},
	}

	// Check if secret already exists
	existing := &corev1.Secret{}
	err := c.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: namespace}, existing)
	if err == nil {
		// Secret exists, update it
		existing.StringData = secret.StringData
		existing.Labels = secret.Labels
		if err := c.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update cluster secret: %w", err)
		}
		log.Info("Updated ArgoCD cluster secret", "cluster", clusterName)
		return nil
	}

	// Create new secret
	if err := c.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create cluster secret: %w", err)
	}

	log.Info("Registered cluster with ArgoCD", "cluster", clusterName, "server", server)
	return nil
}

// UnregisterCluster removes a cluster from ArgoCD by deleting its Secret
func UnregisterCluster(ctx context.Context, c client.Client, clusterName string) error {
	log := log.FromContext(ctx)

	namespace := "argocd"
	secretName := fmt.Sprintf("cluster-%s", clusterName)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	if err := c.Delete(ctx, secret); err != nil {
		return fmt.Errorf("failed to delete cluster secret: %w", err)
	}

	log.Info("Unregistered cluster from ArgoCD", "cluster", clusterName)
	return nil
}

