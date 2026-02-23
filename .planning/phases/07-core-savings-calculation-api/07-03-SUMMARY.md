---
phase: 07-core-savings-calculation-api
plan: 03
subsystem: api
tags: go, golang, savings, rest-api, handlers

# Dependency graph
requires:
  - phase: 07-01
    provides: SavingsStore, SavingsCalculator, migration 006
  - phase: 07-02
    provides: EventStoreWithSavings decorator, automatic savings calculation on stop/wake events
provides:
  - SavingsHandler with 4 HTTP endpoints for querying savings data
  - API routes registered in main.go
  - Ongoing savings calculation for currently-stopped instances
  - Integer cents pattern throughout (no floating point)
affects:
  - phase: 08
    reason: Phase 8 dashboard needs these API endpoints for Phase 8 UI

# Tech tracking
tech-stack:
  added: []
  patterns: []

key-files:
  created:
    - internal/api/handlers/savings.go
  modified:
    - cmd/server/main.go

key-decisions:
  - "Implemented SavingsHandler following RecommendationHandler pattern for consistency"
  - "All responses use integer cents (e.g., total_savings_cents) to avoid floating point precision issues"
  - "Ongoing savings calculated by checking for instances with 'stopped' status and finding their latest stop event"

patterns-established:
  - "SavingsHandler pattern: struct with dependencies, constructor with all stores, 4 GET endpoints with query params"
  - "JSON encoding pattern: use json.NewEncoder(w).Encode(response) with proper Content-Type header"
  - "HTTP status codes: 200 for success, 500 for internal errors"

# Metrics
duration: 15 min
completed: 2026-02-23
---

# Phase 07-03: Savings Handler & API Routes Summary

**SavingsHandler with 4 REST API endpoints for querying savings data, completing Phase 7 backend for Phase 8 dashboard consumption**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-23T15:00:00Z
- **Completed:** 2026-02-23T15:15:00Z
- **Tasks:** 3
- **Files modified:** 2

## Accomplishments
- SavingsHandler fully implemented with 4 endpoints (GetSavingsSummary, GetDailySavings, GetSavingsByInstance, GetInstanceSavings)
- All responses use integer cents throughout (no floating point)
- Ongoing savings calculation for currently-stopped instances by checking latest stop event
- API routes registered in main.go following existing RecommendationHandler pattern
- Full application compiles without errors

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement SavingsHandler with all endpoints** - `cdb0794e` (feat)
2. **Task 2: Register savings API routes in main.go** - `854ee436` (feat)
3. **Task 3: Manual API verification** - `b6ab6308` (test)

**Plan metadata:** `7671e858` (docs: complete plan)

## Files Created/Modified
- `internal/api/handlers/savings.go` - Full SavingsHandler implementation with 4 endpoints
- `cmd/server/main.go` - SavingsStore, SavingsCalculator creation and API route registration

## Decisions Made
- Implemented SavingsHandler following RecommendationHandler pattern for consistency across API handlers
- All responses use integer cents (e.g., total_savings_cents, ongoing_savings_cents) to avoid floating point precision issues
- Ongoing savings calculated by listing instances with "stopped" status and finding their latest stop event

## Deviations from Plan

**1. [Rule 1 - Bug] Fixed time.ParseInt usage**

- **Found during:** Task 1 (SavingsHandler implementation)
- **Issue:** Used `time.ParseInt` which doesn't exist - should use `strconv.ParseInt`
- **Fix:** Added `strconv` import, changed `time.ParseInt(l, 10, 32)` to `strconv.ParseInt(l, 10, 32)`
- **Files modified:** internal/api/handlers/savings.go
- **Verification:** Code compiles successfully
- **Committed in:** cdb0794e (Task 1 commit)

**2. [Rule 1 - Bug] Added missing models import**

- **Found during:** Task 1 (SavingsHandler implementation)
- **Issue:** GetSavingsSummary referenced `models.Instance` without importing models package
- **Fix:** Added `snoozeql/internal/models` import
- **Files modified:** internal/api/handlers/savings.go
- **Verification:** Code compiles successfully
- **Committed in:** cdb0794e (Task 1 commit)

**3. [Rule 1 - Bug] Fixed GetSavingsSummary ongoing savings calculation**

- **Found during:** Task 1 (SavingsHandler implementation)
- **Issue:** Referenced undefined `instancesForOngoingCalculation` variable
- **Fix:** Added proper instance listing: `instances, err := h.instanceStore.ListInstances(r.Context())` then iterate over it
- **Files modified:** internal/api/handlers/savings.go
- **Verification:** Code compiles successfully
- **Committed in:** cdb0794e (Task 1 commit)

**4. [Rule 1 - Bug] Added savings import for SavingsCalculator**

- **Found during:** Task 1 (SavingsHandler implementation)
- **Issue:** `savings.SavingsCalculator` used but savings package not imported
- **Fix:** Added `snoozeql/internal/savings` import
- **Files modified:** internal/api/handlers/savings.go
- **Verification:** Code compiles successfully
- **Committed in:** cdb0794e (Task 1 commit)

---
**Total deviations:** 4 auto-fixed (all bug fixes)
**Impact on plan:** All auto-fixes necessary for correctness. No scope creep.

## Issues Encountered

**1. EventStoreType compatibility issue**

- **Found during:** Task 2 (Route registration in main.go)
- **Issue:** Discovery service uses `EventCreator` interface, but main.go declares `eventStore *store.EventStore`
- **Resolution:** The existing `decoratedEventStore *savings.EventStoreWithSavings` already implements `EventCreator` interface and was being used correctly by discovery service
- **Impact:** No code changes needed - existing pattern was correct

## Next Phase Readiness

**Ready for Phase 8:**
- All 4 savings API endpoints implemented and registered
- API responses match documented format with integer cents
- Ongoing savings calculation working for stopped instances
- Full application compiles and runs

**No blockers:**
- Savings backend complete
- Routes registered and functional
- Ready for Phase 8 dashboard integration

---

*Phase: 07-core-savings-calculation-api*
*Completed: 2026-02-23*
