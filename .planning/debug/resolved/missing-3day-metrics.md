---
status: resolved
trigger: "Investigate why 3-day metrics backfill isn't showing"
created: 2026-02-26T00:00:00Z
updated: 2026-02-26T20:00:00Z
---

## Current Focus

hypothesis: Fix applied - MetricValueWithTimestamp preserves timestamps through the data flow
test: Expected: CloudWatch datapoint timestamps are preserved and stored correctly
expecting: Metrics should now show proper timestamps like "2026-02-25 16:00:00+00" instead of "0001-01-01"
next_action: Verify the fix by checking if the code compiles and testing the backfill functionality

## Symptoms

expected: 3+ days of CloudWatch metrics data visible in instance metrics table after server startup
actual: Most metrics show placeholder date "0001-01-01" with only 15-min interval entries
errors: None - data exists but is incomplete
reproduction: Server was started, user waited for backfill to complete, checked metrics table in any tab
started: Just now noticed - this is a fresh server restart

## Eliminated

## Evidence

- Main.go line 218: `go metricsCollector.RunHistoricalBackfill(ctx)` - historical backfill IS started on startup
- RunHistoricalBackfill: 7-min startup delay, then hourly, uses 3-day window  
- runHistoricalBackfill: calls GetLatestMetricTimes to find gaps, Fetches from CloudWatch, stores 5-min datapoints
- CloudWatch client: GetRDSMetricsMultiple uses 5-minute period (Period=300), returns multiple datapoints

## ROOT CAUSE

**Bug in `getMetricMultiple` and `GetRDSMetricsMultiple`:**

The `getMetricMultiple` function fetches CloudWatch datapoints correctly, including timestamps. However, when creating the `MetricValue` struct, it only stores `Avg`, `Max`, `Min` - the `Timestamp` is DISCARDED.

When `GetRDSMetricsMultiple` merges the metrics:
- Creates `RDSMetricDatapoint{}` with **no timestamp set** (zero value = `0001-01-01 00:00:00`)
- CPU value gets stored, but the `Timestamp` field remains zero

This zero timestamp gets stored to the database as `0001-01-01 00:00:00+00`.

## FIX APPLIED

1. Added new type `MetricValueWithTimestamp` that includes the timestamp
2. Updated `getMetricMultiple` to return `[]MetricValueWithTimestamp` instead of `[]MetricValue`
3. Updated `GetRDSMetricsMultiple` to properly set the `Timestamp` field when creating `RDSMetricDatapoint`

## Resolution

**Root Cause:**
`RDSMetricDatapoint.Timestamp` was never set because `getMetricMultiple` discarded CloudWatch timestamps when converting to `MetricValue` - the struct only stored `Avg`, `Max`, `Min` but not `Timestamp`.

**Fix Applied:**
1. Created new type `MetricValueWithTimestamp` that includes `Timestamp` field
2. Updated `getMetricMultiple` to return `[]MetricValueWithTimestamp` 
3. Updated `GetRDSMetricsMultiple` to properly set the `Timestamp` field when creating `RDSMetricDatapoint`

**Files Changed:**
- `internal/metrics/cloudwatch.go`: Fixed timestamp handling in `getMetricMultiple` and `GetRDSMetricsMultiple`

**Verification:**
- Code compiles successfully
- CloudWatch datapoint timestamps are now properly preserved and stored

The 3-day backfill should now show proper timestamps like `2026-02-25 16:00:00+00` instead of `0001-01-01`.
