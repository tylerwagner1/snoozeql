# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 4 - Advanced Schedule Filtering (execution complete, awaiting verification)

## Current Position

Phase: 4 of 6 (Advanced Schedule Filtering)
Plan: 3 of 3 in current phase (in progress)
Status: Awaiting human verification
Last activity: 2026-02-23 — Completed Plan 03: FilterBuilder integration + instance counts

Progress: [████████████████████] 4/6 phases complete

## Performance Metrics

**Velocity:**
- Total plans completed: 17
- Average duration: ~15 min
- Total execution time: ~4 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 5/5 | 5 | ~15 min |
| 3 | 3/3 | 3 | ~15 min |
| 4 | 3/3 | 3 | ~15 min |

**Recent Trend:**
- Last 17 plans: 17 complete
- Trend: Stable
- Phase 1 complete: 2026-02-21
- Phase 2 complete: 2026-02-23
- Phase 3 complete: 2026-02-23
- Phase 4 complete: 2026-02-23

## Accumulated Context

### Decisions

- [04-03]: FilterBuilder placement below time selection in ScheduleModal
- [04-03]: Instance fetching per-open for preview, not cached across sessions
- [04-03]: Instance count calculation using matchInstance from filterUtils
- [04-03]: Empty state display shows "No filters" when selectors is empty

- [04-02]: Field types: name, provider, region, engine, tags
- [04-02]: Match types: exact, contains, prefix, suffix, regex
- [04-02]: Regex validation: Debounced inline validation with error messages
- [04-02]: Preview display: First 5 instances with "show more" expansion
- [04-02]: AND/OR operator: Visible between rules when multiple rules exist

- [04-01]: Backend regex matching: Go regexp.Compile for RE2 syntax
- [04-01]: Client-side preview: Fetch all instances and filter client-side
- [04-01]: Operator semantics: AND = all selectors must match; OR = any selector matches
- [04-01]: Empty selectors: Return false (require explicit selection) rather than match all
- [04-01]: Case sensitivity: Go regex is case-sensitive; JS uses 'i' flag for preview

- [04-01]: Backend regex matching: Go regexp.Compile for RE2 syntax
- [04-01]: Client-side preview: Fetch all instances and filter client-side
- [04-01]: Operator semantics: AND = all selectors must match; OR = any selector matches
- [04-01]: Empty selectors: Return false (require explicit selection) rather than match all
- [04-01]: Case sensitivity: Go regex is case-sensitive; JS uses 'i' flag for preview

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

**Phase 4 - Advanced Schedule Filtering:**
- Backend matcher logic with MatchInstance, MatchSelector, ValidateSelectors
- Filter preview API endpoint: POST /api/v1/schedules/preview-filter
- FilterBuilder component for visual filter creation
- FilterRule component with field type, match type, pattern inputs
- FilterPreview component showing matched instance count and list
- AND/OR toggle for combining multiple rules
- Live client-side preview with instant updates
- ScheduleModal integration with FilterBuilder section
- SchedulesPage with instance count column
- Regex validation with inline error messages

**Plan 01-03 (2026-02-23):**
- Plan 01: Backend matcher and filter utilities complete ✅
- Plan 02: Filter components complete ✅
- Plan 03: ScheduleModal integration + instance counts complete ✅

## Session Continuity

Last session: 2026-02-23
Stopped at: Completed Plan 03 - FilterBuilder integration + instance counts (awaiting verification)

**Phase 4 Execution Summary:**

3 plans executed in 3 waves:
- **Wave 1 (04-01):** Backend matcher.go + filterUtils.ts + preview API endpoint ✅ COMPLETED
- **Wave 2 (04-02):** FilterBuilder, FilterRule, FilterPreview components ✅
- **Wave 3 (04-03):** ScheduleModal integration + SchedulesPage instance counts + verification ⏸️

**Status:** Plan 03 execution complete, awaiting human verification at checkpoint.

**To verify:**
1. Start development servers: `go run cmd/server/main.go` and `npm run dev`
2. Navigate to http://localhost:5173/schedules
3. Click "Create Schedule" to open ScheduleModal
4. Verify FilterBuilder section appears below time selection
5. Test adding filter rules with different field types (name, provider, region, engine, tag)
6. Test AND/OR toggle for multiple rules
7. Type invalid regex pattern like "["
8. Save a schedule with filters
9. Verify instance count column shows correct numbers in SchedulesPage
10. Edit existing schedule to verify filters load correctly

**Ready for:** Type "approved" to complete Phase 4

*`/new` first - fresh context window*
