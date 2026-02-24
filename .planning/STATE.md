# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-23)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** v1.1 - Enhanced Insights & Savings (milestone complete)
**Recent work:** Debugging - Instance Details page 404 error (UUID lookup), savings page 500 error (type assertion)

## Current Position

Phase: 9 of 9 (Complete Savings Removal & Cache Validation)
Plan: 1 of 1 in current phase (Phase complete)
Status: Phase complete
Last activity: 2026-02-24 - Completed 09-01-PLAN.md (complete savings removal & cache validation)

Progress: [██████████████████████████████████] 9/9 phases complete

**Next Action:** v1.1 milestone complete - run `/gsd-audit-milestone` or `/gsd-complete-milestone`

## Performance Metrics

**Velocity:**
- Total plans completed: 28
- Average duration: ~16 min
- Total execution time: ~7 hours

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
| 9 | 1/1 | 1 | ~10 min |

**Recent Trend:**
- Last 28 plans: 28 complete
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
- Core visualization components: SavingsSummaryCards (2 cards), SavingsChart, SavingsTable
- **Phase 8 Plan 04 (08-04):** SavingsPage integration with all components, route registration (/savings), Navigation.tsx link with PiggyBank icon

**Quick Task #001 - Simplify Savings Page (2026-02-24):**
- Removed "Top Savings Instances" card from SavingsPage
- Removed "Cost Comparison" card from SavingsPage  
- Created simplified SavingsTable component for displaying top savings
- Updated SavingsPage.tsx to use SavingsSummaryCards (2 cards: Total, Ongoing) and SavingsChart
- Removed InstanceSavingsTable and CostProjection components
- Updated SpendingPage to fetch only summary, daily, and top savers data

**Quick Task #002 - Remove Savings Page (2026-02-24):**
- Removed SavingsPage import from web/src/main.tsx
- Removed /savings route from Routes configuration
- Removed PiggyBank icon import from Navigation.tsx
- Removed Savings navigation link from Navigation.tsx
- Deleted web/src/pages/SavingsPage.tsx
- Deleted all components in web/src/components/savings/ directory
- Application builds successfully without any savings-related code

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

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 001 | Simplify savings page - remove Top Savings and Cost Comparison cards | 2026-02-24 | TBA | [001-simplify-savings-page](./quick/001-simplify-savings-page/) |
| 002 | Remove Savings page entirely | 2026-02-24 | 0bdab629, 20cd9762 | [002-remove-savings-page](./quick/002-remove-savings-page/) |

**Quick Task #001 - Savings Page Simplification (2026-02-24):**
- Removed "Top Savings Instances" card (InstanceSavingsTable)
- Removed "Cost Comparison" card (CostProjection)
- Simplified to 2 summary cards: Total Savings, Ongoing Savings
- Added savings table at bottom showing top 5 saving instances

**Quick Task #002 - Remove Savings Page (2026-02-24):**
- Removed SavingsPage import and route from web/src/main.tsx
- Removed PiggyBank import and Savings link from Navigation.tsx
- Deleted SavingsPage.tsx and all components in web/src/components/savings/
- Application builds successfully without savings feature

## Blockers/Concerns Carried Forward

None - All blockers resolved by Phase 9 complete savings removal.

### Roadmap Evolution

- Phase 9 added (2026-02-24): Complete Savings Removal & Cache Validation - systematically remove all remaining savings code and rebuild Docker containers

**Debugging Session (2026-02-24):**
- Fixed `/instances/{id}` endpoint to use `GetInstanceByID` for app-generated UUID lookups
- Created missing `metrics_hourly` table for metrics storage
- Added metrics display to Instance Detail page showing CPU, connections, IOPS

## Session Continuity

Last session: 2026-02-24
Stopped at: Completed 09-01-PLAN.md (Complete Savings Removal & Cache Validation)
Resume file: None

### Phase 9 - Complete Savings Removal & Cache Validation (2026-02-24)

**Quick Task #003 - Complete Savings Removal (Phase 9):**
- Removed savings API methods: getSavingsSummary, getDailySavings, getOngoingCost, getSavingsByInstance, getInstanceSavings
- Removed savings interfaces: SavingsSummary, DailySavingsResponse, InstanceSavingsItem, InstanceSavingsDetail
- Removed backend savings handler: internal/api/handlers/savings.go
- Removed backend savings store: internal/store/savings_store.go
- Removed savings package: internal/savings/calculator.go, event_decorator.go
- Removed savings routes: GET /savings, /savings/daily, /savings/by-instance, /savings/ongoing, /instances/{id}/savings
- Updated CostOverTimeChart to use estimated cost calculation instead of savings API
- Rebuilt Docker containers with --no-cache flag

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

**Quick Task #001 (`web/src/pages/SavingsPage.tsx` & components):**
- Removed `InstanceSavingsTable` component import and usage
- Removed `CostProjection` component import and usage  
- Created new `SavingsTable` component with simplified table layout
- Updated to only fetch: summary, daily, and top savings data

**Quick Task #002 (`web/src/main.tsx`, `web/src/components/Navigation.tsx`):**
- Removed `SavingsPage` import and /savings route
- Removed `PiggyBank` icon import and Savings navigation link
- Deleted SavingsPage.tsx and all web/src/components/savings/ components

## Pending Actions

None - All pending actions completed.

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
