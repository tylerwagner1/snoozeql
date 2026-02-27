# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Debug - Activity Pattern Graph Regression (resolved)
**Recent work:** Completed debug-001: Activity pattern graph flat line fix

## Current Position

Phase: 18 of 18 (Dual-Mode Data Collection)
Plan: 1 of 1 in current phase
Status: Phase complete
Last activity: 2026-02-27 — Completed quick-007: Recommendations table display

Progress: [████████████████████████████████████] 100% (38/38 plans through Phase 18)

**Next Action:** Quick task 007 complete - table-based recommendations display. Ready for next quick task.

## Quick Tasks Completed

| Quick Task | Date | Description |
|------------|------|-------------|
| quick-001-01 | 2026-02-24 | Test Metrics button with backend endpoint, API method, and UI button |
| quick-002-01 | 2026-02-25 | Phase 14-01: Backend pattern grouping (PatternSignature, RecommendationGroup, groupRecommendations) |
| quick-002-02 | 2026-02-25 | Phase 14-02: Frontend grouped recommendations (RecommendationGroup component, UI updates) |
| quick-003-01 | 2026-02-26 | Metrics backfill: BackfillMetrics method + POST /api/v1/instances/{id}/metrics/backfill endpoint |
| quick-004-01 | 2026-02-26 | Instance Details page cleanup: fix title color, remove Quick Stats and Metrics cards, dark mode colors |
| quick-005-01 | 2026-02-27 | Flip Sleep/Wake order to Wake/Sleep in all UI components |
| quick-006-01 | 2026-02-27 | Scheduler daemon: CRON evaluation with cronexpr, one-time execution tracking, event logging |
| quick-007-01 | 2026-02-27 | Table-based recommendations display: RecommendationsTable component with batch dismissal |

**Debug Sessions Completed**

| Session | Date | Description |
|---------|------|-------------|
| debug-001 | 2026-02-27 | Activity pattern graph flat line: Fixed edge case handling and CPU multiplier bugs |


**Performance Metrics**

**Velocity (v1.0 + v1.1 + v1.2):**
- Total plans completed: 35
- Average duration: ~15 min
- Total execution time: ~8.75 hours

**By Milestone:**

| Milestone | Phases | Plans | Status |
|-----------|--------|-------|--------|
| v1.0 MVP | 1-6 | 24 | Shipped 2026-02-23 |
| v1.1 Enhanced Insights | 7-9 | 9 | Shipped 2026-02-24 |
| v1.2 Metrics & Recommendations | 10-15 | 9/9 | COMPLETE (2026-02-25) |

## Accumulated Context

### Shipped Milestones

**v1.0 MVP (Shipped 2026-02-23):**
- Multi-cloud instance discovery (AWS RDS + GCP Cloud SQL)
- Manual sleep/wake with confirmation and audit logging
- Time-based scheduling with visual grid and CRON mode
- Regex-based schedule filtering
- Activity analysis with CloudWatch metrics
- Intelligent schedule recommendations

**v1.1 Enhanced Insights & Savings (Shipped 2026-02-24):**
- Cost savings tracking system built and evaluated
- Savings dashboard with charts and tables
- Instance metrics display added
- Feature removed per product direction change
- Tech debt: formatters.ts deleted, Saving struct removed (Phase 15)

### v1.2 Metrics & Recommendations (COMPLETE 2026-02-25)

**Phase 15: UI Polish & Cleanup (Complete 2026-02-25):**
- Phase 15-01: Navigation active states with useLocation, formatters.ts deleted, Saving struct removed

**Phase 14, Plan 01 (Complete - 2026-02-25):**
- Added PatternSignature and RecommendationGroup types
- Implemented groupRecommendations() function with pattern signatures
- Modified GetAllRecommendations endpoint to return { groups: [...] }
- Backend response returns recommendations organized by similar idle patterns
- Group totals and per-instance savings both displayed
- Requirements REC-02, REC-03 satisfied
- Files modified: `internal/api/handlers/recommendations.go`

**Phase 14, Plan 02 (Complete - 2026-02-25):**
- Added RecommendationGroup and GroupedRecommendationsResponse API types
- Created `RecommendationGroup.tsx` component with expand/collapse
- Updated `RecommendationsPage.tsx` to render groups
- Updated `Dashboard.tsx` to handle grouped response
- All builds pass (Go and TypeScript)

**Phase 14, Plan 01 (Complete - 2026-02-25):**
- Added PatternSignature and RecommendationGroup types
- Implemented groupRecommendations() function with pattern signatures
- Modified GetAllRecommendations endpoint to return { groups: [...] }
- Backend response returns recommendations organized by similar idle patterns
- Group totals and per-instance savings both displayed
- Requirements REC-02, REC-03 satisfied
- Files modified: `internal/api/handlers/recommendations.go`

