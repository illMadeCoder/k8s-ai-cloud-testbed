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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	// Description of the experiment
	// +optional
	Description string `json:"description,omitempty"`

	// Targets to deploy (app, loadgen, etc.)
	// +required
	Targets []Target `json:"targets"`

	// Workflow for validation and lifecycle
	// +required
	Workflow WorkflowSpec `json:"workflow"`

	// Tutorial configuration for interactive learning
	// +optional
	Tutorial *TutorialSpec `json:"tutorial,omitempty"`

	// TTL in days - experiment will be auto-deleted after this many days
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=365
	TTLDays int `json:"ttlDays,omitempty"`
}

// TutorialSpec defines tutorial configuration for interactive experiments
type TutorialSpec struct {
	// Path to tutorial file relative to experiment directory, default "tutorial.yaml"
	// +optional
	Path string `json:"path,omitempty"`

	// Expose kubeconfig for user kubectl access to target clusters
	// +optional
	ExposeKubeconfig bool `json:"exposeKubeconfig,omitempty"`

	// Services to discover on target clusters
	// +optional
	Services []TutorialServiceRef `json:"services,omitempty"`
}

// TutorialServiceRef references a service to discover on a target cluster
type TutorialServiceRef struct {
	// Name is a friendly name for the service (e.g., "grafana")
	// +required
	Name string `json:"name"`

	// Target is the target name from spec.targets
	// +required
	Target string `json:"target"`

	// Service is the Kubernetes service name
	// +required
	Service string `json:"service"`

	// Namespace is the Kubernetes namespace of the service
	// +required
	Namespace string `json:"namespace"`

	// Port is the service port (optional, uses first port if unset)
	// +optional
	Port int `json:"port,omitempty"`
}

// Target defines a deployment target (cluster + components)
type Target struct {
	// Name of the target (app, loadgen, etc.)
	// +required
	Name string `json:"name"`

	// Cluster configuration
	// +required
	Cluster ClusterSpec `json:"cluster"`

	// Components to deploy
	// +optional
	Components []ComponentRef `json:"components,omitempty"`

	// Observability configuration
	// +optional
	Observability *ObservabilitySpec `json:"observability,omitempty"`

	// Dependencies (other target names)
	// +optional
	Depends []string `json:"depends,omitempty"`
}

// ClusterSpec defines cluster configuration
type ClusterSpec struct {
	// Type: gke, hub (existing hub cluster)
	// +required
	// +kubebuilder:validation:Enum=gke;hub
	Type string `json:"type"`

	// Zone (GCP)
	// +optional
	Zone string `json:"zone,omitempty"`

	// Node configuration
	// +optional
	NodeCount int `json:"nodeCount,omitempty"`

	// +optional
	MachineType string `json:"machineType,omitempty"`

	// +optional
	DiskSizeGb int `json:"diskSizeGb,omitempty"`

	// +optional
	Preemptible bool `json:"preemptible,omitempty"`
}

// ComponentRef references a component by name
type ComponentRef struct {
	// App name
	// +optional
	App string `json:"app,omitempty"`

	// Workflow name
	// +optional
	Workflow string `json:"workflow,omitempty"`

	// Config name
	// +optional
	Config string `json:"config,omitempty"`

	// Parameters to pass to component
	// +optional
	Params map[string]string `json:"params,omitempty"`
}

// WorkflowSpec defines the experiment workflow
type WorkflowSpec struct {
	// WorkflowTemplate name
	// +required
	Template string `json:"template"`

	// Completion mode configuration
	// +required
	Completion CompletionSpec `json:"completion"`

	// Parameters
	// +optional
	Params map[string]string `json:"params,omitempty"`
}

// CompletionSpec defines when experiment completes
type CompletionSpec struct {
	// Mode: workflow (auto-complete after workflow), manual (stay Running until user tears down)
	// +required
	// +kubebuilder:validation:Enum=workflow;manual
	Mode string `json:"mode"`
}

// ObservabilitySpec defines observability config
type ObservabilitySpec struct {
	// +required
	Enabled bool `json:"enabled"`

	// Transport: direct, tailscale
	// +required
	// +kubebuilder:validation:Enum=direct;tailscale
	Transport string `json:"transport"`

	// +optional
	Tenant string `json:"tenant,omitempty"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	// Phase of the experiment
	// +optional
	// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Running;Complete;Failed
	Phase ExperimentPhase `json:"phase,omitempty"`

	// Target statuses
	// +optional
	Targets []TargetStatus `json:"targets,omitempty"`

	// Workflow status
	// +optional
	WorkflowStatus *WorkflowStatus `json:"workflowStatus,omitempty"`

	// Tutorial status (populated when spec.tutorial is set)
	// +optional
	TutorialStatus *TutorialStatus `json:"tutorialStatus,omitempty"`

	// CompletedAt is the timestamp when the experiment reached a terminal state
	// +optional
	CompletedAt *metav1.Time `json:"completedAt,omitempty"`

	// ResourcesCleaned indicates whether expensive resources (clusters, apps) have been cleaned up
	// +optional
	ResourcesCleaned bool `json:"resourcesCleaned,omitempty"`

	// Conditions
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TutorialStatus represents the observed state of tutorial resources
type TutorialStatus struct {
	// Discovered services with resolved endpoints
	// +optional
	Services []DiscoveredService `json:"services,omitempty"`

	// KubeconfigSecrets maps target names to kubeconfig secret names in the experiments namespace
	// +optional
	KubeconfigSecrets map[string]string `json:"kubeconfigSecrets,omitempty"`
}

// DiscoveredService represents a service discovered on a target cluster
type DiscoveredService struct {
	// Name is the friendly service name from spec.tutorial.services
	// +required
	Name string `json:"name"`

	// Endpoint is the resolved service endpoint (LoadBalancer IP or ClusterIP)
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Ready indicates whether the service is accessible
	// +optional
	Ready bool `json:"ready,omitempty"`
}

// ExperimentPhase represents the current phase of an experiment
// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Running;Complete;Failed
type ExperimentPhase string

const (
	PhasePending      ExperimentPhase = "Pending"
	PhaseProvisioning ExperimentPhase = "Provisioning"
	PhaseReady        ExperimentPhase = "Ready"
	PhaseRunning      ExperimentPhase = "Running"
	PhaseComplete     ExperimentPhase = "Complete"
	PhaseFailed       ExperimentPhase = "Failed"
)

// TargetStatus represents the status of a deployment target
type TargetStatus struct {
	// +required
	Name string `json:"name"`

	// +optional
	Phase string `json:"phase,omitempty"`

	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// +optional
	Components []string `json:"components,omitempty"`

	// KubeconfigSecret is the name of the secret containing the kubeconfig for this target
	// +optional
	KubeconfigSecret string `json:"kubeconfigSecret,omitempty"`
}

// WorkflowStatus represents the status of the experiment workflow
type WorkflowStatus struct {
	// +required
	Name string `json:"name"`

	// +optional
	Phase string `json:"phase,omitempty"`

	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// +optional
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Targets",type=string,JSONPath=`.spec.targets[*].name`
// +kubebuilder:printcolumn:name="Workflow",type=string,JSONPath=`.status.workflowStatus.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Experiment is the Schema for the experiments API
type Experiment struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Experiment
	// +required
	Spec ExperimentSpec `json:"spec"`

	// status defines the observed state of Experiment
	// +optional
	Status ExperimentStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ExperimentList contains a list of Experiment
type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Experiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
