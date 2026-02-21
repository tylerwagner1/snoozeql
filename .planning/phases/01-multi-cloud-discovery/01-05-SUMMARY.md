---
phase: 01-multi-cloud-discovery
plan: 05
subsystem: ui
tags: [react, router, navigation, dashboard, cta]

# Dependency graph
requires:
  - phase: 01-01
    provides: InstanceStore with ListInstances and ListRecommendations methods
provides:
  - Clickable stats cards with navigation to pre-filtered views
  - InstancesPage URL-based filtering with updateFilter function
  - Prominent CTAs for adding cloud accounts
affects: ui, filtering, navigation

# Tech tracking
tech-stack:
  added: []
  patterns: ["useNavigate for routing", "useSearchParams for URL params", "URL-based filter state"]

key-files:
  created: []
  modified:
    - web/src/pages/Dashboard.tsx
    - web/src/pages/InstancesPage.tsx
    - cmd/server/main.go
    - internal/store/postgres.go
    - web/src/pages/CloudAccountsPage.tsx

key-decisions:
  - "Stats cards navigated to /instances?status=running and /instances?status=stopped for filtered views"
  - "InstancesPage uses URL params for initial filter state, syncs filter changes to URL"
  - "CTAs shown when no accounts exist (Get Started section) and quick actions when accounts exist"
  - "Stats endpoint returns real counts by mapping instance statuses to running/stopped categories"

patterns-established:
  - "useNavigate hook for programmatic navigation in React hooks"
  - "useSearchParams for reading and updating URL search parameters"
  - "URL params drive initial filter state, filter changes update URL for deep linking"

# Metrics
duration: ~15 min
completed: 2026-02-21
---

# Phase 1 Plan 5: Dashboard Clickable Stats and CTAs Summary

**Interactive dashboard with clickable stats cards, URL-based filtering, and prominent account CTAs**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-02-21T02:25:16Z
- **Completed:** 2026-02-21T02:40:00Z
- **Tasks:** 3
- **Files modified:** 5
- **Commits:** 3

## Accomplishments

1. **Dashboard stats cards now clickable** - Running Databases navigates to `/instances?status=running`, Sleeping Databases navigates to `/instances?status=stopped`, Pending Actions navigates to `/recommendations`

2. **InstancesPage URL-based filtering** - Uses `useSearchParams` to read initial filter state from URL, `updateFilter` function syncs filter changes back to URL for deep linking support

3. **Prominent CTAs for cloud accounts** - When no accounts exist, shows "Get Started" section with AWS/GCP account buttons; when accounts exist, shows quick action links

4. **Stats endpoint enhanced** - Returns real instance counts by mapping statuses (available/running/starting → running, stopped/stopping → stopped)

## Task Commits

Each task was committed atomically:

1. **Task 1: Make stats cards clickable** - `44da6ac` (feat)
2. **Task 2: Update InstancesPage to read URL params** - `5443e76` (feat)
3. **Task 3: Add CTAs and enhance stats** - `f304581` (feat)

## Files Created/Modified

- `web/src/pages/Dashboard.tsx` - Added useNavigate, cloudAccounts state, clickable stats cards with onClick handlers, CTAs for adding accounts
- `web/src/pages/InstancesPage.tsx` - Added useSearchParams, updateFilter function, URL-based initial filter state
- `web/src/pages/CloudAccountsPage.tsx` - Fixed handleSubmit function (added missing `const` keyword)
- `cmd/server/main.go` - Enhanced stats endpoint to return real counts from database
- `internal/store/postgres.go` - Added ListRecommendationsByStatus method for stats endpoint

## Decisions Made

- Stats cards naviagte to `/instances?status=running` and `/instances?status=stopped` for filtered views
- InstancesPage uses URL params for initial filter state, syncs filter changes to URL for deep linking
- CTAs shown when no accounts exist (Get Started section) and quick actions when accounts exist
- Stats endpoint maps statuses: available/running/starting → running, stopped/stopping → stopped

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed CloudAccountsPage handleSubmit missing const**

- **Found during:** Task 3 (CloudAccountsPage verification)
- **Issue:** handleSubmit function defined without `const` keyword causing TypeScript error
- **Fix:** Added `const` keyword before `handleSubmit` function declaration
- **Files modified:** web/src/pages/CloudAccountsPage.tsx
- **Verification:** npm run build succeeds
- **Committed in:** f304581 (Task 3 commit)

**2. [Rule 1 - Bug] Stats endpoint returns hardcoded zeros**

- **Found during:** Task 3 (stats endpoint implementation)
- **Issue:** Stats endpoint returned hardcoded zeros instead of real counts
- **Fix:** Implemented real counts by iterating instances and mapping statuses to running/stopped
- **Files modified:** cmd/server/main.go, internal/store/postgres.go
- **Verification:** Go code compiles, npm run build succeeds
- **Committed in:** f304581 (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 bug)
**Impact on plan:** All auto-fixes necessary for correct operation. No scope creep.

## Issues Encountered

- TypeScript error in CloudAccountsPage - missing `const` keyword on handleSubmit function
- Stats endpoint needed database-driven counts instead of hardcoded zeros

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Dashboard navigation foundation complete
- URL-based filtering pattern established
- Stats endpoint returns real counts
- CTAs for cloud accounts ready for user flow

---

*Phase: 01-multi-cloud-discovery*
*Completed: 2026-02-21*
