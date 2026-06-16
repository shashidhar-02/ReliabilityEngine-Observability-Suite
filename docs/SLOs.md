# Service Level Objectives (SLOs)

## Overview
This document defines the reliability targets and error budget burn policies for the Enterprise SRE Platform.

## 1. Orders API

**SLI (Service Level Indicator):** Availability (Success Rate)
- **Metric:** `count(http_status=2xx/3xx) / total(http_requests)` over 30 days
- **SLO Target:** 99.9%
- **Monthly Error Budget:** 43.2 Minutes

## 2. Payments API

**SLI (Service Level Indicator):** Latency
- **Metric:** P95 Latency of incoming API requests over rolling 5-minute windows
- **SLO Target:** < 200ms
- **Monthly Error Budget:** 5% of total requests exceeding 200ms

## Error Budget Burn Policy

1. **Burn Rate > 1x:** Routine monitoring, added to sprint backlog.
2. **Burn Rate > 2x:** Non-urgent ticket generated in Jira.
3. **Burn Rate > 14.4x (5% consumed in 1 hr):** Critical PagerDuty alert triggered.
4. **Budget Depletion (100%):** Feature freeze. All velocity shifts to reliability.
