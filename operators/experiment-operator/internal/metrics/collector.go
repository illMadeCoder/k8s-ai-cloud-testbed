package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
)

// MetricsResult is the top-level container for all collected metrics.
type MetricsResult struct {
	CollectedAt time.Time              `json:"collectedAt"`
	Source      string                 `json:"source,omitempty"`
	TimeRange   TimeRange              `json:"timeRange"`
	Queries     map[string]QueryResult `json:"queries"`
}

// TimeRange captures the experiment time window used for queries.
type TimeRange struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Duration string    `json:"duration"`
	StepSec  int       `json:"stepSeconds"`
}

// QueryResult holds the result of a single PromQL query, flattened for Vega-Lite.
type QueryResult struct {
	Query       string      `json:"query"`
	Type        string      `json:"type"`
	Unit        string      `json:"unit,omitempty"`
	Description string      `json:"description,omitempty"`
	Error       string      `json:"error,omitempty"`
	Data        []DataPoint `json:"data,omitempty"`
}

// DataPoint is a single row in the flat tabular output.
// Instant queries: one DataPoint per series (one label-set).
// Range queries: one DataPoint per (series, timestamp) pair.
type DataPoint struct {
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
}

// promResponse models the Prometheus HTTP API response.
type promResponse struct {
	Status string   `json:"status"`
	Data   promData `json:"data"`
}

type promData struct {
	ResultType string       `json:"resultType"`
	Result     []promResult `json:"result"`
}

type promResult struct {
	Metric map[string]string    `json:"metric"`
	Value  [2]json.RawMessage   `json:"value,omitempty"`
	Values [][2]json.RawMessage `json:"values,omitempty"`
}

// AnalyzerConfigJSON captures the requested analysis sections in the summary JSON.
// The analyzer reads this to decide which passes to run and which sections to keep.
type AnalyzerConfigJSON struct {
	Sections []string `json:"sections"`
}

// ExperimentSummary is the top-level results object stored in S3.
type ExperimentSummary struct {
	Name           string              `json:"name"`
	Namespace      string              `json:"namespace"`
	Description    string              `json:"description"`
	CreatedAt      time.Time           `json:"createdAt"`
	CompletedAt    time.Time           `json:"completedAt"`
	DurationSec    float64             `json:"durationSeconds"`
	Phase          string              `json:"phase"`
	Tags           []string            `json:"tags,omitempty"`
	Hypothesis     *HypothesisContext  `json:"hypothesis,omitempty"`
	AnalyzerConfig *AnalyzerConfigJSON `json:"analyzerConfig,omitempty"`
	Targets        []TargetSummary     `json:"targets"`
	Workflow       WorkflowSummary     `json:"workflow"`
	Metrics        *MetricsResult      `json:"metrics,omitempty"`
	CostEstimate   *CostEstimate       `json:"costEstimate,omitempty"`
	Analysis       *AnalysisResult     `json:"analysis,omitempty"`
}

// HypothesisContext captures the experiment's hypothesis and success criteria for AI analysis.
type HypothesisContext struct {
	Claim           string                    `json:"claim,omitempty"`
	Questions       []string                  `json:"questions,omitempty"`
	Focus           []string                  `json:"focus,omitempty"`
	SuccessCriteria []SuccessCriterionSummary `json:"successCriteria,omitempty"`
	MachineVerdict  string                    `json:"machineVerdict,omitempty"`
}

