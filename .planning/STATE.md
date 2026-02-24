# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-23)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** v1.1 - Enhanced Insights & Savings (roadmap created, ready for phase planning)
**Recent work:** Debugging - Instance Details page 404 error (UUID lookup), savings page 500 error (type assertion)

## Current Position

Phase: 8 of 8 (Dashboard & Visualization)
Plan: 4 of 4 in current phase (checkpoint: human-verify)
Status: All tasks complete, checkpoint reached for human verification
Last activity: 2026-02-24 - Completed debugging fixes for savings page 500 error

Progress: [██████████████████████████████████] 8/8 phases complete

## Performance Metrics

**Velocity:**
- Total plans completed: 27
- Average duration: ~16 min
- Total execution time: ~6.5 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 6/6 | 6 | ~16 min |
| 2 | 5/5 | 5 | ~15 min |
| 3 | 3/3 | 3 | ~15 min |
| 4 | 3/3 | 3 | ~15 min |
| 5 | 3/3 | 3 | ~15 min |
| 6 | 4/4 | 4 | ~15 min |
| 7 | 3/3 | 3 | ~13 min |
| 8 | 4/4 | 4 | ~16 min |

**Recent Trend:**
- Last 26 plans: 26 complete
- Trend: Stable

## Accumulated Context

### Completed Features

**Phase 1 - Multi-Cloud Discovery:**
- Instance persistence with database syncing
- Multi-account provider registration (AWS + GCP)
- Sortable/filterable instances table with account column
- Connection status tracking with chips and toasts

**Phase 2 - Manual Control & Audit:**
- EventStore with CreateEvent, ListEvents methods
- ConfirmDialog component with Headless UI
- Bulk stop/start API endpoints with audit logging

**Phase 3 - Basic Scheduling:**
- WeeklyScheduleGrid component with 7×24 visual grid
- ScheduleModal with grid and CRON mode
- SchedulesPage integration

**Phase 4 - Advanced Schedule Filtering:**
- Backend matcher logic with MatchInstance, MatchSelector
- FilterBuilder, FilterRule, FilterPreview components
- ScheduleModal integration with filter preview

**Phase 5 - Activity Analysis:**
- metrics_hourly table with hourly aggregation
- CloudWatch client with 3 retries and backoff
- Idle period detection algorithms

**Phase 6 - Intelligent Recommendations:**
- Recommendation generation from idle patterns
- RecommendationCard with expand/collapse
- RecommendationModal with ActivityGraph visualization
- Dashboard and RecommendationsPage integration

**Phase 7 - Core Savings Calculation:**
- Migration 006 with time-range indexes and materialized view
- SavingsStore with upsert/query methods and daily accumulation
- SavingsCalculator with CalculateSavings, CalculateOngoingSavings, SplitByDay
- Integer cents pattern for all money calculations
- 7-day cap based on AWS auto-restart limits
- EventStoreWithSavings decorator for automatic savings calculation on stop/wake events
- EventStoreWithSavings intercepts CreateEvent: captures hourly_rate_cents in stop events (AUD-02)
- EventStoreWithSavings calculates and persists savings on start/wake events (AUD-01)
- SavingsHandler with 4 API endpoints for summary, daily, by-instance, and instance detail
- Routes registered: GET /api/v1/savings, /savings/daily, /savings/by-instance, /instances/{id}/savings

**Phase 8 - Dashboard & Visualization:**
- Savings API types: SavingsSummary, DailySavingsResponse, InstanceSavingsItem, InstanceSavingsDetail
- Savings API methods: getSavingsSummary, getDailySavings, getSavingsByInstance, getInstanceSavings
- Currency formatter: formatCurrency using Intl.NumberFormat (centsToDollars)
- DateRangeSelector component: 7d/30d/90d tab navigation
- Three main visualization components: SavingsSummaryCards, SavingsChart, InstanceSavingsTable
- All components handle loading and empty states gracefully
- **Phase 8 Plan 04 (08-04):** SavingsPage integration with all components, route registration (/savings), Navigation.tsx link with PiggyBank icon

