---
phase: 05-activity-analysis
plan: 01
subsystem: database
tags: postgres, metrics, cloudwatch, hourly-aggregation

# Dependency graph
requires:
  - phase: 04-activity-analysis
    provides: Basic infrastructure setup and postgres connection patterns
provides:
  - Database schema for hourly metrics storage with metrics_hourly table
  - HourlyMetric model struct for Go application
  - MetricsStore with CRUD operations for metrics persistence
affects:
  - 02-activity-analysis: Uses this schema for CloudWatch metrics collection
  - 03-activity-analysis: Uses HasSufficientData for pattern detection

# Tech tracking
tech-stack:
  added:
    - metrics_hourly table in PostgreSQL
    - internal/metrics package
  patterns:
    - MetricsStore: Postgres-based store pattern
    - Hourly aggregation: ON CONFLICT UPSERT with incremental averaging

key-files:
  created:
    - deployments/docker/migrations/005_metrics_hourly.sql
    - internal/metrics/store.go
  modified:
    - internal/models/models.go

key-decisions:
  - "Hourly aggregation with incremental averaging for UPSERT"
  - "Unique constraint on (instance_id, metric_name, hour) for idempotent inserts"
  - "Separate metrics_store.go with CRUD operations pattern matching existing stores"

patterns-established:
  - "MetricsStore pattern: Postgres connection wrapper with typed methods for metrics operations"
  - "Hourly aggregation: Incremental averaging formula (avg_value * sample_count + new_value) / (sample_count + 1)"

# Metrics
duration: ~15 min
completed: 2026-02-23
---

# Phase 05: Activity Analysis - Plan 01 Summary

**Database schema and MetricsStore for persisting hourly CloudWatch metric aggregates with UPSERT support**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-02-23T15:00:00Z
- **Completed:** 2026-02-23T15:15:00Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments
- Created metrics_hourly table in PostgreSQL with unique constraint and indexes for efficient time-range queries
- Added HourlyMetric model struct with MetricNames constants for CloudWatch metric types
- Implemented MetricsStore with UpsertHourlyMetric (ON CONFLICT with incremental averaging), GetMetricsByInstance, GetLatestMetrics, DeleteOldMetrics, and HasSufficientData methods

## Task Commits

Each task was committed atomically:

1. **Task 1: Create metrics_hourly migration** - `a5051627` (test)
2. **Task 2: Add HourlyMetric model** - `cf6e4318` (feat)
3. **Task 3: Create MetricsStore** - `3a0725c8` (feat)

**Plan metadata:** Created during plan execution.

## Files Created/Modified
- `deployments/docker/migrations/005_metrics_hourly.sql` - SQL migration for metrics_hourly table with indexes and triggers
- `internal/models/models.go` - Added HourlyMetric struct and MetricNames constants
- `internal/metrics/store.go` - Created MetricsStore with all CRUD operations

## Decisions Made
- Used `date_trunc('hour', $3::timestamptz)` in UPSERT for hour bucketing
- Incremental averaging formula for updating avg_value on conflict: `(metrics_hourly.avg_value * metrics_hourly.sample_count + EXCLUDED.avg_value) / (metrics_hourly.sample_count + 1)`
- GREATEST/LEAST functions for tracking max_value and min_value across samples

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all verification steps passed successfully.

## Next Phase Readiness
- Foundation ready for Phase 5 Plan 02 (CloudWatch client and MetricsCollector)
- HasSufficientData method available for checking 24+ hours requirement per CONTEXT.md
- DeleteOldMetrics method available for 14-day retention cleanup

---

*Phase: 05-activity-analysis*
*Completed: 2026-02-23*
