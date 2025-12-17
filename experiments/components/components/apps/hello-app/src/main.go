package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Message   string `json:"message"`
	App       string `json:"app"`
	Hostname  string `json:"hostname"`
	Timestamp string `json:"timestamp"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "hello-app"
	}

	hostname, _ := os.Hostname()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Message:   fmt.Sprintf("Hello from %s!", appName),
			App:       appName,
			Hostname:  hostname,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	log.Printf("Starting %s on :%s", appName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
