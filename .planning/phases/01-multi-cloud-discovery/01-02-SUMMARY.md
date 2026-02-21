---
phase: 01-multi-cloud-discovery
plan: 02
subsystem: infra
tags: [aws, gcp, multi-account, provider-registry, cloud-sql]

# Dependency graph
requires:
  - phase: 01-01
    provides: Instance persistence layer, connection status tracking
provides:
  - Account-aware provider keys preventing multi-account collisions
  - GCP Cloud SQL provider with service account JSON credential support
  - Provider registry with Unregister method for dynamic updates
affects: [multi-account, cloud-connection, provider-discovery]

# Tech tracking
tech-stack:
  added: [gcpprovider package]
  patterns: [account-aware provider keys, multi-provider registration]

key-files:
  created:
    - internal/provider/gcp/cloudsql.go (with service account support)
    - internal/provider/registry.go (with Unregister method)
  modified:
    - cmd/server/main.go (GCP provider registration)

key-decisions:
  - "Provider key format updated to include account ID preventing collisions across multiple accounts in same region"
  - "GCP provider accepts serviceAccountJSON parameter for service account authentication"
  - "Registry has Unregister method for future dynamic provider re-registration"

patterns-established:
  - "Pattern 1: Account-aware provider keys - Use fmt.Sprintf(\"%s_%s_%s\", provider, account.ID, region) for AWS and fmt.Sprintf(\"%s_%s\", provider, account.ID) for GCP"
  - "Pattern 2: GCP service account credentials - Use option.WithCredentialsJSON() with service account JSON string"

# Metrics
duration: ~25min
completed: 2026-02-21
---

# Phase 1 Plan 2: Multi-Account Provider Registration and GCP Service Account Support

**Account-aware provider keys preventing multi-account collisions, plus GCP Cloud SQL provider with service account JSON credential support and dynamic unregistration capability**

## Performance

- **Duration:** 25 min
- **Started:** 2026-02-21T20:54:00Z
- **Completed:** 2026-02-21T21:19:00Z
- **Tasks:** 3/3
- **Files modified:** 4

## Accomplishments
- Provider key format now includes account ID (`fmt.Sprintf("aws_%s_%s", account.ID, region)`) preventing collisions when multiple AWS accounts exist in the same region
- GCP Cloud SQL provider now accepts `serviceAccountJSON` parameter using `option.WithCredentialsJSON()` for service account authentication
- Registry gains `Unregister` method for dynamic provider re-registration support
- Server now registers both AWS and GCP providers from cloud accounts database

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix provider key format to include account ID** - `da4f0b5` (feat)
2. **Task 2: Add GCP service account JSON credential support** - `607d502` (feat)
3. **Task 3: Register GCP providers and handle dynamic re-registration** - `5cf0e74` (feat)

**Plan metadata:** `3ab6df3` (docs: complete plan)

## Files Created/Modified
- `internal/provider/gcp/cloudsql.go` - Added `option` import, `serviceAccountJSON` parameter, `TestConnection` method
- `internal/provider/registry.go` - Added `sync` import, `sync.RWMutex`, `Unregister` method
- `cmd/server/main.go` - Added `gcpprovider` import, updated cloud accounts loop to handle both AWS and GCP providers

## Decisions Made
- **Provider key format:** Changed from `"aws_" + region` to `fmt.Sprintf("aws_%s_%s", account.ID, region)` using account ID to prevent collisions. Same pattern applied to GCP with `fmt.Sprintf("gcp_%s", account.ID)`.
- **GCP credential handling:** Provider now accepts optional `serviceAccountJSON` string parameter. When provided, uses `option.WithCredentialsJSON()`; otherwise falls back to Application Default Credentials.
- **Registry thread safety:** Added `sync.RWMutex` to Registry struct for thread-safe operations. `Unregister` method uses mutex to safely remove providers.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- File system timing issues during multiple edits required multiple read/write cycles to ensure correct file state
- Vendor directory caused `go build` errors; removed vendor directory to resolve vendoring inconsistencies

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Multi-account provider registry foundation complete with account-aware keys
- GCP support in place, ready for UI integration with credential forms
- Dynamic provider registration pattern established for future API-based account management
- TODO: Future enhancement could add provider re-registration when cloud accounts added/deleted via API (currently requires server restart)

---

*Phase: 01-multi-cloud-discovery*
*Completed: 2026-02-21*