**Phase 10: Metrics Collection Enhancement (Complete):**
- Added FreeableMemory metric to CloudWatch collector
- Memory stored as percentage using instance class mapping (~20 classes)
- Stopped instances get zero metrics (shows "asleep" state)
- "Metrics unavailable" badge on Instance Details page
- Requirements METR-01, METR-02 satisfied

**Quick 001-01: Test Metrics Button (Complete):**
- Added POST /instances/:id/collect-metrics backend endpoint
- Added collectInstanceMetrics() API method
- Added Test Metrics button to Instance Details page
- Button shows loading state, disabled for non-AWS instances
- Public CollectInstance() method exposed on MetricsCollector

**Phase 11: Time-Series Visualization (Complete):**
- Added GET /instances/{id}/metrics/history API endpoint
- Created MetricsChart component with tabs for CPU, Memory, Connections
- Time range selector (1h, 6h, 24h, 7d) with dynamic axis formatting
- Integrated into InstanceDetailPage with "Metrics History" section
- Requirements VIS-01, VIS-02, VIS-03, VIS-04 satisfied

**Phase 12: Metrics Retention (Complete):**
- Created RetentionCleaner service with RunContinuous pattern (149 lines)
- 7-day retention enforced via batched deletes (1000 rows/batch)
- 100ms pauses between batches to prevent table locking
- Last-run timestamp tracked in settings table for skip-if-recent logic
- 7-minute startup delay followed by 24-hour interval cleanup
- Requirements METR-03 satisfied

### Phase 13: Idle Detection (Complete)

**Phase 13: Idle Detection (Complete):**
- Added ConnectionsThreshold field to ActivityThresholds struct
- Updated DefaultThresholds() with CPUPercent: 5.0, ConnectionsThreshold: 0
- Modified findIdleSegments() to check CPU < 5% AND connections == 0
- REC-01 requirement satisfied - compound threshold prevents false positives
- Requirements REC-01 satisfied

**Decisions:**

| Decision | Rationale |
|----------|-----------|
| CPUPercent: 5.0 | Requirement REC-01 specifies CPU < 5% for idle flag |
| ConnectionsThreshold: 0 | Requirement REC-01 specifies connections = 0 for idle flag |
| Use <= for connections check | Ensures negative connection edge case handled properly |

### Phase 14: Grouped Recommendations (Complete)

**Key Decisions:**
| Decision | Rationale |
|----------|-----------|
| PatternSignature struct | Efficient O(n) grouping instead of O(n²) pairwise comparisons |
| 5 time buckets | Balance between granularity and group size; covers typical patterns |
| 80% threshold for day type | Allows "mostly weekdays" patterns to group with weekdays |
| Sort groups by total savings | High-impact patterns shown first for better UX |
| Keep per-instance savings visible | Users need to see individual savings within groups |
| Single-instance groups as regular cards | Avoids awkward "Group: 1 instance" UI overhead |
| Always start expanded | Ensures visibility of all patterns on first load |
| Show first 3 recommendations in Dashboard | Preview without overwhelming dashboard |

**Phase 13: Idle Detection (Complete):**
- Added ConnectionsThreshold field to ActivityThresholds struct
- Updated DefaultThresholds() with CPUPercent: 5.0, ConnectionsThreshold: 0
- Modified findIdleSegments() to check CPU < 5% AND connections == 0
- REC-01 requirement satisfied - compound threshold prevents false positives
- Requirements REC-01 satisfied

**Decisions:**

| Decision | Rationale |
|----------|-----------|
| CPUPercent: 5.0 | Requirement REC-01 specifies CPU < 5% for idle flag |
| ConnectionsThreshold: 0 | Requirement REC-01 specifies connections = 0 for idle flag |
| Use <= for connections check | Ensures negative connection edge case handled properly |

### Tech Debt

Tech debt from savings removal fully cleaned up in Phase 15-01:
- `web/src/lib/formatters.ts` deleted
- `Saving` struct removed from models.go

Migration files kept for history.

### Phase 17: Enhanced Metrics & Data Collection Strategy (COMPLETE 2026-02-26)

**Phase 17-01 (Complete - 5-minute CloudWatch collection):**
- Added MetricPeriod constant (5 * time.Minute)
- Implemented GetRDSMetricsMultiple method for multi-datapoint fetch
- Updated collector to use 5-minute granularity with 3 datapoints per cycle
- Zero entries for stopped instances (3 per 15-minute window)

**Phase 17-02 (Complete - Metrics backfill - superseding to Phase 18-01):**
- Superseded by Phase 18-01 for continuous backfill
- Original synchronous gap detection replaced with background goroutine
- See Phase 18-01 for new 3-day window + hourly backfill implementation

### Phase 18: Dual-Mode Data Collection (COMPLETE 2026-02-26)

