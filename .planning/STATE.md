# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-23)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** v1.1 - Enhanced Insights & Savings (roadmap created, ready for phase planning)

## Current Position

Phase: 8 of 8 (Dashboard & Visualization)
Plan: 4 of 4 in current phase (checkpoint: human-verify)
Status: All tasks complete, checkpoint reached for human verification
Last activity: 2026-02-23 - Completed Phase 8 (all plans)

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

### Decisions Made

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

## Session Continuity

Last session: 2026-02-23T21:48:00Z
Stopped at: Completed Phase 8 (all 4 plans)
Resume file: None

## Pending Actions

- [ ] Human verification of SavingsPage visual correctness (checkpoint:human-verify)

### Server Status

| Service | URL | Status |
|---------|-----|--------|
| Go Backend API | http://localhost:8080 | ✅ Running |
| Next.js Frontend | http://localhost:3002 | ✅ Running |
| PostgreSQL Database | localhost:5432 | ✅ Running |
