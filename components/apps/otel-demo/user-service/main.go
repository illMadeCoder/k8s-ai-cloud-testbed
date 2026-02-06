package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Get OTLP endpoint from environment
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.otel-collector:4317"
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service info
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("user-service"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "tutorial"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer("user-service")
	return tp, nil
}

// User represents a user in the system
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// OrderResponse from order-service
type OrderResponse struct {
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
	PaymentID string  `json:"payment_id,omitempty"`
}

func main() {
	// Initialize tracer
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	// Create HTTP client with tracing
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	// Set up routes with tracing middleware
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)
	mux.Handle("/api/users/", otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleGetUser(w, r, httpClient)
		}),
		"GetUser",
	))
	mux.Handle("/api/users", otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleCreateUser(w, r, httpClient)
		}),
		"CreateUser",
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("User service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "user-service",
		"status":  "running",
		"version": "1.0.0",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func handleGetUser(w http.ResponseWriter, r *http.Request, client *http.Client) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	// Extract user ID from path
	userID := r.URL.Path[len("/api/users/"):]
	span.SetAttributes(attribute.String("user.id", userID))

	// Simulate user lookup with custom span
	ctx, lookupSpan := tracer.Start(ctx, "database.lookup_user")
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
	lookupSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)
	lookupSpan.End()

	// Simulate occasional errors
	if rand.Float32() < 0.05 {
		span.SetStatus(codes.Error, "user not found")
		span.RecordError(fmt.Errorf("user %s not found", userID))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user := User{
		ID:    userID,
		Name:  fmt.Sprintf("User %s", userID),
		Email: fmt.Sprintf("user%s@example.com", userID),
	}

	// Call order-service to get user's orders
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://order-service:8080"
	}

	ctx, orderSpan := tracer.Start(ctx, "call_order_service")
	req, _ := http.NewRequestWithContext(ctx, "GET", orderServiceURL+"/api/orders?user_id="+userID, nil)
	resp, err := client.Do(req)
	if err != nil {
		orderSpan.SetStatus(codes.Error, err.Error())
		orderSpan.RecordError(err)
		orderSpan.End()
		log.Printf("Error calling order service: %v", err)
	} else {
		defer resp.Body.Close()
		orderSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		orderSpan.End()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, client *http.Client) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var user User
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &user); err != nil {
		span.SetStatus(codes.Error, "invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate user ID
	user.ID = fmt.Sprintf("%d", rand.Intn(10000))
	span.SetAttributes(attribute.String("user.id", user.ID))

	// Simulate database insert
	ctx, insertSpan := tracer.Start(ctx, "database.insert_user")
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
	insertSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "users"),
	)
	insertSpan.End()

	// Create initial order for new user
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://order-service:8080"
	}

	orderData := map[string]interface{}{
		"user_id": user.ID,
		"items":   []string{"welcome-gift"},
	}
	orderBody, _ := json.Marshal(orderData)

	ctx, orderSpan := tracer.Start(ctx, "create_welcome_order")
	req, _ := http.NewRequestWithContext(ctx, "POST", orderServiceURL+"/api/orders",
		io.NopCloser(io.Reader(nil)))
	req.Body = io.NopCloser(io.Reader(nil))
	// Note: simplified for demo - real implementation would send orderBody
	_ = orderBody

	resp, err := client.Do(req)
	if err != nil {
		orderSpan.SetStatus(codes.Error, err.Error())
		log.Printf("Error creating welcome order: %v", err)
	} else {
		resp.Body.Close()
		orderSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	}
	orderSpan.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
