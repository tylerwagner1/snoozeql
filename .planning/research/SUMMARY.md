# Project Research Summary

**Project:** SnoozeQL v1.1 - Cost Savings Tracking
**Domain:** Cloud database cost optimization / Infrastructure cost management
**Researched:** 2026-02-23
**Confidence:** HIGH

## Executive Summary

SnoozeQL v1.0 already contains the foundational infrastructure for cost savings tracking: the `Event` model captures all stop/start actions with timestamps, instances store `HourlyCostCents`, and the `Saving` model with `StoppedMinutes` and `EstimatedSavingsCents` is ready to use. The v1.1 cost tracking feature is primarily about **calculating and visualizing** data the system already captures. No new external libraries are required—the existing Go backend handles savings calculation with simple arithmetic, and Recharts (already in use) supports all needed visualizations.

The recommended approach is a **hybrid push/pull model**: calculate savings when events occur (push) using a decorator pattern that wraps the existing EventStore, then aggregate views on-demand (pull) with a materialized view for dashboard performance. The architecture minimizes coupling by using the decorator pattern—savings calculation hooks into the existing event flow without modifying core event storage logic. This allows existing code to remain unchanged while adding new functionality.

The primary risks are **AWS 7-day auto-restart** (instances automatically restart after 7 days stopped, causing overcounted savings if ignored) and **race conditions in event processing** (duplicate or out-of-order events causing double-counted savings). Both are mitigated by state-based calculation from periodic instance status polls rather than relying solely on event timestamps. POC pricing accuracy is acceptable with disclaimers—defer billing API integration to future versions.

## Key Findings

### Recommended Stack

No new dependencies required for POC. The existing Go/React/PostgreSQL stack is sufficient.

**Core technologies:**
- **Go stdlib `time`**: Duration calculations for stop periods — already available
- **pgx/v5**: PostgreSQL aggregation queries with materialized views — already in use
- **Recharts 2.10.0**: Time-series charts (LineChart, BarChart, ComposedChart) — already integrated
- **Integer cents pattern**: Store all currency as `int` cents, not `float64` dollars — already the SnoozeQL pattern

**Explicitly NOT needed for POC:**
- AWS Pricing API / GCP Cloud Billing API — adds complexity for marginal accuracy improvement
- shopspring/decimal — integer cents is sufficient
- Infracost / OpenCost — designed for IaC and K8s, not managed database stop/start

### Expected Features

**Must have (table stakes):**
- **Total savings display** (SAV-02) — summary card showing money saved
- **Per-instance savings table** (SAV-04) — attribution showing which instances contributed
- **Historical time-series chart** (SAV-03) — visualize savings trend over time
- **Actual vs projected comparison** (SAV-02, SAV-05) — show what SnoozeQL saved vs running 24/7
- **Time range selection** (SAV-03) — 7d, 30d, 90d, custom

**Should have (differentiators):**
- **Savings trend indicator** — % change vs previous period
- **Month-to-date rollup** — standard financial reporting view
- **CSV export** — external analysis capability
- **Per-schedule savings** — attribute savings to schedules, not just instances

**Defer (v2+):**
- Real billing API integration (AWS Cost Explorer, GCP Cloud Billing)
- PDF report generation
- Email notifications for savings milestones
- ML-based forecasting
- Multi-currency support

### Architecture Approach

Use an **Event-Driven Savings Calculator** with a hybrid push/pull model. The calculator hooks into existing event flow via the **decorator pattern**: `EventStoreWithSavings` wraps the existing `EventStore`, triggering savings calculation when start/wake events occur (finalizing the previous stop period). Dashboard queries use a **materialized view** (`savings_summary`) refreshed every 15 minutes for performance. Currently-stopped instances calculate ongoing savings in real-time using `time.Now() - last_stop_event`.

**Major components:**
1. **SavingsCalculator** (`internal/savings/calculator.go`) — core calculation logic: duration × hourly_cost_cents
2. **EventStoreWithSavings** (`internal/savings/event_decorator.go`) — decorator that triggers savings on start events
3. **SavingsStore** (`internal/store/savings_store.go`) — PostgreSQL storage with upsert for daily aggregations
4. **SavingsHandler** (`internal/api/handlers/savings.go`) — API endpoints for summary, history, per-instance
5. **Materialized View** (`savings_summary`) — pre-aggregated daily/weekly/monthly rollups

