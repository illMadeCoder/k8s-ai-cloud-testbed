package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// ExperimentSummary is the top-level results object stored in S3.
type ExperimentSummary struct {
	Name         string          `json:"name"`
	Namespace    string          `json:"namespace"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"createdAt"`
	CompletedAt  time.Time       `json:"completedAt"`
	DurationSec  float64         `json:"durationSeconds"`
	Phase        string          `json:"phase"`
	Targets      []TargetSummary `json:"targets"`
	Workflow     WorkflowSummary `json:"workflow"`
	MimirMetrics map[string]any  `json:"mimirMetrics,omitempty"`
	CostEstimate *CostEstimate   `json:"costEstimate,omitempty"`
}

// TargetSummary captures per-target metadata.
type TargetSummary struct {
	Name        string   `json:"name"`
	ClusterName string   `json:"clusterName,omitempty"`
	ClusterType string   `json:"clusterType"`
	MachineType string   `json:"machineType,omitempty"`
	NodeCount   int      `json:"nodeCount,omitempty"`
	Components  []string `json:"components,omitempty"`
}

// WorkflowSummary captures workflow execution metadata.
type WorkflowSummary struct {
	Name       string     `json:"name"`
	Template   string     `json:"template"`
	Phase      string     `json:"phase"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}

// CostEstimate provides a rough GCP cost estimate for the experiment.
type CostEstimate struct {
	TotalUSD     float64           `json:"totalUSD"`
	DurationHrs  float64           `json:"durationHours"`
	PerTarget    map[string]float64 `json:"perTarget,omitempty"`
	Note         string            `json:"note"`
}

// GCP on-demand hourly rates (USD) for common machine types.
var gcpHourlyRates = map[string]float64{
	"e2-medium":    0.0335,
	"e2-standard-2": 0.0670,
	"e2-standard-4": 0.1340,
	"e2-standard-8": 0.2680,
	"n1-standard-1": 0.0475,
	"n1-standard-2": 0.0950,
	"n1-standard-4": 0.1900,
	"n2-standard-2": 0.0971,
	"n2-standard-4": 0.1942,
	"n2-standard-8": 0.3884,
}

// preemptibleDiscount is applied when preemptible/spot is enabled.
const preemptibleDiscount = 0.20

// CollectSummary builds an ExperimentSummary from an Experiment CR.
func CollectSummary(exp *experimentsv1alpha1.Experiment) *ExperimentSummary {
	s := &ExperimentSummary{
		Name:        exp.Name,
		Namespace:   exp.Namespace,
		Description: exp.Spec.Description,
		CreatedAt:   exp.CreationTimestamp.Time,
		Phase:       string(exp.Status.Phase),
	}

	if exp.Status.CompletedAt != nil {
		s.CompletedAt = exp.Status.CompletedAt.Time
		s.DurationSec = exp.Status.CompletedAt.Time.Sub(exp.CreationTimestamp.Time).Seconds()
	}

	// Targets
	for i, target := range exp.Spec.Targets {
		ts := TargetSummary{
			Name:        target.Name,
			ClusterType: target.Cluster.Type,
			MachineType: target.Cluster.MachineType,
			NodeCount:   target.Cluster.NodeCount,
		}
		if i < len(exp.Status.Targets) {
			ts.ClusterName = exp.Status.Targets[i].ClusterName
			ts.Components = exp.Status.Targets[i].Components
		}
		s.Targets = append(s.Targets, ts)
	}

	// Workflow
	s.Workflow = WorkflowSummary{
		Template: exp.Spec.Workflow.Template,
	}
	if ws := exp.Status.WorkflowStatus; ws != nil {
		s.Workflow.Name = ws.Name
		s.Workflow.Phase = ws.Phase
		if ws.StartedAt != nil {
			t := ws.StartedAt.Time
			s.Workflow.StartedAt = &t
		}
		if ws.FinishedAt != nil {
			t := ws.FinishedAt.Time
			s.Workflow.FinishedAt = &t
		}
	}

	return s
}

// CollectMetricsSnapshot queries VictoriaMetrics for key container metrics over the
// experiment lifetime and returns the raw Prometheus query_range response.
// The experimentName parameter is used to filter metrics by the experiment external label.
func CollectMetricsSnapshot(ctx context.Context, metricsURL string, experimentName string, exp *experimentsv1alpha1.Experiment) (map[string]any, error) {
	if metricsURL == "" {
		return nil, nil
	}

	start := exp.CreationTimestamp.Time
	var end time.Time
	if exp.Status.CompletedAt != nil {
		end = exp.Status.CompletedAt.Time
	} else {
		end = time.Now()
	}

	// Don't query if duration is trivially short
	if end.Sub(start) < 30*time.Second {
		return nil, nil
	}

	queries := map[string]string{
		"cpu":    fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{experiment=%q}[1m]))`, experimentName),
		"memory": fmt.Sprintf(`sum(container_memory_working_set_bytes{experiment=%q})`, experimentName),
	}

	result := make(map[string]any)
	for name, query := range queries {
		data, err := queryMetricsRange(ctx, metricsURL, query, start, end)
		if err != nil {
			// Non-fatal: include the error in the result
			result[name] = map[string]any{"error": err.Error()}
			continue
		}
		result[name] = data
	}

	return result, nil
}

// queryMetricsRange executes a Prometheus-compatible range query against VictoriaMetrics.
func queryMetricsRange(ctx context.Context, metricsURL, query string, start, end time.Time) (any, error) {
	step := selectStep(end.Sub(start))

	u, err := url.Parse(metricsURL + "/api/v1/query_range")
	if err != nil {
		return nil, fmt.Errorf("parse metrics URL: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("start", strconv.FormatInt(start.Unix(), 10))
	q.Set("end", strconv.FormatInt(end.Unix(), 10))
	q.Set("step", step)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("metrics query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read metrics response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metrics server returned %d: %s", resp.StatusCode, string(body))
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("unmarshal metrics response: %w", err)
	}
	return data, nil
}

// selectStep picks a reasonable step size for the query duration.
func selectStep(d time.Duration) string {
	switch {
	case d < 5*time.Minute:
		return "15s"
	case d < 1*time.Hour:
		return "60s"
	default:
		return "300s"
	}
}

// EstimateCost produces a rough GCP cost estimate from the experiment spec.
func EstimateCost(exp *experimentsv1alpha1.Experiment) *CostEstimate {
	if exp.Status.CompletedAt == nil {
		return nil
	}

	duration := exp.Status.CompletedAt.Time.Sub(exp.CreationTimestamp.Time)
	hours := duration.Hours()

	est := &CostEstimate{
		DurationHrs: hours,
		PerTarget:   make(map[string]float64),
		Note:        "Rough estimate based on on-demand GCE pricing; actual cost may differ.",
	}

	for _, target := range exp.Spec.Targets {
		if target.Cluster.Type != "gke" {
			continue
		}
		rate, ok := gcpHourlyRates[target.Cluster.MachineType]
		if !ok {
			rate = 0.10 // fallback
		}
		if target.Cluster.Preemptible {
			rate *= preemptibleDiscount
		}
		nodes := target.Cluster.NodeCount
		if nodes == 0 {
			nodes = 1
		}
		cost := rate * float64(nodes) * hours
		est.PerTarget[target.Name] = cost
		est.TotalUSD += cost
	}

	return est
}
