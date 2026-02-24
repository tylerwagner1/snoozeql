---
phase: 09-complete-savings-removal
plan: 01
subsystem: infra
tags: docker, cleanup, frontend, backend

# Dependency graph
requires:
  - phase: 08
    provides: savings API types and handlers for frontend/backend integration
provides:
  - Removed all savings-related code from frontend (web/src/)
  - Removed all savings routes, handlers, stores, and calculators from backend
  - Rebuilt Docker containers with fresh artifacts
  - Frontend navigation now shows only 6 items (no Savings link)
affects:
  - infra: Docker rebuild ensures clean deployment state
  - frontend: no savings data displayed (uses estimated costs instead)
  - backend: savings API endpoints return 404

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Systematic removal of unused code packages
    - Docker container rebuild for cache invalidation
    - Gradual migration pattern (estimated costs instead of savings API)

key-files:
  created: []
  modified:
    - web/src/lib/api.ts
    - web/src/components/CostOverTimeChart.tsx
    - cmd/server/main.go

key-decisions:
  - "Removed entire internal/savings/ package including calculator and event decorator"
  - "Removed savingsStore, savingsCalculator, decoratedEventStore from cmd/server/main.go"
  - "Replaced savings API calls with estimated cost calculation in CostOverTimeChart"
  - "Kept savings_7d in Stats interface for dashboard display (estimated calculation)"

patterns-established:
  - "Pattern 1: Complete package removal for unused features - delete entire directory + clean all imports"
  - "Pattern 2: Docker rebuild with --no-cache to invalidate stale build artifacts"

# Metrics
duration: 10 min
completed: 2026-02-24
---

# Phase 9 Plan 1: Complete Savings Removal & Cache Validation Summary

**Systematic removal of all savings-related code from frontend and backend, with Docker container rebuild for clean cache**

## Performance

- **Duration:** 10 min
- **Started:** 2026-02-24T17:00:00Z
- **Completed:** 2026-02-24T17:10:00Z
- **Tasks:** 8
- **Files modified:** 8
- **Files deleted:** 4

## Accomplishments

- Removed all savings-related TypeScript interfaces and API methods from frontend
- Removed savings routes, handlers, stores, and calculators from backend
- Deleted entire `internal/savings/` package (calculator, event decorator)
- Deleted savings handler `internal/api/handlers/savings.go`
- Deleted savings store `internal/store/savings_store.go`
- Updated CostOverTimeChart to use estimated cost calculation instead of savings API
- Rebuilt frontend and backend Docker containers with fresh artifacts
- Verified `/savings` endpoint returns 404
- Verified frontend navigation shows only 6 items

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove Frontend Savings API Methods and Types** - `832e1f8f` (fix)
2. **Task 2: Remove Backend Savings Routes** - `8704ae85` (feat)
3. **Task 3: Delete Backend Savings Handler** - `7ac020fd` (chore)
4. **Task 4: Delete Backend Savings Store** - `520497c2` (chore)
5. **Task 5: Delete Backend Savings Package** - `437e2461` (chore)
6. **Task 6: Update Event Store** - `5a47dffc` (docs)
7. **Task 7: Rebuild and Restart Docker Containers** - `1b65caeb` (chore)
8. **Task 8: Final Validation** - `bff0793e` (docs)

**Plan metadata:** `bff0793e` (docs: complete plan)

## Files Created/Modified

- `web/src/lib/api.ts` - Removed SavingsSummary, DailySavingsResponse, InstanceSavingsItem, InstanceSavingsDetail interfaces; removed 5 savings API methods; kept savings_7d in Stats and estimated_daily_savings in Recommendation interfaces
- `web/src/components/CostOverTimeChart.tsx` - Removed api import and getDailySavings call; updated to use estimated cost calculation
- `cmd/server/main.go` - Removed savings import, savingsStore, savingsCalculator, decoratedEventStore initialization; removed 5 savings routes
- `internal/api/handlers/savings.go` - Deleted
- `internal/store/savings_store.go` - Deleted
- `internal/savings/calculator.go` - Deleted
- `internal/savings/event_decorator.go` - Deleted

## Decisions Made

- **Removed entire internal/savings/ package** - Clean removal of savings calculation logic instead of conditional compilation
- **Replaced savings API calls with estimated costs in CostOverTimeChart** - Maintains UI functionality without backend dependency
- **Kept savings_7d in Stats interface** - Dashboard still displays estimated daily savings (calculated from instance hourly costs)
- **Kept estimated_daily_savings in Recommendation interfaces** - Recommendations feature (not savings) still uses this field for display

## Deviations from Plan

**None - plan executed exactly as written.**

## Issues Encountered

**None.**

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Backend build: Successful
- Frontend build: Successful
- Docker containers: Rebuilt with fresh artifacts
- `/savings` endpoint: Returns 404 (as expected)
- Frontend navigation: Shows 6 items only (Dashboard, Accounts, Instances, Schedules, Recommendations, Audit Log)

---

*Phase: 09-complete-savings-removal*
*Completed: 2026-02-24*
