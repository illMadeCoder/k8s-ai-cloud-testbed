package argocd

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// argoClusterConfig is the JSON format ArgoCD expects in cluster secrets.
type argoClusterConfig struct {
	TLSClientConfig *argoTLSClientConfig `json:"tlsClientConfig,omitempty"`
	BearerToken     string               `json:"bearerToken,omitempty"`
}

type argoTLSClientConfig struct {
	CAData   []byte `json:"caData,omitempty"`
	CertData []byte `json:"certData,omitempty"`
	KeyData  []byte `json:"keyData,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
}

// RegisterCluster registers a cluster with ArgoCD by creating a Secret
func RegisterCluster(ctx context.Context, c client.Client, clusterName string, kubeconfig []byte, server string) error {
	log := log.FromContext(ctx)

	config, err := buildArgoClusterConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build ArgoCD cluster config: %w", err)
	}

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
			"config": config,
		},
	}

	// Check if secret already exists
	existing := &corev1.Secret{}
	err = c.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: namespace}, existing)
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

// buildArgoClusterConfig parses a kubeconfig and returns the JSON config ArgoCD expects.
func buildArgoClusterConfig(kubeconfig []byte) (string, error) {
	cfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	ctx := cfg.Contexts[cfg.CurrentContext]
	if ctx == nil {
		return "", fmt.Errorf("current-context %q not found", cfg.CurrentContext)
	}

	cluster := cfg.Clusters[ctx.Cluster]
	user := cfg.AuthInfos[ctx.AuthInfo]

	argoConfig := argoClusterConfig{}

	if cluster != nil {
		argoConfig.TLSClientConfig = &argoTLSClientConfig{
			CAData:   cluster.CertificateAuthorityData,
			Insecure: cluster.InsecureSkipTLSVerify,
		}
	}

	if user != nil {
		if len(user.ClientCertificateData) > 0 || len(user.ClientKeyData) > 0 {
			if argoConfig.TLSClientConfig == nil {
				argoConfig.TLSClientConfig = &argoTLSClientConfig{}
			}
			argoConfig.TLSClientConfig.CertData = user.ClientCertificateData
			argoConfig.TLSClientConfig.KeyData = user.ClientKeyData
		}
		if user.Token != "" {
			argoConfig.BearerToken = user.Token
		}
	}

	data, err := json.Marshal(argoConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ArgoCD config: %w", err)
	}

	return string(data), nil
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
		if errors.IsNotFound(err) {
			log.Info("Cluster secret already deleted", "cluster", clusterName)
			return nil
		}
		return fmt.Errorf("failed to delete cluster secret: %w", err)
	}

	log.Info("Unregistered cluster from ArgoCD", "cluster", clusterName)
	return nil
}
