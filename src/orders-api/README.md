# Orders API

Go 1.26 HTTP service. Run `go test ./...`, `go vet ./...`, then `PAYMENTS_URL=http://localhost:8000 go run .`.

- `POST /checkout`: bounded JSON object, calls Payments; returns 502 on downstream failure.
- `GET /health`: liveness/readiness.
- `GET /metrics`: Prometheus business-request counters and latency histograms.

Set `OTEL_EXPORTER_OTLP_ENDPOINT` to a collector URL. `SIMULATE_FAILURES=true` enables `{"fail":true}` fault injection. See the root README and deployment guide for chart values and infrastructure requirements.
