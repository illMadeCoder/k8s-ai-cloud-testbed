package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	opsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "customdb_operations_total",
			Help: "Total key-value operations",
		},
		[]string{"op", "status"},
	)
	opDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "customdb_operation_duration_seconds",
			Help:    "Operation latency",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .005, .01, .05},
		},
		[]string{"op"},
	)
	storeKeys = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "customdb_keys",
		Help: "Number of keys in store",
	})
)

func init() {
	prometheus.MustRegister(opsTotal, opDuration, storeKeys)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := NewMemStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/kv/", kvHandler(store))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("custom-db listening on :%s (engine=memstore)", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func kvHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/kv/")
		if key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			start := time.Now()
			val, ok := store.Get(key)
			opDuration.WithLabelValues("get").Observe(time.Since(start).Seconds())
			if !ok {
				opsTotal.WithLabelValues("get", "miss").Inc()
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			opsTotal.WithLabelValues("get", "hit").Inc()
			w.Write(val)

		case http.MethodPut:
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB max
			if err != nil {
				http.Error(w, "read error", http.StatusBadRequest)
				return
			}
			start := time.Now()
			store.Put(key, body)
			opDuration.WithLabelValues("put").Observe(time.Since(start).Seconds())
			opsTotal.WithLabelValues("put", "ok").Inc()
			storeKeys.Set(float64(store.Len()))
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			start := time.Now()
			ok := store.Delete(key)
			opDuration.WithLabelValues("delete").Observe(time.Since(start).Seconds())
			if ok {
				opsTotal.WithLabelValues("delete", "ok").Inc()
			} else {
				opsTotal.WithLabelValues("delete", "miss").Inc()
			}
			storeKeys.Set(float64(store.Len()))
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, fmt.Sprintf("method %s not allowed", r.Method), http.StatusMethodNotAllowed)
		}
	}
}
