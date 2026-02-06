# Prometheus & Grafana Deep Dive

## CLOUDBREAK STATION - Systems Monitoring Division

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  WEYLAND-ISHIMURA CORP.              ║  CLOUDBREAK STATION - SECTOR 7G       ║
║  SYSTEMS MONITORING DIVISION         ║  CYCLE 2847.3 | DEEP SPACE            ║
╠══════════════════════════════════════════════════════════════════════════════╣
║                                                                              ║
║  TECHNICIAN ORIENTATION - PROMETHEUS MONITORING SYSTEMS                      ║
║  ────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  Welcome aboard, Technician.                                                 ║
║                                                                              ║
║  You've been assigned to Cloudbreak Station's Systems Monitoring Division.  ║
║  Your primary duty: maintain observability of all station subsystems using  ║
║  the Prometheus monitoring array and Grafana visualization terminals.        ║
║                                                                              ║
║  The previous technician... transferred. Suddenly. Their notes were          ║
║  incomplete. This orientation will bring you up to speed.                    ║
║                                                                              ║
║  Current station status: NOMINAL                                             ║
║  Crew complement: 2,847                                                      ║
║  Days since last incident: 12                                                ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

**A comprehensive, SRE-focused tutorial for Platform Engineers**

This tutorial takes you from Prometheus fundamentals through production-grade observability patterns. You'll learn by doing: generating traffic with curl, writing PromQL queries, and building dashboards while understanding the "why" behind each concept.

## Prerequisites

- Basic Kubernetes knowledge (pods, services, deployments)
- Familiarity with command line and curl
- Access to a running prometheus-tutorial environment

```bash
# Initialize station monitoring systems
task kind:conduct -- prometheus-tutorial
```

## Table of Contents

