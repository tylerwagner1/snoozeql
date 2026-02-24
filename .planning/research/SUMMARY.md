# Project Research Summary

**Project:** SnoozeQL v1.2 - Metrics & Recommendations
**Domain:** Cloud database metrics and cost optimization
**Researched:** 2026-02-24
**Confidence:** HIGH

## Executive Summary

SnoozeQL v1.2 requires minimal stack changes because the existing infrastructure—Go 1.24.0 with AWS SDK v2, React 18.2 with Recharts 2.10.0, and PostgreSQL with the `metrics_hourly` table—already supports all core requirements. The primary work is adding one CloudWatch metric (FreeableMemory), building time-series chart components on the Instance Details page, and improving threshold-based idle detection to use a compound rule (CPU < 5% AND connections = 0). No new dependencies are required.

The recommended approach follows the existing codebase patterns: extend the CloudWatch client to collect FreeableMemory, add a `/metrics/history` endpoint for time-series data, create a Recharts-based MetricsChart component, and update the analyzer's threshold logic. The architecture supports these changes without schema migrations—`metrics_hourly` already handles arbitrary metric names.

Key risks are CloudWatch API throttling during batch operations, misleading idle detection (simple thresholds missing batch jobs or replicas), and time-series chart performance with 7 days of hourly data. These are mitigated by leveraging existing retry logic with rate limiting, requiring connections = 0 for true idle detection, and downsampling chart data for display.

## Key Findings

### Recommended Stack

The existing stack requires no additions. All v1.2 features build on current infrastructure:

**Core technologies (unchanged):**
- **Go 1.24.0 + aws-sdk-go-v2**: Backend and CloudWatch integration — add FreeableMemory to existing collector
- **Recharts 2.10.0**: Time-series visualization — ActivityGraph.tsx already proves the pattern
- **PostgreSQL (metrics_hourly table)**: Metrics storage — schema supports new metric names without migration
- **Chi router 5.2.0**: REST API — add one new endpoint for metrics history

**What NOT to add:**
- date-fns, moment.js (use native Date APIs)
- Chart.js (already have Recharts)
- React Query (existing fetch patterns work)
- Enhanced Monitoring (adds cost, overkill for POC)

### Expected Features

**Must have (table stakes):**
- Time-series chart on Instance Details page (CPU, Memory, Connections, IOPS)
- Time range selector (24h, 7d — matches retention)
- Multiple metrics visible together (stacked or tabbed)
- Loading/empty states for charts
- Simple threshold-based idle detection (CPU < 5% AND connections = 0)
- Recommendation grouping by pattern similarity
- Per-instance override within grouped recommendations
- Confidence score and estimated savings display (already built)

**Should have (differentiators):**
- Anomaly highlighting on charts
- Correlation view (CPU vs Connections dual-axis)
- Bulk confirm/dismiss for recommendations
- Recommendation staleness warnings

**Defer (v2+):**
- Real-time streaming metrics
- ML-based pattern detection
- Cross-instance metric comparison
- Pattern evolution tracking
- Export to CSV

### Architecture Approach

All changes extend existing components with no new architectural patterns. The data flow is: CloudWatch -> MetricsCollector (add FreeableMemory) -> metrics_hourly (existing schema) -> new `/metrics/history` endpoint -> MetricsChart.tsx (new Recharts component). Recommendation flow: metrics_hourly -> Analyzer (updated thresholds) -> grouped API response -> UI with per-instance overrides.

**Major integration points:**
1. **CloudWatch client (`cloudwatch.go`)**: Add FreeableMemory fetch alongside existing CPU, Connections, IOPS
2. **Metrics API (`handlers/instances.go`)**: Add time-series endpoint with time range params
3. **Analyzer (`patterns.go`)**: Update ActivityThresholds for compound idle detection
4. **Recommendations API (`handlers/recommendations.go`)**: Add grouping logic to response
5. **Frontend (`InstanceDetailPage.tsx`)**: Integrate new MetricsChart component

### Critical Pitfalls

1. **CloudWatch API throttling** — Add rate limiting for batch operations; existing retry logic handles transient failures but backfill scenarios need self-throttling
2. **FreeableMemory confusion** — CloudWatch reports bytes, not percentage; either display as GB or calculate percentage using instance class lookup table
3. **Timezone confusion in charts** — Metrics stored UTC, must convert to user timezone for display; add chart caption showing timezone
4. **7-day retention vs analysis window** — Changing from 14-day to 7-day retention must align analyzer lookback period; document tradeoff before implementation
5. **Recommendation-schedule conflicts** — Check for existing schedules before generating recommendations; warn users if instance already has a schedule

