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

package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"golang.org/x/oauth2/google"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	"github.com/illmadecoder/experiment-operator/internal/argocd"
	"github.com/illmadecoder/experiment-operator/internal/crossplane"
	ghclient "github.com/illmadecoder/experiment-operator/internal/github"
	"github.com/illmadecoder/experiment-operator/internal/metrics"
	"github.com/illmadecoder/experiment-operator/internal/storage"
	"github.com/illmadecoder/experiment-operator/internal/workflow"
)

const (
	experimentFinalizer = "experiments.illm.io/finalizer"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ClusterManager *crossplane.ClusterManager
	ArgoCD         *argocd.Client
	Workflow       *workflow.Manager
	S3Client       *storage.Client
	GitClient      *ghclient.Client
	MetricsURL     string
	AnalyzerImage  string
	S3Endpoint     string
	GitHubRepo     string
}

// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflowtemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=experiments.illm.io,resources=components,verbs=get;list;watch
// +kubebuilder:rbac:groups=illm.io,resources=gkeclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=create;get;list;watch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch Experiment
	experiment := &experimentsv1alpha1.Experiment{}
	if err := r.Get(ctx, req.NamespacedName, experiment); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Experiment")
		return ctrl.Result{}, err
	}

	// Handle deletion (finalizer)
	if !experiment.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, experiment)
	}

	// Add finalizer if missing
	if !controllerutil.ContainsFinalizer(experiment, experimentFinalizer) {
		controllerutil.AddFinalizer(experiment, experimentFinalizer)
		if err := r.Update(ctx, experiment); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Phase-based reconciliation
	switch experiment.Status.Phase {
	case "", experimentsv1alpha1.PhasePending:
		return r.reconcilePending(ctx, experiment)
	case experimentsv1alpha1.PhaseProvisioning:
		return r.reconcileProvisioning(ctx, experiment)
	case experimentsv1alpha1.PhaseReady:
		return r.reconcileReady(ctx, experiment)
	case experimentsv1alpha1.PhaseRunning:
		return r.reconcileRunning(ctx, experiment)
	case experimentsv1alpha1.PhaseComplete, experimentsv1alpha1.PhaseFailed:
		return r.reconcileComplete(ctx, experiment)
	}

	return ctrl.Result{}, nil
}

