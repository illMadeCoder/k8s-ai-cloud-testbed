package workflow

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

const (
	// DefaultNamespace is the default namespace for Argo Workflows
	DefaultNamespace = "argo-workflows"
)

var (
	// Argo Workflow GVKs
	workflowGVK = schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Workflow",
	}

	workflowTemplateGVK = schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "WorkflowTemplate",
	}
)

// Manager manages Argo Workflows for experiments
type Manager struct {
	client.Client
	Namespace string
}

// NewManager creates a new workflow manager
func NewManager(c client.Client) *Manager {
	return &Manager{
		Client:    c,
		Namespace: DefaultNamespace,
	}
}

// WorkflowResult contains the result of a workflow status check
type WorkflowResult struct {
	Phase      string
	StartedAt  *metav1.Time
	FinishedAt *metav1.Time
	Message    string
}

// SubmitWorkflow creates an Argo Workflow from the experiment's workflow spec
func (m *Manager) SubmitWorkflow(ctx context.Context, experimentName string, spec experimentsv1alpha1.WorkflowSpec) (string, error) {
	logger := log.FromContext(ctx)

	workflowName := fmt.Sprintf("%s-validation", experimentName)

	// Check if WorkflowTemplate exists
	tmpl := &unstructured.Unstructured{}
	tmpl.SetGroupVersionKind(workflowTemplateGVK)
	err := m.Get(ctx, client.ObjectKey{Name: spec.Template, Namespace: m.Namespace}, tmpl)
	if err != nil {
		logger.Info("WorkflowTemplate not found, creating inline workflow", "template", spec.Template)
		// Template not found - create a basic workflow that completes immediately
		// This allows the operator to function even without pre-created templates
		return m.submitInlineWorkflow(ctx, workflowName, experimentName, spec)
	}

	// Build Workflow that references the WorkflowTemplate
	wf := &unstructured.Unstructured{}
	wf.SetGroupVersionKind(workflowGVK)
	wf.SetName(workflowName)
	wf.SetNamespace(m.Namespace)

	// Set labels for tracking
	wf.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by":  "experiment-operator",
		"experiments.illm.io/experiment": experimentName,
	})

	// Build workflow spec referencing the template
	wfSpec := map[string]interface{}{
		"serviceAccountName": "argo-workflow",
		"workflowTemplateRef": map[string]interface{}{
			"name": spec.Template,
		},
	}

	// Add parameters if provided
	if len(spec.Params) > 0 {
		params := []interface{}{}
		for key, value := range spec.Params {
			params = append(params, map[string]interface{}{
				"name":  key,
				"value": value,
			})
		}
		wfSpec["arguments"] = map[string]interface{}{
			"parameters": params,
		}
	}

	if err := unstructured.SetNestedMap(wf.Object, wfSpec, "spec"); err != nil {
		return "", fmt.Errorf("failed to set workflow spec: %w", err)
	}

	// Check if workflow already exists
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(workflowGVK)
	err = m.Get(ctx, client.ObjectKey{Name: workflowName, Namespace: m.Namespace}, existing)
	if err == nil {
		// Workflow already exists, return its name
		logger.Info("Workflow already exists", "name", workflowName)
		return workflowName, nil
	}

	// Create the workflow
	if err := m.Create(ctx, wf); err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}

	logger.Info("Submitted Argo Workflow", "name", workflowName, "template", spec.Template)
	return workflowName, nil
}

