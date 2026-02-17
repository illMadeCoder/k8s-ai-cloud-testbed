package metrics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	experimentsv1alpha1 "github.com/illmadecoder/experiment-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	logger := log.FromContext(ctx)
	logger.Info("Matched monitoring service candidates", "count", len(endpoints))

	// Probe each candidate to verify it serves the Prometheus API.
	restClient := clientset.CoreV1().RESTClient()

	var verified []MonitoringEndpoint
	for _, ep := range endpoints {
		if err := probeEndpoint(ctx, restClient, ep); err != nil {
			logger.Info("Probe failed for endpoint", "service", ep.Service, "namespace", ep.Namespace, "port", ep.Port, "error", err)
		} else {
			logger.Info("Probe succeeded for endpoint", "service", ep.Service, "namespace", ep.Namespace, "port", ep.Port)
			verified = append(verified, ep)
		}
	}

	if len(verified) == 0 {
		logger.Info("All probes failed — no verified monitoring endpoints")
		return nil, nil
	}
	return verified, nil
}

// matchMonitoringService checks if a service matches known monitoring service name patterns.
func matchMonitoringService(svc corev1.Service) (MonitoringEndpoint, bool) {
	name := strings.ToLower(svc.Name)

	// Prometheus (excluding operator, alertmanager, node-exporter, grafana, kube-state-metrics)
	if strings.Contains(name, "prometheus") &&
		!strings.Contains(name, "operator") &&
		!strings.Contains(name, "alertmanager") &&
		!strings.Contains(name, "node-exporter") &&
		!strings.Contains(name, "grafana") &&
		!strings.Contains(name, "kube-state-metrics") {
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

	// Mimir — only match query-capable services (query-frontend, querier, nginx gateway).
	// Distributors, ingesters, compactors, store-gateways do NOT serve the Prometheus read API.
	if strings.Contains(name, "mimir") &&
		(strings.Contains(name, "query-frontend") ||
			strings.Contains(name, "querier") ||
			strings.Contains(name, "nginx")) {
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
func probeEndpoint(ctx context.Context, restClient rest.Interface, ep MonitoringEndpoint) error {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := restClient.Get().
		Namespace(ep.Namespace).
		Resource("services").
		Name(fmt.Sprintf("%s:%d", ep.Service, ep.Port)).
		SubResource("proxy", "api", "v1", "status", "buildinfo").
		Do(probeCtx)

	if result.Error() != nil {
		return fmt.Errorf("proxy request: %w", result.Error())
	}

	raw, err := result.Raw()
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check that the response looks like a Prometheus API response.
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unmarshal response: %w (body: %.200s)", err, string(raw))
	}
	if resp.Status != "success" {
		return fmt.Errorf("unexpected status: %q", resp.Status)
	}
	return nil
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
	logger := log.FromContext(ctx)

	start := exp.CreationTimestamp.Time
	// Use current time as end — resources are still running during metrics collection
	// (cleanup happens after). This captures data Prometheus scraped after the workflow finished.
	end := time.Now()

	duration := end.Sub(start)
	if duration < 30*time.Second {
		return nil, fmt.Errorf("experiment duration too short for metrics: %s", duration)
	}

	stepStr := selectStep(duration)
	stepSec := 60
	fmt.Sscanf(strings.TrimSuffix(stepStr, "s"), "%d", &stepSec)

	// Use custom queries from spec.metrics if defined, otherwise defaults
	queries := exp.Spec.Metrics
	if len(queries) == 0 {
		queries = defaultTargetQueries()
	}

	// Build substitution variables (same as CollectMetricsSnapshot)
	// On target clusters, pods deploy to the experiment-named namespace
	// (e.g., "db-baseline-fsync-b8twf"), not the Experiment CR's namespace ("experiments").
	vars := map[string]string{
		"$EXPERIMENT": exp.Name,
		"$NAMESPACE":  exp.Name,
		"$DURATION":   promDuration(duration),
	}

	// Try each verified endpoint until one works
	for i, ep := range endpoints {
		logger.Info("Trying monitoring endpoint", "index", i, "service", ep.Service, "namespace", ep.Namespace, "port", ep.Port)
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
			resolvedQuery := substituteVars(mq.Query, vars)
			qr := QueryResult{
				Query:       resolvedQuery,
				Type:        mq.Type,
				Unit:        mq.Unit,
				Description: mq.Description,
			}

			var data []DataPoint
			var queryErr error

			switch mq.Type {
			case "range":
				data, queryErr = queryRangeViaProxy(ctx, restClient, ep, resolvedQuery, start, end, stepStr)
			case "instant":
				data, queryErr = queryInstantViaProxy(ctx, restClient, ep, resolvedQuery, end)
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
				logger.Info("Queries succeeded but returned empty data", "service", ep.Service)
			} else {
				logger.Info("Successfully collected metrics", "service", ep.Service, "namespace", ep.Namespace)
			}
			return result, nil
		}
		logger.Info("All queries failed for endpoint", "service", ep.Service, "namespace", ep.Namespace)
	}

	return nil, fmt.Errorf("all %d monitoring endpoints failed", len(endpoints))
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
	sysNS := `kube-system|gke-managed-system|gmp-system|gmp-public|kube-node-lease|kube-public|observability`
	infraPods := `alloy-.*|ts-vm-hub-.*|prometheus-.*|alertmanager-.*|grafana-.*|kube-state-metrics-.*|node-exporter-.*|kube-prometheus-stack-.*|tailscale-operator-.*|operator-.*`
	return []experimentsv1alpha1.MetricsQuery{
		{
			Name:        "cpu_by_pod",
			Query:       fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace!~"%s",pod!~"%s",container!="POD",container!=""}[1m])) by (pod)`, sysNS, infraPods),
			Type:        "range",
			Unit:        "cores",
			Description: "CPU usage by pod",
		},
		{
			Name:        "memory_by_pod",
			Query:       fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace!~"%s",pod!~"%s",container!="POD",container!=""}) by (pod)`, sysNS, infraPods),
			Type:        "range",
			Unit:        "bytes",
			Description: "Memory working set by pod",
		},
		{
			Name:        "cpu_total",
			Query:       fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace!~"%s",pod!~"%s"}[1m]))`, sysNS, infraPods),
			Type:        "range",
			Unit:        "cores",
			Description: "Total CPU usage",
		},
		{
			Name:        "memory_total",
			Query:       fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace!~"%s",pod!~"%s"})`, sysNS, infraPods),
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