// handleDeletion handles experiment cleanup when deleted
func (r *ExperimentReconciler) handleDeletion(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(exp, experimentFinalizer) {
		log.Info("Cleaning up experiment resources")

		if err := r.cleanupResources(ctx, exp); err != nil {
			log.Error(err, "Cleanup failed during deletion, retrying")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(exp, experimentFinalizer)
		if err := r.Update(ctx, exp); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// reconcileComplete handles the Complete/Failed phase — cleans up expensive resources
// while preserving the Experiment CR as a historical record.
func (r *ExperimentReconciler) reconcileComplete(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Idempotency: skip if resources already cleaned
	if exp.Status.ResourcesCleaned {
		return ctrl.Result{}, nil
	}

	// Record completion timestamp before results collection so that
	// summary.json gets a real CompletedAt, durationSeconds, and costEstimate.
	now := metav1.Now()
	if exp.Status.CompletedAt == nil {
		exp.Status.CompletedAt = &now
	}

	// Collect and store experiment results before cleanup
	if exp.Status.ResultsURL == "" {
		if err := r.collectAndStoreResults(ctx, exp); err != nil {
			log.Error(err, "Failed to collect experiment results — will retry")
			if updateErr := r.Status().Update(ctx, exp); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
		}
		if err := r.Status().Update(ctx, exp); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Cleaning up resources for completed experiment", "phase", exp.Status.Phase)

	cleanupErr := r.cleanupResources(ctx, exp)

	// Only mark cleaned if all expensive resources were successfully deleted
	if cleanupErr != nil {
		log.Error(cleanupErr, "Cleanup incomplete — cloud resources may still be running")
		// Requeue to retry cleanup
		if err := r.Status().Update(ctx, exp); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	exp.Status.ResourcesCleaned = true

	if err := r.Status().Update(ctx, exp); err != nil {
		log.Error(err, "Failed to update status after cleanup")
		return ctrl.Result{}, err
	}

	// Remove finalizer — no more sub-resources to protect
	if controllerutil.ContainsFinalizer(exp, experimentFinalizer) {
		controllerutil.RemoveFinalizer(exp, experimentFinalizer)
		if err := r.Update(ctx, exp); err != nil {
			log.Error(err, "Failed to remove finalizer after cleanup")
			return ctrl.Result{}, err
		}
	}

	log.Info("Experiment resources cleaned, CR preserved as history", "completedAt", exp.Status.CompletedAt.Time)
	return ctrl.Result{}, nil
}

// collectAndStoreResults gathers experiment metrics and uploads them to S3.
func (r *ExperimentReconciler) collectAndStoreResults(ctx context.Context, exp *experimentsv1alpha1.Experiment) error {
	log := logf.FromContext(ctx)

	if r.S3Client == nil {
		log.Info("S3 client not configured, skipping results collection")
		exp.Status.ResultsURL = "disabled"
		return nil
	}

	prefix := exp.Name

	// Build summary
	summary := metrics.CollectSummary(exp)

	// Phase 1: Try collecting metrics from target cluster monitoring stacks.
	// Target clusters (GKE) often have Prometheus/VictoriaMetrics deployed as components.
	// Retry discovery because ArgoCD may still be deploying monitoring when the workflow completes.
	var metricsResult *metrics.MetricsResult
	const maxDiscoveryAttempts = 6
	const discoveryRetryInterval = 10 * time.Second

	for i, target := range exp.Spec.Targets {
		if target.Cluster.Type == "hub" {
			continue
		}
		if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		kubeconfig, err := r.ClusterManager.GetClusterKubeconfig(ctx, clusterName, target.Cluster.Type)
		if err != nil {
			log.Error(err, "Failed to get kubeconfig for target metrics", "cluster", clusterName)
			continue
		}

		var endpoints []metrics.MonitoringEndpoint
		for attempt := 1; attempt <= maxDiscoveryAttempts; attempt++ {
			endpoints, err = metrics.DiscoverMonitoringServices(ctx, kubeconfig, exp.Name)
			if err != nil {
				log.Error(err, "Monitoring discovery failed", "cluster", clusterName, "attempt", attempt)
				break
			}
			if len(endpoints) > 0 {
				break
			}
			if attempt < maxDiscoveryAttempts {
				log.Info("No monitoring services found yet, retrying", "cluster", clusterName, "attempt", attempt, "nextRetry", discoveryRetryInterval)
				time.Sleep(discoveryRetryInterval)
			}
		}
		if len(endpoints) == 0 {
			log.Info("No monitoring services found on target cluster", "cluster", clusterName)
			continue
		}

		log.Info("Discovered monitoring endpoints on target", "cluster", clusterName, "count", len(endpoints))

		result, err := metrics.CollectMetricsFromTarget(ctx, kubeconfig, endpoints, exp)
		if err != nil {
			log.Error(err, "Target metrics collection failed", "cluster", clusterName)
			continue
		}

		if !metrics.AllQueriesEmpty(result) {
			metricsResult = result
			log.Info("Collected metrics from target cluster", "cluster", clusterName, "source", result.Source)
			break
		}
		log.Info("Target metrics returned empty data", "cluster", clusterName)
	}

	// Phase 2: Fall back to hub VictoriaMetrics if target collection failed or returned empty.
	if metricsResult == nil || metrics.AllQueriesEmpty(metricsResult) {
		hubResult, err := metrics.CollectMetricsSnapshot(ctx, r.MetricsURL, exp)
		if err != nil {
			log.Error(err, "Hub metrics snapshot failed — continuing without metrics")
		} else if hubResult != nil && !metrics.AllQueriesEmpty(hubResult) {
			metricsResult = hubResult
			log.Info("Collected metrics from hub VictoriaMetrics")
		}
	}

	summary.Metrics = metricsResult

	// Estimate cost
	summary.CostEstimate = metrics.EstimateCost(exp)

	// Upload summary
	if err := r.S3Client.PutJSON(ctx, prefix+"/summary.json", summary); err != nil {
		return fmt.Errorf("upload summary.json: %w", err)
	}

	// Upload metrics snapshot separately for easier tooling consumption
	if metricsResult != nil {
		if err := r.S3Client.PutJSON(ctx, prefix+"/metrics-snapshot.json", metricsResult); err != nil {
			log.Error(err, "Failed to upload metrics-snapshot.json — non-fatal")
		}
	}

	exp.Status.ResultsURL = fmt.Sprintf("s3://experiment-results/%s/", prefix)
	log.Info("Experiment results stored", "url", exp.Status.ResultsURL)

	// Commit results to GitHub and run AI analysis only for publishable experiments.
	// Non-publish experiments are stored in S3 only (no site publish, no AI analysis cost).
	if exp.Spec.Publish && exp.Status.Phase == experimentsv1alpha1.PhaseComplete {
		// Commit results to GitHub for benchmark site (best-effort, non-fatal).
		if r.GitClient != nil {
			if err := r.GitClient.CommitResult(ctx, exp.Name, summary); err != nil {
				log.Error(err, "Failed to commit results to GitHub — non-fatal", "repo", r.GitClient.RepoPath())
			} else {
				exp.Status.Published = true
				log.Info("Results committed to GitHub", "experiment", exp.Name, "repo", r.GitClient.RepoPath())
			}
		}

		// Launch AI analysis Job (best-effort, non-fatal).
		if r.AnalyzerImage != "" {
			if err := r.createAnalysisJob(ctx, exp); err != nil {
				log.Error(err, "Failed to create analysis Job — non-fatal")
			} else {
				jobName := fmt.Sprintf("experiment-analyzer-%s", exp.Name)
				if len(jobName) > 63 {
					jobName = jobName[:63]
				}
				exp.Status.AnalysisJobName = jobName
			}
		}
	} else if !exp.Spec.Publish {
		log.Info("Skipping site publish and AI analysis — spec.publish is false", "experiment", exp.Name)
	}

	return nil
}

// createAnalysisJob creates a Kubernetes Job that runs the experiment analyzer
// (Claude Code CLI) to generate AI analysis of the experiment results.
func (r *ExperimentReconciler) createAnalysisJob(ctx context.Context, exp *experimentsv1alpha1.Experiment) error {
	log := logf.FromContext(ctx)

	// Truncate job name to 63 chars (K8s name limit)
	jobName := fmt.Sprintf("experiment-analyzer-%s", exp.Name)
	if len(jobName) > 63 {
		jobName = jobName[:63]
	}

	namespace := "experiment-operator-system"

	// Check if job already exists
	existing := &batchv1.Job{}
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, existing); err == nil {
		log.Info("Analysis Job already exists", "job", jobName)
		return nil
	}

	s3Endpoint := r.S3Endpoint

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "experiment-analyzer",
				"experiment": exp.Name,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(1),
			TTLSecondsAfterFinished: int32Ptr(3600),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "experiment-analyzer",
						"experiment": exp.Name,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: boolPtr(true),
						RunAsUser:    int64Ptr(1000),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "analyzer",
							Image: r.AnalyzerImage,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: boolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "EXPERIMENT_NAME",
									Value: exp.Name,
								},
								{
									Name:  "S3_ENDPOINT",
									Value: s3Endpoint,
								},
								{
									Name: "CLAUDE_CODE_OAUTH_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "claude-auth",
											},
											Key: "token",
										},
									},
								},
								{
									Name: "GITHUB_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "github-api-token",
											},
											Key:      "token",
											Optional: boolPtr(true),
										},
									},
								},
								{
									Name:  "GITHUB_REPO",
									Value: r.GitHubRepo,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Note: Owner references are not set because the Experiment CR is in a
	// different namespace (experiments) than the Job (experiment-operator-system).
	// The TTLSecondsAfterFinished field handles cleanup instead.

	if err := r.Create(ctx, job); err != nil {
		return fmt.Errorf("create analysis Job %s: %w", jobName, err)
	}

	log.Info("Created analysis Job", "job", jobName, "image", r.AnalyzerImage)
	return nil
}

