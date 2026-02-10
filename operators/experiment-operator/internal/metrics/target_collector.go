package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// MonitoringEndpoint describes a discovered Prometheus-compatible service on a target cluster.
type MonitoringEndpoint struct {
	Service   string
	Namespace string
	Port      int
}

// namespacesToSearch is the ordered list of namespaces to look for monitoring services.
var namespacesToSearch = []string{
	"monitoring",
	"observability",
	"kube-prometheus-stack",
	"victoria-metrics",
	"mimir",
	"default",
}

// DiscoverMonitoringServices finds Prometheus-compatible monitoring services on a target cluster.
// It searches well-known namespaces (plus the experiment name) for services matching common
// monitoring stack naming patterns, then probes each candidate to verify it serves the Prometheus API.
func DiscoverMonitoringServices(ctx context.Context, kubeconfig []byte, experimentName string) ([]MonitoringEndpoint, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}

	// Build namespace search order: experiment name first, then well-known namespaces.
	namespaces := append([]string{experimentName}, namespacesToSearch...)

	var endpoints []MonitoringEndpoint

	for _, ns := range namespaces {
		svcs, err := clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue // namespace may not exist
		}

		for _, svc := range svcs.Items {
			if ep, ok := matchMonitoringService(svc); ok {
				endpoints = append(endpoints, ep)
			}
		}
	}

	// Fallback: search all namespaces if targeted search found nothing.
	if len(endpoints) == 0 {
		allSvcs, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, svc := range allSvcs.Items {
				if ep, ok := matchMonitoringService(svc); ok {
					endpoints = append(endpoints, ep)
				}
			}
		}
	}

	if len(endpoints) == 0 {
		return nil, nil
	}

	// Probe each candidate to verify it serves the Prometheus API.
	restClient := clientset.CoreV1().RESTClient()

	var verified []MonitoringEndpoint
	for _, ep := range endpoints {
		if probeEndpoint(ctx, restClient, ep) {
			verified = append(verified, ep)
		}
	}

	if len(verified) == 0 {
		return endpoints, nil // return unverified as fallback
	}
	return verified, nil
}

// matchMonitoringService checks if a service matches known monitoring service name patterns.
func matchMonitoringService(svc corev1.Service) (MonitoringEndpoint, bool) {
	name := strings.ToLower(svc.Name)

	// Prometheus (excluding operator, alertmanager, node-exporter)
	if strings.Contains(name, "prometheus") &&
		!strings.Contains(name, "operator") &&
		!strings.Contains(name, "alertmanager") &&
		!strings.Contains(name, "node-exporter") {
		return MonitoringEndpoint{
			Service:   svc.Name,
			Namespace: svc.Namespace,
			Port:      findPort(svc, 9090),
		}, true
	}

	// VictoriaMetrics
	if strings.Contains(name, "victoria-metrics") ||
		strings.Contains(name, "vmsingle") ||
		strings.Contains(name, "vmselect") {
		return MonitoringEndpoint{
			Service:   svc.Name,
			Namespace: svc.Namespace,
			Port:      findPort(svc, 8428),
		}, true
	}

	// Mimir
	if strings.Contains(name, "mimir") {
		return MonitoringEndpoint{
			Service:   svc.Name,
			Namespace: svc.Namespace,
			Port:      findPort(svc, 8080),
		}, true
	}

	return MonitoringEndpoint{}, false
}

// findPort returns the given default port if the service has it, or the first service port.
func findPort(svc corev1.Service, defaultPort int) int {
	for _, p := range svc.Spec.Ports {
		if int(p.Port) == defaultPort {
			return defaultPort
		}
	}
	// Fall back to first port if default not found
	if len(svc.Spec.Ports) > 0 {
		return int(svc.Spec.Ports[0].Port)
	}
	return defaultPort
}

// probeEndpoint checks if a monitoring endpoint serves the Prometheus API by hitting /api/v1/status/buildinfo.
func probeEndpoint(ctx context.Context, restClient rest.Interface, ep MonitoringEndpoint) bool {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := restClient.Get().
		Namespace(ep.Namespace).
		Resource("services").
		Name(fmt.Sprintf("%s:%d", ep.Service, ep.Port)).
		SubResource("proxy", "api", "v1", "status", "buildinfo").
		Do(probeCtx)

	if result.Error() != nil {
		return false
	}

	raw, err := result.Raw()
	if err != nil {
		return false
	}

	// Check that the response looks like a Prometheus API response.
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return false
	}
	return resp.Status == "success"
}

