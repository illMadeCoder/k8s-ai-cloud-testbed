package components

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// Resolver resolves component references to actual sources
type Resolver struct {
	client.Client
}

// NewResolver creates a new component resolver
func NewResolver(c client.Client) *Resolver {
	return &Resolver{Client: c}
}

// ResolvedComponent represents a resolved component with its sources
type ResolvedComponent struct {
	Name    string
	Type    string
	Sources []ResolvedSource
}

// ResolvedSource represents a resolved source
type ResolvedSource struct {
	RepoURL        string
	TargetRevision string
	Path           string
	Chart          string
	Helm           *HelmConfig
}

// HelmConfig represents Helm configuration
type HelmConfig struct {
	ReleaseName string
	ValuesFiles []string
	Parameters  map[string]string
}

// ResolveComponentRef resolves a ComponentRef to actual sources
func (r *Resolver) ResolveComponentRef(ctx context.Context, ref experimentsv1alpha1.ComponentRef) (*ResolvedComponent, error) {
	log := log.FromContext(ctx)

	// Determine component name based on ref type
	var componentName string
	var componentType string

	if ref.App != "" {
		componentName = ref.App
		componentType = "app"
	} else if ref.Workflow != "" {
		componentName = ref.Workflow
		componentType = "workflow"
	} else if ref.Config != "" {
		componentName = ref.Config
		componentType = "config"
	} else {
		return nil, fmt.Errorf("component reference has no app, workflow, or config")
	}

	// Try to find Component CR
	component := &experimentsv1alpha1.Component{}
	err := r.Get(ctx, client.ObjectKey{Name: componentName}, component)
	if err != nil {
		// Component CR not found, use fallback
		log.Info("Component CR not found, using fallback", "name", componentName, "type", componentType)
		return r.fallbackComponent(componentName, componentType, ref.Params)
	}

	// Resolve from Component CR
	log.Info("Resolved component from CR", "name", componentName)
	return r.resolveFromCR(component, ref.Params)
}

// resolveFromCR resolves a component from its CR
func (r *Resolver) resolveFromCR(component *experimentsv1alpha1.Component, params map[string]string) (*ResolvedComponent, error) {
	resolved := &ResolvedComponent{
		Name:    component.Name,
		Type:    component.Spec.Type,
		Sources: []ResolvedSource{},
	}

	for _, source := range component.Spec.Sources {
		targetRevision := source.TargetRevision
		if targetRevision == "" {
			targetRevision = "HEAD"
		}

		resolvedSource := ResolvedSource{
			RepoURL:        source.RepoURL,
			TargetRevision: targetRevision,
			Path:           source.Path,
			Chart:          source.Chart,
		}

		// Handle Helm configuration
		if source.Helm != nil {
			helmConfig := &HelmConfig{
				ReleaseName: source.Helm.ReleaseName,
				ValuesFiles: source.Helm.ValuesFiles,
				Parameters:  make(map[string]string),
			}

			// Merge default Helm parameters from Component
			for _, param := range source.Helm.Parameters {
				helmConfig.Parameters[param.Name] = param.Value
			}

			// Override with provided params
			for key, value := range params {
				helmConfig.Parameters[key] = value
			}

			resolvedSource.Helm = helmConfig
		}

		resolved.Sources = append(resolved.Sources, resolvedSource)
	}

	return resolved, nil
}

// componentCategory maps component names to their category directory in the repo.
// Components are organized as components/<category>/<name>/ in the git repo.
var componentCategory = map[string]string{
	// apps
	"cardinality-generator": "apps",
	"custom-db":             "apps",
	"demo-app":              "apps",
	"hello-app":             "apps",
	"log-generator":         "apps",
	"metrics-app":           "apps",
	"nginx":                 "apps",
	"otel-demo":             "apps",
	"station-monitor":       "apps",
	// chaos
	"chaos-mesh": "chaos",
	// core
	"argocd":             "core",
	"cert-manager":       "core",
	"envoy-gateway":      "core",
	"gateway-api":        "core",
	"nginx-ingress":      "core",
	"tailscale-operator": "core",
	"traefik":            "core",
	// messaging
	"kafka":            "messaging",
	"rabbitmq":         "messaging",
	"strimzi-operator": "messaging",
	// observability
	"eck-operator":          "observability",
	"elasticsearch":         "observability",
	"elk-stack":             "observability",
	"fluent-bit":            "observability",
	"jaeger":                "observability",
	"kibana":                "observability",
	"kube-prometheus-stack": "observability",
	"loki":                  "observability",
	"metrics-agent":         "observability",
	"metrics-egress":        "observability",
	"mimir":                 "observability",
	"otel-collector":        "observability",
	"prometheus-stack":      "observability",
	"promtail":              "observability",
	"pyrra":                 "observability",
	"seaweedfs":             "observability",
	"tempo":                 "observability",
	"victoria-metrics":      "observability",
	// storage
	"minio": "storage",
	// testing
	"k6":                    "testing",
	"k6-gateway-loadtest":   "testing",
	// workflows
	"argo-workflows": "workflows",
}

// fallbackComponent creates a fallback component when CR is not found
func (r *Resolver) fallbackComponent(name string, componentType string, params map[string]string) (*ResolvedComponent, error) {
	// Default repository
	defaultRepo := "https://github.com/illMadeCoder/k8s-ai-cloud-testbed.git"

	// Construct path based on component category lookup, then type
	var path string
	if category, ok := componentCategory[name]; ok {
		path = fmt.Sprintf("components/%s/%s", category, name)
	} else {
		// Unknown component — guess based on type
		switch componentType {
		case "app":
			path = fmt.Sprintf("components/apps/%s", name)
		case "workflow":
			path = fmt.Sprintf("components/workflows/%s", name)
		case "config":
			path = fmt.Sprintf("components/configs/%s", name)
		default:
			return nil, fmt.Errorf("unknown component type: %s", componentType)
		}
	}

	resolved := &ResolvedComponent{
		Name: name,
		Type: componentType,
		Sources: []ResolvedSource{
			{
				RepoURL:        defaultRepo,
				TargetRevision: "HEAD",
				Path:           path,
			},
		},
	}

	// If params provided, add basic Helm config
	if len(params) > 0 {
		resolved.Sources[0].Helm = &HelmConfig{
			Parameters: params,
		}
	}

	return resolved, nil
}

// ResolveComponents resolves all component refs for a target
func (r *Resolver) ResolveComponents(ctx context.Context, components []experimentsv1alpha1.ComponentRef) ([]*ResolvedComponent, error) {
	resolved := []*ResolvedComponent{}

	for _, comp := range components {
		resolvedComp, err := r.ResolveComponentRef(ctx, comp)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve component: %w", err)
		}
		resolved = append(resolved, resolvedComp)
	}

	return resolved, nil
}