**Data Model (Minimal Changes):**
- Use existing `savings` table (already defined with `instance_id`, `date`, `stopped_minutes`, `estimated_savings_cents`)
- Add `savings_summary` materialized view for dashboard performance
- Add indexes: `idx_events_instance_time`, `idx_events_type_time`
- Store `hourly_rate_cents` in event metadata at stop time (handles instance type changes)

### Critical Pitfalls

1. **AWS 7-day auto-restart ignored** — AWS RDS automatically restarts stopped instances after 7 consecutive days. Cap `actualStoppedDuration` at 7 days maximum. Detection: calculated savings vs actual billing diverge after 7+ days.

2. **Race conditions in event processing** — Stop/start events can arrive out of order, duplicate, or fail. Use **state-based calculation** from periodic instance status polls, not just event timestamps. Deduplicate consecutive same-state observations.

3. **Calculating on every dashboard load** — Query all events and compute durations on each page load becomes O(n*m) as events grow. Use pre-aggregated `savings` table; calculate once on event, read many times.

4. **Ignoring currently-stopped instances** — Only calculating savings on start event shows $0 for instances stopped for days. Calculate **ongoing savings** separately: `time.Since(last_stop_event) × hourly_cost_cents`.

5. **Floating-point for money** — Rounding errors compound over thousands of operations. Already solved: SnoozeQL uses integer cents. **Maintain this pattern.**

## Implementation Approach

### Data Model Changes

**New migration: `006_cost_tracking.sql`**

```sql
-- Materialized view for dashboard performance
CREATE MATERIALIZED VIEW savings_summary AS
SELECT 
    instance_id,
    DATE_TRUNC('day', date) as day,
    DATE_TRUNC('week', date) as week,
    DATE_TRUNC('month', date) as month,
    SUM(stopped_minutes) as total_stopped_minutes,
    SUM(estimated_savings_cents) as total_savings_cents
FROM savings
GROUP BY instance_id, DATE_TRUNC('day', date), DATE_TRUNC('week', date), DATE_TRUNC('month', date);

CREATE UNIQUE INDEX idx_savings_summary_instance_day ON savings_summary(instance_id, day);

-- Index for time-range queries on events
CREATE INDEX idx_events_instance_time ON events(instance_id, created_at DESC);
CREATE INDEX idx_events_type_time ON events(event_type, created_at);
```

### Service Layer

**New files:**
```
internal/savings/
├── calculator.go      # Core calculation: duration × hourly_cost_cents
├── event_decorator.go # EventStoreWithSavings decorator
├── backfill.go        # One-time historical calculation
└── service.go         # Business logic orchestration

internal/store/
└── savings_store.go   # PostgreSQL CRUD for savings table
```

**Key interfaces:**
```go
type SavingsStore interface {
    UpsertDailySaving(ctx context.Context, saving *models.Saving) error
    GetSavingsByInstance(ctx context.Context, instanceID string, start, end time.Time) ([]models.Saving, error)
    GetTotalSavings(ctx context.Context, start, end time.Time) (int, error)
}
```

### API Endpoints

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `GET /api/v1/savings` | Overall summary | `{total_savings_cents, period, top_savers[], ongoing_savings}` |
| `GET /api/v1/savings/daily?days=30` | Daily breakdown | `{daily_savings: [{date, savings_cents, stopped_minutes}]}` |
| `GET /api/v1/savings/by-instance` | Per-instance attribution | `[{instance_id, name, savings_cents, stopped_hours}]` |
| `GET /api/v1/instances/{id}/savings` | Single instance detail | `{instance_id, total_savings_cents, events[]}` |
| `POST /api/v1/savings/backfill` | Trigger historical calc | `{status: "backfill_started"}` |

### Frontend Components

**New components (using existing Recharts patterns):**
```
web/src/components/
├── SavingsCard.tsx           # Summary stat card (like existing dashboard cards)
├── SavingsLineChart.tsx      # Time series trend using <LineChart>
├── SavingsBreakdown.tsx      # Per-instance attribution table
└── ProjectionChart.tsx       # Actual vs projected using <ComposedChart>
```

**Dashboard.tsx integration:**
- Replace mock `totalSavings` calculation with real API data from `/api/v1/savings`
- Add time range selector (7d/30d/90d/custom)
- Add per-instance savings table below chart

### Integration Points

