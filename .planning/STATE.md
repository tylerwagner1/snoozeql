# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 3 - Basic Scheduling (plan 01 complete)

## Current Position

Phase: 3 of 6 (Basic Scheduling)
Plan: 1 of 3 in current phase
Status: In progress
Last activity: 2026-02-23 — Completed 03-01-PLAN.md (CRON utilities and schedule grid)

Progress: [████████████░░░░░░░] 2/6 phases complete

## Performance Metrics

**Velocity:**
- Total plans completed: 12
- Average duration: ~15 min
- Total execution time: ~2.9 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 5/5 | 5 | ~15 min |
| 3 | 1/3 | 3 | - |

**Recent Trend:**
- Last 12 plans: 12 complete
- Trend: Stable
- Phase 1 complete: 2026-02-21
- Phase 2 complete: 2026-02-23

## Accumulated Context

### Decisions

- [02-05]: AuditLogPage uses client-side filtering (no additional API calls)
- [02-05]: Event icons use moon/sun SVG with color coding
- [02-05]: Navigation link in header nav (Navigation.tsx)

- [02-04]: Used Set<string> for selectedIds state for O(1) lookups
- [02-04]: Bulk action buttons show count of actionable instances
- [02-04]: Confirmation dialog shows exact count of instances being operated on

- [02-03]: Uses instance.ID (UUID) for event logging
- [02-03]: Uses instance.ProviderID for cloud API calls
- [02-03]: Continues processing other instances if one fails (partial success handling)

- [02-01]: Add EventStore after InstanceStore for consistent code organization
- [02-01]: Use limit/offset pagination for events endpoint (50 rows default, 100 max)
- [02-01]: Store Event metadata as JSONB in PostgreSQL

- [01-01]: Use concrete types instead of interfaces for store injection in DiscoveryService to avoid circular imports
- [01-01]: Connection status values: "connected" (success), "syncing" (running), "failed" (error), "unknown" (default)
- [01-01]: Discovery service automatically updates account status to "syncing" before discovery and "connected"/"failed" after

- [01-02]: Provider key format now includes account ID
- [01-02]: GCP provider now accepts serviceAccountJSON parameter
- [01-02]: Registry now has Unregister method for dynamic provider re-registration support

- [01-03]: Filter state uses URL params for better user experience
- [01-03]: Status filter supports multiple value mappings

- [01-04]: Connection status chips with colors for connected/syncing/failed/unknown states
- [01-04]: Toast notifications using react-hot-toast

- [01-05]: Dashboard stats cards are clickable with useNavigate
- [01-05]: CTAs shown when no accounts exist

- [03-01]: Simplified CRON conversion for Phase 3 - assumes single contiguous sleep window per day. Complex multi-block schedules deferred to future phases.

- [Init]: Use existing Go/React/PostgreSQL stack
- [Init]: Single-user POC scope (no RBAC complexity)
- [Init]: Regex-based schedule assignment (user-requested flexible filtering)

### Pending Todos

None yet.

### Blockers/Concerns

From research (deferred to Phase 3 for implementation):
- AWS 7-day auto-restart: implement re-stop mechanism
- Instance state race conditions: implement proper state machine
- Read replicas cannot be stopped: flag in discovery
- Storage charges continue: clear UI messaging

### Completed Features

**Phase 1 - Multi-Cloud Discovery:**
- Instance persistence with database syncing
- Multi-account provider registration (AWS + GCP)
- Sortable/filterable instances table with account column
- Connection status tracking with chips and toasts
- Clickable dashboard stats with URL-based filtering
- CTAs for adding cloud accounts

**Phase 2 - Manual Control & Audit:**
- EventStore with CreateEvent, ListEvents, ListEventsByInstance methods
- GET /api/v1/events endpoint with pagination
- ConfirmDialog component with Headless UI
- POST /api/v1/instances/bulk-stop endpoint with state validation
- POST /api/v1/instances/bulk-start endpoint with state validation
- Multi-select table with checkbox column in InstancesPage
- Bulk action buttons showing count of actionable instances
- Confirmation dialogs for bulk sleep/wake operations
- Toast notifications for operation success/failure
- AuditLogPage with event filtering
- Navigation link to audit log

**Phase 3 - Basic Scheduling (in progress):**
- WeeklyScheduleGrid component with 7×24 visual grid
- Click-drag painting for scheduling multiple cells
- Document-level mouseup listener for reliable drag termination
- cronUtils.ts with grid↔CRON conversion utilities
- Nighttime (overnight) schedule handling
- Active days summary (Weekdays, Weekends, Every day)

## Session Continuity

Last session: 2026-02-23
Stopped at: Completed 03-01-PLAN.md (CRON utilities and weekly schedule grid)
Resume file: None

**Next Phase Readiness:**
- Phase 3 plan 01 complete ✅
- WeeklyScheduleGrid component with click-drag painting ✅
- CRON conversion utilities with grid↔CRON round-trip ✅
- Ready for 03-02-PLAN.md (ScheduleModal with grid and CRON mode)

**Phase 3 Goal:** Users can create time-based sleep/wake schedules
**Phase 3 Success Criteria:**
1. User can create a schedule specifying start time, end time, and days of week
2. Created schedules appear in the schedules list

**Ready for:** 03-02-PLAN.md (ScheduleModal with grid integration)
