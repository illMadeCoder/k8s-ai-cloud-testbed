/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentSpec defines the desired state of Component
type ComponentSpec struct {
	// Description of the component
	// +optional
	Description string `json:"description,omitempty"`

	// Type of component: app, workflow, config
	// +required
	// +kubebuilder:validation:Enum=app;workflow;config
	Type string `json:"type"`

	// Sources define where to find the component
	// +optional
	Sources []ComponentSource `json:"sources,omitempty"`

	// Parameters that can be passed to this component
	// +optional
	Parameters []ComponentParameter `json:"parameters,omitempty"`

	// Observability configuration for this component
	// +optional
	Observability *ComponentObservability `json:"observability,omitempty"`
}

// ComponentSource defines a source location for a component
type ComponentSource struct {
	// RepoURL is the Git repository URL
	// +required
	RepoURL string `json:"repoURL"`

	// TargetRevision is the Git ref (branch, tag, commit SHA)
	// +optional
	// +kubebuilder:default="HEAD"
	TargetRevision string `json:"targetRevision,omitempty"`

	// Path within the repository
	// +required
	Path string `json:"path"`

	// Helm chart configuration
	// +optional
	Helm *HelmConfig `json:"helm,omitempty"`
}

// HelmConfig defines Helm-specific configuration
type HelmConfig struct {
	// ReleaseName override
	// +optional
	ReleaseName string `json:"releaseName,omitempty"`

	// Values file path within the repo
	// +optional
	ValuesFiles []string `json:"valuesFiles,omitempty"`

	// Parameters that can be templated
	// +optional
	Parameters []HelmParameter `json:"parameters,omitempty"`
}

// HelmParameter defines a Helm parameter
type HelmParameter struct {
	// Name of the parameter
	// +required
	Name string `json:"name"`

	// Value or template
	// +required
	Value string `json:"value"`

	// ForceString forces the value to be a string
	// +optional
	ForceString bool `json:"forceString,omitempty"`
}

// ComponentParameter defines a parameter that can be passed to the component
type ComponentParameter struct {
	// Name of the parameter
	// +required
	Name string `json:"name"`

	// Type of the parameter
	// +optional
	// +kubebuilder:validation:Enum=string;int;bool
	// +kubebuilder:default="string"
	Type string `json:"type,omitempty"`

	// Default value
	// +optional
	Default string `json:"default,omitempty"`

	// Description of the parameter
	// +optional
	Description string `json:"description,omitempty"`

	// Required parameter
	// +optional
	Required bool `json:"required,omitempty"`
}

// ComponentObservability defines observability configuration
type ComponentObservability struct {
	// ServiceMonitor generation
	// +optional
	ServiceMonitor bool `json:"serviceMonitor,omitempty"`

	// PodMonitor generation
	// +optional
	PodMonitor bool `json:"podMonitor,omitempty"`

	// PodLabels to identify metrics endpoints
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// MetricsPath override
	// +optional
	// +kubebuilder:default="/metrics"
	MetricsPath string `json:"metricsPath,omitempty"`

	// MetricsPort override
	// +optional
	MetricsPort int `json:"metricsPort,omitempty"`
}

// ComponentStatus defines the observed state of Component
type ComponentStatus struct {
	// Conditions
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Sources",type=integer,JSONPath=`.spec.sources[*]`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Component is the Schema for the components API
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentSpec   `json:"spec,omitempty"`
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComponentList contains a list of Component
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Component{}, &ComponentList{})
}
