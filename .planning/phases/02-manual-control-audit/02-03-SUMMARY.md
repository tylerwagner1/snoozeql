---
phase: 02-manual-control-audit
plan: 03
subsystem: api
tags: [bulk-operations, audit-logging, golang, typescript, aws, gcp]

# Dependency graph
requires:
  - phase: 02
    provides: EventStore, GET /api/v1/events endpoint
provides:
  - POST /api/v1/instances/bulk-stop endpoint
  - POST /api/v1/instances/bulk-start endpoint
  - bulkStopInstances, bulkStartInstances functions in frontend
affects:
  - 02-04: Multi-select and bulk actions in InstancesPage
  - 02-05: AuditLogPage

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Bulk operation endpoint accepting array of instance IDs"
    - "Partial failure handling with success/failed arrays"
    - "Event creation for each successful bulk operation"

key-files:
  created: []
  modified:
    - cmd/server/main.go
    - web/src/lib/api.ts

key-decisions:
  - "Uses instance.ID (UUID) for event logging"
  - "Uses instance.ProviderID for cloud API calls"
  - "Continues processing other instances if one fails (partial success)"

patterns-established:
  - "BulkOperationRequest and Response pattern for multi-instance operations"
  - "State validation before attempting operations"

# Metrics
duration: ~8 min
completed: 2026-02-21
---

# Phase 2 Plan 3: Bulk Stop/Start API Summary

**Bulk stop/start API endpoints with audit logging for multi-instance operations**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-02-21T03:30:00Z
- **Completed:** 2026-02-21T03:38:00Z
- **Tasks:** 2/2 complete
- **Files modified:** 2 (cmd/server/main.go, web/src/lib/api.ts)

## Accomplishments
- BulkOperationRequest and Response types for multi-instance operations
- POST /api/v1/instances/bulk-stop endpoint with state validation
- POST /api/v1/instances/bulk-start endpoint with state validation
- Creates audit events (sleep/wake) for each successful operation
- Frontend has bulkStopInstances and bulkStartInstances API functions
- Partial failure handling with success/failed arrays
- All code compiles without errors (Go build and TypeScript build)

## Task Commits

1. **Task 1: Add bulk stop/start endpoints** - `be20fa3` (feat: add bulk endpoints)
2. **Task 2: Add bulk operations to frontend API client** - `b4b16e9` (feat: add bulk functions)

**Plan metadata:** `8558826` (docs: complete bulk stop/start plan)

## Files Created/Modified
- `cmd/server/main.go` - Added BulkOperationRequest/Response types, bulk-stop and bulk-start endpoints
- `web/src/lib/api.ts` - Added BulkOperationResponse interface, bulkStopInstances and bulkStartInstances functions

## Decisions Made
- Uses `instance.ID` (UUID) for event logging
- Uses `instance.ProviderID` for cloud API calls
- Continues processing other instances if one fails (partial success handling)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - execution was straightforward.

## Next Phase Readiness

- **Ready for:** 02-04-PLAN.md (Multi-select and bulk actions in InstancesPage)
- **No blockers or concerns**
- **Bulk operations foundation complete** - frontend can now use bulkStopInstances and bulkStartInstances

---

*Phase: 02-manual-control-audit*
*Completed: 2026-02-21*
