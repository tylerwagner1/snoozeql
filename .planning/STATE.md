# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.
**Current focus:** Phase 11 - Time-Series Visualization
**Recent work:** Phase 11 Plan 01 (Metrics History API) completed

## Current Position

Phase: 11 of 14 (Time-Series Visualization)
Plan: 1 of 2 in current phase
Status: In progress
Last activity: 2026-02-25 — Completed 11-01-PLAN.md (Metrics History API)

Progress: [███████░░░░░░░░░░░░░░░░░░░░░░░░░░░] 21% (2/7 plans)

**Next Action:** Ready for 11-02-PLAN.md (MetricsChart Component)

## Quick Tasks Completed

| Quick Task | Date | Description |
|------------|------|-------------|
| quick-001-01 | 2026-02-24 | Test Metrics button with backend endpoint, API method, and UI button |

## Performance Metrics

**Velocity (v1.0 + v1.1 + v1.2):**
- Total plans completed: 34
- Average duration: ~15 min
- Total execution time: ~8.5 hours

**By Milestone:**

| Milestone | Phases | Plans | Status |
|-----------|--------|-------|--------|
| v1.0 MVP | 1-6 | 24 | Shipped 2026-02-23 |
| v1.1 Enhanced Insights | 7-9 | 9 | Shipped 2026-02-24 |
| v1.2 Metrics & Recommendations | 10-14 | 1/7 | In progress |

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

**Phase 11-01: Metrics History API (Complete):**
- Added GET /instances/{id}/metrics/history endpoint
- Supports time ranges: 1h, 6h, 24h, 7d (default: 24h)
- Added frontend getMetricsHistory() API method
- Returns empty array on error (consistent with /metrics)

### Key Decisions

All key decisions documented in PROJECT.md.

### Tech Debt

Minimal tech debt from savings removal:
- Orphaned `web/src/lib/formatters.ts` (~50 lines)
- Orphaned `Saving` struct in models.go (~20 lines)
- Migration files kept for history

## Blockers/Concerns

None — ready for next plan (11-02: MetricsChart component).

## Session Continuity

Last session: 2026-02-25
Stopped at: Completed 11-01-PLAN.md (Metrics History API)
Resume file: None

### Archive References

- `.planning/milestones/v1.0-ROADMAP.md` — Full v1.0 phase details
- `.planning/milestones/v1.0-REQUIREMENTS.md` — v1.0 requirements
- `.planning/milestones/v1.1-ROADMAP.md` — Full v1.1 phase details
- `.planning/milestones/v1.1-REQUIREMENTS.md` — v1.1 requirements
- `.planning/milestones/v1.1-MILESTONE-AUDIT.md` — v1.1 audit report

---

*Last updated: 2026-02-25 - Phase 11 Plan 01 completed*