// CollectMetricsFromTarget tries each discovered monitoring endpoint and collects metrics
// using default target queries that work on any Kubernetes cluster (no experiment label needed).
func CollectMetricsFromTarget(ctx context.Context, kubeconfig []byte, endpoints []MonitoringEndpoint, exp *experimentsv1alpha1.Experiment) (*MetricsResult, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no monitoring endpoints provided")
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}
	restClient := clientset.CoreV1().RESTClient()

	start := exp.CreationTimestamp.Time
	var end time.Time
	if exp.Status.CompletedAt != nil {
		end = exp.Status.CompletedAt.Time
	} else {
		end = time.Now()
	}

	duration := end.Sub(start)
	if duration < 30*time.Second {
		return nil, fmt.Errorf("experiment duration too short for metrics: %s", duration)
	}

	stepStr := selectStep(duration)
	stepSec := 60
	fmt.Sscanf(strings.TrimSuffix(stepStr, "s"), "%d", &stepSec)

	queries := defaultTargetQueries()

	// Try each endpoint until one works
	for _, ep := range endpoints {
		result := &MetricsResult{
			CollectedAt: time.Now().UTC(),
			Source:      fmt.Sprintf("target:%s/%s", ep.Namespace, ep.Service),
			TimeRange: TimeRange{
				Start:    start,
				End:      end,
				Duration: duration.String(),
				StepSec:  stepSec,
			},
			Queries: make(map[string]QueryResult),
		}

		anyData := false
		anySuccess := false

		for _, mq := range queries {
			qr := QueryResult{
				Query:       mq.Query,
				Type:        mq.Type,
				Unit:        mq.Unit,
				Description: mq.Description,
			}

			var data []DataPoint
			var queryErr error

			switch mq.Type {
			case "range":
				data, queryErr = queryRangeViaProxy(ctx, restClient, ep, mq.Query, start, end, stepStr)
			case "instant":
				data, queryErr = queryInstantViaProxy(ctx, restClient, ep, mq.Query, end)
			}

			if queryErr != nil {
				qr.Error = queryErr.Error()
			} else {
				anySuccess = true
				qr.Data = data
				if len(data) > 0 {
					anyData = true
				}
			}

			result.Queries[mq.Name] = qr
		}

		// If at least one query succeeded, return this result
		if anySuccess {
			if !anyData {
				// Queries succeeded but returned empty — monitoring exists but has no data yet.
				// Still return the result so the source is recorded.
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("all monitoring endpoints failed")
}

// queryRangeViaProxy executes a range query through the K8s API server proxy.
func queryRangeViaProxy(ctx context.Context, restClient rest.Interface, ep MonitoringEndpoint, query string, start, end time.Time, step string) ([]DataPoint, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	raw, err := restClient.Get().
		Namespace(ep.Namespace).
		Resource("services").
		Name(fmt.Sprintf("%s:%d", ep.Service, ep.Port)).
		SubResource("proxy", "api", "v1", "query_range").
		Param("query", query).
		Param("start", fmt.Sprintf("%d", start.Unix())).
		Param("end", fmt.Sprintf("%d", end.Unix())).
		Param("step", step).
		Do(queryCtx).
		Raw()
	if err != nil {
		return nil, fmt.Errorf("proxy query_range: %w", err)
	}

	var pr promResponse
	if err := json.Unmarshal(raw, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal range response: %w", err)
	}

	if pr.Status != "success" {
		return nil, fmt.Errorf("prometheus returned status: %s", pr.Status)
	}

	return flattenRangeResult(pr.Data.Result)
}

// queryInstantViaProxy executes an instant query through the K8s API server proxy.
func queryInstantViaProxy(ctx context.Context, restClient rest.Interface, ep MonitoringEndpoint, query string, evalTime time.Time) ([]DataPoint, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	raw, err := restClient.Get().
		Namespace(ep.Namespace).
		Resource("services").
		Name(fmt.Sprintf("%s:%d", ep.Service, ep.Port)).
		SubResource("proxy", "api", "v1", "query").
		Param("query", query).
		Param("time", fmt.Sprintf("%d", evalTime.Unix())).
		Do(queryCtx).
		Raw()
	if err != nil {
		return nil, fmt.Errorf("proxy query: %w", err)
	}

	var pr promResponse
	if err := json.Unmarshal(raw, &pr); err != nil {
		return nil, fmt.Errorf("unmarshal instant response: %w", err)
	}

	if pr.Status != "success" {
		return nil, fmt.Errorf("prometheus returned status: %s", pr.Status)
	}

	return flattenInstantResult(pr.Data.Result)
}

// defaultTargetQueries returns PromQL queries that work on any Kubernetes cluster
// without requiring an experiment-specific label. They filter out system namespaces
// to focus on workload metrics.
func defaultTargetQueries() []experimentsv1alpha1.MetricsQuery {
	sysNS := `kube-system|gke-managed-system|gmp-system|gmp-public|kube-node-lease|kube-public`
	return []experimentsv1alpha1.MetricsQuery{
		{
			Name:        "cpu_usage",
			Query:       fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace!~"%s"}[1m])) by (namespace)`, sysNS),
			Type:        "range",
			Unit:        "cores",
			Description: "CPU usage by namespace",
		},
		{
			Name:        "memory_usage",
			Query:       fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace!~"%s"}) by (namespace)`, sysNS),
			Type:        "range",
			Unit:        "bytes",
			Description: "Memory working set by namespace",
		},
		{
			Name:        "cpu_total",
			Query:       fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace!~"%s"}[1m]))`, sysNS),
			Type:        "range",
			Unit:        "cores",
			Description: "Total CPU usage",
		},
		{
			Name:        "memory_total",
			Query:       fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace!~"%s"})`, sysNS),
			Type:        "range",
			Unit:        "bytes",
			Description: "Total memory working set",
		},
	}
}

// AllQueriesEmpty returns true if a MetricsResult has no data points in any query.
func AllQueriesEmpty(mr *MetricsResult) bool {
	if mr == nil {
		return true
	}
	for _, qr := range mr.Queries {
		if len(qr.Data) > 0 {
			return false
		}
	}
	return true
}
