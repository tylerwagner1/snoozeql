# Features Research: v1.2 Metrics & Recommendations

**Researched:** 2026-02-24
**Focus:** Metrics visualization and recommendation features
**Confidence:** HIGH (AWS RDS documentation, existing codebase analysis, industry patterns)

## Summary

Database metrics dashboards universally display time-series charts for CPU, connections, and IOPS with configurable time ranges (1h, 6h, 24h, 7d). Idle detection thresholds are simple and effective when based on CPU < X% AND connections = 0 patterns. Recommendation grouping by pattern similarity (similar idle windows across instances) is the standard approach for scaling recommendations to many instances.

## Existing Implementation (Already Built)

Understanding what's already implemented to avoid duplicating work:

| Feature | Status | Location |
|---------|--------|----------|
| Metrics collection every 15 min | ✅ Built | `internal/metrics/collector.go` |
| HourlyMetric model (avg, min, max, samples) | ✅ Built | `internal/models/models.go` |
| Latest metrics cards on Instance Details | ✅ Built | `web/src/pages/InstanceDetailPage.tsx` |
| API endpoint for instance metrics | ✅ Built | `GET /instances/:id/metrics` |
| Activity pattern analysis | ✅ Built | `internal/analyzer/patterns.go` |
| Idle window detection (8+ hours, 3+ days consistent) | ✅ Built | Thresholds: CPU < 1% |
| Recommendation engine | ✅ Built | `internal/analyzer/recommendation.go` |
| Recommendations UI with confirm/dismiss | ✅ Built | `RecommendationsPage.tsx` |
| Activity graph for recommendations | ✅ Built | `ActivityGraph.tsx` (Recharts) |

---

## Metrics Visualization

### Table Stakes (Must Have)

These features are expected by users viewing database metrics. Missing = product feels incomplete.

- **Time-series chart on Instance Details** — Complexity: Medium
  - Display CPU, Connections, IOPS as line/area charts over time
  - Users expect to see trends, not just latest values
  - Depends on: Existing metrics API, HourlyMetric data

- **Time range selector (1h, 6h, 24h, 7d)** — Complexity: Low
  - Standard ranges users expect for metrics exploration
  - 7-day matches current retention period
  - Depends on: API endpoint modification to accept time range

- **Multiple metrics on single view** — Complexity: Low
  - CPU, Connections, IOPS visible together (stacked or tabbed)
  - Pattern: Separate charts or overlay with toggle
  - Depends on: Chart component supporting multiple series

- **Loading/empty states for charts** — Complexity: Low
  - "No data yet" when metrics haven't been collected
  - Skeleton loading while fetching
  - Depends on: None

- **Responsive charts** — Complexity: Low
  - Charts resize with viewport (ResponsiveContainer pattern)
  - Already implemented in ActivityGraph.tsx
  - Depends on: Recharts (already in stack)

### Differentiators (Nice to Have)

Features that set product apart but aren't expected.

- **Anomaly highlighting on charts** — Complexity: Medium
  - Shade periods of high CPU or unusual connection spikes
  - Visual indication of "interesting" periods
  - Depends on: Threshold configuration

- **Correlation view (CPU vs Connections)** — Complexity: Medium
  - Dual-axis chart showing relationship
  - Helps identify "truly idle" vs "idle but connected"
  - Depends on: Chart component with dual Y-axis

- **Metric comparison across instances** — Complexity: High
  - Select 2-3 instances and compare metrics
  - Useful for identifying outliers
  - Depends on: New UI, API for batch metrics

- **Export metrics as CSV** — Complexity: Low
  - Download raw metric data for external analysis
  - Depends on: API endpoint for CSV format

### Anti-Features (Don't Build)

Features to explicitly NOT build for v1.2 POC.

- **Real-time streaming metrics** — Why not: Complexity, 15-min collection is sufficient for idle detection
  - Instead: Poll API on user-initiated refresh

- **Enhanced Monitoring (OS-level metrics)** — Why not: Requires Enhanced Monitoring setup per AWS, adds cost
  - Instead: Stick with standard CloudWatch metrics

- **Custom alerting/thresholds** — Why not: Requires notification infrastructure (out of scope)
  - Instead: Visual display of current thresholds on chart

- **Performance Insights integration** — Why not: Additional AWS cost, complexity
  - Instead: Standard CloudWatch metrics sufficient for idle detection

---

## Recommendations

### Table Stakes (Must Have)

- **Simple threshold-based idle detection** — Complexity: Low (already implemented partially)
  - Current: CPU < 1%, 8+ hours, 3+ consistent days
  - v1.2 goal: CPU < 5% AND connections = 0
  - Depends on: patterns.go threshold adjustment

- **Recommendation grouping by pattern** — Complexity: Medium
  - Group instances with similar idle windows together
  - "3 instances idle 10pm-6am weekdays" as single recommendation
  - Depends on: Pattern clustering logic

- **Per-instance override within group** — Complexity: Medium
  - Accept recommendation for group but exclude specific instance
  - "Apply to all except production-db"
  - Depends on: UI for instance selection in group

- **Recommendation confidence score display** — Complexity: Low (already built)
  - Show HIGH/MEDIUM/LOW confidence badge
  - Already calculated in patterns.go
  - Depends on: None (polish existing UI)

