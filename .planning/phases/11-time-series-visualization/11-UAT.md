---
status: diagnosed
phase: 11-time-series-visualization
source: 11-01-SUMMARY.md, 11-02-SUMMARY.md
started: 2026-02-25T18:45:00Z
updated: 2026-02-25T18:47:00Z
---

## Current Test

[issue fixed - ready for re-test]

## Tests

### 1. Metrics History Section Visible
expected: Navigate to any Instance Details page. Below the existing "Metrics" section, you should see a new "Metrics History" section with a chart area.
result: issue
reported: "navigating to any instance details page leaves whole browser black screen it requires going to different tab and hard refresh to escape"
severity: blocker

### 2. Tab Switching (CPU, Memory, Connections)
expected: In the Metrics History section, you should see three tabs: CPU, Memory, and Connections. CPU should be selected by default. Clicking each tab should switch which metric is displayed in the chart.
result: skipped
reason: Blocked by Test 1 failure (page won't load)

### 3. Time Range Selector
expected: You should see time range buttons (1h, 6h, 24h, 7d) to the right of the tabs. 24h should be selected by default. Clicking a different range should update the chart to show data for that time period.
result: skipped
reason: Blocked by Test 1 failure (page won't load)

### 4. Loading State
expected: When switching tabs or time ranges, you should briefly see a loading spinner in the chart area while data is being fetched.
result: skipped
reason: Blocked by Test 1 failure (page won't load)

### 5. Empty State
expected: If there's no metrics data for the selected time range (e.g., a new instance or one that's been sleeping), you should see "No data available" message with chart axes still visible.
result: skipped
reason: Blocked by Test 1 failure (page won't load)

### 6. Chart Display with Data
expected: When metrics data exists, the chart should show a line graph. The Y-axis should show 0-100% for CPU/Memory tabs, and auto-scale for Connections. Hovering over the line should show a tooltip with the exact value and timestamp.
result: skipped
reason: Blocked by Test 1 failure (page won't load)

## Summary

total: 6
passed: 0
issues: 1
pending: 0
skipped: 5

## Gaps

- truth: "Instance Details page loads and displays Metrics History section"
  status: failed
  reason: "User reported: navigating to any instance details page leaves whole browser black screen it requires going to different tab and hard refresh to escape"
  severity: blocker
  test: 1
  root_cause: "MetricsChart uses useQuery from @tanstack/react-query but App.tsx has no QueryClientProvider wrapper"
  artifacts:
    - path: "web/src/App.tsx"
      issue: "Missing QueryClientProvider wrapper for react-query"
    - path: "web/src/components/MetricsChart.tsx"
      issue: "Uses useQuery which requires QueryClientProvider in parent tree"
  missing:
    - "Add QueryClientProvider wrapper in App.tsx"
    - "Create QueryClient instance"
  debug_session: "console error: No QueryClient set, use QueryClientProvider to set one"
  fix_commit: "1023ff0f - fix(11): add QueryClientProvider for react-query"
