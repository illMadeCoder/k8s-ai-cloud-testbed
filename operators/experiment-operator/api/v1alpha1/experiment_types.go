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

	// Metrics defines PromQL queries to execute at experiment completion.
	// Results are stored in summary.json for the benchmark site.
	// If omitted, default CPU and memory queries are collected.
	// +optional
	Metrics []MetricsQuery `json:"metrics,omitempty"`

	// Tags for categorization on the benchmark site (e.g., "observability", "networking").
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Publish controls whether results are published to the benchmark site and
	// whether AI analysis is generated. When false (default), results are only
	// stored in S3. Set to true for experiments intended for public display.
	// +optional
	Publish bool `json:"publish,omitempty"`

	// Hypothesis describes the claim being tested and optional success criteria.
	// This is the central organizing principle of the experiment — the AI analyzer
	// uses it to produce focused, goal-aware analysis.
	// +optional
	Hypothesis *HypothesisSpec `json:"hypothesis,omitempty"`

	// AnalyzerConfig configures which AI analysis sections to generate on
	// experiment completion. When publish is true and analyzerConfig is nil,
	// default sections are used. Set analyzerConfig with an empty sections
	// array to explicitly skip analysis on a published experiment.
	//
	// Available sections (grouped by analyzer pass):
	//
	//   Pass 2 — Core analysis:
	//     abstract            Executive summary with hypothesis verdict
	//     targetAnalysis      Per-target infrastructure analysis
	//     performanceAnalysis Key performance findings with data
	//     metricInsights      Per-metric chart annotations
	//
	//   Pass 3 — FinOps + SecOps:
	//     finopsAnalysis      Cost analysis and production projections
	//     secopsAnalysis      Security posture and supply chain assessment
	//
	//   Pass 4 — Deep dive:
	//     body                Research-paper methodology/results/discussion
	//     capabilitiesMatrix  Feature comparison table (comparisons only)
	//     feedback            Recommendations and experiment design improvements
	//
	//   Pass 5 — Diagram:
	//     architectureDiagram ASCII architecture topology diagram
	//
	// +optional
	AnalyzerConfig *AnalyzerConfig `json:"analyzerConfig,omitempty"`
}

// AnalyzerConfig configures AI analysis generation for the experiment.
type AnalyzerConfig struct {
	// Sections is the list of analysis sections to generate. Each value
	// must be one of the recognized section names. The analyzer will only
	// run passes that contain at least one requested section.
	// If nil or empty, default sections are used when publish is true.
	//
	// +optional
	// +kubebuilder:validation:items:Enum=abstract;targetAnalysis;performanceAnalysis;metricInsights;finopsAnalysis;secopsAnalysis;body;capabilitiesMatrix;feedback;architectureDiagram
	Sections []string `json:"sections,omitempty"`
}

// Analysis section constants. Grouped by analyzer pass for documentation.
const (
	// Pass 2: Core analysis
	AnalysisSectionAbstract            = "abstract"
	AnalysisSectionTargetAnalysis      = "targetAnalysis"
	AnalysisSectionPerformanceAnalysis = "performanceAnalysis"
	AnalysisSectionMetricInsights      = "metricInsights"

	// Pass 3: FinOps + SecOps
	AnalysisSectionFinopsAnalysis = "finopsAnalysis"
	AnalysisSectionSecopsAnalysis = "secopsAnalysis"

	// Pass 4: Deep dive
	AnalysisSectionBody               = "body"
	AnalysisSectionCapabilitiesMatrix = "capabilitiesMatrix"
	AnalysisSectionFeedback           = "feedback"

	// Pass 5: Diagram
	AnalysisSectionArchitectureDiagram = "architectureDiagram"
)

// HypothesisSpec describes the claim being tested and optional machine-evaluable
// success criteria. This is the central organizing principle of the experiment.
type HypothesisSpec struct {
	// Claim states the expected outcome being tested.
	// Example: "Loki will use fewer resources than Elasticsearch but offer
	// weaker full-text search capabilities."
	// +required
	Claim string `json:"claim"`

	// Questions are specific things the experiment should answer.
	// Example: "What is the CPU overhead difference at 1000 logs/sec?"
	// +optional
	Questions []string `json:"questions,omitempty"`

	// Focus lists key areas for deep analysis.
	// Example: ["resource efficiency", "query capability", "operational complexity"]
	// +optional
	Focus []string `json:"focus,omitempty"`

	// SuccessCriteria define machine-evaluable thresholds for hypothesis validation.
	// When all criteria pass, the hypothesis is "validated"; if any fail, "invalidated".
	// If criteria are omitted, the AI analyzer decides the verdict.
	// +optional
	SuccessCriteria []SuccessCriterion `json:"successCriteria,omitempty"`
}

