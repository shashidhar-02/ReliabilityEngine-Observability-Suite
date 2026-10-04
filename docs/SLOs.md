# Service level objectives

These are proposed objectives and usable PromQL examples. The repository does not install recording rules, Alertmanager, paging, or a deployment freeze.

## Orders availability

Target: 99.9% successful checkout requests over 30 days. Exclude health/metrics requests; use business-request counters. Availability is request-based, so its budget is 0.1% of requests, not automatically a downtime duration.

```promql
sum(increase(orders_http_requests_total{code=~"2..|3.."}[30d]))
/
sum(increase(orders_http_requests_total[30d]))
```

For time-based 99.9% availability, the equivalent 30-day budget is 43.2 minutes, but a separate time-based SLI is needed.

## Payments latency

Target: at least 95% of payment requests complete within 200 ms over 30 days. The histogram includes an explicit 0.2-second bucket:

```promql
sum(increase(payments_http_request_duration_seconds_bucket{path="/process-payment",le="0.2"}[30d]))
/
sum(increase(payments_http_request_duration_seconds_count{path="/process-payment"}[30d]))
```

Use P95 over five minutes for triage, not as a substitute for the 30-day budget:

```promql
histogram_quantile(0.95,
  sum by (le) (rate(payments_http_request_duration_seconds_bucket{path="/process-payment"}[5m])))
```

## Burn policy

Burn rate is observed bad-request fraction divided by the SLO's allowed bad fraction. A 14.4x burn over one hour consumes 2% of a 30-day budget, not 5%. Pair long and short windows to reduce false positives. Decide ticket/paging routes, no-traffic handling, and deployment policy before enabling enforcement.
