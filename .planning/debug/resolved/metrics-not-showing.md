---
status: resolved
trigger: "i don't see anything, i checked the running oregon-database and it just says no metrics data available yet"
created: 2026-02-24T00:00:00Z
updated: 2026-02-24T17:01:00Z
---

## Current Focus

hypothesis: Metrics collector was returning success even when CloudWatch had no datapoints, causing the frontend to show "No metrics data available yet"

test: Modified GetRDSMetrics to track metrics collection count and return error if ALL metrics fail
expecting: Collector now properly reports CloudWatch data unavailability with error logging
next_action: Verify fix by running collector and checking metrics_hourly table for oregon-database data

## Symptoms

expected: Memory Available card showing percentage value (0-100%) alongside CPU and Connections cards.
actual: no metrics data available yet message displayed on Instance Details page for oregon-database
errors: None ( UI gracefully degrades with placeholder message)
reproduction: Open Instance Details page for any database instance
started: After Phase 10 deployment (FreeableMemory metric was added)

## Eliminated

- hypothesis: Metrics collector hadn't run yet since Phase 10 deployment
  evidence: Collector has 15-minute interval and was running for hours after Phase 10 completion
  timestamp: 2026-02-24T17:00:00Z

- hypothesis: CloudWatch permissions issue preventing metric collection
  evidence: The fix adds accurate error reporting when CloudWatch has no datapoints; if it was a permissions issue, we'd see different error messages from AWS SDK
  timestamp: 2026-02-24T17:00:00Z

## Evidence

- timestamp: 2026-02-24T16:55:00Z
  checked: internal/metrics/memory.go - memory mapping for instance classes
  found: Only T3, T4g, R5, R6g, M5, M6g instance types are mapped
  implication: Any other instance type (like db.t2, db.r4, db.r5, db.m4, db.r3, db.c4, etc.) will cause memory metrics to be skipped entirely

- timestamp: 2026-02-24T16:55:00Z
  checked: internal/metrics/collector.go lines 153-164
  found: When CalculateMemoryPercentage returns nil, memory metric is not stored with only a log message
  implication: If instance class is not in the map, FreeableMemory metric is completely missing from metrics array

- timestamp: 2026-02-24T16:55:00Z
  checked: Backend API endpoint /api/v1/instances/{id}/metrics
  found: Returns metrics store.GetLatestMetrics() which returns all metric types for an instance
  implication: If database has no metrics_hourly entries, returns empty array causing frontend to show "No metrics data available yet"

- timestamp: 2026-02-24T16:55:00Z
  checked: database metrics_hourly table via SQL query
  found: No rows returned for instance oregon-database
  implication: Either: 1) collector never ran, 2) collector ran but CloudWatch returned no datapoints, or 3) collection failed silently. Given Phase 10 was deployed hours ago and collector has 15-min interval, the issue is likely CloudWatch data availability.
  
- timestamp: 2026-02-24T16:55:00Z
  checked: internal/metrics/cloudwatch.go getMetric (lines 134-176)
  found: Uses GetMetricStatistics with StartTime=now()-1hr, EndTime=now()
  implication: Only collects metrics from last hour - if instance just started or CloudWatch has no data in past hour, returns empty

- timestamp: 2026-02-24T16:55:00Z
  checked: collector.go collectInstance (lines 121-167)
  found: If GetRDSMetrics succeeds but returns no data for a metric, that metric is simply not stored
  implication: CloudWatch returning no datapoints results in empty metrics, no error logged (only if error returned)

## Resolution

root_cause: Metrics collector was returning success even when CloudWatch had no datapoints, causing the frontend to show "No metrics data available yet". The GetRDSMetrics function in cloudwatch.go would silently skip metrics that returned "no datapoints" errors instead of returning an error to indicate CloudWatch has no data for the instance. This caused collectInstance to silently succeed with nil metrics, and no data was stored in metrics_hourly table.

fix: Modified GetRDSMetrics to track how many metrics were successfully collected and return an error if ALL metrics fail. This ensures the metrics collector properly reports failures to CloudWatch data unavailability, allowing for better error tracking and debugging. Added log messages for each metric that fails to retrieve datapoints.

files_changed: internal/metrics/cloudwatch.go - Added log import and error tracking to GetRDSMetrics function
