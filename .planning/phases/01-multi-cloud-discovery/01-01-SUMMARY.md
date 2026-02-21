---
phase: 01-multi-cloud-discovery
plan: 01
subsystem: database
tags: postgres, go, aws, gcp, discovery, persistence

# Dependency graph
requires:
  - phase: 01-multi-cloud-discovery
    provides: Basic discovery infrastructure
provides:
  - Instance persistence to PostgreSQL database
  - Connection status tracking for cloud accounts
  - Discovery service wired to database operations
affects:
  - 01-02: GCP credentials support
  - 01-03: Instances table UI enhancements
  - 01-04: Connection status chips UI

# Tech tracking
tech-stack:
  added:
    - SQL column: connection_status
    - SQL column: last_sync_at
    - SQL column: last_error
    - InstanceStore struct with UpsertInstance, ListInstances, GetInstanceByProviderID
    - CloudAccountStore.UpdateConnectionStatus method
  patterns:
    - Discovery service integration with database store
    - Connection status lifecycle: syncing → connected/failed

key-files:
  created:
    - deployments/docker/migrations/002_connection_status.sql
  modified:
    - internal/store/postgres.go
    - internal/models/models.go
    - internal/discovery/discovery.go
    - cmd/server/main.go

key-decisions:
  - Used concrete types instead of interfaces for store injection to avoid circular imports
  - Connection status values: "connected", "syncing", "failed", "unknown"
  - Discovery service automatically updates status before and after each sync run

# Metrics
duration: 15 min
completed: 2026-02-21
---

# Phase 01 Plan 01: Instance Persistence and Connection Status Summary

**Instance persistence layer with database sync from discovery service and connection status tracking for cloud accounts**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-21T01:53:18Z
- **Completed:** 2026-02-21T02:08:18Z
- **Tasks:** 3/3
- **Files modified:** 4

## Accomplishments
- InstanceStore with UpsertInstance, ListInstances, GetInstanceByProviderID methods
- Connection status tracking for CloudAccount (connection_status, last_sync_at, last_error)
- Discovery service automatically persists instances during Run() and updates account status
- API endpoints updated to use database instances instead of live cloud API calls

## Task Commits

Each task was committed atomically:

1. **Task 1: Add InstanceStore with UpsertInstance and ListInstances** - `b618582` (feat)
2. **Task 1: Fix QueryRowContext method** - `258508d` (fix)
3. **Task 2: Add connection status tracking to CloudAccount** - `d2c984a` (feat)
4. **Task 3: Wire discovery service to persist instances and update connection status** - `5f29813` (feat)

**Plan metadata:** Final commit captures plan execution

## Files Created/Modified
- `internal/store/postgres.go` - Added InstanceStore with UpsertInstance, ListInstances, GetInstanceByProviderID; Added QueryRowContext method to Postgres; Updated CloudAccountStore
- `internal/models/models.go` - Added ConnectionStatus, LastSyncAt, LastError fields to CloudAccount model
- `internal/discovery/discovery.go` - Added instanceStore and accountStore fields; Run() now syncs instances and updates status
- `cmd/server/main.go` - Created store instances, wired discovery service, updated instances endpoint
- `deployments/docker/migrations/002_connection_status.sql` - New migration file for connection status columns

## Decisions Made
- Used concrete types (`*store.InstanceStore`, `*store.CloudAccountStore`) instead of interfaces in DiscoveryService to avoid circular import issues
- Connection status values: "connected" (success), "syncing" (running), "failed" (error), "unknown" (default)
- Discovery service automatically updates all accounts to "syncing" before discovery and "connected"/"failed" after
- Start/stop endpoints use instanceStore for lookup with fallback to discovery for compatibility

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added QueryRowContext method to Postgres wrapper**

- **Found during:** Task 1 verification (`go build`)
- **Issue:** Postgres struct was missing `QueryRowContext` method required by InstanceStore.UpsertInstance
- **Fix:** Added `QueryRowContext` method that delegates to `sql.DB.QueryRowContext`, made `QueryRow` delegate to it
- **Files modified:** internal/store/postgres.go
- **Verification:** Build now succeeds without errors
- **Committed in:** `258508d` (fix)

**2. [Rule 3 - Blocking] Fixed circular import by using concrete types**

- **Found during:** Task 3 implementation
- **Issue:** Attempting to use interfaces in discovery package caused circular imports with store package
- **Fix:** Changed DiscoveryService fields to use concrete types `*store.InstanceStore` and `*store.CloudAccountStore`, removed interface definitions
- **Files modified:** internal/discovery/discovery.go
- **Verification:** Build succeeds without import errors
- **Committed in:** `5f29813` (feat)

**3. [Rule 2 - Missing Critical] Removed unused gcpprovider import**

- **Found during:** Task 3 codereview
- **Issue:** `gcpprovider` import was declared but never used, causing compilation warning
- **Fix:** Removed unused import from cmd/server/main.go
- **Files modified:** cmd/server/main.go
- **Verification:** Build succeeds without warnings
- **Committed in:** `5f29813` (feat)

---

**Total deviations:** 3 auto-fixed (2 blocking, 1 missing critical)

## Issues Encountered
- None - all issues were auto-fixed via deviation rules

## Next Phase Readiness
- Instance persistence layer complete, ready for UI enhancements
- Connection status tracking ready for UI chips display
- Discovery service now persistent, background sync continues as designed
- No blockers identified

---
*Phase: 01-multi-cloud-discovery*
*Completed: 2026-02-21*
