---
phase: 07-core-savings-calculation-api
plan: 01
subsystem: database
tags: savings, postgres, materialized-view, calc

# Dependency graph
requires:
  - phase: 06-intelligent-recommendations
    provides: EventStore, InstanceStore for calculating savings from stop events
provides:
  - Migration 006 with indexes for time-range queries
  - SavingsStore with upsert/query methods
  - SavingsCalculator with duration × hourly_cost_cents logic
affects: [07-core-savings-calculation-api, event-decorator, savings-api]

# Tech tracking
tech-stack:
  added:
    - postgresql materialized views for dashboard performance
  patterns:
    - Integer cents for all money calculations (no float64)
    - 7-day cap for savings (AWS auto-restart limits)
    - Upsert patterns for accumulating daily savings

key-files:
  created:
    - deployments/docker/migrations/006_cost_tracking.sql
    - internal/savings/calculator.go
  modified:
    - internal/store/savings_store.go
    - internal/models/models.go

key-decisions:
  - "Use integer cents throughout - no float64 for money to avoid precision issues"
  - "7-day maximum duration based on AWS auto-restart limits for stopped instances"
  - "Materialized view savings_summary for O(1) dashboard aggregations"

patterns-established:
  - "Savings Store pattern: UpserDailySaving with ON CONFLICT accumulator"
  - "Integer math for savings: (stoppedMinutes * hourlyCostCents) / 60"

# Metrics
duration: 12 min
completed: 2026-02-23
---

# Phase 07 Plan 01: Savings Foundation Summary

**Database foundation for savings tracking: migration 006, SavingsStore with upsert/query, and SavingsCalculator with 7-day cap**

## Performance

- **Duration:** 12 min
- **Started:** 2026-02-23T19:57:36Z
- **Completed:** 2026-02-23T20:09:42Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Migration 006 with time-range indexes, hourly_rate_cents column, and savings_summary materialized view
- SavingsStore with 6 query methods following existing store patterns (UpsertDailySaving, GetSavingsByInstance, GetTotalSavings, GetDailySavings, GetTopSavers, RefreshSavingsSummary)
- SavingsCalculator with CalculateSavings, CalculateOngoingSavings, and SplitByDay methods, all with 7-day cap
- HourlyRateCents field added to Savings model for rate tracking (AUD-02 compliance)

## Task Commits

Each task was committed atomically:

1. **Task 1: Migration for cost tracking** - `845b958c` (chore)
   - idx_events_instance_time, idx_events_type_time indexes
   - hourly_rate_cents column on savings table
   - savings_summary materialized view with idx_savings_summary_instance_day unique index

2. **Task 2: Implement SavingsStore** - `d17a687c` (feat)
   - NewSavingsStore, UpsertDailySaving, GetSavingsByInstance
   - GetTotalSavings, GetDailySavings, GetTopSavers, RefreshSavingsSummary

3. **Task 3: Implement SavingsCalculator** - `91d067c5` (feat)
   - NewSavingsCalculator, CalculateSavings, CalculateOngoingSavings, SplitByDay
   - 7-day MaxStoppedDuration constant, integer math for all calculations

**Plan metadata:** [Commit hash for this summary]

## Files Created/Modified

- `deployments/docker/migrations/006_cost_tracking.sql` - Time-range indexes and materialized view for dashboard performance
- `internal/store/savings_store.go` - SavingsStore with 6 CRUD methods following existing store patterns
- `internal/models/models.go` - Added HourlyRateCents field to Saving model
- `internal/savings/calculator.go` - SavingsCalculator with CalculateSavings, CalculateOngoingSavings, SplitByDay

## Decisions Made

- **Integer cents throughout** - No float64 for money to avoid precision issues. All savings calculations use `(minutes * hourlyRate) / 60` with integer math.
- **7-day maximum duration** - Based on AWS auto-restart limits. After 7 days of being stopped, instances may be automatically restarted by AWS.
- **Materialized view for aggregation** - savings_summary materialized view pre-aggregates by day/week/month for O(1) dashboard queries.
- **Upsert accumulator pattern** - ON CONFLICT DO UPDATE accumulates stopped_minutes and estimated_savings_cents for the same day.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- **models.Saving missing HourlyRateCents field** - Add field to Savings model for storing rate at time of calculation (required for AUD-02 compliance)
- **Go vendoring inconsistency** - Used `-mod=mod` flag to ignore vendor directory during build

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Savings foundation complete and ready for event decorator
- Event decorator can use SavingsStore.UpsertDailySaving to register savings from stop events
- SavingsCalculator can be called when instances start to calculate savings since stop
- Rate tracking enabled via hourly_rate_cents column (AUD-02 requires storing rate for reporting)

---
*Phase: 07-core-savings-calculation-api*
*Completed: 2026-02-23*