// systemNamespaces is the set of namespaces to exclude from cadvisor metrics.
var systemNamespaces = map[string]bool{
	"kube-system":        true,
	"gke-managed-system": true,
	"gmp-system":         true,
	"gmp-public":         true,
	"kube-node-lease":    true,
	"kube-public":        true,
	"tailscale":          true,
	"observability":      true,
}

// isInfrastructurePod returns true for pods that are part of the experiment
// infrastructure (observability/monitoring/operators) rather than the workload.
func isInfrastructurePod(pod string) bool {
	infraPrefixes := []string{
		"alloy-", "ts-vm-hub-",                    // harness
		"prometheus-", "alertmanager-", "grafana-", // kube-prometheus-stack
		"kube-state-metrics-", "node-exporter-",    // kube-prometheus-stack exporters
		"kube-prometheus-stack-",                    // catch-all for helm release
		"tailscale-operator-", "operator-",         // operators
	}
	for _, p := range infraPrefixes {
		if strings.HasPrefix(pod, p) {
			return true
		}
	}
	return false
}

// CollectCadvisorMetrics scrapes cadvisor metrics directly from kubelet on target
// cluster nodes via the Kubernetes API server proxy. This works on any cluster
// without requiring Prometheus, VictoriaMetrics, or a remote-write pipeline.
// It returns CPU (cores) and memory (bytes) aggregated by namespace.
func CollectCadvisorMetrics(ctx context.Context, kubeconfig []byte, exp *experimentsv1alpha1.Experiment) (*MetricsResult, error) {
	logger := log.FromContext(ctx)

	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}

	// List nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	now := time.Now()

	// Aggregate CPU (counter seconds) and memory (gauge bytes) by pod
	cpuByPod := make(map[string]float64)
	memByPod := make(map[string]float64)

	restClient := clientset.CoreV1().RESTClient()

	for _, node := range nodes.Items {
		scrapeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		raw, err := restClient.Get().
			Resource("nodes").
			Name(node.Name).
			SubResource("proxy", "metrics", "cadvisor").
			Do(scrapeCtx).
			Raw()
		cancel()
		if err != nil {
			logger.Error(err, "Failed to scrape cadvisor", "node", node.Name)
			continue
		}

		parseCadvisorMetrics(string(raw), cpuByPod, memByPod)
		logger.Info("Scraped cadvisor metrics", "node", node.Name, "cpuPods", len(cpuByPod), "memPods", len(memByPod))
	}

	if len(cpuByPod) == 0 && len(memByPod) == 0 {
		return nil, fmt.Errorf("no cadvisor metrics found on any node")
	}

	// Build MetricsResult with instant-style data points
	result := &MetricsResult{
		CollectedAt: now.UTC(),
		Source:      "target:cadvisor",
		TimeRange: TimeRange{
			Start:    exp.CreationTimestamp.Time,
			End:      now,
			Duration: now.Sub(exp.CreationTimestamp.Time).String(),
			StepSec:  0, // instant snapshot
		},
		Queries: make(map[string]QueryResult),
	}

	// CPU usage by pod (counter total — represents cumulative CPU seconds)
	var cpuPoints []DataPoint
	var cpuTotal float64
	for pod, val := range cpuByPod {
		cpuPoints = append(cpuPoints, DataPoint{
			Labels:    map[string]string{"pod": pod},
			Timestamp: now,
			Value:     val,
		})
		cpuTotal += val
	}
	result.Queries["cpu_by_pod"] = QueryResult{
		Query:       "container_cpu_usage_seconds_total by pod (cadvisor)",
		Type:        "instant",
		Unit:        "cores",
		Description: "CPU usage by pod (cumulative seconds)",
		Data:        cpuPoints,
	}
	result.Queries["cpu_total"] = QueryResult{
		Query:       "sum(container_cpu_usage_seconds_total) (cadvisor)",
		Type:        "instant",
		Unit:        "cores",
		Description: "Total CPU usage (cumulative seconds)",
		Data: []DataPoint{{
			Labels:    map[string]string{"scope": "total"},
			Timestamp: now,
			Value:     cpuTotal,
		}},
	}

	// Memory by pod (gauge — current working set bytes)
	var memPoints []DataPoint
	var memTotal float64
	for pod, val := range memByPod {
		memPoints = append(memPoints, DataPoint{
			Labels:    map[string]string{"pod": pod},
			Timestamp: now,
			Value:     val,
		})
		memTotal += val
	}
	result.Queries["memory_by_pod"] = QueryResult{
		Query:       "container_memory_working_set_bytes by pod (cadvisor)",
		Type:        "instant",
		Unit:        "bytes",
		Description: "Memory working set by pod",
		Data:        memPoints,
	}
	result.Queries["memory_total"] = QueryResult{
		Query:       "sum(container_memory_working_set_bytes) (cadvisor)",
		Type:        "instant",
		Unit:        "bytes",
		Description: "Total memory working set",
		Data: []DataPoint{{
			Labels:    map[string]string{"scope": "total"},
			Timestamp: now,
			Value:     memTotal,
		}},
	}

	return result, nil
}