- [Part 1: Foundations](#part-1-foundations)
  - [Module 1: Your First Metrics](#module-1-your-first-metrics)
  - [Module 2: The Four Metric Types](#module-2-the-four-metric-types)
  - [Module 3: The Prometheus Pull Model](#module-3-the-prometheus-pull-model)
- [Part 2: PromQL Mastery](#part-2-promql-mastery)
  - [Module 4: PromQL Fundamentals](#module-4-promql-fundamentals)
  - [Module 5: Counters and rate()](#module-5-counters-and-rate)
  - [Module 6: Histograms and histogram_quantile()](#module-6-histograms-and-histogram_quantile)
  - [Module 7: Aggregation Operators](#module-7-aggregation-operators)
- [Part 3: SRE Methodology](#part-3-sre-methodology-google-sre-book)
  - [Module 8: The Four Golden Signals](#module-8-the-four-golden-signals)
  - [Module 9: RED vs USE Methods](#module-9-red-vs-use-methods)
  - [Module 10: SLIs, SLOs, and Error Budgets](#module-10-slis-slos-and-error-budgets)
- [Part 4: Production Patterns](#part-4-production-patterns)
  - [Module 11: Alerting That Doesn't Suck](#module-11-alerting-that-doesnt-suck)
  - [Module 12: Recording Rules & Scalability](#module-12-recording-rules--scalability)
- [Part 5: Grafana Dashboards](#part-5-grafana-dashboards)
  - [Module 13: Dashboard Design Principles](#module-13-dashboard-design-principles)
  - [Module 14: Hands-On Dashboard Build](#module-14-hands-on-dashboard-build)
- [Part 6: Sector 12 Investigation](#part-6-sector-12-investigation)
  - [Module 15: The Anomaly](#module-15-the-anomaly)
  - [Module 16: Final Dashboard](#module-16-final-dashboard)

---

# Part 1: Foundations

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  STATION LOG - CYCLE 2847.3                                                 │
│  ─────────────────────────────────────────────────────────────────────────  │
│  All systems nominal. Beginning standard orientation protocol.              │
│  The monitoring terminals are operational. Coffee is cold.                  │
│  Routine shift. Nothing unusual to report.                                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 1: Your First Metrics

**Goal**: See what metrics look like raw, understand Prometheus text format

### Setup: Get Your Environment IPs

When you run `task kind:conduct -- prometheus-tutorial`, you'll see output like:

```
  Prometheus: http://172.19.255.200:9090
  Grafana:    http://172.19.255.201:80
  Metrics App: http://172.19.255.202:80
```

Set your metrics app IP for the exercises:

```bash
# Replace with your actual IP
export METRICS_IP=172.19.255.202
```

### Exercise 1.1: Generate Some Traffic

First, let's create some data for Prometheus to collect:

```bash
# Send 5 requests to the root endpoint
for i in {1..5}; do
  echo "Request $i:"
  curl -s http://$METRICS_IP/
  echo ""
done
```

You should see responses like:
```
Request 1:
Hello from metrics-app! Pod: metrics-app-abc123
Request 2:
Hello from metrics-app! Pod: metrics-app-abc123
...
```

### Exercise 1.2: View Raw Metrics

Now let's see what metrics the app exposes:

```bash
curl -s http://$METRICS_IP/metrics
```

You'll see output like this (truncated):

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/",status="200"} 5
# HELP active_connections Number of currently active connections
# TYPE active_connections gauge
active_connections 0
# HELP request_duration_seconds HTTP request duration in seconds
# TYPE request_duration_seconds histogram
request_duration_seconds_bucket{method="GET",path="/",le="0.005"} 0
request_duration_seconds_bucket{method="GET",path="/",le="0.01"} 0
request_duration_seconds_bucket{method="GET",path="/",le="0.025"} 0
request_duration_seconds_bucket{method="GET",path="/",le="0.05"} 3
request_duration_seconds_bucket{method="GET",path="/",le="0.1"} 5
request_duration_seconds_bucket{method="GET",path="/",le="+Inf"} 5
request_duration_seconds_sum{method="GET",path="/"} 0.234
request_duration_seconds_count{method="GET",path="/"} 5
```

### Understanding the Format

Every metric follows this pattern:

```
# HELP <metric_name> <description>
# TYPE <metric_name> <type>
<metric_name>{<label>=<value>,...} <value>
```

Let's break down what we see:

| Line | Meaning |
|------|---------|
| `# HELP http_requests_total` | Human-readable description |
| `# TYPE http_requests_total counter` | Metric type (counter, gauge, histogram, summary) |
| `http_requests_total{method="GET",...}` | Metric name with labels |
| `5` | The actual value |

### Labels: The Power of Prometheus

Labels let you slice and dice your metrics. Notice how `http_requests_total` has three labels:

```
http_requests_total{method="GET", path="/", status="200"} 5
```

This single metric can answer:
- How many total requests? (sum all)
- How many GET vs POST? (filter by method)
- How many errors? (filter by status=~"5..")
- Which endpoint is busiest? (group by path)

### How Our metrics-app Creates These Metrics

Here's the actual Go code that generates the request counter:

```go
// From components/apps/metrics-app/src/main.go

func instrumentHandler(path string, handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status code
        rw := &responseWriter{ResponseWriter: w, statusCode: 200}

        // Call the actual handler
        handler(rw, r)

        // Record metrics after the request completes
        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", rw.statusCode)

        // Increment the counter with labels
        httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()

        // Record the duration in the histogram
        requestDuration.WithLabelValues(r.Method, path).Observe(duration)
    }
}
```

**Key insight**: Every HTTP request goes through `instrumentHandler`, which:
1. Starts a timer
2. Calls the actual handler
3. Records the request count and duration with appropriate labels

### Exercise 1.3: Filter Metrics with grep

```bash
# See only counter metrics
curl -s http://$METRICS_IP/metrics | grep -E "^http_requests_total"

# See only histogram buckets
curl -s http://$METRICS_IP/metrics | grep "request_duration_seconds_bucket"

# See the HELP and TYPE for a specific metric
curl -s http://$METRICS_IP/metrics | grep -A1 "# HELP http_requests_total"
```

### Key Takeaways

1. Metrics are plain text in a well-defined format
2. Every metric has a name, optional labels, and a value
3. Labels enable powerful filtering and grouping
4. Applications expose metrics on a `/metrics` endpoint
5. The `# HELP` and `# TYPE` lines describe each metric

---

## Module 2: The Four Metric Types

**Goal**: Understand Counter, Gauge, Histogram, Summary with hands-on exercises

```
┌─────────────────────────────────────────────────────────────────┐
│                    The Four Metric Types                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  COUNTER (only goes up)           GAUGE (goes up and down)      │
│  ─────────────────────            ────────────────────────      │
│       ╱                                 ╱╲    ╱╲                │
│      ╱                                ╱   ╲╱    ╲               │
│    ╱                                ╱            ╲              │
│  ╱                                ╱               ╲             │
│  http_requests_total              active_connections            │
│                                                                  │
│  HISTOGRAM (distribution)         SUMMARY (percentiles)         │
│  ────────────────────────         ─────────────────────         │
│     ▂▃▅▇▅▃▂                       p50: 0.05s                    │
│    ▁▂▃▅▇▇▅▃▂▁                     p90: 0.15s                    │
│   bucket boundaries               p99: 0.45s                    │
│   request_duration_seconds        response_size_bytes           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Type 1: Counter

**Definition**: A value that only goes up (or resets to zero on restart)

**Use for**: Things you count - requests, errors, bytes sent, items processed

**Our metric**: `http_requests_total`

#### Exercise 2.1: Watch a Counter Increase

```bash
# Check the current count
curl -s http://$METRICS_IP/metrics | grep "^http_requests_total"

# Generate 10 more requests
for i in {1..10}; do curl -s http://$METRICS_IP/ > /dev/null; done

# Check again - the count increased!
curl -s http://$METRICS_IP/metrics | grep "^http_requests_total"
```

**The code that creates this counter:**

```go
// Counter: Total HTTP requests (with labels for method, path, status)
httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// Usage: increment by 1
httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
```

**Important**: Counters only go up! If you see a counter decrease, the application restarted and the counter reset to zero. This is why we use `rate()` in queries (covered in Module 5).

### Type 2: Gauge

**Definition**: A value that can go up or down

**Use for**: Current state - temperature, memory usage, active connections, queue depth

**Our metric**: `active_connections`

#### Exercise 2.2: Watch a Gauge Change

```bash
# Start 5 slow requests in the background (each takes 0.5-2 seconds)
for i in {1..5}; do
  curl -s http://$METRICS_IP/slow &
done

# QUICKLY check the gauge while requests are in progress
curl -s http://$METRICS_IP/metrics | grep "^active_connections"

# Wait for requests to complete
sleep 3

# Check again - should be back to 0
curl -s http://$METRICS_IP/metrics | grep "^active_connections"
```

**The code that creates this gauge:**

```go
// Gauge: Current active connections
activeConnections = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of currently active connections",
    },
)

// In the handler - increment when request starts, decrement when it ends
atomic.AddInt64(&connCount, 1)
activeConnections.Set(float64(atomic.LoadInt64(&connCount)))
defer func() {
    atomic.AddInt64(&connCount, -1)
    activeConnections.Set(float64(atomic.LoadInt64(&connCount)))
}()
```

### Type 3: Histogram

**Definition**: Samples observations and counts them in configurable buckets

**Use for**: Latency, request size, response time - anything where you want percentiles

**Our metric**: `request_duration_seconds`

#### Understanding Histogram Output

```bash
curl -s http://$METRICS_IP/metrics | grep "request_duration_seconds"
```

You'll see three types of lines:

```
# Buckets: cumulative count of observations <= the le (less than or equal) value
request_duration_seconds_bucket{method="GET",path="/",le="0.005"} 0
request_duration_seconds_bucket{method="GET",path="/",le="0.01"} 0
request_duration_seconds_bucket{method="GET",path="/",le="0.025"} 2
request_duration_seconds_bucket{method="GET",path="/",le="0.05"} 8
request_duration_seconds_bucket{method="GET",path="/",le="0.1"} 15
request_duration_seconds_bucket{method="GET",path="/",le="+Inf"} 15

# Sum of all observed values
request_duration_seconds_sum{method="GET",path="/"} 0.892

# Count of observations
request_duration_seconds_count{method="GET",path="/"} 15
```

```
┌─────────────────────────────────────────────────────────────────┐
│                How Histogram Buckets Work                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Request latencies: 45ms, 120ms, 89ms, 510ms, 95ms, 1200ms      │
│                                                                  │
│  Bucket (le=)     Count (cumulative)    Requests in bucket      │
│  ─────────────────────────────────────────────────────────────  │
│  le="0.05"   →    2                     [45ms, 45ms]            │
│  le="0.1"    →    4                     [+ 89ms, 95ms]          │
│  le="0.25"   →    5                     [+ 120ms]               │
│  le="0.5"    →    5                     (none)                  │
│  le="1"      →    6                     [+ 510ms]               │
│  le="2.5"    →    7                     [+ 1200ms]              │
│  le="+Inf"   →    7                     (all requests)          │
│                                                                  │
│  histogram_quantile(0.95, ...) calculates the 95th percentile   │
│  by interpolating between bucket boundaries                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Key insight**: Buckets are cumulative! `le="0.1"` includes ALL requests that took <= 100ms, not just those between 50ms and 100ms.

#### Exercise 2.3: Generate Different Latencies

```bash
# Generate fast requests (10-100ms)
for i in {1..20}; do curl -s http://$METRICS_IP/ > /dev/null; done

# Generate slow requests (500-2000ms)
for i in {1..10}; do curl -s http://$METRICS_IP/slow > /dev/null; done

# Check the histogram
curl -s http://$METRICS_IP/metrics | grep "request_duration_seconds_bucket"
```

Notice how the slow requests show up in higher buckets!

**The code that creates this histogram:**

```go
// Histogram: Request duration in seconds
requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path"},
)

// Usage: observe a duration
requestDuration.WithLabelValues(r.Method, path).Observe(duration)
```

### Type 4: Summary

**Definition**: Similar to histogram, but calculates quantiles on the client side

**Use for**: When you need exact percentiles and don't need to aggregate across instances

**Our metric**: `response_size_bytes`

```bash
curl -s http://$METRICS_IP/metrics | grep "response_size_bytes"
```

Output:
```
response_size_bytes{method="GET",path="/",quantile="0.5"} 42
response_size_bytes{method="GET",path="/",quantile="0.9"} 42
response_size_bytes{method="GET",path="/",quantile="0.99"} 42
response_size_bytes_sum{method="GET",path="/"} 630
response_size_bytes_count{method="GET",path="/"} 15
```

**The code:**

```go
// Summary: Response size in bytes
responseSize = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "response_size_bytes",
        Help:       "HTTP response size in bytes",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"method", "path"},
)
```

### How Each Endpoint Affects Metrics

Our metrics-app has four endpoints with different behaviors. Here's the actual Go code:

**`/` - Normal response (10-100ms latency)**
```go
func rootHandler(w http.ResponseWriter, r *http.Request) {
    // Simulate variable processing time (10-100ms)
    delay := time.Duration(10+rand.Intn(90)) * time.Millisecond
    time.Sleep(delay)

    podName := os.Getenv("POD_NAME")
    if podName == "" {
        podName = "unknown"
    }
    response := fmt.Sprintf("Hello from metrics-app! Pod: %s\n", podName)
    w.Write([]byte(response))
}
```

**`/slow` - Slow response (500-2000ms latency)**
```go
func slowHandler(w http.ResponseWriter, r *http.Request) {
    // Simulate slow endpoint (500-2000ms)
    delay := time.Duration(500+rand.Intn(1500)) * time.Millisecond
    time.Sleep(delay)

    w.Write([]byte("Slow response complete\n"))
}
```

**`/error` - Random errors (30% failure rate)**
```go
func errorHandler(w http.ResponseWriter, r *http.Request) {
    // Randomly return errors (30% of requests)
    if rand.Float64() < 0.3 {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal Server Error\n"))
        return
    }
    w.Write([]byte("Success\n"))
}
```

**`/process` - Business logic (increments business metric)**
```go
func processHandler(w http.ResponseWriter, r *http.Request) {
    // Simulate processing items (useful for business metrics)
    items := 1 + rand.Intn(10)
    itemsProcessed.Add(float64(items))

    response := fmt.Sprintf("Processed %d items\n", items)
    w.Write([]byte(response))
}
```

These handlers explain why:
- `/slow` shows up in higher histogram buckets
- `/error` generates 5xx status codes ~30% of the time
- `/process` increments the `items_processed_total` business metric

### Histogram vs Summary: When to Use Which

| Aspect | Histogram | Summary |
|--------|-----------|---------|
| **Aggregation** | Can aggregate across instances | Cannot aggregate |
| **Quantile calculation** | Server-side (Prometheus) | Client-side (app) |
| **Bucket configuration** | Choose buckets upfront | Choose quantiles upfront |
| **Cost** | Low client overhead | Higher client overhead |
| **Recommendation** | **Use this by default** | Only when you can't use histogram |

**Rule of thumb**: Always use histograms unless you have a specific reason to use summaries.

### Quick Reference: Choosing the Right Type

| Question | Answer |
|----------|--------|
| Can it only increase? | **Counter** (requests, errors, bytes) |
| Can it go up and down? | **Gauge** (connections, queue depth, temperature) |
| Do you need percentiles? | **Histogram** (latency, request size) |
| Do you need exact quantiles from a single instance? | **Summary** (rarely needed) |

---

## Module 3: The Prometheus Pull Model

**Goal**: Understand how Prometheus scrapes targets

```
┌─────────────────────────────────────────────────────────────────┐
│                    Prometheus Architecture                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐     scrape      ┌──────────────────────┐     │
│   │  metrics-app │ ◄────────────── │     Prometheus       │     │
│   │   :80/metrics│   every 15s     │                      │     │
│   └──────────────┘                 │  ┌────────────────┐  │     │
│                                    │  │   Time Series  │  │     │
│   ┌──────────────┐     scrape      │  │    Database    │  │     │
│   │ node-exporter│ ◄────────────── │  │    (TSDB)      │  │     │
│   │  :9100/metrics│                │  └────────────────┘  │     │
│   └──────────────┘                 │                      │     │
│                                    │  ┌────────────────┐  │     │
│   ┌──────────────┐     scrape      │  │  Rule Engine   │  │     │
│   │ kube-state   │ ◄────────────── │  │  (alerts)      │  │     │
│   │   -metrics   │                 │  └────────────────┘  │     │
│   └──────────────┘                 └──────────┬───────────┘     │
│                                               │                  │
│                                               │ PromQL           │
│                                               ▼                  │
│                                    ┌──────────────────────┐     │
│                                    │      Grafana         │     │
│                                    │   (Visualization)    │     │
│                                    └──────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

### Pull vs Push: Why Prometheus Pulls

**Push model** (e.g., StatsD, Graphite):
- Applications send metrics to a central collector
- If collector is down, metrics are lost
- Hard to know if a target is down vs just not sending data

**Pull model** (Prometheus):
- Prometheus fetches metrics from targets
- If a scrape fails, Prometheus knows the target is down
- Simpler application code - just expose an endpoint
- Central control over scrape frequency and targets

### Service Discovery in Kubernetes

Prometheus doesn't need a static list of targets. In Kubernetes, it discovers them automatically using **ServiceMonitors**.

Our metrics-app has a ServiceMonitor that tells Prometheus: "Scrape any service with label `app: metrics-app` on port `http`."

```yaml
# components/apps/metrics-app/k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: metrics-app
  labels:
    app: metrics-app
spec:
  selector:
    matchLabels:
      app: metrics-app
  endpoints:
    - port: http
      interval: 15s
      path: /metrics
```

### Exercise 3.1: Check Prometheus Targets

Open Prometheus UI at your Prometheus IP and go to **Status → Targets**.

You should see:
- `serviceMonitor/demo/metrics-app/0` - our app
- Various other targets (node-exporter, kube-state-metrics, etc.)

Each target shows:
- **State**: UP or DOWN
- **Last Scrape**: When it was last scraped
- **Scrape Duration**: How long the scrape took
- **Error**: Any error message if DOWN

### The `up` Metric

Prometheus automatically creates an `up` metric for every target:

```promql
# In Prometheus UI, try:
up{job="metrics-app"}
```

- `up == 1` means the target is healthy
- `up == 0` means the scrape failed

This is the most basic health check!

### Scrape Intervals and Staleness

**Scrape interval**: How often Prometheus fetches metrics (default: 15s in our setup)

**Staleness**: If no new sample arrives within 5 minutes, data is considered stale

This means:
- Data points are typically 15 seconds apart
- Queries like `rate(x[1m])` need at least 4 data points (1 minute / 15 seconds)
- If you set a 1-second scrape interval, you get more resolution but more storage

### Why This Matters

1. **rate() needs multiple samples**: `rate(http_requests_total[30s])` won't work well with 15s scrape interval (only 2 samples)

2. **Alerting delays**: A metric must be unhealthy for at least one scrape interval before Prometheus sees it

3. **Storage costs**: More frequent scrapes = more data points = more storage

### Key Takeaways

1. Prometheus **pulls** metrics from targets on a regular interval
2. ServiceMonitors tell Prometheus what to scrape in Kubernetes
3. The `up` metric indicates target health
4. Scrape interval affects query resolution and storage
5. Always ensure your rate() windows contain enough samples

---

# Part 2: PromQL Mastery

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  STATION LOG - CYCLE 2847.4                                                 │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Minor anomaly detected in Sector 12 environmental systems.                 │
│  Probably just sensor drift. Maintenance has been notified.                 │
│  The lights in the corridor outside flickered twice. Old wiring.            │
│                                                                             │
│  Learning to query the monitoring systems properly now. Need to             │
│  understand what these metrics are actually telling us.                     │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  TRAINING PROTOCOL: To understand the threat, we must become the    │   │
│  │  threat. Use the simulation terminal to generate anomalous traffic. │   │
│  │  Observe how the monitoring systems respond. Learn the patterns.    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 4: PromQL Fundamentals

**Goal**: Master instant vectors, range vectors, label selectors

### PromQL: Prometheus Query Language

PromQL is how you ask questions about your metrics. Open Prometheus UI and go to the **Graph** tab to try these queries.

### Instant Vectors

An **instant vector** is a set of time series, each with a single sample at the current time.

```promql
# All HTTP request counters
http_requests_total

# Result:
http_requests_total{method="GET", path="/", status="200"} 150
http_requests_total{method="GET", path="/slow", status="200"} 45
http_requests_total{method="GET", path="/error", status="200"} 28
http_requests_total{method="GET", path="/error", status="500"} 12
```

### Label Selectors

Filter metrics using label selectors:

```promql
# Exact match
http_requests_total{path="/"}

# Not equal
http_requests_total{status!="200"}

# Regex match
http_requests_total{status=~"5.."}

# Regex not match
http_requests_total{path!~"/slow|/error"}
```

**Exercise 4.1**: Try each of these in Prometheus UI

### Range Vectors

A **range vector** is a set of time series, each with a range of samples over time.

```promql
# All samples from the last 5 minutes
http_requests_total[5m]

# Result (multiple values per series):
http_requests_total{...} 100 @1609459200
http_requests_total{...} 110 @1609459215
http_requests_total{...} 125 @1609459230
...
```

**Important**: You cannot graph a range vector directly! You must use a function like `rate()` to convert it to an instant vector.

```promql
# This WORKS - rate() returns an instant vector
rate(http_requests_total[5m])

# This FAILS - can't graph a range vector
http_requests_total[5m]
```

### Time Durations

PromQL uses these suffixes for time:

| Suffix | Meaning |
|--------|---------|
| `s` | seconds |
| `m` | minutes |
| `h` | hours |
| `d` | days |
| `w` | weeks |
| `y` | years |

```promql
http_requests_total[30s]   # Last 30 seconds
http_requests_total[5m]    # Last 5 minutes
http_requests_total[1h]    # Last hour
http_requests_total[7d]    # Last week
```

### Offset Modifier

Look at data from the past:

```promql
# Current value
http_requests_total

# Value from 1 hour ago
http_requests_total offset 1h

# Rate now vs rate 1 hour ago
rate(http_requests_total[5m])
rate(http_requests_total[5m] offset 1h)
```

### Exercise 4.2: Practice Queries

> *TRAINING MODE: Simulate station activity patterns. Normal operations, slow responses (overloaded systems), and errors (system failures). Watch how the metrics respond.*

```bash
# Simulate mixed station activity - normal, slow, and failing systems
for i in {1..50}; do
  curl -s http://$METRICS_IP/ > /dev/null        # Normal operations
  curl -s http://$METRICS_IP/slow > /dev/null &  # Overloaded system
  curl -s http://$METRICS_IP/error > /dev/null   # System failure
  sleep 0.2
done
```

Now try these queries in Prometheus UI:

```promql
# 1. All request counters
http_requests_total

# 2. Only root path requests
http_requests_total{path="/"}

# 3. Only errors (5xx status codes)
http_requests_total{status=~"5.."}

# 4. Everything except /slow
http_requests_total{path!="/slow"}

# 5. Request rate over last 1 minute
rate(http_requests_total[1m])

# 6. Request rate for errors only
rate(http_requests_total{status=~"5.."}[1m])
```

---

## Module 5: Counters and rate()

**Goal**: Understand why rate() is essential for counters

```
┌─────────────────────────────────────────────────────────────────┐
│              Why Use rate() Instead of Raw Counter?              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Raw counter (http_requests_total):                              │
│                                                                  │
│  1000 ─────────────────────────╲                                │
│   800 ─────────────────────╲    (pod restart = counter reset)   │
│   600 ─────────────────╲        ╲                               │
│   400 ─────────────╲             ╲──────────────────            │
│   200 ─────────╲                                                │
│     0 ────╱                                                     │
│        Time →                                                    │
│                                                                  │
│  rate(http_requests_total[1m]):                                  │
│                                                                  │
│   50 ─────────────────────────────────────────────              │
│   40 ─────────────────────────────────────────────              │
│   30 ─────────────────────────────────────────────              │
│        (smooth line - handles resets automatically)              │
│        Time →                                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### The Problem with Raw Counters

Counters only go up... until they don't. When a pod restarts, counters reset to zero:

```promql
# Raw counter - NOT what you want for dashboards
http_requests_total
```

This shows the total accumulated count, which:
- Keeps growing forever
- Resets to zero on restart (looks like negative spike)
- Doesn't show "requests per second"

### rate(): Requests Per Second

`rate()` calculates the per-second rate of increase:

```promql
# Requests per second, averaged over 5 minutes
rate(http_requests_total[5m])
```

**How rate() works:**
1. Takes all samples in the time range
2. Calculates the slope (change over time)
3. Returns per-second increase
4. **Automatically handles counter resets!**

### Exercise 5.1: rate() vs Raw Counter

```bash
# Generate steady traffic for 2 minutes
for i in {1..120}; do
  curl -s http://$METRICS_IP/ > /dev/null
  sleep 1
done &
```

While traffic is running, compare in Prometheus UI:

```promql
# Raw counter (ever-increasing)
http_requests_total{path="/"}

# Rate (requests per second)
rate(http_requests_total{path="/"}[1m])
```

Click **Graph** tab to see the visual difference!

### Choosing the Right Time Window

The time window in `rate()` affects smoothness vs responsiveness:

```promql
# Very responsive, but noisy
rate(http_requests_total[30s])

# Balanced
rate(http_requests_total[1m])

# Very smooth, but slow to show changes
rate(http_requests_total[5m])
```

**Rule of thumb**: Use at least 4x your scrape interval
- 15s scrape interval → use at least [1m] (4 samples)
- For dashboards, [5m] is often a good balance

### rate() vs irate() vs increase()

| Function | Returns | Use When |
|----------|---------|----------|
| `rate()` | Per-second rate, averaged | **Default choice** - dashboards, alerting |
| `irate()` | Per-second rate, last 2 samples | You need instant spikes (rarely) |
| `increase()` | Total increase over time range | You want "count in last hour" |

```promql
# Per-second rate (smoothed)
rate(http_requests_total[5m])

# Instant rate (spiky)
irate(http_requests_total[5m])

# Total increase in last hour
increase(http_requests_total[1h])
```

### The Code That Increments Counters

```go
// From instrumentHandler() in main.go
func instrumentHandler(path string, handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... handler code ...

        // This is what increments the counter
        httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
    }
}
```

Each call to `.Inc()` adds 1 to the counter. The counter value itself is just a number, but Prometheus stores the timestamp of each scrape, enabling `rate()` to calculate the change over time.

### Exercise 5.2: Error Rate Calculation

A common pattern - calculate error percentage:

> *TRAINING MODE: Simulate a cascade failure in Sector 12. Hammer the error endpoint. Watch the error rate climb. This is what it looks like when something goes wrong.*

```bash
# Simulate cascade failure - systems failing repeatedly
for i in {1..100}; do
  curl -s http://$METRICS_IP/error > /dev/null
done
```

```promql
# Error rate as percentage
sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
sum(rate(http_requests_total[5m]))
  * 100
```

This query:
1. Calculates error request rate
2. Divides by total request rate
3. Multiplies by 100 for percentage

---

## Module 6: Histograms and histogram_quantile()

**Goal**: Master latency percentile calculations

### Why Percentiles Matter

**Average latency** can be misleading:
- 99 requests at 10ms + 1 request at 10,000ms = average of 109ms
- But 99% of users experienced 10ms!

**Percentiles** show the real user experience:
- p50 (median): Half of requests are faster than this
- p90: 90% of requests are faster
- p95: 95% of requests are faster
- p99: 99% of requests are faster (often the most important!)

### Understanding Histogram Data

Remember from Module 2, our histogram produces:

```
request_duration_seconds_bucket{le="0.005"} 0
request_duration_seconds_bucket{le="0.01"} 0
request_duration_seconds_bucket{le="0.025"} 2
request_duration_seconds_bucket{le="0.05"} 15
request_duration_seconds_bucket{le="0.1"} 28
request_duration_seconds_bucket{le="0.25"} 35
request_duration_seconds_bucket{le="0.5"} 35
request_duration_seconds_bucket{le="1"} 38
request_duration_seconds_bucket{le="2.5"} 45
request_duration_seconds_bucket{le="+Inf"} 45
request_duration_seconds_sum 23.456
request_duration_seconds_count 45
```

The `le` label means "less than or equal" - buckets are cumulative!

### histogram_quantile(): Calculate Percentiles

```promql
# Median latency (p50)
histogram_quantile(0.50, rate(request_duration_seconds_bucket[5m]))

# 90th percentile
histogram_quantile(0.90, rate(request_duration_seconds_bucket[5m]))

# 95th percentile
histogram_quantile(0.95, rate(request_duration_seconds_bucket[5m]))

# 99th percentile (most important for SLOs!)
histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m]))
```

### Exercise 6.1: Generate Latency Data

> *TRAINING MODE: Simulate a system under strain. Most requests complete normally, but every fifth request hits an overloaded subsystem. Watch how the percentiles diverge - p50 stays low while p99 spikes. This is how you detect problems that only affect some users.*

```bash
# Simulate system strain - mostly normal, some slow
for i in {1..100}; do
  # 80% normal operations
  curl -s http://$METRICS_IP/ > /dev/null

  # 20% hit overloaded subsystem (the thing in Sector 12?)
  if [ $((i % 5)) -eq 0 ]; then
    curl -s http://$METRICS_IP/slow > /dev/null &
  fi
done

# Wait for slow requests to complete
wait
```

Now try these queries:

```promql
# p50 - median
histogram_quantile(0.50, rate(request_duration_seconds_bucket[5m]))

# p99 - 99th percentile
histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m]))
```

Notice how p99 is much higher than p50 because of the slow requests!

### Grouping by Labels

To get percentiles per endpoint:

```promql
# p99 latency by path
histogram_quantile(
  0.99,
  sum(rate(request_duration_seconds_bucket[5m])) by (path, le)
)
```

**Critical**: When aggregating histograms, you MUST include `le` in the `by()` clause!

```promql
# WRONG - missing le
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (path))

# RIGHT - includes le
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (path, le))
```

### Average Latency (Alternative)

If you want average instead of percentiles:

```promql
# Average latency
rate(request_duration_seconds_sum[5m])
  /
rate(request_duration_seconds_count[5m])
```

This divides total duration by request count.

### How Bucket Boundaries Affect Accuracy

Histogram quantiles are **estimated** by interpolating between buckets. The default buckets:

```go
// prometheus.DefBuckets
[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
```

This means:
- Great precision for requests under 10 seconds
- No precision above 10 seconds
- Better precision around common latencies

**For your service**, consider custom buckets:

```go
// Custom buckets optimized for a 300ms SLO
Buckets: []float64{.05, .1, .15, .2, .25, .3, .35, .4, .5, .75, 1, 2, 5}
```

### Exercise 6.2: Compare Paths

```promql
# p99 latency by path - see the difference!
histogram_quantile(
  0.99,
  sum(rate(request_duration_seconds_bucket[5m])) by (path, le)
)
```

You should see:
- `/` around 0.05-0.1s
- `/slow` around 1-2s

---

## Module 7: Aggregation Operators

**Goal**: Slice and dice metrics effectively

### Basic Aggregation Functions

```promql
# Sum all values
sum(http_requests_total)

# Average
avg(http_requests_total)

# Minimum value
min(http_requests_total)

# Maximum value
max(http_requests_total)

# Count of time series
count(http_requests_total)

# Standard deviation
stddev(http_requests_total)
```

### Grouping with by()

Group results by specific labels:

```promql
# Total requests by path
sum(http_requests_total) by (path)

# Total requests by status
sum(http_requests_total) by (status)

# Total by path AND status
sum(http_requests_total) by (path, status)
```

### Grouping with without()

Keep all labels EXCEPT specified ones:

```promql
# Sum without the instance label (aggregate across pods)
sum(http_requests_total) without (instance)

# Equivalent to: sum by (everything except instance)
```

### Exercise 7.1: Aggregation Practice

> *TRAINING MODE: Simulate activity across multiple station subsystems. Normal operations, error-prone systems, and the processing queue. Learn to slice the data by endpoint to identify which systems are failing.*

```bash
# Simulate activity across multiple subsystems
for i in {1..100}; do
  curl -s http://$METRICS_IP/ > /dev/null       # Primary systems
  curl -s http://$METRICS_IP/error > /dev/null  # Unstable systems
  curl -s http://$METRICS_IP/process > /dev/null # Processing queue
done
```

Try these queries:

```promql
# Total request rate across all paths
sum(rate(http_requests_total[5m]))

# Request rate by path
sum(rate(http_requests_total[5m])) by (path)

# Request rate by status
sum(rate(http_requests_total[5m])) by (status)

# Error rate by path
sum(rate(http_requests_total{status=~"5.."}[5m])) by (path)
```

### topk() and bottomk()

Find the highest/lowest values:

```promql
# Top 3 busiest endpoints
topk(3, sum(rate(http_requests_total[5m])) by (path))

# 3 endpoints with lowest traffic
bottomk(3, sum(rate(http_requests_total[5m])) by (path))
```

### Binary Operators

Math with metrics:

```promql
# Error percentage
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100

# Compare current to historical (ratio)
sum(rate(http_requests_total[5m]))
  / sum(rate(http_requests_total[5m] offset 1h))
```

### Comparison Operators

Filter by value:

```promql
# Only series where rate > 1 req/sec
rate(http_requests_total[5m]) > 1

# Only series with exactly 0 errors
sum(http_requests_total{status=~"5.."}) by (path) == 0
```

### Logical Operators

Combine conditions:

```promql
# High rate AND high error rate
(sum(rate(http_requests_total[5m])) by (path) > 10)
  and
(sum(rate(http_requests_total{status=~"5.."}[5m])) by (path) > 0.1)
```

---

# Part 3: SRE Methodology (Google SRE Book)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  STATION LOG - CYCLE 2847.5                                                 │
│  ─────────────────────────────────────────────────────────────────────────  │
│  ⚠ INCIDENT REPORT FILED: Crew member reported unusual sounds from         │
│    maintenance shaft 7-G. Investigation found nothing. Logged as           │
│    "acoustic anomaly - environmental systems."                              │
│                                                                             │
│  Command is asking about our Service Level Objectives. They want to know   │
│  if station systems are "meeting reliability targets." Need to establish   │
│  proper SLIs and SLOs. The previous technician's notes mention something   │
│  about "error budgets" but the pages are... stained.                        │
│                                                                             │
│  Days since last incident: 0                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 8: The Four Golden Signals

**Goal**: Implement Google's monitoring philosophy

```
┌─────────────────────────────────────────────────────────────────┐
│           The Four Golden Signals (Google SRE Book)              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  "If you can only measure four metrics of your user-facing       │
│   system, focus on these four." - Google SRE Book               │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  1. LATENCY                                                 ││
│  │     Time to service a request                               ││
│  │     ─────────────────────────────────────────────────────── ││
│  │     • Distinguish successful vs failed request latency      ││
│  │     • A fast 500 error is still a problem                   ││
│  │     • Track p50, p90, p99 percentiles                       ││
│  │     Query: histogram_quantile(0.99, rate(request_duration...││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  2. TRAFFIC                                                 ││
│  │     Demand on the system                                    ││
│  │     ─────────────────────────────────────────────────────── ││
│  │     • HTTP requests/second for web services                 ││
│  │     • Transactions/second for databases                     ││
│  │     • I/O operations for storage systems                    ││
│  │     Query: sum(rate(http_requests_total[5m]))               ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  3. ERRORS                                                  ││
│  │     Rate of failed requests                                 ││
│  │     ─────────────────────────────────────────────────────── ││
│  │     • Explicit failures (HTTP 5xx)                          ││
│  │     • Implicit failures (HTTP 200 but wrong content)        ││
│  │     • Policy violations (response > SLO threshold)          ││
│  │     Query: sum(rate(http_requests_total{status=~"5.."}[5m]))││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  4. SATURATION                                              ││
│  │     How "full" the service is                               ││
│  │     ─────────────────────────────────────────────────────── ││
│  │     • Most services degrade before 100% utilization         ││
│  │     • CPU, memory, disk, network capacity                   ││
│  │     • Queue depths, connection pools                        ││
│  │     Query: container_memory_usage_bytes / limits            ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Why These Four?

From the Google SRE book:

> "If you can only measure four metrics of your user-facing system, focus on these four."

These signals are **user-centric** - they directly affect user experience:
- Users care about **latency** (how fast)
- Users care about **errors** (does it work)
- Operators care about **traffic** (how much load)
- Operators care about **saturation** (will it break soon)

### Exercise 8.1: Implement All Four Signals

> *INCIDENT SIMULATION: Recreate the conditions from the Sector 12 incident report. High traffic, overloaded systems, cascading errors. Command wants to know - would we have caught it with proper monitoring? Would we have seen it coming?*

```bash
# Simulate the Sector 12 incident pattern
for i in {1..200}; do
  curl -s http://$METRICS_IP/ > /dev/null       # Normal traffic
  curl -s http://$METRICS_IP/slow > /dev/null & # Systems slowing down
  curl -s http://$METRICS_IP/error > /dev/null  # Failures beginning
  sleep 0.1
done
wait
# Check the graphs. This is what it looked like. Before.
```

Now implement each signal:

**1. LATENCY**

```promql
# p50 latency (median)
histogram_quantile(0.50, sum(rate(request_duration_seconds_bucket[5m])) by (le))

# p99 latency (critical for SLOs)
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (le))

# p99 latency for successful requests only
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket{status="200"}[5m])) by (le))
```

**2. TRAFFIC**

```promql
# Total requests per second
sum(rate(http_requests_total[5m]))

# Requests per second by endpoint
sum(rate(http_requests_total[5m])) by (path)
```

**3. ERRORS**

```promql
# Error rate (errors per second)
sum(rate(http_requests_total{status=~"5.."}[5m]))

# Error percentage
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100
```

**4. SATURATION**

Our metrics-app tracks active connections as a simple saturation metric:

```promql
# Active connections (our app's saturation metric)
active_connections

# For Kubernetes pods, you'd also check:
# CPU saturation
# container_cpu_usage_seconds_total / container_spec_cpu_quota

# Memory saturation
# container_memory_working_set_bytes / container_spec_memory_limit_bytes
```

### Latency: A Subtlety

Fast errors can hide problems! Consider:

```promql
# Average latency of all requests
rate(request_duration_seconds_sum[5m]) / rate(request_duration_seconds_count[5m])
```

If errors return immediately, they pull the average down! Always distinguish:

```promql
# Latency of successful requests
histogram_quantile(0.99,
  sum(rate(request_duration_seconds_bucket{status="200"}[5m])) by (le))

# Latency of error requests
histogram_quantile(0.99,
  sum(rate(request_duration_seconds_bucket{status=~"5.."}[5m])) by (le))
```

---

## Module 9: RED vs USE Methods

**Goal**: Know which method to apply where

```
┌─────────────────────────────────────────────────────────────────┐
│                   RED vs USE Methods                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌───────────────────────┐    ┌───────────────────────┐         │
│  │    RED Method         │    │    USE Method         │         │
│  │  (Request-Driven)     │    │  (Resource-Scoped)    │         │
│  ├───────────────────────┤    ├───────────────────────┤         │
│  │                       │    │                       │         │
│  │  R - Rate             │    │  U - Utilization      │         │
│  │      requests/sec     │    │      % time busy      │         │
│  │                       │    │                       │         │
│  │  E - Errors           │    │  S - Saturation       │         │
│  │      failures/sec     │    │      queue depth      │         │
│  │                       │    │                       │         │
│  │  D - Duration         │    │  E - Errors           │         │
│  │      latency (p99)    │    │      error events     │         │
│  │                       │    │                       │         │
│  └───────────────────────┘    └───────────────────────┘         │
│                                                                  │
│  When to use which:                                              │
│  ─────────────────────────────────────────────────────────────  │
│  RED  → Microservices, APIs, request handlers                   │
│  USE  → Databases, caches, nodes, infrastructure                │
│                                                                  │
│  Our metrics-app uses RED:                                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Rate:     sum(rate(http_requests_total[5m]))                ││
│  │ Errors:   sum(rate(http_requests_total{status=~"5.."}[5m])) ││
│  │ Duration: histogram_quantile(0.99, rate(request_duration...))│
│  └─────────────────────────────────────────────────────────────┘│
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### RED Method: For Services

Created by Tom Wilkie at Weaveworks. Use for anything that handles requests:

| Signal | Description | Query |
|--------|-------------|-------|
| **R**ate | Requests per second | `sum(rate(http_requests_total[5m]))` |
| **E**rrors | Errors per second | `sum(rate(http_requests_total{status=~"5.."}[5m]))` |
| **D**uration | Latency (p99) | `histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m]))` |

**Use RED for**:
- Web servers
- APIs
- Microservices
- Message consumers
- Any request/response system

### USE Method: For Resources

Created by Brendan Gregg. Use for physical/virtual resources:

| Signal | Description | Example Query |
|--------|-------------|---------------|
| **U**tilization | % time busy | `avg(rate(node_cpu_seconds_total{mode!="idle"}[5m]))` |
| **S**aturation | Queue depth | `node_load1` (1-minute load average) |
| **E**rrors | Error events | `rate(node_disk_io_time_seconds_total[5m])` |

**Use USE for**:
- CPU
- Memory
- Disk
- Network interfaces
- Database connections

### Exercise 9.1: Complete RED Dashboard Queries

For our metrics-app:

```promql
# RATE: Requests per second
sum(rate(http_requests_total[5m]))

# ERRORS: Error percentage
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100

# DURATION: p50, p90, p99 latency
histogram_quantile(0.50, sum(rate(request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.90, sum(rate(request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (le))
```

### Combining RED and USE

For a complete picture, combine both:

1. **RED** for your service (how is it performing?)
2. **USE** for underlying infrastructure (why is it performing that way?)

Example: If RED shows high latency, USE might show high CPU utilization explaining it.

---

## Module 10: SLIs, SLOs, and Error Budgets

**Goal**: Define and measure reliability objectives

```
┌─────────────────────────────────────────────────────────────────┐
│              SLI → SLO → SLA → Error Budget                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│                         Defines                                  │
│  ┌──────────┐  what to   ┌──────────┐  contractual  ┌─────────┐ │
│  │   SLI    │ ─────────► │   SLO    │ ────────────► │   SLA   │ │
│  │ Indicator│   measure  │ Objective│    promise    │Agreement│ │
│  └──────────┘            └──────────┘               └─────────┘ │
│       │                       │                                  │
│       │                       │ determines                       │
│       │                       ▼                                  │
│       │               ┌──────────────┐                          │
│       │               │ Error Budget │                          │
│       │               │  (100% - SLO)│                          │
│       │               └──────────────┘                          │
│       │                                                          │
│  Example Stack:                                                  │
│  ─────────────────────────────────────────────────────────────  │
│                                                                  │
│  SLI:  % of requests completing in < 300ms                      │
│        Query: sum(rate(request_duration_seconds_bucket{le="0.3"}│
│               [30d])) / sum(rate(request_duration_seconds_count │
│               [30d]))                                            │
│                                                                  │
│  SLO:  99.9% of requests must complete in < 300ms               │
│        (3 nines = 43.2 minutes/month allowed failure)           │
│                                                                  │
│  SLA:  If we miss the SLO, customer gets service credits        │
│        (external commitment with consequences)                   │
│                                                                  │
│  Error Budget:  0.1% = ~43 minutes/month                        │
│        ┌────────────────────────────────────────┐               │
│        │██████████████████████████████████████░░│ 95% remaining │
│        └────────────────────────────────────────┘               │
│        "We can deploy risky changes while budget remains"       │
│                                                                  │
│  Common SLO Targets:                                             │
│  ─────────────────────────────────────────────────────────────  │
│  99%     = 7.3 hours/month downtime   (two nines)               │
│  99.9%   = 43.8 minutes/month         (three nines)             │
│  99.99%  = 4.4 minutes/month          (four nines)              │
│  99.999% = 26.3 seconds/month         (five nines)              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Definitions

**SLI (Service Level Indicator)**: A measurable quantity
- "% of requests completing successfully"
- "% of requests completing in < 300ms"
- "% of time the service returns 200"

**SLO (Service Level Objective)**: A target for your SLI
- "99.9% of requests complete successfully"
- "99% of requests complete in < 300ms"

**SLA (Service Level Agreement)**: External commitment with consequences
- "If we miss 99.9%, customer gets 10% credit"
- Usually more lenient than internal SLO

**Error Budget**: The inverse of SLO
- If SLO is 99.9%, error budget is 0.1%
- In a 30-day month: 0.1% × 43,200 minutes = 43.2 minutes

### Exercise 10.1: Define SLIs for metrics-app

**Availability SLI**: % of requests that succeed

```promql
# SLI: Availability (last 30 minutes for this exercise)
sum(rate(http_requests_total{status!~"5.."}[30m]))
  / sum(rate(http_requests_total[30m]))
```

**Latency SLI**: % of requests under threshold

```promql
# SLI: Latency (% of requests under 500ms)
sum(rate(request_duration_seconds_bucket{le="0.5"}[30m]))
  / sum(rate(request_duration_seconds_count[30m]))
```

### Exercise 10.2: Calculate SLO Compliance

Let's say our SLOs are:
- Availability: 99.9%
- Latency (p99): < 1 second

```promql
# Are we meeting availability SLO?
sum(rate(http_requests_total{status!~"5.."}[30m]))
  / sum(rate(http_requests_total[30m]))
  >= 0.999

# Are we meeting latency SLO?
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[30m])) by (le))
  < 1
```

### Exercise 10.3: Calculate Error Budget Consumption

> *WARNING: This simulation will consume error budget rapidly. In production, this is what a real incident looks like. The error budget is your buffer against chaos. Watch it burn.*

```bash
# SIMULATION: Catastrophic failure event
# This is what consumed our error budget during the Sector 12 incident
for i in {1..500}; do
  curl -s http://$METRICS_IP/error > /dev/null
done
# 500 requests. 30% failure rate. That's 150 errors.
# At 99.9% SLO, that's... significant.
```

```promql
# Error budget consumed (as percentage of total budget)
# If SLO is 99.9%, error budget is 0.1%

# Current error rate
(
  sum(rate(http_requests_total{status=~"5.."}[30m]))
  / sum(rate(http_requests_total[30m]))
)
/
0.001  # Error budget (0.1%)
* 100  # Convert to percentage
```

If this returns > 100, you've exhausted your error budget!

### Multi-Window, Multi-Burn-Rate Alerts

Simple threshold alerts are noisy. Instead, use **burn rate**:

```
Burn Rate = Error Rate / Error Budget

If burn rate = 1: You'll exactly exhaust budget at period end
If burn rate = 14.4: You'll exhaust budget in 1 hour
If burn rate = 6: You'll exhaust budget in 6 hours
```

Alert when burn rate is high over multiple windows:

```yaml
# Alert: 1-hour burn rate of 14.4x (exhausts budget in 1 hour)
# Combined with 5-minute burn rate for confirmation
groups:
  - name: slo-alerts
    rules:
      - alert: HighErrorBudgetBurn
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[1h]))
            / sum(rate(http_requests_total[1h]))
          ) > (14.4 * 0.001)
          and
          (
            sum(rate(http_requests_total{status=~"5.."}[5m]))
            / sum(rate(http_requests_total[5m]))
          ) > (14.4 * 0.001)
        labels:
          severity: critical
        annotations:
          summary: "Error budget being consumed rapidly"
```

### Choosing the Right SLO

**Don't just pick "99.99%" because it sounds good!**

Consider:
1. **User expectations**: What do users actually need?
2. **Cost**: Higher SLOs cost more (redundancy, on-call, etc.)
3. **Dependencies**: Your SLO can't exceed your dependencies' SLOs
4. **History**: What have you actually achieved?

| Target | Monthly Downtime | Cost Level |
|--------|-----------------|------------|
| 99% | 7.3 hours | $ |
| 99.9% | 43.8 minutes | $$ |
| 99.99% | 4.4 minutes | $$$$ |
| 99.999% | 26.3 seconds | $$$$$$$$ |

---

# Part 4: Production Patterns

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ██ PRIORITY ALERT - CYCLE 2847.6 ██                                        │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Multiple systems showing anomalous readings. Sector 12 environmental       │
│  completely offline. Crew complement reports: 2,844 (−3 from last count).  │
│                                                                             │
│  The old alerting rules are firing constantly - nobody can tell what's      │
│  real anymore. Maintenance team sent to Sector 12 hasn't reported back.     │
│                                                                             │
│  Need better alerting. Need to know what actually requires action vs        │
│  what's just noise. The station can't afford to miss a real emergency.      │
│                                                                             │
│  Found a note in the previous technician's desk: "BURN RATE. WATCH THE      │
│  BURN RATE." Underlined three times.                                        │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 11: Alerting That Doesn't Suck

**Goal**: Design alerts that are actionable, not noisy

### The Problem with Bad Alerts

Bad alerting leads to:
- **Alert fatigue**: Too many alerts = people ignore them
- **Missed incidents**: Important alerts lost in noise
- **False positives**: Alerts that don't require action
- **Burnout**: On-call becomes unsustainable

### Symptom vs Cause-Based Alerting

**Cause-based** (avoid):
```yaml
# BAD: Alerts on every possible cause
- alert: HighCPU
  expr: cpu_usage > 80%
- alert: HighMemory
  expr: memory_usage > 80%
- alert: HighDiskIO
  expr: disk_io > 1000
# ... endless causes
```

**Symptom-based** (prefer):
```yaml
# GOOD: Alerts on what users experience
- alert: HighErrorRate
  expr: error_rate > SLO_threshold
- alert: HighLatency
  expr: p99_latency > SLO_threshold
```

### Alert Severity Levels

| Severity | Response | Examples |
|----------|----------|----------|
| **Page** | Wake someone up | SLO breach, complete outage |
| **Ticket** | Fix during business hours | Degraded performance, approaching limits |
| **Log** | Informational | Expected events, successful deployments |

**Rule**: If an alert doesn't require human action, it's not an alert.

### Exercise 11.1: Convert Bad Alert to Good Alert

**Bad**: Noisy threshold alert

```yaml
# BAD: Fires on any spike
- alert: HighErrorRate
  expr: sum(rate(http_requests_total{status=~"5.."}[5m])) > 10
  for: 1m
```

Problems:
- Fires on short spikes
- Doesn't consider normal traffic levels
- No relation to SLO

**Good**: Burn-rate alert

```yaml
# GOOD: Only fires when error budget is being consumed significantly
- alert: ErrorBudgetBurn
  expr: |
    # 1-hour window showing 14.4x burn rate
    (
      sum(rate(http_requests_total{status=~"5.."}[1h]))
      / sum(rate(http_requests_total[1h]))
    ) > (14.4 * 0.001)
    and
    # 5-minute window confirming the problem is current
    (
      sum(rate(http_requests_total{status=~"5.."}[5m]))
      / sum(rate(http_requests_total[5m]))
    ) > (14.4 * 0.001)
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Error budget burn rate is 14.4x - will exhaust in 1 hour"
    runbook: "https://wiki/runbooks/error-budget-burn"
```

### Multi-Window Alerting Strategy

Different windows catch different problems:

| Window | Burn Rate | Catches | Severity |
|--------|-----------|---------|----------|
| 1h + 5m | 14.4x | Fast-burning incidents | Page |
| 6h + 30m | 6x | Slow-burning issues | Page |
| 3d + 6h | 1x | Chronic problems | Ticket |

```yaml
# Fast burn - page immediately
- alert: ErrorBudgetFastBurn
  expr: |
    (job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.001))
    and
    (job:slo_errors_per_request:ratio_rate5m > (14.4 * 0.001))
  labels:
    severity: critical

# Slow burn - page during business hours
- alert: ErrorBudgetSlowBurn
  expr: |
    (job:slo_errors_per_request:ratio_rate6h > (6 * 0.001))
    and
    (job:slo_errors_per_request:ratio_rate30m > (6 * 0.001))
  labels:
    severity: warning
```

### Essential Alert Annotations

Every alert should have:

```yaml
annotations:
  summary: "One-line description of what's wrong"
  description: "Detailed description with {{ $value }} of the metric"
  runbook: "https://wiki/runbooks/alert-name"
  dashboard: "https://grafana/d/xxx?var-job={{ $labels.job }}"
```

---

## Module 12: Recording Rules & Scalability

**Goal**: Optimize for multi-cluster, high-cardinality environments

### When to Use Recording Rules

Recording rules pre-compute expensive queries:

**Without recording rules**:
```promql
# Every dashboard panel runs this expensive query
histogram_quantile(0.99,
  sum(rate(request_duration_seconds_bucket[5m])) by (job, le))
```

**With recording rules**:
```yaml
# Computed once, stored efficiently
groups:
  - name: slo-rules
    rules:
      - record: job:request_duration_seconds:p99_5m
        expr: |
          histogram_quantile(0.99,
            sum(rate(request_duration_seconds_bucket[5m])) by (job, le))
```

```promql
# Dashboard queries the pre-computed metric
job:request_duration_seconds:p99_5m
```

### Recording Rule Naming Convention

Use the format: `level:metric:operations`

```yaml
# Level 1: By job
- record: job:http_requests:rate5m
  expr: sum(rate(http_requests_total[5m])) by (job)

# Level 2: By job and path
- record: job_path:http_requests:rate5m
  expr: sum(rate(http_requests_total[5m])) by (job, path)
```

### Exercise 12.1: Write Recording Rules for metrics-app

```yaml
groups:
  - name: metrics-app-recording-rules
    interval: 15s
    rules:
      # Traffic (RED - Rate)
      - record: job:http_requests_total:rate5m
        expr: sum(rate(http_requests_total[5m])) by (job)

      - record: job_path:http_requests_total:rate5m
        expr: sum(rate(http_requests_total[5m])) by (job, path)

      # Errors (RED - Errors)
      - record: job:http_requests_errors:rate5m
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) by (job)

      - record: job:http_requests:error_ratio_5m
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (job)
          / sum(rate(http_requests_total[5m])) by (job)

      # Latency (RED - Duration)
      - record: job:request_duration_seconds:p50_5m
        expr: |
          histogram_quantile(0.50,
            sum(rate(request_duration_seconds_bucket[5m])) by (job, le))

      - record: job:request_duration_seconds:p99_5m
        expr: |
          histogram_quantile(0.99,
            sum(rate(request_duration_seconds_bucket[5m])) by (job, le))
```

### High Cardinality: The Silent Killer

**Cardinality** = number of unique time series

Each unique combination of labels creates a new series:

```
http_requests_total{method="GET", path="/", status="200", user_id="123"}
http_requests_total{method="GET", path="/", status="200", user_id="124"}
http_requests_total{method="GET", path="/", status="200", user_id="125"}
... millions of user_ids = millions of series
```

**Symptoms of high cardinality**:
- Slow queries
- High memory usage
- Prometheus crashes

**Prevention**:
- Never use unbounded labels (user_id, request_id, IP address)
- Use recording rules to aggregate away high-cardinality labels
- Set cardinality limits in Prometheus

### Multi-Cluster Monitoring

```
┌─────────────────────────────────────────────────────────────────┐
│           Multi-Cluster Monitoring (AKS Fleet Pattern)          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  AKS Cluster 1  │  │  AKS Cluster 2  │  │  AKS Cluster N  │  │
│  │   (West US)     │  │  (Central US)   │  │    (East US)    │  │
│  │                 │  │                 │  │                 │  │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │
│  │ │ Prometheus  │ │  │ │ Prometheus  │ │  │ │ Prometheus  │ │  │
│  │ │  (local)    │ │  │ │  (local)    │ │  │ │  (local)    │ │  │
│  │ └──────┬──────┘ │  │ └──────┬──────┘ │  │ └──────┬──────┘ │  │
│  └────────┼────────┘  └────────┼────────┘  └────────┼────────┘  │
│           │                    │                    │           │
│           │   remote_write     │   remote_write     │           │
│           └────────────────────┼────────────────────┘           │
│                                ▼                                 │
│           ┌────────────────────────────────────────┐            │
│           │     Central Metrics Store              │            │
│           │  (Thanos / Cortex / Azure Monitor)     │            │
│           │                                        │            │
│           │  • Long-term storage (months/years)    │            │
│           │  • Global queries across clusters      │            │
│           │  • Deduplication & downsampling        │            │
│           └────────────────────┬───────────────────┘            │
│                                │                                 │
│                                ▼                                 │
│           ┌────────────────────────────────────────┐            │
│           │           Central Grafana              │            │
│           │    (Single pane of glass)              │            │
│           │                                        │            │
│           │  • Cross-cluster dashboards            │            │
│           │  • Unified alerting                    │            │
│           │  • Fleet-wide SLO tracking             │            │
│           └────────────────────────────────────────┘            │
│                                                                  │
│  Recording Rules (run locally, reduce query load):              │
│  ─────────────────────────────────────────────────────────────  │
│  - record: job:http_requests:rate5m                             │
│    expr: sum(rate(http_requests_total[5m])) by (job)            │
│                                                                  │
│  - record: job:request_latency:p99_5m                           │
│    expr: histogram_quantile(0.99, sum(rate(request_duration_    │
│          seconds_bucket[5m])) by (job, le))                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Pattern**: Each cluster has local Prometheus that:
1. Scrapes local targets
2. Runs recording rules to pre-aggregate
3. Sends aggregated metrics to central store via `remote_write`

This reduces:
- Central query load (aggregated, not raw data)
- Network traffic (fewer series)
- Storage costs (downsampled over time)

---

# Part 5: Grafana Dashboards

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ██ EMERGENCY PROTOCOL ACTIVE - CYCLE 2847.7 ██                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Sectors 12, 13, 14 sealed. Command has ordered a full monitoring           │
│  dashboard for all station systems. They need to see everything.            │
│  They need to see it NOW.                                                   │
│                                                                             │
│  Whatever is happening down there, the metrics will tell us. They           │
│  have to. Build the dashboard. Monitor the signals.                         │
│                                                                             │
│  Crew complement: 2,831                                                     │
│                                                                             │
│  The lights are flickering again.                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 13: Dashboard Design Principles

**Goal**: Build dashboards that answer questions, not create them

### The Problem with Bad Dashboards

Bad dashboards:
- Show raw metrics without context
- Have too many panels
- Use confusing colors
- Don't answer "is there a problem?"

### The Four-Panel Rule

Start with four panels answering these questions:

1. **Is there a problem?** (Big status indicator)
2. **How bad is it?** (Current value)
3. **How is it trending?** (Time series)
4. **Where is the problem?** (Breakdown by dimension)

### Progressive Disclosure

Organize dashboards in layers:

1. **Overview**: Single-glance health check
2. **Service Dashboard**: RED metrics for one service
3. **Debug Dashboard**: Detailed breakdowns for troubleshooting

### Color Rules

- **Red** = Bad (errors, problems)
- **Yellow/Orange** = Warning (approaching limits)
- **Green** = Good (healthy)
- **Blue** = Informational (neutral data)

**Never** use red decoratively!

### Template Variables

Use variables for multi-tenancy:

```yaml
# Dashboard variable definition
variable:
  name: namespace
  query: label_values(http_requests_total, namespace)

# Usage in queries
sum(rate(http_requests_total{namespace="$namespace"}[5m]))
```

---

## Module 14: Hands-On Dashboard Build

**Goal**: Step-by-step dashboard creation

### Access Grafana

Open Grafana at your Grafana IP (shown when you started the tutorial).
- Login: admin / admin

### Step 1: Create New Dashboard

1. Click **+** → **Dashboard**
2. Click **Add visualization**

### Step 2: Request Rate Panel

**Panel type**: Time Series

**Query**:
```promql
sum(rate(http_requests_total[5m]))
```

**Settings**:
- Title: "Request Rate"
- Unit: requests/sec
- Legend: `{{path}}`

For per-path breakdown:
```promql
sum(rate(http_requests_total[5m])) by (path)
```

### Step 3: Error Rate Panel

**Panel type**: Stat

**Query**:
```promql
sum(rate(http_requests_total{status=~"5.."}[5m]))
  / sum(rate(http_requests_total[5m])) * 100
```

**Settings**:
- Title: "Error Rate"
- Unit: percent (0-100)
- Thresholds:
  - 0 (green)
  - 1 (yellow)
  - 5 (red)

### Step 4: Latency Percentiles Panel

**Panel type**: Time Series

**Queries** (add multiple):

Query A:
```promql
histogram_quantile(0.50, sum(rate(request_duration_seconds_bucket[5m])) by (le))
```
Legend: `p50`

Query B:
```promql
histogram_quantile(0.90, sum(rate(request_duration_seconds_bucket[5m])) by (le))
```
Legend: `p90`

Query C:
```promql
histogram_quantile(0.99, sum(rate(request_duration_seconds_bucket[5m])) by (le))
```
Legend: `p99`

**Settings**:
- Title: "Latency Percentiles"
- Unit: seconds

### Step 5: Top Endpoints Table

**Panel type**: Table

**Query**:
```promql
topk(5, sum(rate(http_requests_total[5m])) by (path))
```

**Settings**:
- Title: "Top Endpoints"
- Format: Table
- Columns: path, Value

### Step 6: Active Connections Gauge

**Panel type**: Gauge

**Query**:
```promql
active_connections
```

**Settings**:
- Title: "Active Connections"
- Min: 0
- Max: 100 (adjust based on expected load)
- Thresholds:
  - 0 (green)
  - 50 (yellow)
  - 80 (red)

### Step 7: SLO Compliance Panel

**Panel type**: Stat

**Query** (Availability SLI):
```promql
sum(rate(http_requests_total{status!~"5.."}[1h]))
  / sum(rate(http_requests_total[1h])) * 100
```

**Settings**:
- Title: "Availability (SLO: 99.9%)"
- Unit: percent (0-100)
- Thresholds:
  - 99.9 (green) - meeting SLO
  - 99 (yellow) - close to SLO
  - 0 (red) - missing SLO

### Final Dashboard Layout

Arrange panels in this order (top to bottom, left to right):

```
┌─────────────────────┬─────────────────────┬─────────────────────┐
│  Request Rate       │  Error Rate (Stat)  │  SLO Compliance     │
│  (Time Series)      │                     │  (Stat)             │
├─────────────────────┴─────────────────────┴─────────────────────┤
│                    Latency Percentiles                          │
│                    (Time Series - Full Width)                   │
├─────────────────────────────────────────────┬───────────────────┤
│  Top Endpoints (Table)                      │ Active Connections│
│                                             │ (Gauge)           │
└─────────────────────────────────────────────┴───────────────────┘
```

### Save Your Dashboard

1. Click the save icon (disk)
2. Name: "metrics-app RED Dashboard"
3. Save

---

# Part 6: Sector 12 Investigation

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ██ CLASSIFIED - AUTHORIZATION LEVEL 7 - CYCLE 2847.9 ██                    │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Command has authorized access to Sector 12 monitoring data.                │
│                                                                             │
│  The station-monitor system tracks real-time telemetry from                 │
│  Cloudbreak Station's internal systems. You are now cleared to              │
│  investigate the anomalies that began 72 hours ago.                         │
│                                                                             │
│  Previous investigator status: MISSING                                      │
│  Sector 12 access: RESTRICTED                                               │
│  Beacon signal: ACTIVE (unknown source)                                     │
│                                                                             │
│  Your mission: Use your monitoring skills to determine what happened.       │
│                                                                             │
│  Note: The station-monitor application runs in the 'station' namespace.     │
│        Metrics use the 'station_' prefix.                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module 15: The Anomaly

**Goal**: Investigate Sector 12 using the station telemetry system

### Setup: Station Monitor Access

When you started the tutorial, you received two endpoints:
- **metrics-app**: Training simulator (demo namespace)
- **station-monitor**: Live Cloudbreak telemetry (station namespace)

Set your station monitor IP:

```bash
# Replace with your actual Station Monitor IP
export STATION_IP=172.19.255.203
```

### Exercise 15.1: Initial Station Assessment

Query the station's core metrics in Prometheus:

```promql
# Current crew complement (should be 2847)
station_crew_complement

# Containment integrity (should be 1.0 = 100%)
station_containment_integrity

# Anomaly detection count
station_anomaly_detections_total

# Mysterious beacon signal
station_beacon_signal_strength
```

**What you should notice**: The metrics may already be degraded. The station has been running since you started the tutorial.

### Exercise 15.2: Sector Comparison

Compare the sectors. Something is different about Sector 12:

```promql
# Life support efficiency by sector (1.0 = optimal)
station_life_support_efficiency

# Sector status (1.0 = nominal, <0.7 = degraded, <0.5 = critical)
station_sector_status
```

Query for degraded sectors:

```promql
station_sector_status < 0.7
```

### Exercise 15.3: Become the Investigator

Generate traffic to the station endpoints:

```bash
# Check normal sectors (these should work fine)
curl http://$STATION_IP/sectors/1
curl http://$STATION_IP/sectors/2
curl http://$STATION_IP/sectors/3

# Now... access Sector 12
curl http://$STATION_IP/sector12
```

Run it multiple times:

```bash
# Sector 12 has a 70% failure rate
for i in {1..20}; do
  echo "Attempt $i:"
  curl -s http://$STATION_IP/sector12
  echo ""
done
```

**Notice**: The responses from Sector 12 are... unusual.

### Exercise 15.4: The Beacon

Something is transmitting from Sector 12:

```bash
# Check the beacon
curl http://$STATION_IP/beacon
```

Monitor the beacon signal in Prometheus:

```promql
# Watch the signal fluctuate
station_beacon_signal_strength

# Rate of change (is it getting stronger?)
deriv(station_beacon_signal_strength[5m])
```

### Exercise 15.5: Track the Degradation

The station state degrades over time. Calculate the rates:

```promql
# Crew loss rate (crew per minute) - this should be zero
deriv(station_crew_complement[5m])

# Containment decay rate
deriv(station_containment_integrity[5m])

# Life support degradation by sector
deriv(station_life_support_efficiency[5m])
```

### Exercise 15.6: Error Rate Analysis

Apply your RED method knowledge to Sector 12:

```promql
# Sector 12 error rate
sum(rate(station_sector_requests_total{sector="12", status=~"5.."}[5m]))
  / sum(rate(station_sector_requests_total{sector="12"}[5m])) * 100

# Compare to normal sectors
sum(rate(station_sector_requests_total{sector!="12", status=~"5.."}[5m]))
  / sum(rate(station_sector_requests_total{sector!="12"}[5m])) * 100
```

### Exercise 15.7: The Four Golden Signals - Sector 12

Apply everything you've learned:

**Latency** - Is Sector 12 responding slowly?
```promql
histogram_quantile(0.99,
  sum(rate(station_sector_response_seconds_bucket{sector="12"}[5m])) by (le))
```

**Traffic** - How many requests to Sector 12?
```promql
sum(rate(station_sector_requests_total{sector="12"}[5m]))
```

**Errors** - What's the failure rate?
```promql
sum(rate(station_sector_requests_total{sector="12", status="500"}[5m]))
  / sum(rate(station_sector_requests_total{sector="12"}[5m])) * 100
```

**Saturation** - Is Sector 12 overwhelmed? (Check anomaly count growth)
```promql
rate(station_anomaly_detections_total[5m])
```

---

## Module 16: Final Dashboard

**Goal**: Build a crisis dashboard for Sector 12

### Exercise 16.1: Station Status Dashboard

Create a new Grafana dashboard with these panels:

**Panel 1: Crew Complement** (Stat)
```promql
station_crew_complement
```
- Thresholds: 2800 (green), 2700 (yellow), 2600 (red)

**Panel 2: Containment Integrity** (Gauge)
```promql
station_containment_integrity * 100
```
- Min: 0, Max: 100
- Thresholds: 80 (green), 60 (yellow), 40 (red)
- Unit: percent (0-100)

**Panel 3: Life Support by Sector** (Time Series)
```promql
station_life_support_efficiency * 100
```
- Legend: `Sector {{sector}}`
- Unit: percent (0-100)

**Panel 4: Sector Status** (Table)
```promql
station_sector_status
```

**Panel 5: Beacon Signal** (Time Series)
```promql
station_beacon_signal_strength
```
- Add annotation: "Signal source: UNKNOWN"

**Panel 6: Anomaly Count** (Stat with sparkline)
```promql
station_anomaly_detections_total
```

### Exercise 16.2: Crisis Alerts

Define alerts for the station:

```yaml
# Crew loss alert
- alert: CrewComplementDecreasing
  expr: deriv(station_crew_complement[5m]) < 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Crew complement is decreasing"

# Containment breach
- alert: ContainmentBreach
  expr: station_containment_integrity < 0.8
  for: 0m
  labels:
    severity: critical
  annotations:
    summary: "Containment integrity below 80%"

# Sector 12 anomaly
- alert: Sector12Anomaly
  expr: station_sector_status{sector="12"} < 0.5
  for: 0m
  labels:
    severity: critical
  annotations:
    summary: "Sector 12 status is critical"
```

### The Investigation Concludes

You've applied every technique in this tutorial to investigate Sector 12:

- **Metrics fundamentals** to understand the data format
- **PromQL queries** to slice and analyze the signals
- **The Four Golden Signals** to assess service health
- **SRE methodology** to define what "healthy" means
- **Dashboards** to visualize the crisis

The metrics tell the story. The crew complement is falling. Containment is failing. Something in Sector 12 is responding to requests, but the responses suggest it's no longer... entirely... the station system.

And the beacon keeps transmitting. Getting stronger.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  STATION MONITOR - CYCLE 2847.9                                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  > The metrics don't lie.                                                   │
│  > If you can see it in the graphs, it's happening.                         │
│  > If the numbers are changing, something is changing them.                 │
│                                                                             │
│  > Sector 12 error rate: 70%                                                │
│  > But the endpoint responds.                                               │
│  > Something responds.                                                      │
│                                                                             │
│  > Current status: MONITORING                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

# Quick Reference

## PromQL Cheat Sheet

### Selectors
```promql
metric_name                     # All series
metric_name{label="value"}      # Exact match
metric_name{label!="value"}     # Not equal
metric_name{label=~"regex"}     # Regex match
metric_name{label!~"regex"}     # Regex not match
```

### Functions
```promql
rate(counter[5m])               # Per-second rate
irate(counter[5m])              # Instant rate
increase(counter[1h])           # Total increase
histogram_quantile(0.99, ...)   # 99th percentile
sum(), avg(), min(), max()      # Aggregations
topk(5, ...), bottomk(5, ...)   # Top/bottom N
```

### Common Patterns
```promql
# Error rate percentage
sum(rate(errors[5m])) / sum(rate(total[5m])) * 100

# p99 latency by path
histogram_quantile(0.99, sum(rate(histogram_bucket[5m])) by (path, le))

# SLO compliance
sum(rate(requests{status!~"5.."}[1h])) / sum(rate(requests[1h]))
```

## SLO Quick Reference

| Target | Monthly Downtime | Error Budget |
|--------|-----------------|--------------|
| 99% | 7.3 hours | 1% |
| 99.9% | 43.8 minutes | 0.1% |
| 99.99% | 4.4 minutes | 0.01% |
| 99.999% | 26.3 seconds | 0.001% |

## Burn Rate Reference

| Window | Burn Rate | Exhausts Budget In |
|--------|-----------|-------------------|
| 1h | 14.4x | 1 hour |
| 6h | 6x | 6 hours |
| 1d | 3x | 1 day |
| 3d | 1x | 30 days |

---

# Additional Resources

## Books
- [Google SRE Book](https://sre.google/sre-book/table-of-contents/) - Free online
- [The SRE Workbook](https://sre.google/workbook/table-of-contents/) - Free online
- [Prometheus: Up & Running](https://www.oreilly.com/library/view/prometheus-up/9781492034131/)

## Documentation
- [Prometheus Documentation](https://prometheus.io/docs/)
- [PromQL Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Documentation](https://grafana.com/docs/)

## Related Tutorials
- [metrics-app source code](../../components/apps/metrics-app/)
- [kube-prometheus-stack component](../../components/observability/prometheus-stack/)

---

## Troubleshooting

### Prometheus not scraping metrics-app

Check ServiceMonitor exists and labels match:
```bash
kubectl get servicemonitor -n demo
kubectl describe servicemonitor metrics-app -n demo
```

Check Prometheus targets:
- In Prometheus UI: Status → Targets
- Look for `serviceMonitor/demo/metrics-app/0`

### Metrics not appearing

Ensure the app is exposing metrics:
```bash
kubectl port-forward -n demo svc/metrics-app 8080:80
curl http://localhost:8080/metrics
```

### Grafana login issues

Default credentials are `admin` / `admin`. If changed:
```bash
kubectl get secret -n observability kube-prometheus-stack-grafana \
  -o jsonpath='{.data.admin-password}' | base64 -d
```

### Slow queries

If queries are slow:
1. Use recording rules for complex aggregations
2. Reduce the time range
3. Check for high cardinality labels

---

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  END OF ORIENTATION MATERIALS                                                ║
╠══════════════════════════════════════════════════════════════════════════════╣
║                                                                              ║
║  FINAL STATION LOG - CYCLE 2847.8                                            ║
║  ─────────────────────────────────────────────────────────────────────────── ║
║                                                                              ║
║  The dashboard is complete. All systems are now visible.                     ║
║  The metrics are telling us something. The error rate in Sector 12 is...     ║
║  that can't be right. 100% errors. No successful requests.                   ║
║                                                                              ║
║  But the endpoint is still responding. Something is responding.              ║
║                                                                              ║
║  The previous technician's final note, found wedged behind the terminal:     ║
║                                                                              ║
║      "The metrics never lie. If you can see it in the graphs,                ║
║       it's already too late. But at least you'll know."                      ║
║                                                                              ║
║  ───────────────────────────────────────────────────────────────────────────  ║
║                                                                              ║
║  Crew complement: [RECALCULATING]                                            ║
║  Station status: [REQUIRES MANUAL VERIFICATION]                              ║
║                                                                              ║
║  Your shift begins now, Technician. Good luck.                               ║
║                                                                              ║
║  Press ENTER to end this session.                                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```
