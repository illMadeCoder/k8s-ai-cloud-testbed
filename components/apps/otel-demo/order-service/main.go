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

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.otel-collector:4317"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("order-service"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "tutorial"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

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

	tracer = tp.Tracer("order-service")
	return tp, nil
}

// Order represents an order in the system
type Order struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	Items     []string `json:"items"`
	Status    string   `json:"status"`
	Total     float64  `json:"total"`
	PaymentID string   `json:"payment_id,omitempty"`
}

// PaymentResponse from payment-service
type PaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Amount    float64 `json:"amount"`
}

func main() {
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

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)
	mux.Handle("/api/orders", otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				handleCreateOrder(w, r, httpClient)
			} else {
				handleGetOrders(w, r)
			}
		}),
		"Orders",
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "order-service",
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

func handleGetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	userID := r.URL.Query().Get("user_id")
	span.SetAttributes(attribute.String("user.id", userID))

	// Simulate database query
	ctx, dbSpan := tracer.Start(ctx, "database.query_orders")
	time.Sleep(time.Duration(15+rand.Intn(25)) * time.Millisecond)
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "orders"),
	)
	dbSpan.End()

	// Return mock orders
	orders := []Order{
		{
			ID:     fmt.Sprintf("ord-%d", rand.Intn(10000)),
			UserID: userID,
			Items:  []string{"item-1", "item-2"},
			Status: "completed",
			Total:  99.99,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request, client *http.Client) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	// Parse request
	var orderReq struct {
		UserID string   `json:"user_id"`
		Items  []string `json:"items"`
	}
	body, _ := io.ReadAll(r.Body)
	if len(body) > 0 {
		json.Unmarshal(body, &orderReq)
	}

	// Generate order
	order := Order{
		ID:     fmt.Sprintf("ord-%d", rand.Intn(100000)),
		UserID: orderReq.UserID,
		Items:  orderReq.Items,
		Status: "pending",
		Total:  float64(len(orderReq.Items)) * 29.99,
	}

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("user.id", order.UserID),
		attribute.Float64("order.total", order.Total),
		attribute.Int("order.item_count", len(order.Items)),
	)

	// Validate inventory
	ctx, inventorySpan := tracer.Start(ctx, "validate_inventory")
	time.Sleep(time.Duration(10+rand.Intn(15)) * time.Millisecond)
	inventorySpan.SetAttributes(attribute.Int("items.count", len(order.Items)))

	// Simulate occasional inventory issues
	if rand.Float32() < 0.03 {
		inventorySpan.SetStatus(codes.Error, "insufficient inventory")
		inventorySpan.RecordError(fmt.Errorf("insufficient inventory for items"))
		inventorySpan.End()
		span.SetStatus(codes.Error, "order failed - inventory")
		http.Error(w, "Insufficient inventory", http.StatusConflict)
		return
	}
	inventorySpan.End()

	// Save order to database
	ctx, dbSpan := tracer.Start(ctx, "database.insert_order")
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "orders"),
	)
	dbSpan.End()

	// Call payment service
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://payment-service:8080"
	}

	paymentReq := map[string]interface{}{
		"order_id": order.ID,
		"amount":   order.Total,
		"user_id":  order.UserID,
	}
	paymentBody, _ := json.Marshal(paymentReq)

	ctx, paymentSpan := tracer.Start(ctx, "call_payment_service")
	req, _ := http.NewRequestWithContext(ctx, "POST", paymentServiceURL+"/api/payments",
		io.NopCloser(io.Reader(nil)))
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(io.Reader(nil))
	_ = paymentBody // simplified for demo

	resp, err := client.Do(req)
	if err != nil {
		paymentSpan.SetStatus(codes.Error, err.Error())
		paymentSpan.RecordError(err)
		paymentSpan.End()
		log.Printf("Error calling payment service: %v", err)
		order.Status = "payment_pending"
	} else {
		defer resp.Body.Close()
		paymentSpan.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		paymentSpan.End()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var paymentResp PaymentResponse
			json.NewDecoder(resp.Body).Decode(&paymentResp)
			order.PaymentID = paymentResp.PaymentID
			order.Status = "paid"
		} else {
			order.Status = "payment_failed"
		}
	}

	// Update order status
	ctx, updateSpan := tracer.Start(ctx, "database.update_order")
	time.Sleep(time.Duration(10+rand.Intn(15)) * time.Millisecond)
	updateSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "UPDATE"),
		attribute.String("order.status", order.Status),
	)
	updateSpan.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