**Phase 8 Debugging & Fixes (2026-02-24):**
- Instance Details page 404 error: Fixed by changing endpoint from `GetInstanceByProviderID` to `GetInstanceByID` (backend uses app-generated UUID, not provider ID)
- Added metrics_hourly table to database (was missing from schema)
- Added `/api/v1/instances/{id}/metrics` endpoint to fetch latest metrics for an instance
- Updated GetLatestMetrics to return empty slice instead of nil when no metrics exist
- Added metrics display to InstanceDetailPage.tsx showing CPU, connections, IOPS with stats
- Added `HourlyMetric` interface and `getInstanceMetrics` API method to frontend

### Decisions Made

**Phase 8 Debugging & Fixes (2026-02-24):**

| Decision | Rationale |
|----------|-----------|
| Use `GetInstanceByID` instead of `GetInstanceByProviderID` for `/instances/{id}` | Frontend uses app-generated UUID (`id` column), not provider identifier (`provider_id` column) |
| Create `metrics_hourly` table in database | Table was referenced in code but missing from schema |
| Return empty slice instead of nil for missing metrics | Consistent JSON serialization (empty array `[]` vs null) |
| Add metrics display to InstanceDetailPage | Shows data used for recommendations (CPU, connections, IOPS stats) |
| Rewrite GetDailySavings to use intermediate variable | Fix circular type assertion panic - build slice separately before assignment |

**Phase 7 - Core Savings Calculation**

| Decision | Rationale |
|----------|-----------|
| EventStoreWithSavings decorator pattern | Automatic savings calculation on event creation instead of dashboard load time (push model) |
| EventCreator interface | Flexible event store types - allows decorator wrapping |
| Full SavingsHandler with 4 endpoints | Required for Phase 8 dashboard API consumption |

**Phase 8 - Dashboard & Visualization:**

| Decision | Rationale |
|----------|-----------|
| API methods use GET requests to Phase 7 backend | No new backend changes needed; reuses existing endpoints |
| Currency formatting uses Intl.NumberFormat | Proper locale handling (thousands separators, currency symbols) |
| Date range selector uses tab-style design | Matches existing Dashboard.tsx card styling pattern |
| SavingsSummaryCards follows Dashboard.tsx card styling | UI consistency with existing summary cards |
| SavingsChart uses green theme (#10b981) | Visual consistency with ActivityGraph |
| InstanceSavingsTable ranked rows | Clear attribution of top-saving instances |
| CostProjection component follows existing styling pattern | Visual consistency for cost comparison section |
| Yellow/yellow-400 theme for disclaimer | Warning styling for SAV-05 requirement adherence |
| SavingsPage imports and assembles all Phase 8 visualization components | Complete dashboard integration |
| Navigation.tsx added PiggyBank icon for Savings link | Clear visual indicator for savings-related navigation |

## Blockers/Concerns Carried Forward

None - Phase 8 complete (all 4 plans executed). SavingsPage integrated with all visualization components, route registration complete, Navigation link added.

**Debugging Session (2026-02-24):**
- Fixed `/instances/{id}` endpoint to use `GetInstanceByID` for app-generated UUID lookups
- Created missing `metrics_hourly` table for metrics storage
- Added metrics display to Instance Detail page showing CPU, connections, IOPS

## Session Continuity

Last session: 2026-02-24
Stopped at: Completed Instance Details page debugging and metrics display
Resume file: None

### Changes Made (2026-02-24)

**Backend (`cmd/server/main.go`):**
- Line 329: Changed `instanceStore.GetInstanceByProviderID(ctx, "", instanceID)` to `instanceStore.GetInstanceByID(ctx, instanceID)`
- Added new metrics endpoint at line 683-702 for `/instances/{id}/metrics`

**Frontend (`web/src/lib/api.ts`):**
- Added `HourlyMetric` interface (lines 171-181)
- Added `getInstanceMetrics` API method (line 293)

**Frontend (`web/src/pages/InstanceDetailPage.tsx`):**
- Added metrics state and API call to load metrics
- Added Metrics section displaying CPU, connections, IOPS with avg/min/max values

## Pending Actions

- [ ] Human verification of SavingsPage visual correctness (checkpoint:human-verify)

### Server Status

| Service | URL | Status |
|---------|-----|--------|
| Go Backend API | http://localhost:8080 | ✅ Running (with metrics endpoint fix) |
| Next.js Frontend | http://localhost:3002 | ✅ Running |
| PostgreSQL Database | localhost:5432 | ✅ Running (metrics_hourly table created) |

### Database Changes

| Table | Change |
|-------|--------|
| metrics_hourly | Created new table with columns: id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count, created_at, updated_at |