// parseCadvisorMetrics parses Prometheus text format from cadvisor and accumulates
// CPU and memory values by namespace, filtering out system namespaces and pause containers.
func parseCadvisorMetrics(text string, cpuByPod, memByPod map[string]float64) {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Match container_cpu_usage_seconds_total or container_memory_working_set_bytes
		var isCPU, isMem bool
		if strings.HasPrefix(line, "container_cpu_usage_seconds_total{") {
			isCPU = true
		} else if strings.HasPrefix(line, "container_memory_working_set_bytes{") {
			isMem = true
		} else {
			continue
		}

		ns := extractLabel(line, "namespace")
		pod := extractLabel(line, "pod")
		container := extractLabel(line, "container")

		// Skip system namespaces, empty containers (pause), and POD containers
		if ns == "" || systemNamespaces[ns] || container == "" || container == "POD" || pod == "" {
			continue
		}

		// Skip infrastructure pods deployed to experiment namespace
		if isInfrastructurePod(pod) {
			continue
		}

		// Extract value (last space-separated token before optional timestamp)
		val := extractValue(line)
		if val == 0 {
			continue
		}

		if isCPU {
			cpuByPod[pod] += val
		} else if isMem {
			memByPod[pod] += val
		}
	}
}

// extractLabel extracts a label value from a Prometheus metric line.
// e.g., extractLabel(`container_cpu{namespace="foo",container="bar"} 1.5`, "namespace") returns "foo".
func extractLabel(line, label string) string {
	key := label + `="`
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return ""
	}
	return line[start : start+end]
}

// extractValue extracts the numeric value from a Prometheus metric line.
// The value is the first token after the closing "} ".
func extractValue(line string) float64 {
	idx := strings.LastIndex(line, "} ")
	if idx < 0 {
		return 0
	}
	valStr := strings.TrimSpace(line[idx+2:])
	// Handle optional timestamp (space-separated after value)
	if spaceIdx := strings.IndexByte(valStr, ' '); spaceIdx > 0 {
		valStr = valStr[:spaceIdx]
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	return val
}
