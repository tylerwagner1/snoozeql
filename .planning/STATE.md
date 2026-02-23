# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 6 - Intelligent Recommendations (plan 04 checkpoint pending)

## Current Position

Phase: 6 of 6 (Intelligent Recommendations)
Plan: 04 (final - checkpoint)
Status: In progress
Last activity: 2026-02-23 - Completed 06-04-PLAN.md checkpoint (human verification)

Progress: [██████████████████████████████] 6/6 phases complete

## Performance Metrics

**Velocity:**
- Total plans completed: 19
- Average duration: ~15 min
- Total execution time: ~4.75 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 5/5 | 5 | ~15 min |
| 3 | 3/3 | 3 | ~15 min |
| 4 | 3/3 | 3 | ~15 min |
| 5 | 3/3 | 3 | ~15 min |

**Recent Trend:**
- Last 20 plans: 20 complete
- Trend: Stable
- Phase 1 complete: 2026-02-21
- Phase 2 complete: 2026-02-23
- Phase 3 complete: 2026-02-23
- Phase 4 complete: 2026-02-23
- Phase 5 complete: 2026-02-23
- Phase 6 progress: Plan 04 complete (checkpoint awaiting human verification)

## Accumulated Context

### Decisions

- [06-03]: Recommendations section shows up to 3 cards on dashboard with "View all" link
- [06-03]: Empty state on dashboard shows "Need 24+ hours of activity data" message
- [06-03]: Both dashboard and RecommendationsPage track dismissed recommendations count
- [06-03]: Dismissed count displayed in RecommendationsPage subtitle and empty state
- [05-02]: GetMetricStatistics API for single-metric queries (simpler than GetMetricData for this use case)
- [05-02]: 3 retry attempts with exponential backoff for throttling and transient errors
- [05-02]: 1-hour period fetching with hourly aggregation, returning hourly averages
- [05-02]: Client caching by account+region combination to minimize AWS configuration overhead
- [05-02]: MetricsCollector background service pattern following existing DiscoveryService

- [05-01]: Hourly aggregation with incremental averaging for UPSERT
- [05-01]: Unique constraint on (instance_id, metric_name, hour) for idempotent inserts
- [05-01]: MetricsStore pattern matching existing store implementations (InstanceStore, EventStore)

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

**Phase 5 - Activity Analysis:**
- Plan 01 (05-01): Database schema (metrics_hourly table), HourlyMetric model, MetricsStore ✅ COMPLETED
- Plan 02 (05-02): CloudWatch client, MetricsCollector service, integration in main.go ✅ COMPLETED
- Plan 03 (05-03): Idle period detection algorithms ✅ COMPLETED

**Phase 6 Plans Complete:**
- Plan 01 (06-01): Backend recommendation generation and API handlers ✅ COMPLETED
- Plan 02 (06-02): Frontend components (RecommendationCard, RecommendationModal, ActivityGraph) ✅ COMPLETED
- Plan 03 (06-03): Dashboard and RecommendationsPage integration ✅ COMPLETED

## Session Continuity

Last session: 2026-02-23T17:10:33Z
Stopped at: Completed 06-04-PLAN.md checkpoint (human verification)

**Phase 6 Execution Summary:**

Plan 01 (06-01) - Backend recommendation generation and API handlers ✅ COMPLETED
- Added GenerateRecommendations() method to Analyzer service
- Added /recommendations/generate, /recommendations/:id/dismiss, /recommendations/:id/confirm endpoints
- Updated RecommendationStore with ListRecommendationsByStatus method

Plan 02 (06-02) - Frontend recommendation components ✅ COMPLETED
- Created RecommendationEnriched TypeScript interface
- Implemented RecommendationCard with expand/collapse for list display
- Implemented RecommendationModal with schedule confirmation flow
- Implemented ActivityGraph for 24-hour CPU visualization
- Updated Dashboard.tsx and RecommendationsPage.tsx to use new types

Plan 03 (06-03) - Dashboard and RecommendationsPage integration ✅ COMPLETED
- Added AI Recommendations section to Dashboard with generate button
- Dashboard displays up to 3 recommendation cards with expand/collapse
- RecommendationsPage shows full list with generate button and dismissed count
- Both pages implement dismiss (remove + increment dismissed count) and confirm (creates schedule) workflows
- Proper empty states with data requirement message and generate button

Plan 04 (06-04) - Human verification checkpoint ⏸️ AWAITING
- Human-verify complete recommendation workflow
- 7 test scenarios covering dashboard, recommendations page, and schedule creation
- Verification requires running end-to-end workflow
- ** waiting for user approval to complete Phase 6**

**Phase 6: In progress - awaiting human verification**

**Files created:**
- internal/analyzer/patterns.go (Phase 5)
- internal/metrics/cloudwatch.go (05-02)
- internal/metrics/collector.go (05-02)
- internal/analyzer/recommendation.go (06-01)
- web/src/components/ActivityGraph.tsx (06-02)
- web/src/components/RecommendationCard.tsx (06-02)
- web/src/components/RecommendationModal.tsx (06-02)

**Files modified:**
- internal/analyzer/analyzer.go (06-01 - added GenerateRecommendations)
- cmd/server/main.go (05-02 - added metrics initialization)
- internal/models/models.go (05-01 - added HourlyMetric)
- web/src/lib/api.ts (06-01 - added RecommendationEnriched type)
- web/src/pages/Dashboard.tsx (06-03 - added recommendations section)
- web/src/pages/RecommendationsPage.tsx (06-03 - refactored with new components)

*Phase 6: In progress - awaiting human verification*
