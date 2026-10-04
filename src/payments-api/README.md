# Payments API

Python 3.11 FastAPI reference service. Install `requirements-dev.lock` with `--only-binary=:all: --require-hashes` in a virtual environment, run `python -m pytest`, `flake8 .`, and `uvicorn main:app --host 127.0.0.1 --port 8000`.

- `POST /process-payment`: returns a synthetic UUID transaction ID; does not perform a financial transaction.
- `GET /health`: liveness/readiness.
- `GET /metrics`: bounded-cardinality Prometheus request and latency metrics.

Set `OTEL_EXPORTER_OTLP_ENDPOINT` to a collector URL. `SIMULATE_FAILURES=true` opts into random delay/failure injection. It is disabled by default.