## Implementation Approach

Based on dependencies and pitfall avoidance, suggested build order:

### Phase 1: Metrics Collection Enhancement
**Rationale:** Foundation for all other changes; minimal risk, isolated change
**Delivers:** FreeableMemory collection and storage
**Addresses:** Memory utilization visibility (table stakes)
**Avoids:** Pitfall #2 (memory % confusion) by deciding display format upfront

### Phase 2: Time-Series API & Charts
**Rationale:** Backend must exist before frontend; highest visible impact
**Delivers:** `/metrics/history` endpoint + MetricsChart.tsx
**Addresses:** Time-series visualization (table stakes), time range selector
**Avoids:** Pitfall #6 (performance) by testing with 7 days of data, Pitfall #7 (timezone) by converting to local time

### Phase 3: Retention Configuration
**Rationale:** Simple config change, do after charts work to avoid data loss during dev
**Delivers:** 7-day retention enforcement
**Avoids:** Pitfall #4 (retention mismatch) by aligning analyzer config

### Phase 4: Idle Detection Improvement
**Rationale:** Isolated analyzer change, improves recommendation quality
**Delivers:** Compound threshold (CPU < 5% AND connections = 0)
**Addresses:** Better idle detection accuracy
**Avoids:** Pitfall #10 (simplistic detection) by requiring zero connections

### Phase 5: Grouped Recommendations
**Rationale:** Depends on improved detection; API change + UI update
**Delivers:** Grouped recommendation API, per-instance override UI
**Addresses:** Recommendation grouping (table stakes), per-instance overrides
**Avoids:** Pitfall #11 (conflicts) by checking existing schedules, Pitfall #12 (wrong schedules) by allowing overrides

### Phase Ordering Rationale

- **Phase 1 before Phase 2**: Charts need memory metric available
- **Phase 2 before Phase 3**: Verify charts work before reducing data retention
- **Phase 4 independent**: Can run parallel to Phase 2-3 if needed
- **Phase 5 last**: Depends on Phase 4's improved detection for better groupings

### Research Flags

**Standard patterns (skip research-phase):**
- **Phase 1 (Metrics)**: Well-documented CloudWatch API, existing collector pattern
- **Phase 2 (Charts)**: Recharts patterns proven in ActivityGraph.tsx
- **Phase 3 (Retention)**: Simple config change

**May need validation during implementation:**
- **Phase 4 (Idle Detection)**: Test threshold values with real data; may need tuning
- **Phase 5 (Grouped Recommendations)**: API response format change may need frontend coordination

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Verified against codebase; no additions needed |
| Features | HIGH | Based on AWS docs, existing implementation, industry patterns |
| Architecture | HIGH | All integration points verified in current codebase |
| Pitfalls | MEDIUM | General patterns verified; some SnoozeQL-specific risks need validation |

**Overall confidence:** HIGH

### Gaps to Address

- **Memory percentage calculation**: Decide whether to show FreeableMemory as GB or calculate percentage; percentage requires instance class lookup table
- **Threshold tuning**: CPU < 5% AND connections = 0 is the starting point; may need adjustment based on real user data
- **Chart performance thresholds**: Recharts performance with actual data volume needs testing; downsampling parameters TBD
- **7-day AWS auto-restart**: Deferred in PROJECT.md but affects long-running schedules; ensure schedule validation warns users

## Sources

### Primary (HIGH confidence)
- AWS RDS CloudWatch Metrics documentation (FreeableMemory specification)
- SnoozeQL codebase: `internal/metrics/`, `internal/analyzer/`, `web/src/components/`
- Recharts v2.x API documentation

### Secondary (MEDIUM confidence)
- CloudWatch GetMetricStatistics rate limits (account-tier dependent)
- Common time-series visualization performance patterns

### Tertiary (needs validation)
- Threshold values for idle detection (CPU < 5%, connections = 0) may need tuning
- Recommendation staleness window (3 days suggested, may need user feedback)

---
*Research completed: 2026-02-24*
*Ready for roadmap: yes*
