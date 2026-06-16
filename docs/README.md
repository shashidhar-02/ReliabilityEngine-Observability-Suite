# Documentation Index

Welcome to the Enterprise SRE & Observability Platform documentation. This directory contains the architectural decisions, system design specs, reliability targets, and operational runbooks necessary to maintain and operate the platform at scale.

## Table of Contents

### 1. System Architecture
- **[System Design](system-design.md):** Comprehensive architecture and data flow diagrams showing how EKS, OpenTelemetry, and the observability backends interact.
- **[Architecture Decision Records (ADRs)](ADRs/):** Historical decisions outlining why specific technologies were chosen (e.g., [ADR-001: Why OpenTelemetry](ADRs/ADR-001-why-otel.md)).

### 2. SRE Governance
- **[Service Level Objectives (SLOs)](SLOs.md):** Defines our SLIs, SLO targets, and strict error budget burn policies for the Orders and Payments APIs.

### 3. Operational Runbooks
Detailed, step-by-step guides for on-call engineers to triage and mitigate production incidents:
- **[High Latency Triage](runbooks/high-latency-triage.md):** Troubleshooting steps for when the Payments API P95 latency exceeds 200ms.
- **[Pod CrashLoopBackOff](runbooks/pod-crash-loop.md):** Debugging guide for when Kubernetes pods fail to start or continually crash.

## How to Contribute
When implementing new features or making architectural changes, please submit a new ADR. If a new class of alert is created in Prometheus/Alertmanager, it **must** be accompanied by a new Runbook in this directory.
