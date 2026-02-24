---
status: diagnosed
phase: 10-metrics-collection-enhancement
source: 10-01-SUMMARY.md
started: 2026-02-24T21:45:00Z
updated: 2026-02-24T21:52:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Memory Available Card on Instance Details
expected: On the Instance Details page for a running instance, you should see a "Memory Available" card showing a percentage value (0-100%) alongside the existing CPU and Connections cards.
result: issue
reported: "i don't see anything, i checked the running oregon-database and it just says 'no metrics data available yet'"
severity: major

### 2. Metrics Unavailable Badge (No Data)
expected: For an instance with no recent metrics (or a newly added instance), the Instance Details page should show a yellow "Metrics unavailable" badge near the Metrics section header.
result: issue
reported: "no, there's no badge. i would guess something is cached or stale"
severity: major

### 3. Stopped Instance Behavior
expected: For a stopped instance, the "Metrics unavailable" badge should NOT appear (stopped instances are expected to have no active metrics). The Memory card may show 0% or no data.
result: skipped
reason: Cannot verify — badge functionality broken (Test 2), so absence is inconclusive

### 4. Backend Compilation
expected: Running `go build ./...` in the project root should complete without errors.
result: pass

### 5. Frontend Build
expected: Running `cd web && npm run build` should complete without TypeScript errors.
result: pass

## Summary

total: 5
passed: 2
issues: 2
pending: 0
skipped: 1

## Gaps

- truth: "Memory Available card showing percentage value on Instance Details page"
  status: failed
  reason: "User reported: i don't see anything, i checked the running oregon-database and it just says 'no metrics data available yet'"
  severity: major
  test: 1
  root_cause: "No metrics data in database - CloudWatch may not have returned datapoints yet, or instance class not in memory mapping"
  artifacts:
    - path: "internal/metrics/collector.go"
      issue: "collectInstance() stores nothing if CloudWatch returns no datapoints"
    - path: "internal/metrics/memory.go"
      issue: "instanceClassMemoryGB mapping incomplete - missing many RDS instance types"
  missing:
    - "Add support for more RDS instance types to memory mapping"
    - "Consider fallback mechanism when CloudWatch returns no datapoints"
  debug_session: ".planning/debug/metrics-not-showing.md"

- truth: "Yellow 'Metrics unavailable' badge appears near Metrics section header when no recent metrics"
  status: failed
  reason: "User reported: no, there's no badge. i would guess something is cached or stale"
  severity: major
  test: 2
  root_cause: "Badge condition incorrectly required instance.status !== 'stopped' - badge was being hidden"
  artifacts:
    - path: "web/src/pages/InstanceDetailPage.tsx"
      issue: "Line 232 had incorrect condition preventing badge render"
  missing:
    - "Remove instance.status !== 'stopped' from badge rendering condition"
  debug_session: ".planning/debug/resolved/metrics-badge-not-showing.md"
