# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 13 - Idle Detection (next)
**Recent work:** Phase 12 (Metrics Retention) completed

## Current Position

Phase: 12 of 14 (Metrics Retention) - COMPLETE
Plan: 1 of 1 complete
Status: Phase complete
Last activity: 2026-02-25 — Completed 12-01-PLAN.md (Metrics Retention)

Progress: [████████████████████████░░░░░░░░] 71% (5/7 plans)

**Next Action:** Phase 13 - Idle Detection (compound threshold for CPU + connections)

## Quick Tasks Completed

| Quick Task | Date | Description |
|------------|------|-------------|
| quick-001-01 | 2026-02-24 | Test Metrics button with backend endpoint, API method, and UI button |

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
| v1.2 Metrics & Recommendations | 10-14 | 2/7 | In progress (Phase 12 complete) |

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

### v1.2 Progress

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

### Phase 13: Idle Detection (Next - not started)

### Key Decisions

| Decision | Rationale |
|----------|-----------|
| Hard-coded 7-day retention | Per CONTEXT.md requirement |
| 1000 rows per batch | Reasonable default for PostgreSQL |
| 7-minute startup delay | Within 5-10 minute range per CONTEXT.md |
| Settings key: metrics_retention_last_run | Standard key-value tracking in settings table |
| UTC timestamps for comparisons | Per RESEARCH.md pitfalls |
| Subquery pattern for batched deletes | PostgreSQL standard approach |

### Tech Debt

Minimal tech debt from savings removal:
- Orphaned `web/src/lib/formatters.ts` (~50 lines)
- Orphaned `Saving` struct in models.go (~20 lines)
- Migration files kept for history

## Blockers/Concerns

None — ready for Phase 13.

## Session Continuity

Last session: 2026-02-25
Stopped at: Completed Phase 12 (12-01), metrics retention automated
Resume file: None

## Accumulated Context

### Key Decisions

| Decision | Rationale |
|----------|-----------|
| Hard-coded 7-day retention | Per CONTEXT.md requirement |
| 1000 rows per batch | Reasonable default for PostgreSQL |
| 7-minute startup delay | Within 5-10 minute range per CONTEXT.md |
| Settings key: metrics_retention_last_run | Standard key-value tracking |
| UTC timestamps for comparisons | Per RESEARCH.md pitfalls |
| Subquery pattern for batched deletes | PostgreSQL standard approach |

### Archive References

- `.planning/milestones/v1.0-ROADMAP.md` — Full v1.0 phase details
- `.planning/milestones/v1.0-REQUIREMENTS.md` — v1.0 requirements
- `.planning/milestones/v1.1-ROADMAP.md` — Full v1.1 phase details
- `.planning/milestones/v1.1-REQUIREMENTS.md` — v1.1 requirements
- `.planning/milestones/v1.1-MILESTONE-AUDIT.md` — v1.1 audit report

---

*Last updated: 2026-02-25 - Phase 12 completed, 7-day retention automated*
