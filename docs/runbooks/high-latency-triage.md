# High Latency Triage Runbook

## Description
This runbook is triggered when the Payments API P95 latency exceeds the **200ms** threshold over a rolling 5-minute window, threatening the monthly Error Budget.

## Triage Workflow Diagram

```mermaid
stateDiagram-v2
    [*] --> AlertReceived
    AlertReceived --> CheckGrafana: Acknowledge PagerDuty
    CheckGrafana --> IsolateComponent: View "Payments API Latency"
    
    IsolateComponent --> SystemWideLatency: CPU/Memory Spikes?
    IsolateComponent --> SpecificEndpoint: Only /process-payment?
    
    SystemWideLatency --> ScalePods: Horizontal Pod Autoscaler failed?
    ScalePods --> Resolve
    
    SpecificEndpoint --> CheckJaeger: Find Trace ID > 200ms
    CheckJaeger --> DatabaseBottleneck: DB Span > 150ms
    CheckJaeger --> UpstreamTimeout: Third-party API failed
    
    DatabaseBottleneck --> FixDBPools: Increase connection pool
    UpstreamTimeout --> ImplementCircuitBreaker: Block bad upstream
    
    FixDBPools --> Resolve
    ImplementCircuitBreaker --> Resolve
    Resolve --> [*]
```

## Step-by-Step Triage
1. **Acknowledge Alert:** Immediately acknowledge the PagerDuty alert to prevent escalation.
2. **Isolate Component:** Check the Grafana dashboard for the "Payments API" panel. Determine if the latency is isolated to a specific endpoint (e.g., `/process-payment`) or is system-wide.
3. **Trace Analysis (Jaeger):** 
   - Open the Jaeger UI.
   - Filter for traces in the `payments-api` service where `Duration > 200ms`.
   - Inspect the span dependency graph to identify the exact bottleneck (e.g., a slow `SELECT` query or a timed-out external API call).
4. **Resource Verification (Prometheus):** 
    - Check kubelet/cAdvisor metrics for CPU throttling (`container_cpu_cfs_throttled_seconds_total`). kube-state-metrics does not provide container CPU throttling counters.
5. **Mitigation Strategies:**
   - **If CPU throttled:** Review `src/payments-api/chart/values-prod.yaml` and the HPA resource metrics API before changing limits through a reviewed release.
   - **If DB bound:** Check Postgres/MySQL dashboards for connection pool exhaustion or deadlocks.
   - **If Upstream API bound:** Ensure circuit breakers are functioning properly to fail-fast rather than hanging.
