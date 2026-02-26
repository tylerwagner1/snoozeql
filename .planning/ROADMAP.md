# Roadmap: SnoozeQL

## Milestones

- ✅ **v1.0 MVP** - Phases 1-6 (shipped 2026-02-23)
- ✅ **v1.1 Enhanced Insights** - Phases 7-9 (shipped 2026-02-24)
- ✅ **v1.2 Metrics & Recommendations** - Phases 10-15 (complete 2026-02-25)

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

### ✅ v1.2 Metrics & Recommendations - COMPLETE

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

#### Phase 12: Metrics Retention ✓ COMPLETE
**Goal**: Metrics data stays manageable with automatic cleanup
**Depends on**: Phase 10 (needs data to clean up)
**Requirements**: METR-03
**Success Criteria** (what must be TRUE):
  1. Metrics older than 7 days are automatically deleted ✓
  2. Cleanup runs without affecting application performance ✓
**Plans**: 1 plan

Plans:
- [x] 12-01-PLAN.md — Create RetentionCleaner service with batched deletes and last-run tracking

**Delivered:** RetentionCleaner service with batched deletes (1000 rows/batch), 7-day retention, 7-min startup delay, 24h interval, settings-based last-run tracking.

#### Phase 13: Idle Detection ✓ COMPLETE
**Goal**: Idle detection accurately identifies truly inactive instances
**Depends on**: Phase 10 (needs Memory metric for complete picture)
**Requirements**: REC-01
**Success Criteria** (what must be TRUE):
  1. Instance only flagged idle when CPU < 5% AND connections = 0 ✓
  2. Instances with active connections never flagged as idle ✓
  3. Recommendations use compound threshold for pattern detection ✓
**Plans**: 1 plan

Plans:
- [x] 13-01-PLAN.md — Update ActivityThresholds and findIdleSegments with compound threshold

**Delivered:** Compound idle threshold: CPU < 5% AND connections = 0 via ConnectionsThreshold field, DefaultThresholds() with CPUPercent: 5.0, findIdleSegments() compound check.

#### Phase 14: Grouped Recommendations
**Goal**: Recommendations show patterns and savings clearly
**Depends on**: Phase 13 (needs improved idle detection)
**Requirements**: REC-02, REC-03
**Success Criteria** (what must be TRUE):
  1. Similar idle patterns are grouped in recommendations view
  2. Each recommendation shows estimated daily savings in dollars
  3. User can see which instances share similar idle patterns
**Plans**: 2 plans

**Delivered:** Pattern signature generation, RecommendationGroup struct, groupRecommendations() function, modified API endpoint to return `{ groups: [...] }`, RecommendationGroup React component with expand/collapse, grouped display in RecommendationsPage and Dashboard.

Plans:
- [x] 14-01-PLAN.md — Add pattern signature generation and grouping logic to recommendations API
- [x] 14-02-PLAN.md — Update recommendations UI with grouped display and expand/collapse

#### Phase 15: UI Polish & Cleanup
**Goal**: Fix visual issues, improve styling, and ensure consistent UI/UX across the application
**Depends on**: Phase 14
**Plans**: 1 plan

**Delivered:** Navigation active states using useLocation hook, formatters.ts deleted, Saving struct removed from models.go.

Plans:
- [x] 15-01-PLAN.md — Add navigation active states and remove orphaned code

#### Phase 16: UI Changes Per Tab with Server Restart and Verification
**Goal**: Address UI inconsistencies per tab and verify changes via server restart
**Depends on**: Phase 15
**Plans**: 0 plans

Plans:
- [ ] TBD (run /gsd-plan-phase 16 to break down)

**Details:**
Tab-specific UI changes needed per user request. Plan will require:
- Tab-by-tab analysis of inconsistencies
- Fixes for visual issues in each tab
- Server restart and frontend verification

#### Phase 17: Enhanced Metrics & Data Collection Strategy ✓ COMPLETE
**Goal**: CloudWatch scraped at 5-min intervals, 3 datapoints per 15-min collection, with intelligent gap detection and interpolated backfill
**Depends on**: Phase 16
**Plans**: 2 plans

Plans:
- [x] 17-01-PLAN.md — High-resolution CloudWatch collection (Period=300, 3 datapoints/cycle)
- [x] 17-02-PLAN.md — Gap detection and interpolation on startup

**Delivered:** 5-minute CloudWatch collection with GetRDSMetricsMultiple method, MetricPeriod constant, 3 datapoints per 15-min cycle, DetectAndFillGaps on startup with 7-day CloudWatch historical fetch.

#### Phase 18: Dual-Mode Data Collection ✓ COMPLETE
**Goal**: Reliable metrics data via real-time collection + hourly historical backfill
**Depends on**: Phase 17
**Requirements**: DATA-01, DATA-02
**Success Criteria** (what must be TRUE):
  1. Real-time collection continues every 15 minutes (unchanged) ✓
  2. Historical backfill runs on startup (7 min delay) and then hourly ✓
  3. Historical backfill fetches 3-day CloudWatch window and upserts (gaps filled automatically) ✓
  4. Phase 17-02's startup-only gap detection is removed/replaced ✓
**Plans**: 1 plan

Plans:
- [x] 18-01-PLAN.md — Add RunHistoricalBackfill method and update main.go startup

**Delivered:** Dual-mode collection with RunHistoricalBackfill method (7-min delay + hourly interval), runHistoricalBackfill private method (3-day window), non-blocking server startup with background goroutines.

## Progress

**Execution Order:**
Phases execute in numeric order: 10 → 11 → 12 → 13 → 14

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 10. Metrics Collection Enhancement | 1/1 | Complete | 2026-02-24 |
| 11. Time-Series Visualization | 2/2 | Complete | 2026-02-25 |
| 12. Metrics Retention | 1/1 | Complete | 2026-02-25 |
| 13. Idle Detection | 1/1 | Complete | 2026-02-25 |
| 14. Grouped Recommendations | 2/2 | Complete | 2026-02-25 |
| 15. UI Polish & Cleanup | 1/1 | Complete | 2026-02-25 |
| 16. UI Changes Per Tab | 0/0 | Skipped | - |
| 17. Enhanced Metrics & Data Collection | 2/2 | Complete | 2026-02-26 |
| 18. Dual-Mode Data Collection | 1/1 | Complete | 2026-02-26 |

---

*Roadmap created: 2026-02-24*
*Last updated: 2026-02-26 - Phase 18 complete (dual-mode data collection with RunHistoricalBackfill)*