// cleanupResources deletes expensive sub-resources (ArgoCD apps, cluster secrets,
// GKE clusters, kubeconfig secrets, Argo Workflows). Shared by handleDeletion and
// reconcileComplete.
func (r *ExperimentReconciler) cleanupResources(ctx context.Context, exp *experimentsv1alpha1.Experiment) error {
	log := logf.FromContext(ctx)
	var clusterDeleteErr error

	// Delete ArgoCD Applications
	for _, target := range exp.Spec.Targets {
		if err := r.ArgoCD.AppManager.DeleteApplication(ctx, exp.Name, target.Name); err != nil {
			log.Error(err, "Failed to delete application", "target", target.Name)
		}
	}

	// Unregister clusters from ArgoCD
	clusterNames := []string{}
	for i := range exp.Spec.Targets {
		if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
			continue
		}
		clusterNames = append(clusterNames, exp.Status.Targets[i].ClusterName)
	}
	if err := r.ArgoCD.DeleteClusterAndApps(ctx, exp.Name, exp.Spec.Targets, clusterNames); err != nil {
		log.Error(err, "Failed to unregister clusters from ArgoCD")
	}

	// Delete clusters — this is the expensive resource, errors are fatal
	for i, target := range exp.Spec.Targets {
		if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		log.Info("Deleting cluster", "cluster", clusterName, "type", target.Cluster.Type)

		if err := r.ClusterManager.DeleteCluster(ctx, clusterName, target.Cluster.Type); err != nil {
			log.Error(err, "Failed to delete cluster", "cluster", clusterName)
			clusterDeleteErr = err
		}
	}

	// Delete kubeconfig secrets in experiments namespace
	if exp.Status.TutorialStatus != nil {
		for targetName, secretName := range exp.Status.TutorialStatus.KubeconfigSecrets {
			secret := &corev1.Secret{}
			if err := r.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: exp.Namespace,
			}, secret); err == nil {
				if err := r.Delete(ctx, secret); err != nil {
					log.Error(err, "Failed to delete kubeconfig secret", "target", targetName, "secret", secretName)
				}
			}
		}
	}

	// Delete Argo Workflows
	if exp.Status.WorkflowStatus != nil && exp.Status.WorkflowStatus.Name != "" {
		if err := r.Workflow.DeleteWorkflow(ctx, exp.Status.WorkflowStatus.Name); err != nil {
			log.Error(err, "Failed to delete workflow", "workflow", exp.Status.WorkflowStatus.Name)
		}
	}

	return clusterDeleteErr
}

