package argocd

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	"github.com/illmadecoder/experiment-operator/internal/components"
)

var (
	// ArgoCD Application GVK
	applicationGVK = schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	}
)

// ApplicationManager manages ArgoCD Applications
type ApplicationManager struct {
	client.Client
	Resolver *components.Resolver
}

// NewApplicationManager creates a new ApplicationManager
func NewApplicationManager(c client.Client) *ApplicationManager {
	return &ApplicationManager{
		Client:   c,
		Resolver: components.NewResolver(c),
	}
}

// CreateApplication creates an ArgoCD Application for a target
func (m *ApplicationManager) CreateApplication(ctx context.Context, experimentName string, target experimentsv1alpha1.Target, clusterServer string) error {
	log := log.FromContext(ctx)

	appName := fmt.Sprintf("%s-%s", experimentName, target.Name)

	// Build ArgoCD Application
	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(applicationGVK)
	app.SetName(appName)
	app.SetNamespace("argocd")

	// Set labels
	app.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by": "experiment-operator",
		"experiments.illm.io/experiment": experimentName,
		"experiments.illm.io/target":     target.Name,
	})

	// Resolve components using the resolver
	resolvedComponents, err := m.Resolver.ResolveComponents(ctx, target.Components)
	if err != nil {
		return fmt.Errorf("failed to resolve components: %w", err)
	}

	// Build sources from resolved components
	sources := []interface{}{}
	for _, resolved := range resolvedComponents {
		for _, source := range resolved.Sources {
			argoSource := map[string]interface{}{
				"repoURL":        source.RepoURL,
				"targetRevision": source.TargetRevision,
				"path":           source.Path,
			}

			// Add Helm configuration if present
			if source.Helm != nil {
				helmConfig := map[string]interface{}{}

				if source.Helm.ReleaseName != "" {
					helmConfig["releaseName"] = source.Helm.ReleaseName
				}

				if len(source.Helm.ValuesFiles) > 0 {
					helmConfig["valueFiles"] = source.Helm.ValuesFiles
				}

				if len(source.Helm.Parameters) > 0 {
					helmParams := []interface{}{}
					for key, value := range source.Helm.Parameters {
						helmParams = append(helmParams, map[string]interface{}{
							"name":  key,
							"value": value,
						})
					}
					helmConfig["parameters"] = helmParams
				}

				if len(helmConfig) > 0 {
					argoSource["helm"] = helmConfig
				}
			}

			sources = append(sources, argoSource)
		}
	}

	// If no sources, use a placeholder
	if len(sources) == 0 {
		sources = append(sources, map[string]interface{}{
			"repoURL":        "https://github.com/argoproj/argocd-example-apps.git",
			"targetRevision": "HEAD",
			"path":           "guestbook",
		})
	}

	// Build Application spec
	spec := map[string]interface{}{
		"project": "default",
		"sources":  sources,
		"destination": map[string]interface{}{
			"server":    clusterServer,
			"namespace": "default",
		},
		"syncPolicy": map[string]interface{}{
			"automated": map[string]interface{}{
				"prune":    true,
				"selfHeal": true,
			},
			"syncOptions": []interface{}{
				"CreateNamespace=true",
			},
		},
	}

	if err := unstructured.SetNestedMap(app.Object, spec, "spec"); err != nil {
		return fmt.Errorf("failed to set application spec: %w", err)
	}

	// Check if application already exists
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(applicationGVK)
	err = m.Get(ctx, client.ObjectKey{Name: appName, Namespace: "argocd"}, existing)
	if err == nil {
		// Application exists, update it
		existing.Object["spec"] = app.Object["spec"]
		if err := m.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update application: %w", err)
		}
		log.Info("Updated ArgoCD Application", "name", appName)
		return nil
	}

	// Create new application
	if err := m.Create(ctx, app); err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	log.Info("Created ArgoCD Application", "name", appName, "target", target.Name)
	return nil
}

// DeleteApplication deletes an ArgoCD Application
func (m *ApplicationManager) DeleteApplication(ctx context.Context, experimentName string, targetName string) error {
	log := log.FromContext(ctx)

	appName := fmt.Sprintf("%s-%s", experimentName, targetName)

	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(applicationGVK)
	app.SetName(appName)
	app.SetNamespace("argocd")

	if err := m.Delete(ctx, app); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	log.Info("Deleted ArgoCD Application", "name", appName)
	return nil
}

// IsApplicationHealthy checks if an Application is healthy
func (m *ApplicationManager) IsApplicationHealthy(ctx context.Context, experimentName string, targetName string) (bool, error) {
	appName := fmt.Sprintf("%s-%s", experimentName, targetName)

	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(applicationGVK)

	if err := m.Get(ctx, client.ObjectKey{Name: appName, Namespace: "argocd"}, app); err != nil {
		return false, fmt.Errorf("failed to get application: %w", err)
	}

	// Check health status
	healthStatus, found, err := unstructured.NestedString(app.Object, "status", "health", "status")
	if err != nil || !found {
		return false, nil // Not ready yet
	}

	// Check sync status
	syncStatus, found, err := unstructured.NestedString(app.Object, "status", "sync", "status")
	if err != nil || !found {
		return false, nil
	}

	// Application is healthy if health is "Healthy" and sync is "Synced"
	healthy := healthStatus == "Healthy" && syncStatus == "Synced"
	return healthy, nil
}

// GetApplicationComponents returns the list of deployed components
func (m *ApplicationManager) GetApplicationComponents(ctx context.Context, experimentName string, targetName string) ([]string, error) {
	appName := fmt.Sprintf("%s-%s", experimentName, targetName)

	app := &unstructured.Unstructured{}
	app.SetGroupVersionKind(applicationGVK)

	if err := m.Get(ctx, client.ObjectKey{Name: appName, Namespace: "argocd"}, app); err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// Get resources from status
	resources, found, err := unstructured.NestedSlice(app.Object, "status", "resources")
	if err != nil || !found {
		return []string{}, nil
	}

	components := []string{}
	for _, res := range resources {
		resource, ok := res.(map[string]interface{})
		if !ok {
			continue
		}

		kind, _, _ := unstructured.NestedString(resource, "kind")
		name, _, _ := unstructured.NestedString(resource, "name")
		if kind != "" && name != "" {
			components = append(components, fmt.Sprintf("%s/%s", kind, name))
		}
	}

	return components, nil
}
