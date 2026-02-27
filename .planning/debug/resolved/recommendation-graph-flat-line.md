---
status: resolved
trigger: "recommendation-graph-flat-line"
created: 2026-02-26T00:00:00Z
updated: 2026-02-26T00:00:00Z
---

## Current Focus

hypothesis: ActivityGraph.tsx generates synthetic data with minimal CPU variation (1% vs ~15%), creating a flat line appearance
test: Updated ActivityGraph to accept optional metrics and plot real CPU data
expecting: Graph now displays actual CPU utilization from metrics_hourly table when available
next_action: Verification complete

## Symptoms

expected: Line graph showing last 24hrs of CPU utilization with proper fluctuation/data points
actual: Line graph shows flat line (no fluctuation) - only start/stop vertical lines visible
errors: None reported yet
reproduction: Open recommendation details view
started: User experiencing issue now

## Evidence

- ActivityGraph.tsxline 30-42: CPU is calculated as isIdle ? avg_cpu : avg_cpu + 14
- This creates only 2 values: ~1% (idle) and ~15% (active)  
- Both "idle" and "activity" areas plot the SAME data series
- No actual CPU metrics fetched - purely synthetic based on detected pattern
- The 24hr graph shows only 2 distinct CPU levels, appearing as flat line

## Resolution

root_cause: ActivityGraph.tsx generates synthetic data with minimal CPU variation (1% vs ~15%), creating a flat line appearance. The component should visualize actual CPU utilization metrics from metrics_hourly table, not generated pattern data.
fix: 
- backend: N/A (used existing API)
- frontend: Updated ActivityGraph.tsx to accept optional metrics prop
- frontend: Updated RecommendationModal.tsx to fetch CPU metrics when modal opens
files_changed: 
  - /Users/tylerwagner/snoozeql/web/src/components/ActivityGraph.tsx
  - /Users/tylerwagner/snoozeql/web/src/components/RecommendationModal.tsx
verification: 
- ActivityGraph now displays real CPU metrics when provided, falls back to pattern data otherwise
- RecommendationModal fetches 24h CPU metrics from api.getMetricsHistory(instanceId, '24h')
- Loading state shown while metrics are being fetched

