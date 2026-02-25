# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 14 - Grouped Recommendations (complete)
**Recent work:** Phase 12 (Metrics Retention) completed

## Current Position

Phase: 14 of 14 (Grouped Recommendations) - COMPLETE
Plan: 2 of 2 complete
Status: v1.2 COMPLETE
Last activity: 2026-02-25 — Phase 14 complete, v1.2 milestone complete
Next Phase: 15 (UI Polish & Cleanup) - Not planned (added 2026-02-25)

Progress: [████████████████████████████████████] 100% (8/8 plans)

**Next Action:** Phase 15 (if any) or v1.2 release

## Quick Tasks Completed

| Quick Task | Date | Description |
|------------|------|-------------|
| quick-001-01 | 2026-02-24 | Test Metrics button with backend endpoint, API method, and UI button |
| quick-002-01 | 2026-02-25 | Phase 14-01: Backend pattern grouping (PatternSignature, RecommendationGroup, groupRecommendations) |
| quick-002-02 | 2026-02-25 | Phase 14-02: Frontend grouped recommendations (RecommendationGroup component, UI updates) |

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
| v1.2 Metrics & Recommendations | 10-14 | 7/7 | COMPLETE (2026-02-25) |

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

### v1.2 Metrics & Recommendations (COMPLETE 2026-02-25)

**Phase 15: UI Polish & Cleanup (Added 2026-02-25):**
- Fix visual issues, improve styling, ensure consistent UI/UX

**Phase 14: Grouped Recommendations (Complete):**

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

Minimal tech debt from savings removal:
- Orphaned `web/src/lib/formatters.ts` (~50 lines)
- Orphaned `Saving` struct in models.go (~20 lines)
- Migration files kept for history

## Blockers/Concerns

None — v1.2 complete.

## Session Continuity

Last session: 2026-02-25
Stopped at: v1.2 Metrics & Recommendations complete (Phases 10-14)
Resume file: None

---
*Last updated: 2026-02-25 - v1.2 Metrics & Recommendations COMPLETE (Phases 10-14), Phase 15 added*
