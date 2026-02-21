# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 2 - Manual Control & Audit (planning complete, ready for execution)

## Current Position

Phase: 2 of 6 (Manual Control & Audit)
Plan: 1 of 5 in current phase
Status: In progress
Last activity: 2026-02-21 — Completed 02-01-PLAN.md (EventStore and Events API)

Progress: [███████░░░] 50%

## Performance Metrics

**Velocity:**
- Total plans completed: 5
- Average duration: ~16 min
- Total execution time: 1.4 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 1/5 | 1 | ~15 min |

**Recent Trend:**
- Last 7 plans: 7 complete
- Trend: Stable
- Phase 1 complete: 2026-02-21
- Phase 2 in progress: 02-01 complete

*Updated after plan 02-01 completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [02-01]: Add EventStore after InstanceStore for consistent code organization
- [02-01]: Use limit/offset pagination for events endpoint (50 rows default, 100 max)
- [02-01]: Store Event metadata as JSONB in PostgreSQL

- [01-01]: Use concrete types instead of interfaces for store injection in DiscoveryService to avoid circular imports
- [01-01]: Connection status values: "connected" (success), "syncing" (running), "failed" (error), "unknown" (default)
- [01-01]: Discovery service automatically updates account status to "syncing" before discovery and "connected"/"failed" after

- [01-02]: Provider key format now includes account ID: fmt.Sprintf("aws_%s_%s", account.ID, region) and fmt.Sprintf("gcp_%s", account.ID)
- [01-02]: GCP provider now accepts serviceAccountJSON parameter using option.WithCredentialsJSON()
- [01-02]: Registry now has Unregister method for dynamic provider re-registration support

- [01-03]: Filter state uses URL params for better user experience (filters persist on refresh)
- [01-03]: Status filter supports multiple value mappings: 'running' maps to ['available', 'running', 'starting']

- [01-04]: Connection status chips with colors for connected/syncing/failed/unknown states
- [01-04]: Failed accounts show truncated error message on hover
- [01-04]: Skeleton loading cards appear during initial data fetch
- [01-04]: Toast notifications using react-hot-toast replace inline error display
- [01-04]: Toaster styled to dark theme (background #1e293b, border #334155)

- [01-05]: Dashboard stats cards are clickable with useNavigate for routing
- [01-05]: Running Databases card navigates to /instances?status=running
- [01-05]: Sleeping Databases card navigates to /instances?status=stopped
- [01-05]: Pending Actions card navigates to /recommendations
- [01-05]: Stats endpoint returns real counts from database (running/stopped mapping)
- [01-05]: CTAs shown when no accounts exist ("Get Started") and quick actions when accounts exist
- [01-05]: InstancesPage uses useSearchParams for URL-based filter state

- [Init]: Use existing Go/React/PostgreSQL stack (leverage working foundation)
- [Init]: Single-user POC scope (no RBAC complexity)
- [Init]: Regex-based schedule assignment (user-requested flexible filtering)

### Pending Todos

None yet.

### Blockers/Concerns

From research (deferred to Phase 3 for implementation):
- AWS 7-day auto-restart: implement re-stop mechanism (deferred to Phase 3)
- Instance state race conditions: implement proper state machine (deferred to Phase 3)
- Read replicas cannot be stopped: flag in discovery (deferred to Phase 3)
- Storage charges continue: clear UI messaging (deferred to Phase 3)

New from Plan 01-02:
- Multi-account provider registry now uses account-aware keys preventing collisions
- GCP provider supports service account JSON credentials
- Registry supports dynamic unregistration for future dynamic updates

New from Plan 01-03:
- Filter state synced with URL parameters for filter persistence
- Instance type enhanced with optional account_name field

New from Plan 01-04:
- Connection status chips with colors and icons for visual feedback
- Skeleton loading cards during data fetch for perceived performance
- Toast notifications using react-hot-toast library

New from Plan 01-05:
- Dashboard stats cards now clickable with useNavigate
- Stats endpoint returns real counts from database
- CTAs for cloud accounts implemented

New from Plan 01-06:
- Phase 1 verification checkpoint created
- Database migration 002_connection_status.sql applied successfully
- All components verified for end-to-end flow

New from Plan 02-01:
- EventStore with CreateEvent, ListEvents, ListEventsByInstance methods
- GET /api/v1/events endpoint with pagination
- Event interface and API client functions

## Session Continuity

Last session: 2026-02-21
Stopped at: Completed 02-01-PLAN.md (EventStore and Events API endpoint)
Resume file: None
Phase 2 in progress: Plan 02-01 complete, 4 remaining plans

**Next Phase Readiness:**
- Phase 2 in progress (1/5 plans complete)
- EventStore foundation ready for future plans
- Events API ready to be used by audit logging features
- Ready for plan 02-02 (ConfirmDialog component)

**Phase 2 Plans:**
- 02-01: EventStore and Events API endpoint - COMPLETE
- 02-02: ConfirmDialog component with Headless UI
- 02-03: Bulk stop/start API endpoints with audit logging
- 02-04: Multi-select and bulk actions in InstancesPage
- 02-05: AuditLogPage, navigation, and human verification checkpoint

**Ready for:** plan 02-02
