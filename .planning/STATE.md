# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 4 - Advanced Schedule Filtering (planning complete, ready for execution)

## Current Position

Phase: 4 of 6 (Advanced Schedule Filtering)
Plan: 0 of 3 in current phase (planned, not started)
Status: Ready for execution
Last activity: 2026-02-23 — Completed Phase 4 planning

Progress: [████████████████████] 3/6 phases complete

## Performance Metrics

**Velocity:**
- Total plans completed: 14
- Average duration: ~15 min
- Total execution time: ~3.5 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 5/5 | 5 | ~15 min |
| 3 | 3/3 | 3 | ~15 min |
| 4 | 0/3 | - | - |

**Recent Trend:**
- Last 14 plans: 14 complete
- Trend: Stable
- Phase 1 complete: 2026-02-21
- Phase 2 complete: 2026-02-23
- Phase 3 complete: 2026-02-23

## Accumulated Context

### Decisions

- [04-planning]: Filter builder uses visual chips for rules with AND/OR toggle
- [04-planning]: Client-side preview for instant feedback, backend validates on save
- [04-planning]: Filters stored in schedule.selectors array (existing model)
- [04-planning]: Matcher types: exact, contains, prefix, suffix, regex

- [03-03]: Schedule creation uses modal dialog for better UX
- [03-03]: Schedule list shows active days and sleep hours summaries
- [03-03]: Grid is source of truth, CRON mode is for power users

- [03-02]: ScheduleModal uses Headless UI Dialog matching ConfirmDialog styling
- [03-02]: Modal supports create and edit modes
- [03-02]: Toggle between grid and CRON modes with data preservation
- [03-02]: cronstrue for human-readable CRON descriptions

- [03-01]: gridToCron handles overnight schedules (sleep crosses midnight)
- [03-01]: formatGridSummary provides active days (Weekdays, Weekends, Every day)
- [03-01]: Simple CRON format for Phase 3 (single contiguous window per day)

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

- [01-01]: Use concrete types instead of interfaces for store injection
- [01-01]: Connection status values: "connected", "syncing", "failed", "unknown"
- [01-01]: Discovery service automatically updates account status

- [01-02]: Provider key format includes account ID
- [01-02]: GCP provider accepts serviceAccountJSON parameter
- [01-02]: Registry has Unregister method for dynamic provider re-registration

- [01-03]: Filter state uses URL params
- [01-03]: Status filter supports multiple value mappings

- [01-04]: Connection status chips with colors
- [01-04]: Toast notifications using react-hot-toast

- [01-05]: Dashboard stats cards are clickable with useNavigate
- [01-05]: CTAs shown when no accounts exist

- [Init]: Use existing Go/React/PostgreSQL stack
- [Init]: Single-user POC scope (no RBAC complexity)

### Pending Todos

None yet.

### Blockers/Concerns

From Phase 3 research (deferred to future phases):
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

**Phase 3 - Basic Scheduling:**
- WeeklyScheduleGrid component with 7×24 visual grid
- Click-drag painting for scheduling multiple cells
- Document-level mouseup listener for reliable drag termination
- cronUtils.ts with grid↔CRON conversion utilities (gridToCron, cronToGrid, formatGridSummary, describeCron)
- Nighttime (overnight) schedule handling
- Active days summary (Weekdays, Weekends, Every day)
- ScheduleModal with create/edit modes
- Grid and CRON mode toggle with data preservation
- cronstrue integration for human-readable CRON descriptions
- SchedulesPage integration with modal for create/edit

## Session Continuity

Last session: 2026-02-23
Stopped at: Phase 4 planning complete
Resume file: None

**Phase 4 Planning Summary:**

3 plans created in 3 waves:
- **Wave 1 (04-01):** Backend matcher.go + filterUtils.ts + preview API endpoint
- **Wave 2 (04-02):** FilterBuilder, FilterRule, FilterPreview components
- **Wave 3 (04-03):** ScheduleModal integration + SchedulesPage instance counts + verification

**Ready for:** `/gsd-execute-phase 4`

*`/new` first - fresh context window*
