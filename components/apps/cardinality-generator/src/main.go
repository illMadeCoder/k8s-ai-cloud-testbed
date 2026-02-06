package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Cardinality levels define how many unique time series to generate
var cardinalityLevels = map[string]int{
	"low":    1000,   // 1k series - baseline
	"medium": 10000,  // 10k series - moderate stress
	"high":   50000,  // 50k series - heavy stress
}

// Global state
var (
	currentLevel    string = "low"
	currentSeries   int    = 1000
	mu              sync.RWMutex
	registry        *prometheus.Registry
	lastMetricsBuild time.Time
)

// Metrics (rebuilt when cardinality changes)
var (
	sensorReading *prometheus.GaugeVec
	sensorCounter *prometheus.CounterVec
	sensorLatency *prometheus.HistogramVec
)

func buildMetrics() {
	mu.Lock()
	defer mu.Unlock()

	// Create a new registry each time to allow re-registration
	registry = prometheus.NewRegistry()

	// Gauge: sensor_reading - main cardinality driver
	sensorReading = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensor_reading",
			Help: "Current sensor reading value",
		},
		[]string{"deck", "section", "sensor_type"},
	)

	// Counter: sensor_events_total
	sensorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensor_events_total",
			Help: "Total sensor events recorded",
		},
		[]string{"deck", "section", "event_type"},
	)

	// Histogram: sensor_response_seconds
	sensorLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sensor_response_seconds",
			Help:    "Sensor response latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"deck", "sensor_type"},
	)

	registry.MustRegister(sensorReading)
	registry.MustRegister(sensorCounter)
	registry.MustRegister(sensorLatency)

	// Also register some basic process metrics
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())

	lastMetricsBuild = time.Now()
	log.Printf("Metrics registry rebuilt for %d series", currentSeries)
}

// Generate label combinations up to targetSeries
func generateMetrics() {
	mu.RLock()
	target := currentSeries
	mu.RUnlock()

	// Calculate dimensions needed for target cardinality
	// deck * section * type = target
	// We'll use: deck=1-50, section=a-z (26), type=1-N
	// 50 * 26 * N = target, so N = target / 1300

	decks := 50
	sections := 26
	types := (target / (decks * sections)) + 1

	if types < 1 {
		types = 1
	}
	if types > 100 {
		types = 100
		decks = target / (sections * types)
		if decks > 100 {
			decks = 100
		}
	}

	generated := 0
	for d := 1; d <= decks && generated < target; d++ {
		for s := 0; s < sections && generated < target; s++ {
			section := string(rune('a' + s))
			for t := 1; t <= types && generated < target; t++ {
				deck := strconv.Itoa(d)
				sensorType := fmt.Sprintf("type_%d", t)

				// Gauge: random sensor reading
				sensorReading.WithLabelValues(deck, section, sensorType).Set(
					rand.Float64()*100 + float64(d),
				)

				// Counter: occasional events
				if rand.Float64() < 0.1 {
					eventType := []string{"alert", "warning", "info"}[rand.Intn(3)]
					sensorCounter.WithLabelValues(deck, section, eventType).Add(1)
				}

				// Histogram: latency observation
				if rand.Float64() < 0.3 {
					sensorLatency.WithLabelValues(deck, sensorType).Observe(
						rand.Float64() * 0.1,
					)
				}

				generated++
			}
		}
	}
}

func setCardinality(level string) error {
	series, ok := cardinalityLevels[level]
	if !ok {
		return fmt.Errorf("unknown cardinality level: %s (use low/medium/high)", level)
	}

	mu.Lock()
	currentLevel = level
	currentSeries = series
	mu.Unlock()

	// Rebuild metrics with new cardinality
	buildMetrics()
	generateMetrics()

	log.Printf("Cardinality set to %s (%d series)", level, series)
	return nil
}

func cardinalityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		mu.RLock()
		response := map[string]interface{}{
			"level":       currentLevel,
			"series":      currentSeries,
			"levels":      cardinalityLevels,
			"last_update": lastMetricsBuild.Format(time.RFC3339),
		}
		mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if r.Method == http.MethodPost {
		level := r.URL.Query().Get("level")
		if level == "" {
			http.Error(w, "missing 'level' query parameter", http.StatusBadRequest)
			return
		}

		if err := setCardinality(level); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.RLock()
		response := map[string]interface{}{
			"status":  "updated",
			"level":   currentLevel,
			"series":  currentSeries,
		}
		mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	mu.RLock()
	response := fmt.Sprintf(`Cardinality Generator
=====================
Current Level: %s
Active Series: %d

Available Levels:
- low:    1,000 series
- medium: 10,000 series
- high:   50,000 series

Endpoints:
- GET  /cardinality     - View current settings
- POST /cardinality?level=<level> - Change cardinality
- GET  /metrics         - Prometheus metrics
- GET  /health          - Health check
- GET  /ready           - Readiness check

Example:
  curl -X POST "http://localhost:8080/cardinality?level=high"
`, currentLevel, currentSeries)
	mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(response))
}

func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Regenerate metrics on each scrape to simulate changing values
		generateMetrics()

		mu.RLock()
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		mu.RUnlock()

		handler.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	initialLevel := os.Getenv("CARDINALITY_LEVEL")
	if initialLevel == "" {
		initialLevel = "low"
	}

	// Seed random
	rand.Seed(time.Now().UnixNano())

	// Initialize metrics
	buildMetrics()
	if err := setCardinality(initialLevel); err != nil {
		log.Printf("Warning: %v, using default 'low'", err)
		setCardinality("low")
	}

	// HTTP handlers
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/cardinality", cardinalityHandler)
	http.Handle("/metrics", metricsHandler())

	log.Printf("Cardinality Generator starting on port %s", port)
	log.Printf("Initial cardinality: %s (%d series)", currentLevel, currentSeries)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
