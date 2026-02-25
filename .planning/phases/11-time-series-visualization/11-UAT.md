---
status: diagnosed
phase: 11-time-series-visualization
source: 11-01-SUMMARY.md, 11-02-SUMMARY.md
started: 2026-02-25T18:45:00Z
updated: 2026-02-25T18:47:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Metrics History Section Visible
expected: Navigate to any Instance Details page. Below the existing "Metrics" section, you should see a new "Metrics History" section with a chart area.
result: pass
note: "Initially failed (blocker), fixed with QueryClientProvider in commit 1023ff0f"

### 2. Tab Switching (CPU, Memory, Connections)
expected: In the Metrics History section, you should see three tabs: CPU, Memory, and Connections. CPU should be selected by default. Clicking each tab should switch which metric is displayed in the chart.
result: pass

### 3. Time Range Selector
expected: You should see time range buttons (1h, 6h, 24h, 7d) to the right of the tabs. 24h should be selected by default. Clicking a different range should update the chart to show data for that time period.
result: issue
reported: "the x axis is broken on all of them. I would expect uniform gaps between times (right now it goes 1:00, 1:05, 1:20 on 1hr view) and correct hours or days for 24hr and 7day views (24hr shows 11am, 12pm 5 times, 1pm 5 times & 7 day view just says 'Feb 25' 11 times). If there isn't data to go back 6hr, 24hr, or 7 days, i would expect correct axis but no data until the first datapoint i have registered today"
severity: major

### 4. Loading State
expected: When switching tabs or time ranges, you should briefly see a loading spinner in the chart area while data is being fetched.
result: pass

### 5. Empty State
expected: If there's no metrics data for the selected time range (e.g., a new instance or one that's been sleeping), you should see "No data available" message with chart axes still visible.
result: pass

### 6. Chart Display with Data
expected: When metrics data exists, the chart should show a line graph. The Y-axis should show 0-100% for CPU/Memory tabs, and auto-scale for Connections. Hovering over the line should show a tooltip with the exact value and timestamp.
result: pass

## Summary

total: 6
passed: 5
issues: 1
pending: 0
skipped: 0

## Gaps

- truth: "X-axis shows uniform time intervals appropriate for selected range"
  status: fixed
  reason: "User reported: X-axis shows irregular gaps (1:00, 1:05, 1:20 on 1hr), duplicate labels (12pm 5 times on 24hr), and same date repeated (Feb 25 x11 on 7d). Expected uniform intervals with gaps where no data exists."
  severity: major
  test: 3
  root_cause: "XAxis was using categorical data (string hours) instead of numeric timestamps with scale='time'. Each data point created its own tick label."
  artifacts:
    - path: "web/src/components/MetricsChart.tsx"
      issue: "XAxis dataKey='hour' treated as categorical, not time-based"
  missing:
    - "Convert hour strings to timestamps"
    - "Use type='number' and scale='time' on XAxis"
    - "Set domain to [now - range, now] for full time span"
    - "Add tickCount for appropriate intervals"
  fix_commit: "f8195d11 - fix(11): use time scale for X-axis with uniform intervals"