- **Estimated savings per recommendation** — Complexity: Low (already built)
  - Calculate daily savings based on idle hours × hourly cost
  - Already implemented in handlers/recommendations.go
  - Depends on: None (verify accuracy)

### Differentiators (Nice to Have)

- **Recommendation history/audit** — Complexity: Medium
  - Track which recommendations were confirmed, dismissed, and when
  - "Last recommendation generated: 2 hours ago"
  - Depends on: New table or status tracking

- **Bulk confirm/dismiss** — Complexity: Low
  - Select multiple recommendations and confirm/dismiss at once
  - Depends on: UI checkboxes, batch API endpoint

- **Recommendation refresh scheduling** — Complexity: Low
  - "Check for new patterns every 6 hours"
  - Background job, not user-triggered
  - Depends on: Scheduler infrastructure (already exists for metrics)

- **Pattern evolution tracking** — Complexity: High
  - "This instance's idle pattern shifted from weekdays to all week"
  - Useful for ongoing optimization
  - Depends on: Historical pattern storage

### Anti-Features (Don't Build)

- **ML-based pattern detection** — Why not: Complex, training data needed, overkill for POC
  - Instead: Simple threshold rules are effective and explainable

- **Auto-apply recommendations** — Why not: User explicitly wants confirmation workflow
  - Instead: Confirm/dismiss UI (already built)

- **Cross-schedule conflict detection** — Why not: Edge case, adds complexity
  - Instead: Show warning if instance already has schedule

- **Recommendation re-generation for dismissed** — Why not: If user dismissed, respect that decision
  - Instead: Clear dismissed after 30 days, or allow manual re-enable

---

## Idle Detection Thresholds

### Table Stakes

Current implementation uses CPU < 1% threshold. v1.2 requirement is simpler:

| Threshold | Current | v1.2 Target | Notes |
|-----------|---------|-------------|-------|
| CPU | < 1% | < 5% | More lenient, catches more idle |
| Connections | Not checked | = 0 | Critical: zero connections = truly idle |
| IOPS | Not checked | Not needed | CPU + connections sufficient |
| Duration | 8+ hours | 8+ hours | Keep same |
| Consistency | 3+ days | 3+ days | Keep same |

**Recommendation:** Add connections check to idle detection logic. Pattern: `CPU < 5% AND connections = 0` for `8+ hours` on `3+ consistent days`.

### Complexity: Low

Change in `internal/analyzer/patterns.go`:
- Add connections check to `findIdleSegments()`
- Currently only checks CPU values
- ConnValues already populated in buckets

---

## Recommendation Grouping

### Pattern Grouping Logic

Group recommendations by similar idle windows:

| Grouping Criteria | Example |
|-------------------|---------|
| Same start hour (±1h tolerance) | All instances sleeping at 10pm |
| Same end hour (±1h tolerance) | All instances waking at 6am |
| Same days of week | All weekday-idle instances grouped |

**Already implemented in** `groupSimilarWindows()` in patterns.go.

### UI Representation

```
┌─────────────────────────────────────────────────────────────┐
│ 3 instances idle weekdays 10pm-6am                          │
│ ───────────────────────────────────────────────────────────│
│ ☐ dev-postgres-01    $2.40/day savings                     │
│ ☑ dev-postgres-02    $2.40/day savings                     │
│ ☑ staging-mysql-01   $1.80/day savings                     │
│                                                             │
│ [Confirm Selected] [Dismiss All]                           │
└─────────────────────────────────────────────────────────────┘
```

**Complexity:** Medium — requires grouped recommendation API response and new UI component.

---

## Feature Dependencies

```
Existing:
  HourlyMetric collection → Activity patterns → Recommendations
  
v1.2 Additions:
  Time-series charts ← HourlyMetric API (needs time range param)
  Recommendation groups ← Pattern grouping (partially exists)
  Per-instance overrides ← Group UI component (new)
```

### Dependency Order

1. **Phase 1:** Add time range to metrics API + time-series chart component
2. **Phase 2:** Add connections threshold to idle detection
3. **Phase 3:** Expose grouped recommendations in API
4. **Phase 4:** Per-instance override UI within groups

---

## MVP Recommendation

For v1.2 MVP, prioritize:

1. **Time-series chart on Instance Details** (table stakes, visible impact)
2. **Connections threshold in idle detection** (core improvement)
3. **Recommendation grouping display** (table stakes for scaling)
4. **Per-instance override within group** (table stakes for usability)

Defer to post-MVP:

- Anomaly highlighting (differentiator, not expected)
- Pattern evolution tracking (high complexity)
- Export to CSV (low priority for POC)

---

## Sources

### High Confidence (Official Documentation)

- AWS RDS CloudWatch Metrics: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-metrics.html
- AWS CloudWatch monitoring for RDS: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/monitoring-cloudwatch.html

### Medium Confidence (Codebase Analysis)

- Existing patterns.go thresholds and grouping logic
- ActivityGraph.tsx Recharts implementation
- RecommendationsPage.tsx workflow

### Feature Complexity Estimates

| Feature | Complexity | Effort |
|---------|------------|--------|
| Time-series chart | Medium | 1-2 days |
| Time range selector | Low | 0.5 day |
| Connections threshold | Low | 0.5 day |
| Recommendation grouping API | Medium | 1 day |
| Per-instance override UI | Medium | 1-2 days |

---

*Feature landscape research: 2026-02-24*
