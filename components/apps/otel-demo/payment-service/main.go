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
			semconv.ServiceName("payment-service"),
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

	tracer = tp.Tracer("payment-service")
	return tp, nil
}

// Payment represents a payment transaction
type Payment struct {
	ID         string  `json:"payment_id"`
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	Provider   string  `json:"provider"`
	ProviderID string  `json:"provider_id,omitempty"`
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

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ready", handleReady)
	mux.Handle("/api/payments", otelhttp.NewHandler(
		http.HandlerFunc(handleProcessPayment),
		"ProcessPayment",
	))
	mux.Handle("/api/payments/", otelhttp.NewHandler(
		http.HandlerFunc(handleGetPayment),
		"GetPayment",
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Payment service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "payment-service",
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

func handleGetPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	paymentID := r.URL.Path[len("/api/payments/"):]
	span.SetAttributes(attribute.String("payment.id", paymentID))

	// Simulate database lookup
	_, dbSpan := tracer.Start(ctx, "database.get_payment")
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "payments"),
	)
	dbSpan.End()

	payment := Payment{
		ID:       paymentID,
		OrderID:  "ord-12345",
		Amount:   99.99,
		Status:   "completed",
		Provider: "stripe",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func handleProcessPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var paymentReq struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
		UserID  string  `json:"user_id"`
	}
	body, _ := io.ReadAll(r.Body)
	if len(body) > 0 {
		json.Unmarshal(body, &paymentReq)
	}

	// Generate payment
	payment := Payment{
		ID:       fmt.Sprintf("pay-%d", rand.Intn(100000)),
		OrderID:  paymentReq.OrderID,
		UserID:   paymentReq.UserID,
		Amount:   paymentReq.Amount,
		Status:   "processing",
		Provider: "stripe",
	}

	span.SetAttributes(
		attribute.String("payment.id", payment.ID),
		attribute.String("order.id", payment.OrderID),
		attribute.Float64("payment.amount", payment.Amount),
		attribute.String("payment.provider", payment.Provider),
	)

	// Validate payment details
	ctx, validateSpan := tracer.Start(ctx, "validate_payment")
	time.Sleep(time.Duration(5+rand.Intn(10)) * time.Millisecond)
	validateSpan.SetAttributes(attribute.Float64("amount", payment.Amount))
	validateSpan.End()

	// Fraud check
	ctx, fraudSpan := tracer.Start(ctx, "fraud_check")
	time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)
	fraudSpan.SetAttributes(
		attribute.String("check.type", "ml_model"),
		attribute.Float64("risk_score", rand.Float64()*0.3),
	)

	// Simulate occasional fraud detection
	if rand.Float32() < 0.02 {
		fraudSpan.SetStatus(codes.Error, "fraud detected")
		fraudSpan.RecordError(fmt.Errorf("payment flagged as potential fraud"))
		fraudSpan.End()
		span.SetStatus(codes.Error, "payment rejected - fraud")
		payment.Status = "rejected"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(payment)
		return
	}
	fraudSpan.End()

	// Call external payment provider (simulated)
	ctx, providerSpan := tracer.Start(ctx, "call_payment_provider")
	providerSpan.SetAttributes(
		attribute.String("provider", "stripe"),
		attribute.String("provider.endpoint", "api.stripe.com"),
	)

	// Simulate variable latency for payment processing
	processingTime := time.Duration(50+rand.Intn(150)) * time.Millisecond
	time.Sleep(processingTime)

	// Simulate occasional payment failures
	if rand.Float32() < 0.05 {
		providerSpan.SetStatus(codes.Error, "payment declined")
		providerSpan.RecordError(fmt.Errorf("card declined by issuer"))
		providerSpan.End()
		span.SetStatus(codes.Error, "payment failed")
		payment.Status = "declined"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(payment)
		return
	}

	payment.ProviderID = fmt.Sprintf("pi_%s", randomString(24))
	payment.Status = "completed"
	providerSpan.SetAttributes(attribute.String("provider.transaction_id", payment.ProviderID))
	providerSpan.End()

	// Save payment to database
	ctx, dbSpan := tracer.Start(ctx, "database.insert_payment")
	time.Sleep(time.Duration(15+rand.Intn(20)) * time.Millisecond)
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.table", "payments"),
	)
	dbSpan.End()

	// Send confirmation (async simulation)
	ctx, notifySpan := tracer.Start(ctx, "send_confirmation")
	notifySpan.SetAttributes(
		attribute.String("notification.type", "email"),
		attribute.String("notification.channel", "async"),
	)
	time.Sleep(time.Duration(5+rand.Intn(10)) * time.Millisecond)
	notifySpan.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
