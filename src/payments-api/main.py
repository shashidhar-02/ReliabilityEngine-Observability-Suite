import os
import random
import asyncio
from fastapi import FastAPI, Request, HTTPException
from opentelemetry import trace
from opentelemetry.trace.status import Status, StatusCode
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

# Initialize OpenTelemetry
resource = Resource(attributes={"service.name": "payments-api"})
trace.set_tracer_provider(TracerProvider(resource=resource))

# Configure OTLP Exporter
otel_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector.observability.svc.cluster.local:4317")
otlp_exporter = OTLPSpanExporter(endpoint=otel_endpoint, insecure=True)
span_processor = BatchSpanProcessor(otlp_exporter)
trace.get_tracer_provider().add_span_processor(span_processor)

app = FastAPI()

# Instrument FastAPI
FastAPIInstrumentor.instrument_app(app)

tracer = trace.get_tracer(__name__)

@app.post("/process-payment")
async def process_payment(request: Request):
    with tracer.start_as_current_span("process_payment") as span:
        # Simulate processing delay
        await asyncio.sleep(random.uniform(0.05, 0.2))
        
        # Simulate occasional failure (SLO burn test target)
        if random.random() < 0.05: # 5% failure rate
            span.set_status(Status(StatusCode.ERROR, "Payment processing failed due to upstream timeout"))
            span.record_exception(Exception("Payment processing failed due to upstream timeout"))
            raise HTTPException(status_code=500, detail="Internal Server Error")

        span.set_attribute("payment_status", "success")
        return {"status": "success", "transaction_id": f"txn_{random.randint(1000, 9999)}"}

@app.get("/health")
def health_check():
    return {"status": "healthy"}
