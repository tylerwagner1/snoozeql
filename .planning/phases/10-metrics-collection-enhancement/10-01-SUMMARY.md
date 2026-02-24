---
phase: 10-metrics-collection-enhancement
plan: 01
subsystem: metrics
tags: aws, cloudwatch, memory, percentage, metrics-collection

# Dependency graph
requires:
  - phase: 09
    provides: Instance discovery and metrics storage foundation
provides:
  - FreeableMemory metric collection alongside CPU and Connections
  - Memory displayed as percentage on Instance Details page
  - Zero metrics stored for stopped instances
  - Metrics unavailable badge for stale/missing metrics
affects:
  - 11-time-series-visualization
  - 12-metrics-retention
  - 13-idle-detection

# Tech tracking
tech-stack:
  added:
    - internal/metrics/memory.go (instance class to GB mapping)
  patterns:
    - Memory percentage calculation using hardcoded instance mapping
    - Storing zeros for stopped instances to show "asleep" state
    - Graceful handling of unknown instance classes with warning logs

key-files:
  created:
    - internal/metrics/memory.go
  modified:
    - internal/models/models.go
    - internal/metrics/cloudwatch.go
    - internal/metrics/collector.go
    - web/src/pages/InstanceDetailPage.tsx

key-decisions:
  - "Memory stored as percentage (0-100) using hardcoded instance class mapping instead of raw bytes"
  - "Unknown instance classes log warning but don't crash - return nil and skip storage"
  - "Stopped instances get explicit zero metrics stored (not skipped) to show 'asleep' state"

patterns-established:
  - "Memory percentage calculation: freeableMemoryBytes / totalBytes * 100"
  - "Zero metrics for stopped instances ensures consistent data patterns"
  - "Badge logic checks if latest metric older than 30 minutes"

# Metrics
duration: 1m 11s
completed: 2026-02-24
---

# Phase 10 Plan 01: FreeableMemory Metric Collection Summary

**FreeableMemory metric collection with percentage calculation and Metrics unavailable badge for stale data**

## Performance

- **Duration:** 1m 11s
- **Started:** 2026-02-24T21:37:34Z
- **Completed:** 2026-02-24T21:38:45Z
- **Tasks:** 2
- **Files modified:** 5
- **Files created:** 1

## Accomplishments
- FreeableMemory metric now collected alongside CPU and Connections from CloudWatch
- Memory displayed as percentage (0-100) on Instance Details page
- 20 instance class mappings cover T3, T4g, R5, R6g, M5, M6g families
- Stopped instances get explicit zero metrics stored (not skipped)
- Yellow "Metrics unavailable" badge appears when metrics stale/missing
- Unknown instance classes log warning but don't crash (return nil)

## Task Commits

1. **Task 1: Add FreeableMemory metric to CloudWatch collector** - `7a8c2ed` (feat)
2. **Task 2: Add Metrics unavailable badge to Instance Details UI** - `3e65c67` (feat)

**Plan metadata:** (docs: complete plan)

## Files Created/Modified
- `internal/metrics/memory.go` - Instance class to GB mapping (20 entries) and CalculateMemoryPercentage function
- `internal/models/models.go` - Added MetricFreeableMemory constant
- `internal/metrics/cloudwatch.go` - Added FreeMemory field and fetch in GetRDSMetrics
- `internal/metrics/collector.go` - Added storeZeroMetrics function and FreeableMemory storage in collectInstance
- `web/src/pages/InstanceDetailPage.tsx` - Added isMetricsStale() helper, badge, and updated Memory card to 'freeablememory'

## Decisions Made
- Memory stored as percentage (0-100) using hardcoded instance class mapping instead of raw bytes - enables consistent UI display
- Unknown instance classes log warning but don't crash - graceful degradation for new instance types
- Stopped instances get explicit zero metrics stored (not skipped) - shows "asleep" state in metrics

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- None

## Next Phase Readiness
- FreeableMemory data now available in metrics_hourly table alongside CPU and Connections
- Stopped instance zero metrics provide clear "asleep" state for pattern analysis
- Unknown instance class handling logs warnings for debugging
- Ready for Phase 11 time-series visualization

---
*Phase: 10-metrics-collection-enhancement*
*Completed: 2026-02-24*
