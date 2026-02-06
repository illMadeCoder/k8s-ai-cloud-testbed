package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter: Total HTTP requests (with labels for method, path, status)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Gauge: Current active connections
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of currently active connections",
		},
	)

	// Histogram: Request duration in seconds
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		},
		[]string{"method", "path"},
	)

	// Summary: Response size in bytes
	responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "response_size_bytes",
			Help:       "HTTP response size in bytes",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "path"},
	)

	// Custom business metric: Items processed
	itemsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "items_processed_total",
			Help: "Total number of items processed by the application",
		},
	)

	// Connection counter for tracking
	connCount int64
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(responseSize)
	prometheus.MustRegister(itemsProcessed)
}

func instrumentHandler(path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Track active connections
		atomic.AddInt64(&connCount, 1)
		activeConnections.Set(float64(atomic.LoadInt64(&connCount)))
		defer func() {
			atomic.AddInt64(&connCount, -1)
			activeConnections.Set(float64(atomic.LoadInt64(&connCount)))
		}()

		start := time.Now()

		// Wrap response writer to capture size and status
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Call the actual handler
		handler(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", rw.statusCode)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		requestDuration.WithLabelValues(r.Method, path).Observe(duration)
		responseSize.WithLabelValues(r.Method, path).Observe(float64(rw.size))
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

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

func slowHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate slow endpoint (500-2000ms)
	delay := time.Duration(500+rand.Intn(1500)) * time.Millisecond
	time.Sleep(delay)

	w.Write([]byte("Slow response complete"))
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	// Randomly return errors (30% of requests)
	if rand.Float64() < 0.3 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	w.Write([]byte("Success"))
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate processing items (useful for business metrics)
	items := 1 + rand.Intn(10)
	itemsProcessed.Add(float64(items))

	response := fmt.Sprintf("Processed %d items\n", items)
	w.Write([]byte(response))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}

	// Seed random for simulated delays
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Health endpoints (not instrumented)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)

	// Application endpoints (instrumented)
	http.HandleFunc("/", instrumentHandler("/", rootHandler))
	http.HandleFunc("/slow", instrumentHandler("/slow", slowHandler))
	http.HandleFunc("/error", instrumentHandler("/error", errorHandler))
	http.HandleFunc("/process", instrumentHandler("/process", processHandler))

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("metrics-app %s starting on port %s", version, port)
	log.Printf("Endpoints: /, /slow, /error, /process, /metrics, /health, /ready")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