// SuccessCriterion defines a machine-evaluable threshold for a named metric.
type SuccessCriterion struct {
	// Metric is the metric query name to evaluate (must match a key in spec.metrics).
	// +required
	// +kubebuilder:validation:Pattern=`^[a-z][a-z0-9_]*$`
	Metric string `json:"metric"`

	// Operator is the comparison operator.
	// +required
	// +kubebuilder:validation:Enum=lt;lte;gt;gte
	Operator string `json:"operator"`

	// Value is the threshold to compare against (string to support float parsing).
	// +required
	Value string `json:"value"`

	// Description is a human-readable explanation of what this criterion tests.
	// +optional
	Description string `json:"description,omitempty"`
}

// MetricsQuery defines a PromQL query to execute at experiment completion.
type MetricsQuery struct {
	// Name is the key for this metric in output JSON (e.g., "cpu_peak", "p99_latency").
	// +required
	// +kubebuilder:validation:Pattern=`^[a-z][a-z0-9_]*$`
	Name string `json:"name"`

	// Query is a PromQL expression. Variable substitution:
	//   $EXPERIMENT — experiment name
	//   $NAMESPACE  — experiment namespace
	//   $DURATION   — experiment duration as Prometheus duration (e.g., "15m", "2h")
	// +required
	Query string `json:"query"`

	// Type: "instant" (single value for bar charts) or "range" (time-series for line charts).
	// +optional
	// +kubebuilder:validation:Enum=instant;range
	// +kubebuilder:default="instant"
	Type string `json:"type,omitempty"`

	// Unit is a display hint for chart axis labels (e.g., "bytes", "cores", "req/s").
	// +optional
	Unit string `json:"unit,omitempty"`

	// Description is a human-readable chart title.
	// +optional
	Description string `json:"description,omitempty"`

	// Group is an optional grouping label for organizing metrics in the UI.
	// +optional
	Group string `json:"group,omitempty"`
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

	// Depends lists other target names that must have healthy ArgoCD apps
	// before this target's apps are created. Used for ordered multi-target deployment.
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

	// ResultsURL is the S3 path where experiment results are stored
	// +optional
	ResultsURL string `json:"resultsURL,omitempty"`

	// Published indicates whether results were successfully committed to the benchmark site
	// +optional
	Published bool `json:"published,omitempty"`

	// PublishBranch is the git branch where experiment results were committed
	// +optional
	PublishBranch string `json:"publishBranch,omitempty"`

	// PublishPRNumber is the GitHub PR number for reviewing experiment results
	// +optional
	PublishPRNumber int `json:"publishPRNumber,omitempty"`

	// PublishPRURL is the GitHub PR URL for reviewing experiment results
	// +optional
	PublishPRURL string `json:"publishPRURL,omitempty"`

	// AnalysisJobName is the name of the AI analyzer Job (empty if not launched)
	// +optional
	AnalysisJobName string `json:"analysisJobName,omitempty"`

	// AnalysisPhase tracks the AI analyzer Job lifecycle.
	// Empty when no analysis was requested. Set to Pending when the Job is created,
	// Running when the Job starts, Succeeded/Failed on completion.
	// +optional
	AnalysisPhase AnalysisPhase `json:"analysisPhase,omitempty"`

	// HypothesisResult is the machine-evaluated verdict from success criteria.
	// Values: "validated" (all criteria pass), "invalidated" (any fail),
	// "insufficient" (missing/errored metrics), or empty (no criteria / AI decides).
	// +optional
	HypothesisResult string `json:"hypothesisResult,omitempty"`

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

// AnalysisPhase tracks the lifecycle of the AI analyzer Job.
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
type AnalysisPhase string

const (
	AnalysisPhasePending   AnalysisPhase = "Pending"
	AnalysisPhaseRunning   AnalysisPhase = "Running"
	AnalysisPhaseSucceeded AnalysisPhase = "Succeeded"
	AnalysisPhaseFailed    AnalysisPhase = "Failed"
	AnalysisPhaseSkipped   AnalysisPhase = "Skipped"
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

	// MachineType is the effective machine type (with defaults applied)
	// +optional
	MachineType string `json:"machineType,omitempty"`

	// NodeCount is the effective node count (with defaults applied)
	// +optional
	NodeCount int `json:"nodeCount,omitempty"`

	// +optional
	Components []string `json:"components,omitempty"`

	// KubeconfigSecret is the name of the secret containing the kubeconfig for this target
	// +optional
	KubeconfigSecret string `json:"kubeconfigSecret,omitempty"`

	// DeployedLayers tracks which ArgoCD application layers have been created
	// for this target. Values: "infra", "obs", "workload".
	// +optional
	DeployedLayers []string `json:"deployedLayers,omitempty"`

	// AppsCreated tracks whether ArgoCD apps have been created for this target.
	// Used for dependency gating in multi-target experiments.
	// +optional
	AppsCreated bool `json:"appsCreated,omitempty"`
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
// +kubebuilder:printcolumn:name="Published",type=boolean,JSONPath=`.status.published`
// +kubebuilder:printcolumn:name="Cleaned",type=boolean,JSONPath=`.status.resourcesCleaned`
// +kubebuilder:printcolumn:name="Analysis",type=string,JSONPath=`.status.analysisPhase`
// +kubebuilder:printcolumn:name="Results",type=string,JSONPath=`.status.resultsURL`,priority=1
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
