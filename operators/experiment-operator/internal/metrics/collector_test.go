package metrics

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSubstituteVars(t *testing.T) {
	vars := map[string]string{
		"$EXPERIMENT": "tsdb-comparison-abc",
		"$NAMESPACE":  "experiments",
		"$DURATION":   "30m",
	}

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "all three vars",
			query: `sum(rate(cpu{experiment="$EXPERIMENT", namespace="$NAMESPACE"}[$DURATION]))`,
			want:  `sum(rate(cpu{experiment="tsdb-comparison-abc", namespace="experiments"}[30m]))`,
		},
		{
			name:  "no vars",
			query: `up`,
			want:  `up`,
		},
		{
			name:  "repeated var",
			query: `$EXPERIMENT-$EXPERIMENT`,
			want:  `tsdb-comparison-abc-tsdb-comparison-abc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := substituteVars(tt.query, vars)
			if got != tt.want {
				t.Errorf("substituteVars() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0s"},
		{"sub-second", 500 * time.Millisecond, "0s"},
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes only", 15 * time.Minute, "15m"},
		{"hours only", 2 * time.Hour, "2h"},
		{"hours and minutes", 2*time.Hour + 30*time.Minute, "2h30m"},
		{"all components", 1*time.Hour + 5*time.Minute + 10*time.Second, "1h5m10s"},
		{"negative duration", -15 * time.Minute, "15m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := promDuration(tt.d)
			if got != tt.want {
				t.Errorf("promDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFlattenRangeResult(t *testing.T) {
	// Mock Prometheus range query response
	rawJSON := `{
		"status": "success",
		"data": {
			"resultType": "matrix",
			"result": [
				{
					"metric": {"pod": "prometheus-0", "__name__": "cpu"},
					"values": [
						[1707350400, "0.12"],
						[1707350460, "0.15"]
					]
				},
				{
					"metric": {"pod": "victoria-metrics-0"},
					"values": [
						[1707350400, "0.05"],
						[1707350460, "0.06"]
					]
				}
			]
		}
	}`

	var pr promResponse
	if err := json.Unmarshal([]byte(rawJSON), &pr); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}

	points, err := flattenRangeResult(pr.Data.Result)
	if err != nil {
		t.Fatalf("flattenRangeResult() error: %v", err)
	}

	if len(points) != 4 {
		t.Fatalf("expected 4 data points, got %d", len(points))
	}

	// First series, first point
	if points[0].Labels["pod"] != "prometheus-0" {
		t.Errorf("point[0] pod = %q, want %q", points[0].Labels["pod"], "prometheus-0")
	}
	if points[0].Value != 0.12 {
		t.Errorf("point[0] value = %f, want 0.12", points[0].Value)
	}

	// Second series, second point
	if points[3].Labels["pod"] != "victoria-metrics-0" {
		t.Errorf("point[3] pod = %q, want %q", points[3].Labels["pod"], "victoria-metrics-0")
	}
	if points[3].Value != 0.06 {
		t.Errorf("point[3] value = %f, want 0.06", points[3].Value)
	}

	// Check timestamps are parsed correctly
	expectedTS := time.Unix(1707350400, 0).UTC()
	if !points[0].Timestamp.Equal(expectedTS) {
		t.Errorf("point[0] timestamp = %v, want %v", points[0].Timestamp, expectedTS)
	}
}

func TestFlattenInstantResult(t *testing.T) {
	rawJSON := `{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [
				{
					"metric": {"instance": "node1"},
					"value": [1707350400, "524288000"]
				},
				{
					"metric": {"instance": "node2"},
					"value": [1707350400, "157286400"]
				}
			]
		}
	}`

	var pr promResponse
	if err := json.Unmarshal([]byte(rawJSON), &pr); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}

	points, err := flattenInstantResult(pr.Data.Result)
	if err != nil {
		t.Fatalf("flattenInstantResult() error: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("expected 2 data points, got %d", len(points))
	}

	if points[0].Labels["instance"] != "node1" {
		t.Errorf("point[0] instance = %q, want %q", points[0].Labels["instance"], "node1")
	}
	if points[0].Value != 524288000 {
		t.Errorf("point[0] value = %f, want 524288000", points[0].Value)
	}
	if points[1].Value != 157286400 {
		t.Errorf("point[1] value = %f, want 157286400", points[1].Value)
	}
}

func TestNaNInfHandling(t *testing.T) {
	rawJSON := `{
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [
				{
					"metric": {"name": "valid"},
					"value": [1707350400, "42.5"]
				},
				{
					"metric": {"name": "nan"},
					"value": [1707350400, "NaN"]
				},
				{
					"metric": {"name": "inf"},
					"value": [1707350400, "+Inf"]
				},
				{
					"metric": {"name": "neg_inf"},
					"value": [1707350400, "-Inf"]
				}
			]
		}
	}`

	var pr promResponse
	if err := json.Unmarshal([]byte(rawJSON), &pr); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}

	points, err := flattenInstantResult(pr.Data.Result)
	if err != nil {
		t.Fatalf("flattenInstantResult() error: %v", err)
	}

	// Only the valid point should survive
	if len(points) != 1 {
		t.Fatalf("expected 1 data point (NaN/Inf skipped), got %d", len(points))
	}
	if points[0].Labels["name"] != "valid" {
		t.Errorf("surviving point name = %q, want %q", points[0].Labels["name"], "valid")
	}
	if points[0].Value != 42.5 {
		t.Errorf("surviving point value = %f, want 42.5", points[0].Value)
	}
}

func TestDefaultQueries(t *testing.T) {
	queries := defaultQueries()

	if len(queries) != 2 {
		t.Fatalf("expected 2 default queries, got %d", len(queries))
	}

	if queries[0].Name != "cpu_usage" {
		t.Errorf("first default query name = %q, want %q", queries[0].Name, "cpu_usage")
	}
	if queries[0].Type != "range" {
		t.Errorf("first default query type = %q, want %q", queries[0].Type, "range")
	}
	if queries[1].Name != "memory_usage" {
		t.Errorf("second default query name = %q, want %q", queries[1].Name, "memory_usage")
	}
}

func TestDefaultQueriesHaveExperimentVar(t *testing.T) {
	for _, q := range defaultQueries() {
		vars := map[string]string{
			"$EXPERIMENT": "my-exp",
			"$NAMESPACE":  "experiments",
			"$DURATION":   "15m",
		}
		resolved := substituteVars(q.Query, vars)
		if resolved == q.Query {
			t.Errorf("default query %q was not modified by substituteVars — missing $EXPERIMENT?", q.Name)
		}
		// Ensure $EXPERIMENT was replaced
		if contains(resolved, "$EXPERIMENT") {
			t.Errorf("default query %q still contains $EXPERIMENT after substitution", q.Name)
		}
	}
}

func TestCopyLabelsNil(t *testing.T) {
	result := copyLabels(nil)
	if result != nil {
		t.Errorf("copyLabels(nil) = %v, want nil", result)
	}
}

func TestCopyLabelsEmpty(t *testing.T) {
	result := copyLabels(map[string]string{})
	if result != nil {
		t.Errorf("copyLabels(empty) = %v, want nil", result)
	}
}

func TestCopyLabelsIsolation(t *testing.T) {
	original := map[string]string{"pod": "test-pod", "namespace": "default"}
	copied := copyLabels(original)

	// Mutate original
	original["pod"] = "mutated"

	if copied["pod"] != "test-pod" {
		t.Errorf("copyLabels did not create independent copy: got %q", copied["pod"])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