| v1.0 Component | Integration | Changes |
|----------------|-------------|---------|
| `EventStore` | Wrap with `EventStoreWithSavings` decorator | None to existing code |
| `DiscoveryService` | Use decorated EventStore in constructor | One-line change in main.go |
| `Dashboard.tsx` | Fetch from new `/api/v1/savings` endpoint | Replace mock data |
| PostgreSQL | Add migration 006, materialized view | Additive only |

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Core Savings Calculation (SAV-01)
**Rationale:** Foundation that all other features depend on. Must work correctly before visualization.
**Delivers:** SavingsCalculator service, savings calculation on events, daily aggregation
**Addresses:** SAV-01 (calculate savings from stop/start events)
**Avoids:** Pitfall #1 (7-day restart), Pitfall #2 (race conditions), Pitfall #5 (floating-point)
**Estimated time:** 1-1.5 days

### Phase 2: Savings API Layer (SAV-02 partial)
**Rationale:** Backend must expose data before frontend can consume it
**Delivers:** REST endpoints (`/api/v1/savings/*`), SavingsStore, SavingsHandler
**Uses:** Existing handler patterns, Chi router
**Implements:** SavingsStore, SavingsHandler
**Estimated time:** 0.5-1 day

### Phase 3: Dashboard Integration (SAV-02, SAV-03)
**Rationale:** Users see value through the UI—this is the primary deliverable
**Delivers:** Summary cards, time-series chart, time range selector
**Addresses:** SAV-02 (savings dashboard), SAV-03 (historical charts)
**Avoids:** Pitfall #9 (too much data), Pitfall #10 (too little context)
**Estimated time:** 1-1.5 days

### Phase 4: Per-Instance Attribution (SAV-04)
**Rationale:** Attribution answers "which instances contribute most?"—key for optimization
**Delivers:** Per-instance savings table, sortable columns, drill-down links
**Addresses:** SAV-04 (per-instance savings attribution)
**Avoids:** Pitfall #6 (N+1 queries)
**Estimated time:** 0.5-1 day

### Phase 5: Cost Projection (SAV-05)
**Rationale:** Comparison validates SnoozeQL's value proposition
**Delivers:** Actual vs projected chart, "what you would have spent" baseline
**Addresses:** SAV-05 (cost projection), SAV-02 (actual vs expected)
**Avoids:** Pitfall #12 (confusing stopped with saving—add disclaimer)
**Estimated time:** 0.5-1 day

### Phase Ordering Rationale

- **Calculation before visualization:** Phase 1 must complete before phases 3-5 have data to display
- **API before frontend:** Phase 2 exposes endpoints that phases 3-5 consume
- **Core dashboard before details:** Phase 3 delivers the primary value; phases 4-5 add depth
- **Group by architectural layer:** Backend (1-2), then frontend (3-5)

### Research Flags

Phases with standard patterns (skip research-phase):
- **Phase 1:** Calculation logic is simple arithmetic; existing Event/Instance models sufficient
- **Phase 2:** Follow existing handler patterns from `internal/api/handlers/`
- **Phase 3:** Recharts already in use; follow existing Dashboard.tsx patterns
- **Phase 4:** Standard table component with sorting
- **Phase 5:** Composed chart with two data series

No phases require additional deep research—architecture and patterns are well-documented in the codebase.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | No new dependencies; all existing libraries sufficient |
| Features | HIGH | Clear requirements from PROJECT.md; industry patterns verified |
| Architecture | HIGH | Existing codebase provides patterns; decorator approach minimizes risk |
| Pitfalls | MEDIUM | AWS 7-day behavior documented; race condition mitigation needs testing |

**Overall confidence:** HIGH

### Gaps to Address

- **AWS 7-day auto-restart verification:** Test that savings calculation correctly caps at 7 days. Manual testing required with long-stopped instance.
- **Event ordering edge cases:** Unit tests needed for out-of-order events, duplicates, and missing start events.
- **Pricing accuracy disclaimer:** UI must clearly communicate that estimates may differ from actual billing.
- **Historical backfill:** One-time backfill job needed to calculate savings for events that occurred before v1.1 deployment.

## Sources

### Primary (HIGH confidence)
- SnoozeQL v1.0 codebase — existing models, handlers, and UI patterns
- `internal/models/models.go` — Saving, Event, Instance models already defined
- `deployments/docker/migrations/001_base_schema.sql` — existing schema
- `Dashboard.tsx` — existing Recharts usage patterns

### Secondary (MEDIUM confidence)
- AWS RDS Documentation — confirms 7-day auto-restart behavior
- AWS RDS Pricing — validates cost estimation approach
- FinOps Foundation Reporting & Analytics — industry standard metrics

### Tertiary (LOW confidence)
- Community patterns for cloud cost dashboards — informed feature prioritization

---
*Research completed: 2026-02-23*
*Ready for roadmap: yes*
