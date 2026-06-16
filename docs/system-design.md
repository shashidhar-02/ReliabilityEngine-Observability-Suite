# System Design & Architecture

This document outlines the detailed system design of the Enterprise SRE Observability Platform.

## High-Level Architecture

The platform relies on a decoupled, microservice-based architecture deployed on AWS EKS, utilizing an OpenTelemetry (OTel) pipeline for vendor-agnostic telemetry ingestion.

```mermaid
flowchart TB
    subgraph AWS Cloud [AWS Cloud Environment]
        subgraph VPC [Multi-AZ VPC]
            subgraph Public Subnets
                NAT[NAT Gateway]
                IGW[Internet Gateway]
            end
            
            subgraph Private Subnets [Private Subnets - EKS Node Group]
                subgraph App_Namespace [Namespace: default]
                    OrdersAPI[Orders API - Go]
                    PaymentsAPI[Payments API - Python]
                end
                
                subgraph Obs_Namespace [Namespace: observability]
                    OTelCollector[OpenTelemetry Collector]
                    Prometheus[(Prometheus TSDB)]
                    Jaeger[(Jaeger Traces)]
                    Loki[(Grafana Loki)]
                    Grafana[Grafana Dashboards]
                end
            end
        end
    end

    %% Traffic Flow
    Internet((Internet)) --> IGW
    IGW --> OrdersAPI
    OrdersAPI -->|gRPC/HTTP| PaymentsAPI

    %% Telemetry Flow
    OrdersAPI -.->|OTLP Traces/Metrics| OTelCollector
    PaymentsAPI -.->|OTLP Traces/Metrics| OTelCollector
    
    OTelCollector -.->|Metrics| Prometheus
    OTelCollector -.->|Traces| Jaeger
    OTelCollector -.->|Logs| Loki
    
    Grafana --> Prometheus
    Grafana --> Jaeger
    Grafana --> Loki
```

## Component Details

1. **Microservices (Orders & Payments):** Instrumented using the OpenTelemetry SDKs (`go.opentelemetry.io` and `opentelemetry-instrumentation-fastapi`). They export telemetry data via the OTLP gRPC protocol to the local collector.
2. **OpenTelemetry Collector:** Deployed as a DaemonSet or Deployment in the `observability` namespace. It receives OTLP data, processes it (batching, filtering), and exports it to the specific backends.
3. **Storage Backends:**
   - **Prometheus:** Stores time-series metrics for SLI/SLO tracking and Alertmanager evaluation.
   - **Jaeger:** Stores distributed traces to provide visualizations of microservice dependency graphs and latency bottlenecks.
   - **Loki:** Stores structured JSON logs via Promtail scraping.
4. **Grafana:** The single pane of glass for visualizing all three pillars of observability (Metrics, Logs, Traces) through unified dashboards.
