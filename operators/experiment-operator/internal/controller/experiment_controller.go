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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	"github.com/illmadecoder/experiment-operator/internal/crossplane"
)

const (
	experimentFinalizer = "experiments.illm.io/finalizer"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ClusterManager *crossplane.ClusterManager
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

		// Delete clusters for each target
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

		// TODO Phase 3: Delete ArgoCD Applications
		// TODO Phase 5: Delete Argo Workflows

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

	// TODO Phase 3: Register clusters with ArgoCD
	// TODO Phase 3: Create ArgoCD Applications for each target

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

	// TODO Phase 3: Check ArgoCD Application health
	// TODO Phase 5: Submit Argo Workflow
	// For now, transition to Running
	exp.Status.Phase = experimentsv1alpha1.PhaseRunning
	// exp.Status.WorkflowStatus = &experimentsv1alpha1.WorkflowStatus{
	// 	Name:  fmt.Sprintf("%s-workflow", exp.Name),
	// 	Phase: "Running",
	// }

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

	// TODO Phase 5: Watch workflow status and update phase
	// For now, simulate completion after a delay
	log.Info("Workflow would be running here, check back later")

	// Requeue after 15 seconds to check workflow status
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.Experiment{}).
		Named("experiment").
		Complete(r)
}