// reconcilePending handles the Pending phase - creates cluster resources
func (r *ExperimentReconciler) reconcilePending(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Pending phase")

	// Initialize target status if needed
	if len(exp.Status.Targets) == 0 {
		exp.Status.Targets = make([]experimentsv1alpha1.TargetStatus, len(exp.Spec.Targets))
		for i, target := range exp.Spec.Targets {
			exp.Status.Targets[i] = experimentsv1alpha1.TargetStatus{
				Name:  target.Name,
				Phase: "Pending",
			}
		}
	}

	// Create clusters for each target (respecting dependencies)
	// For simplicity, create all clusters in parallel for now
	// TODO: Implement proper dependency graph traversal
	for i, target := range exp.Spec.Targets {
		// Check if cluster already exists in status
		if exp.Status.Targets[i].ClusterName != "" {
			log.Info("Cluster already created", "target", target.Name, "cluster", exp.Status.Targets[i].ClusterName)
			continue
		}

		// Create the cluster
		clusterName, err := r.ClusterManager.CreateCluster(ctx, exp.Name, target)
		if err != nil {
			log.Error(err, "Failed to create cluster", "target", target.Name)
			// Continue with other targets, will retry on next reconcile
			exp.Status.Targets[i].Phase = "Failed"
			continue
		}

		// Update target status
		exp.Status.Targets[i].ClusterName = clusterName
		exp.Status.Targets[i].Phase = "Provisioning"
		log.Info("Created cluster", "target", target.Name, "cluster", clusterName)
	}

	// Transition to Provisioning phase
	exp.Status.Phase = experimentsv1alpha1.PhaseProvisioning
	if err := r.Status().Update(ctx, exp); err != nil {
		log.Error(err, "Failed to update status to Provisioning")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// reconcileProvisioning handles the Provisioning phase - waits for clusters to be ready
func (r *ExperimentReconciler) reconcileProvisioning(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Provisioning phase")

	allReady := true

	// Check each cluster's readiness
	for i, target := range exp.Spec.Targets {
		if exp.Status.Targets[i].ClusterName == "" {
			log.Info("Target has no cluster name", "target", target.Name)
			allReady = false
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		ready, err := r.ClusterManager.IsClusterReady(ctx, clusterName, target.Cluster.Type)
		if err != nil {
			log.Error(err, "Failed to check cluster readiness", "cluster", clusterName)
			allReady = false
			continue
		}

		if !ready {
			log.Info("Cluster not ready yet", "cluster", clusterName)
			allReady = false
			continue
		}

		// Get cluster endpoint if available
		endpoint, err := r.ClusterManager.GetClusterEndpoint(ctx, clusterName, target.Cluster.Type)
		if err != nil {
			log.Error(err, "Failed to get cluster endpoint", "cluster", clusterName)
		} else if endpoint != "" {
			exp.Status.Targets[i].Endpoint = endpoint
		}

		// Mark target as ready
		exp.Status.Targets[i].Phase = "Ready"
		log.Info("Cluster is ready", "cluster", clusterName, "endpoint", endpoint)
	}

	if !allReady {
		// Requeue after 10 seconds to check again
		if err := r.Status().Update(ctx, exp); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Register clusters with ArgoCD and create Applications
	for i, target := range exp.Spec.Targets {
		if exp.Status.Targets[i].Phase != "Ready" {
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		endpoint := exp.Status.Targets[i].Endpoint

		// For hub cluster, use in-cluster server; for GKE, prefix with https://
		var server string
		if target.Cluster.Type == "hub" {
			server = "https://kubernetes.default.svc"
		} else {
			server = "https://" + endpoint
		}

		// Get kubeconfig for non-hub clusters
		var kubeconfig []byte
		var err error
		if target.Cluster.Type != "hub" {
			kubeconfig, err = r.ClusterManager.GetClusterKubeconfig(ctx, clusterName, target.Cluster.Type)
			if err != nil {
				log.Error(err, "Failed to get kubeconfig", "cluster", clusterName)
				// Use a placeholder kubeconfig for now
				kubeconfig = []byte("# Placeholder kubeconfig")
			}
		} else {
			// Hub cluster: ArgoCD already has in-cluster access via https://kubernetes.default.svc.
			// Skip cluster secret registration — just create the ArgoCD Application.
			if err := r.ArgoCD.AppManager.CreateApplication(ctx, exp.Name, target, server); err != nil {
				log.Error(err, "Failed to create application for hub target", "target", target.Name)
				continue
			}
			log.Info("Created ArgoCD Application for hub target (no cluster registration needed)", "target", target.Name)
			continue
		}

		// Bootstrap RBAC on the target cluster (grant client cert user cluster-admin)
		gcpKey := r.getGCPProviderKey(ctx)
		if err := bootstrapClusterRBAC(ctx, kubeconfig, gcpKey); err != nil {
			log.Error(err, "Failed to bootstrap RBAC on target cluster", "cluster", clusterName)
			// Continue anyway — the cluster may already have RBAC configured
		}

		// Register cluster and create applications
		if err := r.ArgoCD.RegisterClusterAndCreateApps(ctx, exp.Name, target, clusterName, kubeconfig, server); err != nil {
			log.Error(err, "Failed to register cluster and create apps", "cluster", clusterName)
			continue
		}

		log.Info("Registered cluster with ArgoCD and created applications", "cluster", clusterName)
	}

	// Copy Tailscale OAuth secret to target clusters that need tailscale transport
	for i, target := range exp.Spec.Targets {
		if target.Observability == nil || !target.Observability.Enabled || target.Observability.Transport != "tailscale" {
			continue
		}
		if target.Cluster.Type == "hub" {
			continue // Hub cluster already has Tailscale
		}
		if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		kubeconfigBytes, kcErr := r.ClusterManager.GetClusterKubeconfig(ctx, clusterName, target.Cluster.Type)
		if kcErr != nil {
			log.Error(kcErr, "Failed to get kubeconfig for Tailscale secret copy", "cluster", clusterName)
			continue
		}

		if err := r.copyTailscaleSecret(ctx, kubeconfigBytes); err != nil {
			log.Error(err, "Failed to copy Tailscale OAuth secret to target cluster", "cluster", clusterName)
			// Non-fatal: Tailscale operator will fail to start but experiment can still run
		} else {
			log.Info("Copied Tailscale OAuth secret to target cluster", "cluster", clusterName)
		}
	}

	// Copy kubeconfigs to experiments namespace for tutorial access
	if exp.Spec.Tutorial != nil && exp.Spec.Tutorial.ExposeKubeconfig {
		if err := r.copyKubeconfigSecrets(ctx, exp); err != nil {
			log.Error(err, "Failed to copy kubeconfig secrets for tutorial")
			// Non-fatal: continue with phase transition
		}
	}

	// All clusters ready, transition to Ready phase
	exp.Status.Phase = experimentsv1alpha1.PhaseReady
	if err := r.Status().Update(ctx, exp); err != nil {
		log.Error(err, "Failed to update status to Ready")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// reconcileReady handles the Ready phase - waits for apps to be healthy, then submits workflow
func (r *ExperimentReconciler) reconcileReady(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Ready phase")

	// Check if all ArgoCD Applications are healthy
	allHealthy := true
	for i, target := range exp.Spec.Targets {
		if len(target.Components) == 0 {
			// No components to deploy, mark as ready
			continue
		}

		healthy, err := r.ArgoCD.AppManager.IsApplicationHealthy(ctx, exp.Name, target.Name)
		if err != nil {
			log.Error(err, "Failed to check application health", "target", target.Name)
			allHealthy = false
			continue
		}

		if !healthy {
			log.Info("Application not healthy yet", "target", target.Name)
			allHealthy = false
			continue
		}

		// Get deployed components
		components, err := r.ArgoCD.AppManager.GetApplicationComponents(ctx, exp.Name, target.Name)
		if err != nil {
			log.Error(err, "Failed to get application components", "target", target.Name)
		} else {
			exp.Status.Targets[i].Components = components
			log.Info("Application is healthy", "target", target.Name, "components", len(components))
		}
	}

	if !allHealthy {
		// Requeue after 15 seconds to check again
		if err := r.Status().Update(ctx, exp); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	// Discover tutorial services on target clusters
	if exp.Spec.Tutorial != nil && len(exp.Spec.Tutorial.Services) > 0 {
		if err := r.discoverTutorialServices(ctx, exp); err != nil {
			log.Error(err, "Failed to discover tutorial services")
			// Non-fatal: services may become available later
		}
	}

	// Build workflow params: inject experiment-name and first target endpoint
	// so WorkflowTemplates can reach the deployed application.
	wfSpec := exp.Spec.Workflow.DeepCopy()
	if wfSpec.Params == nil {
		wfSpec.Params = make(map[string]string)
	}
	wfSpec.Params["experiment-name"] = exp.Name
	for i, target := range exp.Spec.Targets {
		if i < len(exp.Status.Targets) && exp.Status.Targets[i].Endpoint != "" {
			wfSpec.Params["target-endpoint"] = exp.Status.Targets[i].Endpoint
			wfSpec.Params["target-name"] = target.Name
			break
		}
	}

	// Submit Argo Workflow for validation
	workflowName, err := r.Workflow.SubmitWorkflow(ctx, exp.Name, *wfSpec)
	if err != nil {
		log.Error(err, "Failed to submit workflow")
		// Don't fail the experiment - requeue and try again
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	now := metav1.Now()
	exp.Status.WorkflowStatus = &experimentsv1alpha1.WorkflowStatus{
		Name:      workflowName,
		Phase:     "Pending",
		StartedAt: &now,
	}
	exp.Status.Phase = experimentsv1alpha1.PhaseRunning

	if err := r.Status().Update(ctx, exp); err != nil {
		log.Error(err, "Failed to update status to Running")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// reconcileRunning handles the Running phase - watches workflow status
func (r *ExperimentReconciler) reconcileRunning(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Running phase")

	if exp.Status.WorkflowStatus == nil || exp.Status.WorkflowStatus.Name == "" {
		log.Info("No workflow status found, transitioning to Complete")
		exp.Status.Phase = experimentsv1alpha1.PhaseComplete
		return ctrl.Result{}, r.Status().Update(ctx, exp)
	}

	// Get workflow status from Argo
	result, err := r.Workflow.GetWorkflowStatus(ctx, exp.Status.WorkflowStatus.Name)
	if err != nil {
		log.Error(err, "Failed to get workflow status", "workflow", exp.Status.WorkflowStatus.Name)
		// Requeue - workflow might not be visible yet
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Update workflow status
	exp.Status.WorkflowStatus.Phase = result.Phase
	if result.StartedAt != nil {
		exp.Status.WorkflowStatus.StartedAt = result.StartedAt
	}
	if result.FinishedAt != nil {
		exp.Status.WorkflowStatus.FinishedAt = result.FinishedAt
	}

	// Check if workflow reached a terminal state
	if workflow.IsTerminal(result.Phase) {
		if workflow.IsSucceeded(result.Phase) {
			log.Info("Workflow succeeded", "workflow", exp.Status.WorkflowStatus.Name)
			// In manual mode, stay in Running after workflow succeeds
			// User controls lifecycle via hub:down; TTL is the safety net
			if exp.Spec.Workflow.Completion.Mode == "manual" {
				log.Info("Manual completion mode: staying in Running phase")
				if err := r.Status().Update(ctx, exp); err != nil {
					log.Error(err, "Failed to update workflow status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
			}
			exp.Status.Phase = experimentsv1alpha1.PhaseComplete
		} else {
			log.Info("Workflow failed", "workflow", exp.Status.WorkflowStatus.Name, "phase", result.Phase, "message", result.Message)
			exp.Status.Phase = experimentsv1alpha1.PhaseFailed
		}
		return ctrl.Result{}, r.Status().Update(ctx, exp)
	}

	// Workflow still running, requeue to check again
	log.Info("Workflow still running", "workflow", exp.Status.WorkflowStatus.Name, "phase", result.Phase)
	if err := r.Status().Update(ctx, exp); err != nil {
		log.Error(err, "Failed to update workflow status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// copyKubeconfigSecrets copies kubeconfig secrets from crossplane-system to experiments namespace
// so that labctl can read them without needing access to crossplane-system.
func (r *ExperimentReconciler) copyKubeconfigSecrets(ctx context.Context, exp *experimentsv1alpha1.Experiment) error {
	log := logf.FromContext(ctx)

	if exp.Status.TutorialStatus == nil {
		exp.Status.TutorialStatus = &experimentsv1alpha1.TutorialStatus{
			KubeconfigSecrets: make(map[string]string),
		}
	}
	if exp.Status.TutorialStatus.KubeconfigSecrets == nil {
		exp.Status.TutorialStatus.KubeconfigSecrets = make(map[string]string)
	}

	for i, target := range exp.Spec.Targets {
		if target.Cluster.Type == "hub" {
			continue // Hub cluster uses in-cluster config
		}
		if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
			continue
		}

		clusterName := exp.Status.Targets[i].ClusterName
		dstSecretName := fmt.Sprintf("%s-%s-kubeconfig", exp.Name, target.Name)

		// Get kubeconfig via ClusterManager (handles XR name resolution)
		kubeconfigBytes, err := r.ClusterManager.GetClusterKubeconfig(ctx, clusterName, target.Cluster.Type)
		if err != nil {
			log.Error(err, "Failed to get kubeconfig for tutorial", "cluster", clusterName)
			continue
		}

		// Create destination secret in experiments namespace
		dstSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dstSecretName,
				Namespace: exp.Namespace,
				Labels: map[string]string{
					"experiment": exp.Name,
					"target":     target.Name,
				},
			},
			Data: map[string][]byte{
				"kubeconfig": kubeconfigBytes,
			},
		}

		// Set owner reference for garbage collection
		if err := controllerutil.SetOwnerReference(exp, dstSecret, r.Scheme); err != nil {
			log.Error(err, "Failed to set owner reference on kubeconfig secret", "secret", dstSecretName)
			continue
		}

		// Create or update the secret
		existing := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      dstSecretName,
			Namespace: exp.Namespace,
		}, existing); err != nil {
			if errors.IsNotFound(err) {
				if err := r.Create(ctx, dstSecret); err != nil {
					log.Error(err, "Failed to create kubeconfig secret", "secret", dstSecretName)
					continue
				}
			} else {
				log.Error(err, "Failed to check for existing kubeconfig secret", "secret", dstSecretName)
				continue
			}
		} else {
			existing.Data = dstSecret.Data
			if err := r.Update(ctx, existing); err != nil {
				log.Error(err, "Failed to update kubeconfig secret", "secret", dstSecretName)
				continue
			}
		}

		exp.Status.Targets[i].KubeconfigSecret = dstSecretName
		exp.Status.TutorialStatus.KubeconfigSecrets[target.Name] = dstSecretName
		log.Info("Copied kubeconfig secret", "cluster", clusterName, "dst", dstSecretName)
	}

	return nil
}

// copyTailscaleSecret reads the operator-oauth Secret from the tailscale namespace
// on the hub and creates a matching secret + namespace on the target cluster.
func (r *ExperimentReconciler) copyTailscaleSecret(ctx context.Context, targetKubeconfig []byte) error {
	log := logf.FromContext(ctx)

	// Read operator-oauth secret from hub
	hubSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      "operator-oauth",
		Namespace: "tailscale",
	}, hubSecret); err != nil {
		return fmt.Errorf("failed to read operator-oauth secret from hub: %w", err)
	}

	// Build client for target cluster
	targetCfg, err := clientcmd.RESTConfigFromKubeConfig(targetKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to parse target kubeconfig: %w", err)
	}
	targetClient, err := kubernetes.NewForConfig(targetCfg)
	if err != nil {
		return fmt.Errorf("failed to create target clientset: %w", err)
	}

	// Ensure tailscale namespace exists on target
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tailscale",
		},
	}
	if _, err := targetClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create tailscale namespace on target: %w", err)
		}
	}

	// Create or update the secret on target
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "operator-oauth",
			Namespace: "tailscale",
		},
		Data: hubSecret.Data,
		Type: hubSecret.Type,
	}
	if _, err := targetClient.CoreV1().Secrets("tailscale").Create(ctx, targetSecret, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			existing, getErr := targetClient.CoreV1().Secrets("tailscale").Get(ctx, "operator-oauth", metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("failed to get existing secret on target: %w", getErr)
			}
			existing.Data = hubSecret.Data
			if _, updateErr := targetClient.CoreV1().Secrets("tailscale").Update(ctx, existing, metav1.UpdateOptions{}); updateErr != nil {
				return fmt.Errorf("failed to update secret on target: %w", updateErr)
			}
			log.Info("Updated operator-oauth secret on target cluster")
			return nil
		}
		return fmt.Errorf("failed to create operator-oauth secret on target: %w", err)
	}

	return nil
}

// discoverTutorialServices queries target clusters for services listed in spec.tutorial.services
// and populates status.tutorialStatus.services with resolved endpoints.
func (r *ExperimentReconciler) discoverTutorialServices(ctx context.Context, exp *experimentsv1alpha1.Experiment) error {
	log := logf.FromContext(ctx)

	if exp.Status.TutorialStatus == nil {
		exp.Status.TutorialStatus = &experimentsv1alpha1.TutorialStatus{}
	}

	var discovered []experimentsv1alpha1.DiscoveredService

	for _, svcRef := range exp.Spec.Tutorial.Services {
		ds := experimentsv1alpha1.DiscoveredService{
			Name:  svcRef.Name,
			Ready: false,
		}

		// Find the target and its kubeconfig
		var targetIdx int = -1
		for i, t := range exp.Spec.Targets {
			if t.Name == svcRef.Target {
				targetIdx = i
				break
			}
		}

		if targetIdx < 0 || targetIdx >= len(exp.Status.Targets) {
			log.Info("Target not found for service discovery", "service", svcRef.Name, "target", svcRef.Target)
			discovered = append(discovered, ds)
			continue
		}

		target := exp.Spec.Targets[targetIdx]

		// For hub cluster, query using the reconciler's client directly
		if target.Cluster.Type == "hub" {
			svc := &corev1.Service{}
			if err := r.Get(ctx, types.NamespacedName{
				Name:      svcRef.Service,
				Namespace: svcRef.Namespace,
			}, svc); err != nil {
				log.Error(err, "Failed to get service", "service", svcRef.Service, "namespace", svcRef.Namespace)
				discovered = append(discovered, ds)
				continue
			}
			ds.Endpoint = resolveServiceEndpoint(svc, svcRef.Port)
			ds.Ready = ds.Endpoint != ""
		} else {
			// For remote clusters, we'd need a client built from the kubeconfig.
			// For now, mark as pending — labctl does live discovery against the target.
			log.Info("Service discovery for remote clusters deferred to labctl",
				"service", svcRef.Name, "target", svcRef.Target)
		}

		discovered = append(discovered, ds)
	}

	exp.Status.TutorialStatus.Services = discovered
	return nil
}

// resolveServiceEndpoint returns the best endpoint for a service (LoadBalancer IP preferred, then ClusterIP).
func resolveServiceEndpoint(svc *corev1.Service, port int) string {
	var host string

	// Prefer LoadBalancer IP
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				host = ingress.IP
				break
			}
			if ingress.Hostname != "" {
				host = ingress.Hostname
				break
			}
		}
	}

	// Fall back to ClusterIP
	if host == "" {
		host = svc.Spec.ClusterIP
	}

	if host == "" || host == "None" {
		return ""
	}

	// Determine port
	if port == 0 && len(svc.Spec.Ports) > 0 {
		port = int(svc.Spec.Ports[0].Port)
	}

	if port != 0 && port != 80 && port != 443 {
		return fmt.Sprintf("http://%s:%d", host, port)
	}
	return fmt.Sprintf("http://%s", host)
}

