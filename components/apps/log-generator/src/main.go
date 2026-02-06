package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	Service    string `json:"service"`
	Endpoint   string `json:"endpoint"`
	Method     string `json:"method"`
	Status     int    `json:"status"`
	DurationMs int    `json:"duration_ms"`
	TraceID    string `json:"trace_id"`
	Message    string `json:"message"`
}

// Config holds the generator configuration
type Config struct {
	Rate        string `json:"rate"`        // low, medium, high
	Cardinality string `json:"cardinality"` // low, medium, high
	LogsPerSec  int    `json:"logs_per_sec"`
	Services    int    `json:"services"`
	mu          sync.RWMutex
}

// weightedItem for weighted random selection
type weightedItem struct {
	name   string
	weight int
}

var (
	config = &Config{
		Rate:        "low",
		Cardinality: "low",
		LogsPerSec:  10,
		Services:    5,
	}

	// Prometheus metrics
	logsGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_generator_logs_total",
			Help: "Total number of logs generated",
		},
		[]string{"level", "service"},
	)

	logsPerSecond = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "log_generator_rate",
			Help: "Current log generation rate per second",
		},
	)

	activeServices = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "log_generator_active_services",
			Help: "Number of active services generating logs",
		},
	)

	// Log levels with weights
	levels = []weightedItem{
		{"debug", 40},
		{"info", 40},
		{"warn", 15},
		{"error", 5},
	}

	// HTTP methods with weights
	methods = []weightedItem{
		{"GET", 60},
		{"POST", 25},
		{"PUT", 10},
		{"DELETE", 5},
	}

	// Endpoints
	endpoints = []string{
		"/api/v1/users",
		"/api/v1/orders",
		"/api/v1/products",
		"/api/v1/payments",
		"/api/v1/inventory",
		"/api/v1/notifications",
		"/api/v1/auth",
		"/api/v1/search",
		"/health",
		"/ready",
	}

	// Message templates
	messages = map[string][]string{
		"debug": {
			"Processing request parameters",
			"Cache lookup initiated",
			"Database query executed",
			"Response serialization started",
		},
		"info": {
			"Request processed successfully",
			"User authenticated",
			"Order created",
			"Payment processed",
		},
		"warn": {
			"Slow query detected",
			"Cache miss",
			"Rate limit approaching",
			"Deprecated API version used",
		},
		"error": {
			"Database connection failed",
			"Invalid request payload",
			"Authentication failed",
			"Service unavailable",
		},
	}
)

func init() {
	prometheus.MustRegister(logsGenerated)
	prometheus.MustRegister(logsPerSecond)
	prometheus.MustRegister(activeServices)
}

func main() {
	// Load initial config from environment
	if rate := os.Getenv("LOG_RATE"); rate != "" {
		setRate(rate)
	}
	if cardinality := os.Getenv("LOG_CARDINALITY"); cardinality != "" {
		setCardinality(cardinality)
	}

	// Start log generation
	go generateLogs()

	// HTTP handlers
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Log generator starting on port %s", port)
	log.Printf("Initial config: rate=%s, cardinality=%s, logs/sec=%d, services=%d",
		config.Rate, config.Cardinality, config.LogsPerSec, config.Services)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func setRate(rate string) {
	config.mu.Lock()
	defer config.mu.Unlock()

	switch rate {
	case "low":
		config.LogsPerSec = 10
	case "medium":
		config.LogsPerSec = 100
	case "high":
		config.LogsPerSec = 1000
	default:
		return
	}
	config.Rate = rate
	logsPerSecond.Set(float64(config.LogsPerSec))
}

func setCardinality(cardinality string) {
	config.mu.Lock()
	defer config.mu.Unlock()

	switch cardinality {
	case "low":
		config.Services = 5
	case "medium":
		config.Services = 20
	case "high":
		config.Services = 100
	default:
		return
	}
	config.Cardinality = cardinality
	activeServices.Set(float64(config.Services))
}

func generateLogs() {
	for {
		config.mu.RLock()
		logsPerSec := config.LogsPerSec
		services := config.Services
		config.mu.RUnlock()

		// Calculate interval between logs
		if logsPerSec <= 0 {
			time.Sleep(time.Second)
			continue
		}

		interval := time.Second / time.Duration(logsPerSec)

		// Generate one log entry
		entry := generateLogEntry(services)
		outputLog(entry)

		// Update metrics
		logsGenerated.WithLabelValues(entry.Level, entry.Service).Inc()

		time.Sleep(interval)
	}
}

func generateLogEntry(services int) LogEntry {
	level := weightedChoice(levels)
	method := weightedChoice(methods)
	service := fmt.Sprintf("service-%d", rand.Intn(services)+1)
	endpoint := endpoints[rand.Intn(len(endpoints))]
	status := generateStatus(level)
	duration := generateDuration(level)
	traceID := generateTraceID()
	message := messages[level][rand.Intn(len(messages[level]))]

	return LogEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Level:      level,
		Service:    service,
		Endpoint:   endpoint,
		Method:     method,
		Status:     status,
		DurationMs: duration,
		TraceID:    traceID,
		Message:    message,
	}
}

func weightedChoice(items []weightedItem) string {
	total := 0
	for _, item := range items {
		total += item.weight
	}

	r := rand.Intn(total)
	for _, item := range items {
		r -= item.weight
		if r < 0 {
			return item.name
		}
	}
	return items[0].name
}

func generateStatus(level string) int {
	switch level {
	case "error":
		statuses := []int{400, 401, 403, 404, 500, 502, 503}
		return statuses[rand.Intn(len(statuses))]
	case "warn":
		if rand.Float32() < 0.3 {
			return 429 // Rate limited
		}
		return 200
	default:
		return 200
	}
}

func generateDuration(level string) int {
	switch level {
	case "error":
		return rand.Intn(5000) + 1000 // 1000-6000ms
	case "warn":
		return rand.Intn(2000) + 500 // 500-2500ms
	default:
		return rand.Intn(200) + 10 // 10-210ms
	}
}

func generateTraceID() string {
	const chars = "abcdef0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func outputLog(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshaling log: %v", err)
		return
	}
	fmt.Println(string(data))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config.mu.RLock()
		defer config.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		rate := r.URL.Query().Get("rate")
		cardinality := r.URL.Query().Get("cardinality")

		if rate != "" {
			setRate(rate)
		}
		if cardinality != "" {
			setCardinality(cardinality)
		}

		config.mu.RLock()
		defer config.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rate":        config.Rate,
			"cardinality": config.Cardinality,
			"logs_per_sec": config.LogsPerSec,
			"services":    config.Services,
			"status":      "updated",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}
