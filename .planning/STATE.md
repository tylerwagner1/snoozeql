# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-23)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** v1.1 - Enhanced Insights & Savings (roadmap created, ready for phase planning)

## Current Position

Phase: 7 of 8 (Core Savings Calculation & API)
Plan: 3
Status: Complete
Last activity: 2026-02-23 - Completed Phase 7 (all plans)

Progress: [█████████████████████████████████░] 7/8 phases complete (v1.0 done, v1.1 3/3 plans complete)

## Performance Metrics

**Velocity:**
- Total plans completed: 25
- Average duration: ~15 min
- Total execution time: ~6.2 hours

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

**Recent Trend:**
- Last 24 plans: 24 complete
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

### Decisions Made

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 07-02 | Implemented EventStoreWithSavings decorator pattern | Automatic savings calculation on event creation instead of dashboard load time |
| 07-02 | Added EventCreator interface to DiscoveryService | Flexible event store types (allows decorator wrapping) |
| 07-03 | Implemented full SavingsHandler with 4 endpoints | Required for Phase 8 dashboard API consumption |

**Phase 7 - Core Savings Calculation**

| Decision | Rationale |
|----------|-----------|
| EventStoreWithSavings decorator pattern | Automatic savings calculation on event creation instead of dashboard load time (push model) |
| EventCreator interface | Flexible event store types - allows decorator wrapping |
| Full SavingsHandler with 4 endpoints | Required for Phase 8 dashboard API consumption |

## Blockers/Concerns Carried Forward

None - savings backend is complete and ready for Phase 8 integration.

## Session Continuity

Last session: 2026-02-23T18:30:00Z
Stopped at: Completed Phase 7 (all 3 plans)
Resume file: None

---

*Project: SnoozeQL v1.1 (Enhanced Insights & Savings)*
*Started: 2026-02-23*
