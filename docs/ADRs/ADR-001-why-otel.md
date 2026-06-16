# ADR-001: Adoption of OpenTelemetry for Unified Observability

**Status:** Accepted
**Date:** 2026-06-16

## Context
Our previous architecture relied on fragmented, vendor-specific agents for observability:
- Datadog agent for metrics
- FluentBit for logs mapping to ElasticSearch
- AWS X-Ray daemon for traces

This fragmentation resulted in high operational overhead, vendor lock-in, and an inability to seamlessly correlate a specific trace with its corresponding log entry and metric spike within a single pane of glass.

## Decision
We will standardize on **OpenTelemetry (OTel)** as the single, vendor-agnostic standard for collecting Metrics, Logs, and Traces across all microservices. 

1. All applications (Go, Python) will use native OTel SDKs to generate OTLP data.
2. We will deploy the OpenTelemetry Collector in the EKS cluster to receive, process, and route this data.

## Consequences
### Positive
- **Vendor Neutrality:** We can switch storage backends (e.g., from Prometheus to Datadog) by simply changing the OTel Collector exporter config, requiring zero code changes in the applications.
- **Unified Context:** W3C Trace Context is propagated automatically, allowing us to attach `trace_id` to both logs and metrics for instant correlation.
- **Performance:** A single agent (OTel Collector) reduces the resource overhead on our EKS worker nodes compared to running multiple vendor-specific daemonsets.

### Negative
- **Learning Curve:** Developers must learn the OpenTelemetry API constructs (TracerProvider, Span Processors) instead of using simpler, proprietary SDKs.
- **Log Maturity:** While OTel Traces and Metrics are stable, the Logging signal is still maturing, requiring some custom parsing in the collector.