**Phase 18-01 (Complete - Dual-mode collection architecture):**
- Added RunHistoricalBackfill method with 7-minute startup delay + 1-hour interval
- Created runHistoricalBackfill private method with 3-day CloudWatch window
- Updated server startup to run historical backfill as background goroutine
- Removed synchronous gap detection that blocked server startup
- Real-time collection (15-min) continues unchanged
- Server startup no longer blocks on gap detection

### Roadmap Evolution

- Phase 17 added: Enhanced Metrics & Data Collection Strategy (5-min CloudWatch intervals, 3 datapoints/collection, interpolated gap backfill)
- Phase 17-01 complete (2026-02-26): 5-minute CloudWatch collection with GetRDSMetricsMultiple, MetricPeriod constant, 3 datapoints per 15-min cycle
- Phase 17-02 superseded by Phase 18-01: Continuous background backfill instead of synchronous startup call

### Phase 18: Dual-Mode Data Collection

- Background historical backfill with 7-min startup delay + hourly interval
- 3-day CloudWatch window (not 7) for faster processing and reduced API cost
- Non-blocking server startup with dual background goroutines
- Real-time collection (15-min) continues unchanged

## Blockers/Concerns

Phase 18 complete. No blockers carried forward.

### Accumulated Decisions

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 17-01 | Store timestamps pre-truncated in Go, SQL as-is | Maintain backward compatibility with existing UpsertHourlyMetric |
| 17-01 | Store 3 zero entries for stopped instances (one per 5-min interval) | Match 15-minute collection window |
| 17-01 | Keep existing methods unchanged for BackfillMetrics backward compatibility | Preserve existing callers |
| 17-02 | Call CloudWatch for up to 7 days of historical data on startup | New approach for gap detection |
| 17-02 | Skip existing rows automatically via ON CONFLICT | Avoid duplicate entries |
| 17-02 | Batch query GetLatestMetricTimes for efficient instance processing | Avoid N queries for N instances |

## Decisions

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 15-01 | Navigation active states use bg-blue-500/30 text-blue-400 for most links, bg-purple-500/30 text-purple-400 for Accounts | Visual distinction between navigation items |
| 15-01 | Active path matching: exact for /, prefix for others | Simple and intuitive matching strategy |
| 15-01 | formatters.ts and Saving struct removal | Orphaned code from Phase 9 savings feature removal |
| 17-01 | MetricPeriod = 5 * time.Minute with TruncateToMetricPeriod helper | Consistent timestamp truncation for 5-minute granularity |
| 17-01 | Truncate timestamps in Go, SQL as-is | Maintain backward compatibility with existing UpsertHourlyMetric |
| 17-01 | storeZeroMetrics generates 3 entries (one per 5-min interval) | Match 15-minute collection window for stopped instances |
| 17-02 | Call CloudWatch for up to 7 days of historical data on startup | New approach for gap detection |
| 17-02 | Skip existing rows automatically via ON CONFLICT | Avoid duplicate entries |
| 17-02 | Batch query GetLatestMetricTimes for efficient instance processing | Avoid N queries for N instances |

## Accumulated Decisions

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 18-01 | 7-minute startup delay for historical backfill | Consistent with RetentionCleaner pattern |
| 18-01 | 3-day CloudWatch window (not 7) | Faster processing, lower API cost, continuous hourly healing |
| 18-01 | Background goroutine for historical backfill | Non-blocking startup, real-time collection continues immediately |

## Decisions

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 15-01 | Navigation active states use bg-blue-500/30 text-blue-400 for most links, bg-purple-500/30 text-purple-400 for Accounts | Visual distinction between navigation items |
| 15-01 | Active path matching: exact for /, prefix for others | Simple and intuitive matching strategy |
| 15-01 | formatters.ts and Saving struct removal | Orphaned code from Phase 9 savings feature removal |
| 17-01 | MetricPeriod = 5 * time.Minute with TruncateToMetricPeriod helper | Consistent timestamp truncation for 5-minute granularity |
| 17-01 | Truncate timestamps in Go, SQL as-is | Maintain backward compatibility with existing UpsertHourlyMetric |
| 17-01 | storeZeroMetrics generates 3 entries (one per 5-min interval) | Match 15-minute collection window for stopped instances |
| 17-02 | Call CloudWatch for up to 7 days of historical data on startup | New approach for gap detection |
| 17-02 | Skip existing rows automatically via ON CONFLICT | Avoid duplicate entries |
| 17-02 | Batch query GetLatestMetricTimes for efficient instance processing | Avoid N queries for N instances |

## Session Continuity

Last session: 2026-02-27
Stopped at: Completed quick-007: Recommendations table display with batch dismissal
Resume file: None

---

*Last updated: 2026-02-27 - Recommendations table: Pattern rows with batch dismissal
