# Roadmap: SnoozeQL

## Milestones

- ✅ **v1.0 MVP** - Phases 1-6 (shipped 2026-02-23)
- 🚧 **v1.1 Enhanced Insights & Savings** - Phases 7-8 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-6) - SHIPPED 2026-02-23</summary>

### Phase 1: Multi-Cloud Discovery
**Goal**: Users can view and manage database instances across AWS and GCP accounts
**Plans**: 6 plans

Plans:
- [x] 01-01: Instance persistence with database syncing
- [x] 01-02: Multi-account provider registration
- [x] 01-03: Sortable/filterable instances table
- [x] 01-04: Connection status tracking
- [x] 01-05: Account column display
- [x] 01-06: Toast notifications

### Phase 2: Manual Control & Audit
**Goal**: Users can manually stop/start instances with full audit trail
**Plans**: 5 plans

Plans:
- [x] 02-01: EventStore implementation
- [x] 02-02: ConfirmDialog component
- [x] 02-03: Bulk stop/start API endpoints
- [x] 02-04: Audit logging
- [x] 02-05: UI integration

### Phase 3: Basic Scheduling
**Goal**: Users can create time-based schedules for automated sleep/wake
**Plans**: 3 plans

Plans:
- [x] 03-01: WeeklyScheduleGrid component
- [x] 03-02: ScheduleModal with grid and CRON mode
- [x] 03-03: SchedulesPage integration

### Phase 4: Advanced Schedule Filtering
**Goal**: Users can target schedules to instances using flexible regex patterns
**Plans**: 3 plans

Plans:
- [x] 04-01: Backend matcher logic
- [x] 04-02: FilterBuilder components
- [x] 04-03: ScheduleModal filter integration

### Phase 5: Activity Analysis
**Goal**: System detects idle periods from CloudWatch metrics
**Plans**: 3 plans

Plans:
- [x] 05-01: metrics_hourly table and aggregation
- [x] 05-02: CloudWatch client implementation
- [x] 05-03: Idle period detection algorithms

### Phase 6: Intelligent Recommendations
**Goal**: Users receive schedule suggestions based on activity patterns
**Plans**: 4 plans

Plans:
- [x] 06-01: Recommendation generation
- [x] 06-02: RecommendationCard component
- [x] 06-03: RecommendationModal with ActivityGraph
- [x] 06-04: Dashboard and RecommendationsPage integration

</details>

### 🚧 v1.1 Enhanced Insights & Savings (In Progress)

**Milestone Goal:** Extend the v1.0 foundation with cost savings tracking, usage analytics, and historical visualization to validate SnoozeQL's value and improve user confidence.

#### Phase 7: Core Savings Calculation & API
**Goal**: System calculates and exposes cost savings data from stop/start events
**Depends on**: Phase 6 (v1.0 foundation)
**Requirements**: SAV-01, SAV-02, AUD-01, AUD-02
**Success Criteria** (what must be TRUE):
  1. User can see total estimated savings in cents for any time range via API
  2. User can retrieve daily savings breakdown with stopped minutes and savings per day
  3. User can query savings attributed to specific instances
  4. All savings calculations are logged with instance ID, date, stopped minutes, and rate
  5. System captures hourly rate at stop event time (handles instance resizing)
**Plans**: 3 plans

Plans:
- [x] 07-01-PLAN.md — Database migration, SavingsStore, SavingsCalculator
- [x] 07-02-PLAN.md — EventStoreWithSavings decorator + main.go integration
- [x] 07-03-PLAN.md — SavingsHandler API endpoints

#### Phase 8: Dashboard & Visualization
**Goal**: Users can visualize savings trends, per-instance attribution, and cost projections
**Depends on**: Phase 7
**Requirements**: SAV-03, SAV-04, SAV-05
**Success Criteria** (what must be TRUE):
  1. User can view time-series chart of savings over configurable ranges (7d, 30d, 90d, custom)
  2. User can see per-instance savings table showing which instances contributed most
  3. User can compare actual costs vs projected "always-on" costs with clear disclaimers
  4. Dashboard displays summary cards with total savings and period-over-period trends
**Plans**: 4 plans

Plans:
- [ ] 08-01-PLAN.md — API types, formatters, DateRangeSelector component
- [ ] 08-02-PLAN.md — SavingsSummaryCards, SavingsChart, InstanceSavingsTable components
- [ ] 08-03-PLAN.md — CostProjection component with SAV-05 disclaimer
- [ ] 08-04-PLAN.md — SavingsPage integration, routing, navigation, visual verification

## Progress

**Execution Order:**
Phases execute in numeric order. v1.0 (1-6) complete. v1.1 (7-8) in progress.

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Multi-Cloud Discovery | v1.0 | 6/6 | Complete | 2026-02-23 |
| 2. Manual Control & Audit | v1.0 | 5/5 | Complete | 2026-02-23 |
| 3. Basic Scheduling | v1.0 | 3/3 | Complete | 2026-02-23 |
| 4. Advanced Schedule Filtering | v1.0 | 3/3 | Complete | 2026-02-23 |
| 5. Activity Analysis | v1.0 | 3/3 | Complete | 2026-02-23 |
| 6. Intelligent Recommendations | v1.0 | 4/4 | Complete | 2026-02-23 |
| 7. Core Savings Calculation & API | v1.1 | 3/3 | Complete | 2026-02-23 |
| 8. Dashboard & Visualization | v1.1 | 0/? | Not started | - |

---

*Roadmap created: 2026-02-23*
*Last updated: 2026-02-23 (Phase 7 complete)*
