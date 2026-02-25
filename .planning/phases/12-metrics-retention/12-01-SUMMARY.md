---
phase: 12-metrics-retention
plan: 01
subsystem: database
tags: metrics, retention, cleanup, batch, postgres, background-jobs

# Dependency graph
requires:
  - phase: 10
    provides: metrics_hourly table with hour column and index
  - phase: 11
    provides: metrics collection infrastructure
provides:
  - Automatic 7-day retention cleanup for metrics_hourly table
  - Background RetentionCleaner service with batched deletes
  - Last-run timestamp tracking in settings table
affects: metrics-retention, database-management

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Background service pattern: RunContinuous with startup delay and ticker"
    - "Batched deletes: Subquery with LIMIT to avoid table locks"
    - "State tracking: JSONB settings table for cross-restart persistence"

key-files:
  created:
    - internal/metrics/retention.go (149 lines)
  modified:
    - internal/metrics/store.go (DeleteOldMetrics updated with limit param)
    - cmd/server/main.go (RetentionCleaner initialization added)

key-decisions:
  - "Hard-coded 7-day retention (not configurable per CONTEXT.md)"
  - "1000 rows per batch with 100ms pause between batches"
  - "7-minute startup delay within 5-10 minute range per CONTEXT.md"
  - "24-hour cleanup interval (fixed)"
  - "Settings key: metrics_retention_last_run for last-run tracking"
  - "UTC timestamps for all comparisons (per RESEARCH.md pitfalls)"

patterns-established:
  - "RetentionCleaner: Background service pattern with context-aware RunContinuous"
  - "Batched deletes: Subquery pattern (DELETE WHERE id IN (SELECT ... LIMIT))"
  - "Idempotent cleanup: Skip if last run < 24 hours ago"

# Metrics
duration: 8 min
completed: 2026-02-25
---

# Phase 12: Metrics Retention Summary

**Automatic 7-day retention cleanup with batched deletes and settings-based state tracking**

## Performance

- **Duration:** 8 min
- **Started:** 2026-02-25T20:12:20Z
- **Completed:** 2026-02-25T20:20:00Z
- **Tasks:** 3/3
- **Files modified:** 3

## Accomplishments
- Created RetentionCleaner service following existing MetricsCollector.RunContinuous() pattern
- Batched deletes (1000 rows/batch) with 100ms pauses prevent table locking
- Last-run timestamp stored in settings table enables skip-if-recent logic
- Integration into main.go starts cleanup automatically with startup message

## Task Commits

Each task was committed atomically:

1. **Task 1 & 2: RetentionCleaner and batched delete support** - `9b3cec49` (feat)
   - Created `internal/metrics/retention.go` with RetentionCleaner struct
   - Implemented `RunContinuous()` with 7-min startup delay and 24h ticker
   - Added `runCleanup()`, `deleteInBatches()`, `getLastRunTime()`, `setLastRunTime()`
   - Updated `DeleteOldMetrics()` in `store.go` with limit parameter for batch support
2. **Task 3: Integration in main.go** - `5d955d71` (feat)
   - Added `retentionCleaner := metrics.NewRetentionCleaner(metricsStore, db)`
   - Started background goroutine: `go retentionCleaner.RunContinuous(ctx)`
   - Added startup log message confirming 7-day retention and 24h interval

**Plan metadata:** `82583ead` (docs: create phase plan)

## Files Created/Modified
- `internal/metrics/retention.go` - RetentionCleaner service with RunContinuous pattern, batched deletion, last-run tracking
- `internal/metrics/store.go` - Updated DeleteOldMetrics with limit parameter for batched subquery deletes
- `cmd/server/main.go` - Added RetentionCleaner initialization and startup goroutine

## Decisions Made
- Hard-coded 7-day retention (not configurable per CONTEXT.md requirement)
- 1000 rows per batch with 100ms pause between batches
- 7-minute startup delay (within 5-10 minute range per CONTEXT.md)
- 24-hour cleanup interval (fixed)
- Using "metrics_retention_last_run" key in settings table for tracking
- UTC timestamps for all time comparisons (per RESEARCH.md pitfalls)
- Subquery pattern for batched deletes (PostgreSQL standard: DELETE WHERE id IN (SELECT ... LIMIT))

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - no blocking issues or unexpected problems.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The metrics retention foundation is complete. The cleanup runs automatically in the background with:
- 7-day retention window enforced
- Batched deletes (1000 rows/batch) to avoid table locking
- Last-run tracking to skip if already ran within 24 hours
- Startup after 7-minute delay, then every 24 hours

Ready for Phase 13 (Idle Detection) which will rely on this 7-day data retention.

---

*Phase: 12-metrics-retention*
*Completed: 2026-02-25*
