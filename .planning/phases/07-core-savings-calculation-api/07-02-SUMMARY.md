---
phase: 07-core-savings-calculation-api
plan: 02
subsystem: savings
tags: savings, event-decorator, audit-logging

# Dependency graph
requires:
  - phase: 07-core-savings-calculation-api
    provides: SavingsCalculator, SavingsStore with upsert/query methods
provides:
  - EventStoreWithSavings decorator that intercepts CreateEvent
  - Automatic savings calculation on start/wake events
  - Hourly rate capture at stop event time in event metadata
affects:
  - 07-core-savings-calculation-api: other plans may reference this decorator
  - 08: future plans can use this decorated EventStore

# Tech tracking
tech-stack:
  added: []
  patterns: [decorator-pattern, audit-logging]

key-files:
  created:
    - internal/savings/event_decorator.go
  modified:
    - cmd/server/main.go
    - internal/discovery/discovery.go

key-decisions:
  - "Used EventStoreWithSavings decorator pattern to wrap EventStore with automatic savings calculation"
  - "Added EventCreator interface to DiscoveryService for flexible event store types"
  - "Stop/sleep events capture hourly_rate_cents in metadata for use during start events"

# Metrics
duration: ~10 min
completed: 2026-02-23
---

# Phase 7 Plan 2: EventStoreWithSavings Implementation Summary

**EventStoreWithSavings decorator that intercepts CreateEvent for automatic savings calculation and hourly rate capture**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-02-23T20:04:12Z
- **Completed:** 2026-02-23T20:14:XXZ
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- EventStoreWithSavings decorator implemented with automatic savings calculation on start/wake events
- Stop/sleep events capture hourly_rate_cents in event metadata (satisfies AUD-02)
- Start/wake events calculate and persist savings via SavingsStore (satisfies AUD-01)
- DiscoveryService updated to use decorated EventStore via EventCreator interface
- Per-day savings split and persisted for accurate daily reporting

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement EventStoreWithSavings decorator** - `b5b5b8e` (feat)
2. **Task 2: Wire decorated EventStore into main.go** - `7671e85` (feat)

**Plan metadata:** (auto-generated(docs): complete plan)

## Files Created/Modified
- `internal/savings/event_decorator.go` - New file implementing EventStoreWithSavings decorator with CreateEvent, ListEventsByInstance, and internal helper methods
- `cmd/server/main.go` - Modified to import savings package, create decorated EventStore, and wire it into DiscoveryService
- `internal/discovery/discovery.go` - Added EventCreator interface and updated DiscoveryService to use it for flexible event store types

## Decisions Made
- Used decorator pattern to wrap EventStore with automatic savings calculation without modifying existing behavior
- Added EventCreator interface to DiscoveryService to accept any event store implementation
- Stop/sleep events capture hourly_rate_cents in event metadata for use during subsequent start events
- Fallback to instance.HourlyCostCents if metadata doesn't exist (for migration compatibility)
- Per-day savings split using calculator.SplitByDay for accurate daily aggregation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Vendor directory had inconsistencies after go mod tidy - restored vendor/ from git before builds

## Next Phase Readiness
- EventStoreWithSavings decorator is ready for use by future plans
- Savings calculation and persistence is automatic when events are created
- AUD-01 (logging with instance_id, date, stopped_minutes, rate) is satisfied
- AUD-02 (hourly rate capture at stop event time) is satisfied
- Existing event creation flow continues to work unchanged (decorator pattern)
