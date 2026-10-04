package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestCheckout(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	for _, test := range []struct {
		name          string
		body          string
		paymentStatus int
		simulate      bool
		wantStatus    int
		wantCalls     int
	}{
		{"success", `{}`, 200, false, 200, 1},
		{"bad json", `{`, 200, false, 400, 0},
		{"null json", `null`, 200, false, 400, 0},
		{"multiple objects", `{} {}`, 200, false, 400, 0},
		{"unknown field", `{"unexpected":1}`, 200, false, 400, 0},
		{"simulation disabled", `{"fail":true}`, 200, false, 200, 1},
		{"simulation enabled", `{"fail":true}`, 200, true, 500, 0},
		{"downstream error", `{}`, 500, false, 502, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			payment := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodPost || r.URL.Path != "/process-payment" {
					t.Errorf("unexpected payment request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("traceparent") == "" {
					t.Error("missing distributed trace context")
				}
				w.WriteHeader(test.paymentStatus)
				_, _ = w.Write([]byte(`{"transaction_id":"txn_test"}`))
			}))
			defer payment.Close()
			app := application{
				client:      &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
				paymentsURL: payment.URL, simulate: test.simulate,
			}
			spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, TraceFlags: trace.FlagsSampled,
			})
			req := httptest.NewRequest("POST", "/checkout", strings.NewReader(test.body))
			req = req.WithContext(trace.ContextWithSpanContext(req.Context(), spanCtx))
			rr := httptest.NewRecorder()
			app.handler().ServeHTTP(rr, req)
			if rr.Code != test.wantStatus || int(calls.Load()) != test.wantCalls {
				t.Fatalf("status=%d calls=%d; want status=%d calls=%d", rr.Code, calls.Load(), test.wantStatus, test.wantCalls)
			}
		})
	}
}

func TestEndpointsAndMetrics(t *testing.T) {
	app := application{}
	handler := app.handler()
	for _, path := range []string{"/health", "/metrics"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		if rr.Code != 200 {
			t.Fatalf("%s returned %d", path, rr.Code)
		}
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("POST", "/checkout", strings.NewReader(`{`)))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(rr.Body.String(), `orders_http_requests_total{code="400",method="post"} 1`) {
		t.Fatalf("request metric missing: %s", rr.Body.String())
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/checkout", nil))
	if rr.Code != 405 {
		t.Fatalf("GET checkout must return 405, got %d", rr.Code)
	}
}

func TestTracerDoesNotBlockOnUnavailableCollector(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tp, err := initTracer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
