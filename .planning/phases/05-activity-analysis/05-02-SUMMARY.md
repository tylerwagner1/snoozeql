---
phase: 05-activity-analysis
plan: 02
subsystem: metrics
tags: cloudwatch, aws-sdk, golang, background-service, hourly-aggregation

# Dependency graph
requires:
  - phase: 05-activity-analysis
    provides: Database schema for hourly metrics storage
provides:
  - CloudWatch metrics collection from AWS RDS instances
  - MetricsCollector background service with 15-minute interval
  - Hourly aggregation of CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS
affects:
  - 03-activity-analysis: Collected metrics will be used by pattern detection

# Tech tracking
tech-stack:
  added:
    - github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.54.0
    - CloudWatchClient wrapper
  patterns:
    - CloudWatchClient: AWS SDK v2 wrapper with retry logic and exponential backoff
    - MetricsCollector: Background service pattern following existing DiscoveryService
    - Client caching: Per-account+region caching with RWMutex

key-files:
  created:
    - internal/metrics/cloudwatch.go
    - internal/metrics/collector.go
  modified:
    - cmd/server/main.go
    - go.mod

key-decisions:
  - "GetMetricStatistics API for single-metric queries (simpler than GetMetricData for this use case)"
  - "3 retry attempts with exponential backoff for throttling and transient errors"
  - "1-hour period fetching with hourly aggregation, returning hourly averages"
  - "Client caching by account+region combination to minimize AWS configuration overhead"

patterns-established:
  - "CloudWatchClient pattern: AWS SDK v2 wrapper with GetRDSMetrics returning RDSMetrics struct"
  - "MetricsCollector pattern: Background service on 15-minute ticker with CollectAll iterating instances"
  - "Retry strategy: 3 attempts with exponential backoff, special handling for LimitExceededException"

# Metrics
duration: ~20 min
completed: 2026-02-23
---

# Phase 05: Activity Analysis - Plan 02 Summary

**CloudWatch client and MetricsCollector for background AWS RDS metrics collection every 15 minutes with 3-retry exponential backoff**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-02-23T16:24:52Z
- **Completed:** 2026-02-23T16:44:52Z
- **Tasks:** 4
- **Files modified:** 4

## Accomplishments
- Created CloudWatchClient with GetRDSMetrics returning all 4 metric types (CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS)
- Implemented 3-retry exponential backoff with special handling for throttling (LimitExceededException)
- Created MetricsCollector background service running on 15-minute interval
- Added MetricsCollector initialization and startup in cmd/server/main.go
- Skips non-AWS and stopped instances (no metrics available per CONTEXT.md)
- Stores metrics via MetricsStore.UpsertHourlyMetric with hourly aggregation

## Task Commits

Each task was committed atomically:

1. **Task 1: Add CloudWatch SDK dependency** - `05-02-task1` (chore)
2. **Task 2: Create CloudWatch client wrapper** - `05-02-task2` (feat)
3. **Task 3: Create MetricsCollector service** - `05-02-task3` (feat)
4. **Task 4: Integrate MetricsCollector into main.go** - `05-02-task4` (feat)

**Plan metadata:** `05-02-docs` (docs)

## Files Created/Modified
- `internal/metrics/cloudwatch.go` - Created CloudWatchClient with GetRDSMetrics, retry logic, and 4 metric types
- `internal/metrics/collector.go` - Created MetricsCollector with RunContinuous, CollectAll, per-instance collection, and client caching
- `cmd/server/main.go` - Added metrics import, metricsStore and metricsCollector globals, initialization, and goroutine startup
- `go.mod` - Added github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.54.0 dependency
- `internal/metrics/store.go` - Referenced (existing, used via UpsertHourlyMetric)

## Decisions Made
- Used GetMetricStatistics API (simpler than GetMetricData for single-metric queries)
- 3 retry attempts with exponential backoff for throttling and transient errors
- Fetches last 1 hour of data with 1-hour period for hourly aggregation
- Client caching by account+region combination to minimize AWS configuration
- Skips non-AWS instances (GCP not yet implemented) and stopped instances (no metrics available)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all verification steps passed successfully.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- MetricsCollector running and collecting CloudWatch metrics every 15 minutes
- Hourly aggregates stored in metrics_hourly table for pattern detection
- HasSufficientData method available for checking 24+ hours requirement
- Failed requests properly handled with 3 retries before marking as failed
- Stopped instances correctly skipped as per CONTEXT.md requirements

---

*Phase: 05-activity-analysis*
*Completed: 2026-02-23*
