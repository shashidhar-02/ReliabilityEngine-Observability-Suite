# System design

Orders serves POST `/checkout` on port 8080, validates a bounded JSON body, and calls Payments at POST `/process-payment` on port 8000. Requests propagate W3C trace context with an instrumented HTTP client. Downstream failures produce a 502 response rather than a false successful checkout. Payment IDs are generated UUIDs; there is no financial provider or order database.

Both services expose business-request Prometheus counters and histograms with bounded labels. Health probes and metric scrapes do not contribute to business SLOs. OTLP tracing is asynchronous. Orders handles SIGTERM; Payments flushes its provider in the ASGI lifespan shutdown.

The Kubernetes base creates the restricted observability namespace and a two-replica OpenTelemetry collector. Its trace pipeline applies memory limiting and batching and uses the debug exporter. Prometheus, Grafana, Jaeger, Loki, durable storage, and alert routing are integration points rather than pre-installed resources.

Application Helm charts create real service accounts, HPAs, and PDBs. They apply a non-root user, RuntimeDefault seccomp, dropped capabilities, read-only filesystems, probes, and bounded resources. The default namespace denies network traffic unless an allow rule applies: DNS, Orders→Payments, OTLP to the collector, and scraping from observability are explicitly allowed. External ingress requires a gateway and corresponding policy.

Terraform defines separate production/staging VPCs and state, a private EKS control plane, managed worker nodes, secret encryption, flow logs, and IRSA support. Production ECR repositories have immutable tags and scan-on-push. The private control plane requires a VPC-connected deployment runner.

Repository Quality checks source, dependencies, infrastructure, manifests, workflows, and image builds/scans. Deployment workflows are dispatched on main and consume that validation workflow. AWS credentials use environment-scoped OIDC. GitHub settings, AWS bootstrap roles, state backend resources, cluster authorization, metrics-server, and durable observability backends must be supplied as documented in [deployment.md](deployment.md).