// getGCPProviderKey reads the GCP service account key from the Crossplane provider secret.
func (r *ExperimentReconciler) getGCPProviderKey(ctx context.Context) []byte {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      "gcp-credentials",
		Namespace: "crossplane-system",
	}, secret); err != nil {
		return nil
	}
	return secret.Data["credentials"]
}

// bootstrapClusterRBAC connects to a target GKE cluster using GCP Application
// Default Credentials and creates a ClusterRoleBinding granting the "client"
// user (from the GKE client certificate) cluster-admin permissions.
// GCP IAM identity has automatic cluster-admin on GKE, so we use that to
// bootstrap RBAC for the client cert user.
func bootstrapClusterRBAC(ctx context.Context, kubeconfig []byte, gcpServiceAccountKey []byte) error {
	log := logf.FromContext(ctx)

	// Parse kubeconfig to get server URL and CA data
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Get GCP access token — try ADC first, fall back to Crossplane provider secret
	tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		// Fall back to Crossplane GCP provider credentials
		log.Info("ADC not available, trying Crossplane provider credentials")
		creds, credErr := google.CredentialsFromJSON(ctx, gcpServiceAccountKey, "https://www.googleapis.com/auth/cloud-platform")
		if credErr != nil {
			return fmt.Errorf("failed to get GCP credentials: ADC: %v, Crossplane SA: %v", err, credErr)
		}
		tokenSource = creds.TokenSource
	}
	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get GCP access token: %w", err)
	}

	// Build a REST config using GCP token instead of client cert
	gcpCfg := &restclient.Config{
		Host:        cfg.Host,
		BearerToken: token.AccessToken,
		TLSClientConfig: restclient.TLSClientConfig{
			CAData:   cfg.TLSClientConfig.CAData,
			Insecure: cfg.TLSClientConfig.Insecure,
		},
	}

	clientset, err := kubernetes.NewForConfig(gcpCfg)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "experiment-client-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: "client",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Info("ClusterRoleBinding already exists on target cluster")
			return nil
		}
		return fmt.Errorf("failed to create ClusterRoleBinding: %w", err)
	}

	log.Info("Created ClusterRoleBinding for client user on target cluster")
	return nil
}

// int32Ptr returns a pointer to an int32 value.
func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.Experiment{}).
		Named("experiment").
		Complete(r)
}
