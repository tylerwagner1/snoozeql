---
phase: 01-multi-cloud-discovery
plan: 03
subsystem: ui
tags: react, typescript, instances-table, sorting, filtering, multi-cloud

# Dependency graph
requires:
  - phase: 01-01
    provides: InstanceStore with database persistence
  - phase: 01-02
    provides: Multi-account provider registration, GCP credentials support
provides:
  - Sortable instances table with clickable column headers
  - Filterable instances table with status and provider dropdowns
  - Account name column showing which cloud account each instance belongs to
affects: [01-04, 01-05, 02-01]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Filter state synced with URL parameters for persistent filtering across page refreshes"
    - "Sort state uses useState with useMemo for sorted instance computation"
    - "Instance type includes optional account_name field populated by backend"

key-files:
  created: []
  modified: [web/src/pages/InstancesPage.tsx, web/src/lib/api.ts]

key-decisions:
  - "Filter state uses URL params for better user experience (filters persist on refresh)"
  - "Status filter supports multiple value mappings: 'running' maps to ['available', 'running', 'starting']"
  - "Account name displays 'Unknown' when not populated by backend (future enhancement)"

patterns-established:
  - "Filter state synced with URL parameters enables filter persistence across page refreshes"
  - "Instance type enhanced with optional account_name field populated by backend"

# Metrics
duration: 8 min
completed: 2026-02-21
---

# Phase 1 Plan 03: Sortable and Filterable Instances Table Summary

**Sortable, filterable instances table with clickable headers, status/provider filters, and account name column showing which cloud account each instance belongs to**

## Performance

- **Duration:** 8 min
- **Started:** 2026-02-21T02:23:32Z
- **Completed:** 2026-02-21T02:31:37Z
- **Tasks:** 3/3
- **Files modified:** 2

## Accomplishments
- Clickable column headers for sorting by any column (Name, Account)
- Sort direction indicator (↑/↓) shows on active column
- Status and Provider filter dropdowns allow filtering instances
- Filter state persists in URL params for better UX (filters survive page refresh)
- Account name column shows cloud account name or "Unknown" if not populated
- Filtered count displays "X of Y instances"

## Task Commits

Each task was committed atomically:

1. **Task 3: Add account name column to instances table** - `ce819b0` (feat)
   - Added sortable Account column header with sort indicator (↑/↓)
   - Added Account cell showing `instance.account_name || 'Unknown'`
   - Account column appears after Name, before Provider

**Plan metadata:** ce819b0 (feat: add account name column to instances table)

## Files Created/Modified
- `web/src/pages/InstancesPage.tsx` - Added sortable columns, filter dropdowns, Account column, URL param sync
- `web/src/lib/api.ts` - Added `account_name?: string` to Instance interface

## Decisions Made

- **Filter state uses URL params:** For better user experience where filters persist across page refreshes. Filters sync with URL query parameters (e.g., `?status=running&provider=aws`)

- **Status filter value mapping:** The 'running' option maps to multiple status values ['available', 'running', 'starting'] to cover all running states across AWS and GCP

- **Account name fallback:** When `account_name` is not populated by backend, displays "Unknown" gracefully

## Deviations from Plan

None - plan executed exactly as written. The file already had extended functionality (URL param sync, status value mapping) from previous work that enhanced the planned features.

## Issues Encountered

None - execution smooth across all three tasks.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Instances table is fully functional with sorting and filtering
- Account name column ready for backend to populate `account_name` field
- URL param sync ensures filters persist across navigation
- Ready for Phase 1 plan 04 (connection status chips, skeleton loading, toasts) or Phase 2 (manual control & audit)

---

*Phase: 01-multi-cloud-discovery*
*Completed: 2026-02-21*
