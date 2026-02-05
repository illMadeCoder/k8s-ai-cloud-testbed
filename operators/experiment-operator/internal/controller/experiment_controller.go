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
)

const (
	experimentFinalizer = "experiments.illm.io/finalizer"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.illm.io,resources=experiments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete

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
		// Terminal states, no action needed
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// handleDeletion handles experiment cleanup when deleted
func (r *ExperimentReconciler) handleDeletion(ctx context.Context, exp *experimentsv1alpha1.Experiment) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(exp, experimentFinalizer) {
		log.Info("Cleaning up experiment resources")

		// TODO: Clean up clusters, ArgoCD apps, workflows
		// This will be implemented in later phases

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

	// TODO Phase 2: Build dependency graph and create clusters
	// For now, just transition to Provisioning
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

	// TODO Phase 2: Check cluster readiness
	// TODO Phase 3: Register clusters with ArgoCD and create Applications
	// For now, simulate waiting and transition to Ready

	// Initialize target status if empty
	if len(exp.Status.Targets) == 0 {
		exp.Status.Targets = make([]experimentsv1alpha1.TargetStatus, len(exp.Spec.Targets))
		for i, target := range exp.Spec.Targets {
			exp.Status.Targets[i] = experimentsv1alpha1.TargetStatus{
				Name:  target.Name,
				Phase: "Provisioning",
			}
		}
	}

	// Simulate cluster provisioning - in reality, would check Crossplane XRDs
	log.Info("Clusters would be provisioning here, transitioning to Ready for now")
	exp.Status.Phase = experimentsv1alpha1.PhaseReady
	for i := range exp.Status.Targets {
		exp.Status.Targets[i].Phase = "Ready"
	}

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