// submitInlineWorkflow creates a simple inline workflow when no template exists
func (m *Manager) submitInlineWorkflow(ctx context.Context, workflowName string, experimentName string, spec experimentsv1alpha1.WorkflowSpec) (string, error) {
	logger := log.FromContext(ctx)

	wf := &unstructured.Unstructured{}
	wf.SetGroupVersionKind(workflowGVK)
	wf.SetName(workflowName)
	wf.SetNamespace(m.Namespace)

	wf.SetLabels(map[string]string{
		"app.kubernetes.io/managed-by":  "experiment-operator",
		"experiments.illm.io/experiment": experimentName,
	})

	// Build steps for the workflow
	steps := []interface{}{
		// Step 1: Log that experiment is ready
		[]interface{}{
			map[string]interface{}{
				"name":     "log-ready",
				"template": "log",
				"arguments": map[string]interface{}{
					"parameters": []interface{}{
						map[string]interface{}{
							"name":  "message",
							"value": fmt.Sprintf("Experiment %s is running", experimentName),
						},
					},
				},
			},
		},
	}

	// Only add suspend step for manual completion mode
	if spec.Completion.Mode == "manual" {
		steps = append(steps, []interface{}{
			map[string]interface{}{
				"name":     "wait",
				"template": "suspend-step",
			},
		})
	}

	templates := []interface{}{
		map[string]interface{}{
			"name":  "experiment-lifecycle",
			"steps": steps,
		},
		map[string]interface{}{
			"name": "log",
			"inputs": map[string]interface{}{
				"parameters": []interface{}{
					map[string]interface{}{
						"name": "message",
					},
				},
			},
			"container": map[string]interface{}{
				"image":   "alpine:3.19",
				"command": []interface{}{"echo"},
				"args":    []interface{}{"{{inputs.parameters.message}}"},
			},
		},
	}

	// Only include suspend-step template when needed
	if spec.Completion.Mode == "manual" {
		templates = append(templates, map[string]interface{}{
			"name": "suspend-step",
			"suspend": map[string]interface{}{
				// No duration = wait for manual resume
				// Users can: argo resume <workflow-name>
			},
		})
	}

	wfSpec := map[string]interface{}{
		"serviceAccountName": "argo-workflow",
		"entrypoint":         "experiment-lifecycle",
		"templates":          templates,
	}

	// Add parameters as arguments if provided
	if len(spec.Params) > 0 {
		params := []interface{}{}
		for key, value := range spec.Params {
			params = append(params, map[string]interface{}{
				"name":  key,
				"value": value,
			})
		}
		wfSpec["arguments"] = map[string]interface{}{
			"parameters": params,
		}
	}

	if err := unstructured.SetNestedMap(wf.Object, wfSpec, "spec"); err != nil {
		return "", fmt.Errorf("failed to set workflow spec: %w", err)
	}

	// Check if workflow already exists
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(workflowGVK)
	err := m.Get(ctx, client.ObjectKey{Name: workflowName, Namespace: m.Namespace}, existing)
	if err == nil {
		logger.Info("Inline workflow already exists", "name", workflowName)
		return workflowName, nil
	}

	if err := m.Create(ctx, wf); err != nil {
		return "", fmt.Errorf("failed to create inline workflow: %w", err)
	}

	logger.Info("Submitted inline Argo Workflow", "name", workflowName)
	return workflowName, nil
}

// GetWorkflowStatus checks the status of a workflow
func (m *Manager) GetWorkflowStatus(ctx context.Context, workflowName string) (*WorkflowResult, error) {
	wf := &unstructured.Unstructured{}
	wf.SetGroupVersionKind(workflowGVK)

	if err := m.Get(ctx, client.ObjectKey{Name: workflowName, Namespace: m.Namespace}, wf); err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	result := &WorkflowResult{}

	// Get phase
	phase, found, err := unstructured.NestedString(wf.Object, "status", "phase")
	if err != nil || !found {
		result.Phase = "Pending"
		return result, nil
	}
	result.Phase = phase

	// Get message
	message, _, _ := unstructured.NestedString(wf.Object, "status", "message")
	result.Message = message

	// Get startedAt
	startedAtStr, found, _ := unstructured.NestedString(wf.Object, "status", "startedAt")
	if found && startedAtStr != "" {
		t, err := time.Parse(time.RFC3339, startedAtStr)
		if err == nil {
			result.StartedAt = &metav1.Time{Time: t}
		}
	}

	// Get finishedAt
	finishedAtStr, found, _ := unstructured.NestedString(wf.Object, "status", "finishedAt")
	if found && finishedAtStr != "" {
		t, err := time.Parse(time.RFC3339, finishedAtStr)
		if err == nil {
			result.FinishedAt = &metav1.Time{Time: t}
		}
	}

	return result, nil
}

// DeleteWorkflow deletes an Argo Workflow
func (m *Manager) DeleteWorkflow(ctx context.Context, workflowName string) error {
	logger := log.FromContext(ctx)

	wf := &unstructured.Unstructured{}
	wf.SetGroupVersionKind(workflowGVK)
	wf.SetName(workflowName)
	wf.SetNamespace(m.Namespace)

	if err := m.Delete(ctx, wf); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	logger.Info("Deleted Argo Workflow", "name", workflowName)
	return nil
}

// IsTerminal returns true if the workflow phase is a terminal state
func IsTerminal(phase string) bool {
	switch phase {
	case "Succeeded", "Failed", "Error":
		return true
	default:
		return false
	}
}

// IsSucceeded returns true if the workflow succeeded
func IsSucceeded(phase string) bool {
	return phase == "Succeeded"
}
