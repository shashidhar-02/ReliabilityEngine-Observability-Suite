"""Sample payments API with bounded-cardinality metrics and OTLP tracing."""

import asyncio
from contextlib import asynccontextmanager
import os
import secrets
import time
import uuid

from fastapi import FastAPI, HTTPException, Request, Response
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.trace.status import Status, StatusCode
from prometheus_client import (
    CONTENT_TYPE_LATEST, CollectorRegistry, Counter, Histogram, generate_latest,
)


def create_app(*, enable_tracing=True, simulate_failures=None):
    """Create an independent application and registry, including for HTTP tests."""
    simulate = (
        os.getenv("SIMULATE_FAILURES", "false").lower() == "true"
        if simulate_failures is None else simulate_failures
    )
    provider = TracerProvider(resource=Resource.create({"service.name": "payments-api"}))

    @asynccontextmanager
    async def lifespan(_app):
        if enable_tracing:
            provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(timeout=3)))
        try:
            yield
        finally:
            await asyncio.to_thread(provider.shutdown)

    app = FastAPI(lifespan=lifespan)
    registry = CollectorRegistry()
    requests = Counter(
        "payments_http_requests_total", "Business HTTP requests by status code.",
        ("method", "path", "status"), registry=registry,
    )
    duration = Histogram(
        "payments_http_request_duration_seconds", "Business HTTP request duration.",
        ("method", "path"), buckets=(0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5), registry=registry,
    )
    tracer = provider.get_tracer(__name__)

    @app.middleware("http")
    async def record_metrics(request: Request, call_next):
        if request.url.path in {"/health", "/metrics"}:
            return await call_next(request)
        started = time.perf_counter()
        status = 500
        try:
            response = await call_next(request)
            status = response.status_code
            return response
        finally:
            route = request.scope.get("route")
            path = getattr(route, "path", "unmatched")
            method = request.method if request.method in {
                "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
            } else "OTHER"
            requests.labels(method, path, str(status)).inc()
            duration.labels(method, path).observe(time.perf_counter() - started)

    @app.post("/process-payment", responses={500: {"description": "Simulated payment failure"}})
    async def process_payment():
        with tracer.start_as_current_span("process_payment") as span:
            if simulate:
                await asyncio.sleep((50 + secrets.randbelow(151)) / 1000)
                if secrets.randbelow(100) < 5:
                    message = "simulated payment failure"
                    span.set_status(Status(StatusCode.ERROR, message))
                    raise HTTPException(status_code=500, detail=message)
            span.set_attribute("payment_status", "success")
            return {"status": "success", "transaction_id": f"txn_{uuid.uuid4().hex}"}

    @app.get("/health")
    def health_check():
        return {"status": "healthy"}

    @app.get("/metrics", include_in_schema=False)
    def metrics():
        return Response(content=generate_latest(registry), media_type=CONTENT_TYPE_LATEST)

    if enable_tracing:
        FastAPIInstrumentor.instrument_app(
            app, tracer_provider=provider, excluded_urls="health,metrics",
        )
    return app


app = create_app()
