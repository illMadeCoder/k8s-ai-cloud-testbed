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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	"github.com/illmadecoder/experiment-operator/internal/argocd"
	"github.com/illmadecoder/experiment-operator/internal/crossplane"
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
}

// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=container.gcp.upbound.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=vclusters,verbs=get;list;watch;create;update;patch;delete

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

	// Check TTL and auto-delete if exceeded
	ttlDays := experiment.Spec.TTLDays
	if ttlDays == 0 {
		ttlDays = 1 // Default to 1 day
	}
	if crossplane.ShouldDeleteCluster(experiment.CreationTimestamp.Time, ttlDays) {
		log.Info("Experiment TTL exceeded, deleting", "ttlDays", ttlDays, "age", time.Since(experiment.CreationTimestamp.Time))
		if err := r.Delete(ctx, experiment); err != nil {
			log.Error(err, "Failed to delete expired experiment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
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
		// Terminal states - requeue to check TTL periodically
		// Check TTL every hour for completed experiments
		return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
	}

	return ctrl.Result{}, nil
}

// handleDeletion handles experiment cleanup when deleted
func (r *ExperimentReconciler) handleDeletion(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(exp, experimentFinalizer) {
		log.Info("Cleaning up experiment resources")

		// Delete ArgoCD Applications
		for _, target := range exp.Spec.Targets {
			if err := r.ArgoCD.AppManager.DeleteApplication(ctx, exp.Name, target.Name); err != nil {
				log.Error(err, "Failed to delete application", "target", target.Name)
				// Continue with other deletions
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

		// Delete clusters
		for i, target := range exp.Spec.Targets {
			if i >= len(exp.Status.Targets) || exp.Status.Targets[i].ClusterName == "" {
				continue
			}

			clusterName := exp.Status.Targets[i].ClusterName
			log.Info("Deleting cluster", "cluster", clusterName, "type", target.Cluster.Type)

			if err := r.ClusterManager.DeleteCluster(ctx, clusterName, target.Cluster.Type); err != nil {
				log.Error(err, "Failed to delete cluster", "cluster", clusterName)
				// Continue with other clusters, don't block finalizer removal
			}
		}

		// Delete kubeconfig secrets in experiments namespace (owned by experiment, so
		// normally GC'd, but explicit delete is cleaner for fast cleanup)
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
				// Continue - don't block finalizer removal
			}
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

		// For hub cluster, use in-cluster server
		server := endpoint
		if target.Cluster.Type == "hub" {
			server = "https://kubernetes.default.svc"
		} else if server == "" {
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

		// Register cluster and create applications
		if err := r.ArgoCD.RegisterClusterAndCreateApps(ctx, exp.Name, target, clusterName, kubeconfig, server); err != nil {
			log.Error(err, "Failed to register cluster and create apps", "cluster", clusterName)
			continue
		}

		log.Info("Registered cluster with ArgoCD and created applications", "cluster", clusterName)
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

	// Submit Argo Workflow for validation
	workflowName, err := r.Workflow.SubmitWorkflow(ctx, exp.Name, exp.Spec.Workflow)
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
		srcSecretName := fmt.Sprintf("%s-kubeconfig", clusterName)
		dstSecretName := fmt.Sprintf("%s-%s-kubeconfig", exp.Name, target.Name)

		// Read source secret from crossplane-system
		srcSecret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      srcSecretName,
			Namespace: "crossplane-system",
		}, srcSecret); err != nil {
			log.Error(err, "Failed to read kubeconfig secret", "secret", srcSecretName)
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
			Data: srcSecret.Data,
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
			existing.Data = srcSecret.Data
			if err := r.Update(ctx, existing); err != nil {
				log.Error(err, "Failed to update kubeconfig secret", "secret", dstSecretName)
				continue
			}
		}

		exp.Status.Targets[i].KubeconfigSecret = dstSecretName
		exp.Status.TutorialStatus.KubeconfigSecrets[target.Name] = dstSecretName
		log.Info("Copied kubeconfig secret", "src", srcSecretName, "dst", dstSecretName)
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

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.Experiment{}).
		Named("experiment").
		Complete(r)
}