// SuccessCriterionSummary captures a success criterion and its evaluation result.
type SuccessCriterionSummary struct {
	Metric      string `json:"metric"`
	Operator    string `json:"operator"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Passed      *bool  `json:"passed,omitempty"`
	ActualValue string `json:"actualValue,omitempty"`
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
	TotalUSD    float64            `json:"totalUSD"`
	DurationHrs float64           `json:"durationHours"`
	PerTarget   map[string]float64 `json:"perTarget,omitempty"`
	Note        string             `json:"note"`
}

// AnalysisResult holds AI-generated analysis of experiment results.
type AnalysisResult struct {
	// Backward-compatible fields
	Summary        string            `json:"summary"`
	MetricInsights map[string]string `json:"metricInsights"`
	GeneratedAt    time.Time         `json:"generatedAt"`
	Model          string            `json:"model"`

	// Hypothesis verdict — structured enum for at-a-glance display in stats row.
	// Values: "validated", "invalidated", "insufficient"
	HypothesisVerdict string `json:"hypothesisVerdict,omitempty"`

	// Structured analysis sections
	Abstract            string              `json:"abstract,omitempty"`
	CapabilitiesMatrix  *CapabilitiesMatrix `json:"capabilitiesMatrix,omitempty"`
	Body                *AnalysisBody       `json:"body,omitempty"`
	Feedback            *AnalysisFeedback   `json:"feedback,omitempty"`
	ArchitectureDiagram string              `json:"architectureDiagram,omitempty"`
}

// CapabilitiesMatrix is a feature comparison table for comparison experiments.
type CapabilitiesMatrix struct {
	Technologies []string               `json:"technologies"`
	Categories   []CapabilitiesCategory `json:"categories"`
	Summary      string                 `json:"summary,omitempty"`
}

// CapabilitiesCategory groups related capabilities for comparison.
type CapabilitiesCategory struct {
	Name         string            `json:"name"`
	Capabilities []CapabilityEntry `json:"capabilities"`
}

// CapabilityEntry is a single capability with per-technology values.
type CapabilityEntry struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

// AnalysisBody contains the rich narrative body as an ordered array of typed blocks.
// The analyzer generates blocks; the operator only needs JSON round-tripping.
type AnalysisBody struct {
	Blocks []BodyBlock `json:"blocks,omitempty"`
}

// BodyBlock is a discriminated union of content blocks. The operator never generates
// blocks (the analyzer does), so we use a generic map for JSON round-tripping.
type BodyBlock map[string]interface{}

// AnalysisFeedback provides actionable recommendations and experiment design improvements.
type AnalysisFeedback struct {
	Recommendations  []string `json:"recommendations,omitempty"`
	ExperimentDesign []string `json:"experimentDesign,omitempty"`
}

// GCP on-demand hourly rates (USD) for common machine types.
var gcpHourlyRates = map[string]float64{
	"e2-medium":      0.0335,
	"e2-standard-2":  0.0670,
	"e2-standard-4":  0.1340,
	"e2-standard-8":  0.2680,
	"n1-standard-1":  0.0475,
	"n1-standard-2":  0.0950,
	"n1-standard-4":  0.1900,
	"n2-standard-2":  0.0971,
	"n2-standard-4":  0.1942,
	"n2-standard-8":  0.3884,
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
		Tags:        exp.Spec.Tags,
	}

	// Hypothesis context — pass through to summary for AI analyzer
	if exp.Spec.Hypothesis != nil {
		s.Hypothesis = &HypothesisContext{
			Claim:     exp.Spec.Hypothesis.Claim,
			Questions: exp.Spec.Hypothesis.Questions,
			Focus:     exp.Spec.Hypothesis.Focus,
		}
		for _, sc := range exp.Spec.Hypothesis.SuccessCriteria {
			s.Hypothesis.SuccessCriteria = append(s.Hypothesis.SuccessCriteria, SuccessCriterionSummary{
				Metric:      sc.Metric,
				Operator:    sc.Operator,
				Value:       sc.Value,
				Description: sc.Description,
			})
		}
	}

	// Analyzer config — pass requested sections through to summary for analyzer
	if exp.Spec.AnalyzerConfig != nil && len(exp.Spec.AnalyzerConfig.Sections) > 0 {
		s.AnalyzerConfig = &AnalyzerConfigJSON{
			Sections: exp.Spec.AnalyzerConfig.Sections,
		}
	}

	if exp.Status.CompletedAt != nil {
		s.CompletedAt = exp.Status.CompletedAt.Time
		s.DurationSec = exp.Status.CompletedAt.Time.Sub(exp.CreationTimestamp.Time).Seconds()
	}

	// Targets — read effective values from status (single source of truth)
	for i, target := range exp.Spec.Targets {
		ts := TargetSummary{
			Name:        target.Name,
			ClusterType: target.Cluster.Type,
		}
		if i < len(exp.Status.Targets) {
			ts.ClusterName = exp.Status.Targets[i].ClusterName
			ts.MachineType = exp.Status.Targets[i].MachineType
			ts.NodeCount = exp.Status.Targets[i].NodeCount
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

// defaultQueries returns the built-in metrics queries used when spec.metrics is empty.
func defaultQueries() []experimentsv1alpha1.MetricsQuery {
	sysNS := `kube-system|gke-managed-system|gmp-system|gmp-public|kube-node-lease|kube-public|observability`
	infraPods := `alloy-.*|ts-vm-hub-.*|prometheus-.*|alertmanager-.*|grafana-.*|kube-state-metrics-.*|node-exporter-.*|kube-prometheus-stack-.*|tailscale-operator-.*|operator-.*`
	return []experimentsv1alpha1.MetricsQuery{
		{Name: "cpu_by_pod", Query: fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{experiment="$EXPERIMENT",namespace!~"%s",pod!~"%s",container!="POD",container!=""}[1m])) by (pod)`, sysNS, infraPods), Type: "range", Unit: "cores", Description: "CPU usage by pod"},
		{Name: "memory_by_pod", Query: fmt.Sprintf(`sum(container_memory_working_set_bytes{experiment="$EXPERIMENT",namespace!~"%s",pod!~"%s",container!="POD",container!=""}) by (pod)`, sysNS, infraPods), Type: "range", Unit: "bytes", Description: "Memory working set by pod"},
		{Name: "cpu_total", Query: fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{experiment="$EXPERIMENT",namespace!~"%s",pod!~"%s"}[1m]))`, sysNS, infraPods), Type: "range", Unit: "cores", Description: "Total CPU usage"},
		{Name: "memory_total", Query: fmt.Sprintf(`sum(container_memory_working_set_bytes{experiment="$EXPERIMENT",namespace!~"%s",pod!~"%s"})`, sysNS, infraPods), Type: "range", Unit: "bytes", Description: "Total memory working set"},
	}
}

// substituteVars replaces $EXPERIMENT, $NAMESPACE, and $DURATION placeholders in a query.
func substituteVars(query string, vars map[string]string) string {
	for k, v := range vars {
		query = strings.ReplaceAll(query, k, v)
	}
	return query
}

// promDuration converts a Go duration to a Prometheus-compatible duration string (e.g., "15m", "2h30m").
func promDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Second {
		return "0s"
	}

	var parts []string
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	if len(parts) == 0 {
		return "0s"
	}
	return strings.Join(parts, "")
}

// CollectMetricsSnapshot queries VictoriaMetrics for metrics defined in the experiment
// spec (or defaults) and returns structured, chart-ready results.
func CollectMetricsSnapshot(ctx context.Context, metricsURL string, exp *experimentsv1alpha1.Experiment) (*MetricsResult, error) {
	if metricsURL == "" {
		return nil, nil
	}

	start := exp.CreationTimestamp.Time
	// Always use time.Now() as end time. This function is called while resources
	// are still alive (cleanup happens after it returns), so Now() captures data
	// that Alloy wrote after workflow completion but before results collection.
	end := time.Now()

	duration := end.Sub(start)

	// Don't query if duration is trivially short
	if duration < 30*time.Second {
		return nil, nil
	}

	stepStr := selectStep(duration)
	stepSec, _ := strconv.Atoi(strings.TrimSuffix(stepStr, "s"))

	// Pick queries: spec.metrics or defaults
	queries := exp.Spec.Metrics
	if len(queries) == 0 {
		queries = defaultQueries()
	}

	// Build substitution variables
	vars := map[string]string{
		"$EXPERIMENT": exp.Name,
		"$NAMESPACE":  exp.Namespace,
		"$DURATION":   promDuration(duration),
	}

	result := &MetricsResult{
		CollectedAt: time.Now().UTC(),
		TimeRange: TimeRange{
			Start:    start,
			End:      end,
			Duration: duration.String(),
			StepSec:  stepSec,
		},
		Queries: make(map[string]QueryResult),
	}

	for _, mq := range queries {
		resolvedQuery := substituteVars(mq.Query, vars)
		queryType := mq.Type
		if queryType == "" {
			queryType = "instant"
		}

		qr := QueryResult{
			Query:       resolvedQuery,
			Type:        queryType,
			Unit:        mq.Unit,
			Description: mq.Description,
		}

		var data []DataPoint
		var err error

		switch queryType {
		case "instant":
			data, err = queryMetricsInstant(ctx, metricsURL, resolvedQuery, end)
		case "range":
			data, err = queryMetricsRange(ctx, metricsURL, resolvedQuery, start, end)
		default:
			err = fmt.Errorf("unknown query type: %s", queryType)
		}

		if err != nil {
			qr.Error = err.Error()
		} else {
			qr.Data = data
		}

		result.Queries[mq.Name] = qr
	}

	return result, nil
}

// queryMetricsInstant executes a Prometheus-compatible instant query.
func queryMetricsInstant(ctx context.Context, metricsURL, query string, evalTime time.Time) ([]DataPoint, error) {
	u, err := url.Parse(metricsURL + "/api/v1/query")
	if err != nil {
		return nil, fmt.Errorf("parse metrics URL: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("time", strconv.FormatInt(evalTime.Unix(), 10))
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

	var pr promResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal metrics response: %w", err)
	}

	return flattenInstantResult(pr.Data.Result)
}

// queryMetricsRange executes a Prometheus-compatible range query and returns flat DataPoints.
func queryMetricsRange(ctx context.Context, metricsURL, query string, start, end time.Time) ([]DataPoint, error) {
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

	var pr promResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal metrics response: %w", err)
	}

	return flattenRangeResult(pr.Data.Result)
}

// flattenInstantResult converts Prometheus instant query results to flat DataPoints.
func flattenInstantResult(results []promResult) ([]DataPoint, error) {
	var points []DataPoint
	for _, r := range results {
		if len(r.Value) < 2 {
			continue
		}
		ts, val, err := parseTimestampValue(r.Value[0], r.Value[1])
		if err != nil {
			continue // skip NaN/Inf
		}
		points = append(points, DataPoint{
			Labels:    copyLabels(r.Metric),
			Timestamp: ts,
			Value:     val,
		})
	}
	return points, nil
}

// flattenRangeResult converts Prometheus range query results to flat DataPoints.
func flattenRangeResult(results []promResult) ([]DataPoint, error) {
	var points []DataPoint
	for _, r := range results {
		labels := copyLabels(r.Metric)
		for _, pair := range r.Values {
			ts, val, err := parseTimestampValue(pair[0], pair[1])
			if err != nil {
				continue // skip NaN/Inf
			}
			points = append(points, DataPoint{
				Labels:    labels,
				Timestamp: ts,
				Value:     val,
			})
		}
	}
	return points, nil
}

// parseTimestampValue parses a Prometheus [timestamp, value] pair.
// Returns error for NaN/Inf values which should be skipped.
func parseTimestampValue(rawTS, rawVal json.RawMessage) (time.Time, float64, error) {
	var tsFloat float64
	if err := json.Unmarshal(rawTS, &tsFloat); err != nil {
		return time.Time{}, 0, fmt.Errorf("parse timestamp: %w", err)
	}

	var valStr string
	if err := json.Unmarshal(rawVal, &valStr); err != nil {
		return time.Time{}, 0, fmt.Errorf("parse value: %w", err)
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("parse float: %w", err)
	}
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return time.Time{}, 0, fmt.Errorf("non-finite value: %s", valStr)
	}

	sec := int64(tsFloat)
	nsec := int64((tsFloat - float64(sec)) * 1e9)
	ts := time.Unix(sec, nsec).UTC()

	return ts, val, nil
}

// copyLabels returns a copy of the label map, or nil if empty.
func copyLabels(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
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

// DefaultAnalyzerSections is the default set of analysis sections used when
// publish is true and analyzerConfig is nil.
var DefaultAnalyzerSections = []string{
	"abstract", "targetAnalysis", "performanceAnalysis", "metricInsights",
	"body", "feedback", "architectureDiagram",
}

// EvaluateSuccessCriteria evaluates success criteria against collected metrics.
// Returns "" if no criteria, "insufficient" if metrics are missing/errored,
// "validated" if all pass, "invalidated" if any fail.
func EvaluateSuccessCriteria(summary *ExperimentSummary) string {
	if summary.Hypothesis == nil || len(summary.Hypothesis.SuccessCriteria) == 0 {
		return ""
	}
	if summary.Metrics == nil || len(summary.Metrics.Queries) == 0 {
		return "insufficient"
	}

	allPassed := true
	anyEvaluated := false

	for i, sc := range summary.Hypothesis.SuccessCriteria {
		qr, ok := summary.Metrics.Queries[sc.Metric]
		if !ok || qr.Error != "" || len(qr.Data) == 0 {
			summary.Hypothesis.SuccessCriteria[i].Passed = nil
			allPassed = false
			continue
		}

		// Calculate the value to compare: for instant queries use the single value,
		// for range queries use the mean.
		var actual float64
		if qr.Type == "instant" && len(qr.Data) == 1 {
			actual = qr.Data[0].Value
		} else {
			var sum float64
			for _, dp := range qr.Data {
				sum += dp.Value
			}
			actual = sum / float64(len(qr.Data))
		}

		threshold, err := strconv.ParseFloat(sc.Value, 64)
		if err != nil {
			summary.Hypothesis.SuccessCriteria[i].Passed = nil
			allPassed = false
			continue
		}

		var passed bool
		switch sc.Operator {
		case "lt":
			passed = actual < threshold
		case "lte":
			passed = actual <= threshold
		case "gt":
			passed = actual > threshold
		case "gte":
			passed = actual >= threshold
		default:
			summary.Hypothesis.SuccessCriteria[i].Passed = nil
			allPassed = false
			continue
		}

		anyEvaluated = true
		summary.Hypothesis.SuccessCriteria[i].Passed = &passed
		summary.Hypothesis.SuccessCriteria[i].ActualValue = strconv.FormatFloat(actual, 'f', -1, 64)
		if !passed {
			allPassed = false
		}
	}

	if !anyEvaluated {
		return "insufficient"
	}
	if allPassed {
		return "validated"
	}
	return "invalidated"
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
