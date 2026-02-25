# Roadmap: SnoozeQL

## Milestones

- ✅ **v1.0 MVP** - Phases 1-6 (shipped 2026-02-23)
- ✅ **v1.1 Enhanced Insights** - Phases 7-9 (shipped 2026-02-24)
- 🚧 **v1.2 Metrics & Recommendations** - Phases 10-14 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-6) - SHIPPED 2026-02-23</summary>

See: `.planning/milestones/v1.0-ROADMAP.md` for full details.

Delivered: Multi-cloud discovery, manual sleep/wake, scheduling, activity analysis, recommendations.

</details>

<details>
<summary>✅ v1.1 Enhanced Insights (Phases 7-9) - SHIPPED 2026-02-24</summary>

See: `.planning/milestones/v1.1-ROADMAP.md` for full details.

Delivered: Savings tracking system (built, evaluated, removed), instance metrics display.

</details>

### 🚧 v1.2 Metrics & Recommendations (In Progress)

**Milestone Goal:** Metrics working flawlessly with time-series visualization, powering intelligent schedule recommendations.

#### Phase 10: Metrics Collection Enhancement
**Goal**: System reliably collects and persists CPU, Memory, and Connections metrics
**Depends on**: Nothing (foundation for v1.2)
**Requirements**: METR-01, METR-02
**Success Criteria** (what must be TRUE):
  1. FreeableMemory metric appears alongside CPU and Connections in metrics_hourly table
  2. Metrics collection runs every 15 minutes without errors
  3. New metrics data persists across application restarts
**Plans**: 1 plan

Plans:
- [x] 10-01-PLAN.md — Add FreeableMemory to CloudWatch collector with memory percentage and UI badge

#### Phase 11: Time-Series Visualization
**Goal**: User can view metrics history on Instance Details page
**Depends on**: Phase 10
**Requirements**: VIS-01, VIS-02, VIS-03, VIS-04
**Success Criteria** (what must be TRUE):
  1. User sees CPU, Memory, and Connections charts on Instance Details page
  2. User can switch between 1h, 6h, 24h, and 7d time ranges
  3. All three metrics are visible together on the page
  4. Charts show loading spinner while fetching and "No data" when empty
**Plans**: 2 plans

Plans:
- [x] 11-01-PLAN.md — Add metrics history API endpoint with time range support
- [x] 11-02-PLAN.md — Build MetricsChart component with tabs and time range selector

#### Phase 12: Metrics Retention
**Goal**: Metrics data stays manageable with automatic cleanup
**Depends on**: Phase 10 (needs data to clean up)
**Requirements**: METR-03
**Success Criteria** (what must be TRUE):
  1. Metrics older than 7 days are automatically deleted
  2. Cleanup runs without affecting application performance
**Plans**: 1 plan

Plans:
- [ ] 12-01-PLAN.md — Create RetentionCleaner service with batched deletes and last-run tracking

#### Phase 13: Idle Detection
**Goal**: Idle detection accurately identifies truly inactive instances
**Depends on**: Phase 10 (needs Memory metric for complete picture)
**Requirements**: REC-01
**Success Criteria** (what must be TRUE):
  1. Instance only flagged idle when CPU < 5% AND connections = 0
  2. Instances with active connections never flagged as idle
  3. Recommendations use compound threshold for pattern detection
**Plans**: TBD

Plans:
- [ ] 13-01: Update analyzer with compound idle threshold

#### Phase 14: Grouped Recommendations
**Goal**: Recommendations show patterns and savings clearly
**Depends on**: Phase 13 (needs improved idle detection)
**Requirements**: REC-02, REC-03
**Success Criteria** (what must be TRUE):
  1. Similar idle patterns are grouped in recommendations view
  2. Each recommendation shows estimated daily savings in dollars
  3. User can see which instances share similar idle patterns
**Plans**: TBD

Plans:
- [ ] 14-01: Add grouping and savings to recommendations API
- [ ] 14-02: Update recommendations UI with groups and savings

## Progress

**Execution Order:**
Phases execute in numeric order: 10 → 11 → 12 → 13 → 14

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 10. Metrics Collection Enhancement | 1/1 | Complete | 2026-02-24 |
| 11. Time-Series Visualization | 2/2 | Complete | 2026-02-25 |
| 12. Metrics Retention | 0/1 | Not started | - |
| 13. Idle Detection | 0/1 | Not started | - |
| 14. Grouped Recommendations | 0/2 | Not started | - |

---
*Roadmap created: 2026-02-24*
*Last updated: 2026-02-25 - Phase 11 completed*
