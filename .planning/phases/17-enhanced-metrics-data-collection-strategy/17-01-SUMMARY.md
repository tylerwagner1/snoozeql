---
phase: 17-enhanced-metrics-data-collection-strategy
plan: 01
subsystem: metrics
tags: aws, cloudwatch, cloudwatch-apis, time-series

# Dependency graph
requires:
  - phase: 10
    provides: CloudWatch metrics collection infrastructure
  - phase: 11
    provides: Metrics API endpoints
  - phase: 12
    provides: Metrics retention and cleanup
provides:
  - 5-minute CloudWatch collection via GetRDSMetricsMultiple method
  - MetricPeriod constant for 5-minute timestamp truncation
  - Updated metrics collector processing 3 datapoints per 15-min cycle
  - Zero metric entries for 15-minute windows
affects:
  - 17-02
  - metrics-visualization

# Tech tracking
tech-stack:
  added:
    - MetricPeriod constant
    - RDSMetricDatapoint struct
    - GetRDSMetricsMultiple method
    - getMetricMultiple helper
  patterns:
    - 5-minute granularity metrics storage
    - Multi-datapoint fetch and merge pattern

key-files:
  created: []
  modified:
    - internal/metrics/store.go
    - internal/metrics/cloudwatch.go
    - internal/metrics/collector.go

key-decisions:
  - Store timestamps pre-truncated in Go, SQL as-is for backward compatibility
  - Rely on caller to pass pre-truncated timestamps for 5-minute granularity
  - Store 3 zero entries for stopped instances (1 per 5-minute interval)
  - Keep existing methods unchanged for BackfillMetrics backward compatibility

# Metrics
duration: 3 min
completed: 2026-02-26
---

# Phase 17 Plan 01: 5-Minute CloudWatch Collection Summary

**CloudWatch API updated from 1-hour (Period=3600) to 5-minute (Period=300) intervals, fetching 3 datapoints per 15-minute collection cycle**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-26T02:56:59Z
- **Completed:** 2026-02-26T02:59:46Z
- **Tasks:** 3/3
- **Files modified:** 3

## Accomplishments
- Added MetricPeriod constant (5 * time.Minute) with TruncateToMetricPeriod helper
- Implemented GetRDSMetricsMultiple method returning []RDSMetricDatapoint with Period=300
- Updated collector to call GetRDSMetricsMultiple and process 3 datapoints per cycle
- storeZeroMetrics now generates 3 zero entries for current 15-minute window
- Existing methods preserved for BackfillMetrics backward compatibility

## Task Commits

Each task was committed atomically:

1. **Task 1: Add MetricPeriod constant and update truncation in store.go** - `4995cc33` (feat)
2. **Task 2: Add GetRDSMetricsMultiple to cloudwatch.go** - `89f7bb35` (feat)
3. **Task 3: Update collector to use multi-datapoint collection** - `c0aefc4c` (feat)

**Plan metadata:** metadata commit follows plan completion

## Files Created/Modified
- `internal/metrics/store.go` - Added MetricPeriod constant and TruncateToMetricPeriod helper
- `internal/metrics/cloudwatch.go` - Added RDSMetricDatapoint struct and GetRDSMetricsMultiple method
- `internal/metrics/collector.go` - Updated to use GetRDSMetricsMultiple with 3 datapoints per cycle

## Decisions Made
- Store timestamps pre-truncated in Go, keep SQL as-is for backward compatibility
- Rely on caller to pass pre-truncated timestamps for 5-minute granularity
- Store 3 zero entries for stopped instances (one per 5-minute interval)
- Keep existing methods unchanged for BackfillMetrics backward compatibility

## Deviations from Plan

**None - plan executed exactly as written.**

## Issues Encountered
- None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 5-minute CloudWatch collection infrastructure complete
- Ready for 17-02: Gap detection and interpolation
- Backward compatibility maintained for existing callers

---
*Phase: 17-enhanced-metrics-data-collection-strategy*
*Completed: 2026-02-26*
