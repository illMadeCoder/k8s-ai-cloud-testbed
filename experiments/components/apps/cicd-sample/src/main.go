package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Build-time variables (injected via -ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		podName := os.Getenv("POD_NAME")
		if podName == "" {
			podName = "unknown"
		}
		fmt.Fprintf(w, "Hello from cicd-sample v1! Pod: %s\n", podName)
	})

	// Liveness probe
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Readiness probe
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	// Version endpoint with build info
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"version":   Version,
			"buildTime": BuildTime,
		})
	})

	log.Printf("Starting cicd-sample server on :%s (version=%s, build=%s)", port, Version, BuildTime)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
