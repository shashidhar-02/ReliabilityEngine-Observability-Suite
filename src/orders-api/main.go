package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// SDK construction is non-blocking: telemetry is not a startup dependency.
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String("orders-api"))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

type OrderRequest struct {
	Fail bool `json:"fail"`
}

type application struct {
	client      *http.Client
	paymentsURL string
	simulate    bool
}

func (a application) checkoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("orders-api").Start(r.Context(), "checkout")
	defer span.End()
	var order *OrderRequest
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&order); err != nil || order == nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}
	if decoder.Decode(new(any)) != io.EOF {
		http.Error(w, "expected a single JSON object", http.StatusBadRequest)
		return
	}
	if order.Fail && a.simulate {
		span.SetStatus(codes.Error, "simulated failure")
		http.Error(w, "simulated failure", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.paymentsURL+"/process-payment", nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid payment endpoint")
		http.Error(w, "payment service unavailable", http.StatusBadGateway)
		return
	}
	resp, err := a.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment request failed")
		http.Error(w, "payment service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "payment service returned an error")
		http.Error(w, "payment failed", http.StatusBadGateway)
		return
	}
	var payment struct {
		TransactionID string `json:"transaction_id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&payment); err != nil || payment.TransactionID == "" {
		span.SetStatus(codes.Error, "invalid payment response")
		http.Error(w, "invalid payment response", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "Order processed successfully", "transaction_id": payment.TransactionID,
	}); err != nil {
		log.Printf("write checkout response: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("OK"))
}

func (a application) handler() http.Handler {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_http_requests_total", Help: "Business HTTP requests by status code.",
	}, []string{"code", "method"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "orders_http_request_duration_seconds", Help: "Checkout request duration.",
		Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5},
	}, []string{"method"})
	registry.MustRegister(requests, duration)
	mux := http.NewServeMux()
	var checkout http.Handler = otelhttp.NewHandler(http.HandlerFunc(a.checkoutHandler), "checkout")
	checkout = promhttp.InstrumentHandlerDuration(duration,
		promhttp.InstrumentHandlerCounter(requests, checkout))
	mux.Handle("POST /checkout", checkout)
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	return mux
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	tp, err := initTracer(ctx)
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(flushCtx); err != nil {
			log.Printf("flush traces: %v", err)
		}
	}()
	paymentsURL := os.Getenv("PAYMENTS_URL")
	if paymentsURL == "" {
		paymentsURL = "http://localhost:8000"
	}
	simulate, _ := strconv.ParseBool(os.Getenv("SIMULATE_FAILURES"))
	app := application{
		client:      &http.Client{Timeout: 3 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)},
		paymentsURL: paymentsURL, simulate: simulate,
	}
	srv := &http.Server{
		Addr: ":8080", Handler: app.handler(), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- srv.ListenAndServe() }()
	log.Print("orders-api listening on :8080")
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
