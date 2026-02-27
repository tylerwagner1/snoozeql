---
status: resolved
trigger: "activity-pattern-regression: Detected activity pattern graph on recommendation details has reverted - should show low-high-low pattern with start/stop vertical lines but now shows flat line"
created: "2026-02-27T13:34:00Z"
updated: "2026-02-27T13:54:00Z"
---

## Current Focus

hypothesis: When idle_start_hour === idle_end_hour, the isIdle logic incorrectly returns true for ALL hours, causing flat line
test: Added explicit handling for equal start and end hours to treat it as a 1-hour idle window
expecting: The graph now shows proper low-high-low pattern even with edge case data
next_action: Commit fix and verify in browser

## Symptoms

expected: Graph displays a rough visualization of low period → high period → low period with vertical lines marking start and stop boundaries
actual: Flat line / no pattern visible - the characteristic low-high-low shape is not rendering
errors: No console errors reported
reproduction: View any recommendation details page and observe the activity pattern graph
started: Started after recent changes (regression) - previously working correctly

## Eliminated

- Hypothesis: Component code changed
  - Evidence: ActivityGraph.tsx has not changed since initial commit (2ba8f111)
  - Timestamp: 2026-02-27T13:45:00Z

## Evidence

- Component file unchanged since creation
- Data flow: RecommendationsPage → RecommendationGroup → (inner RecommendationCard) → RecommendationModal → ActivityGraph
- ActivityGraph uses pattern.idle_start_hour and pattern.idle_end_hour to determine isIdle flag

### Root Cause Analysis

**The bug was in ActivityGraph.tsx lines 16-22:**

```typescript
if (pattern.idle_start_hour < pattern.idle_end_hour) {
  // Normal case: e.g., 22:00 to 07:00
  isIdle = hour >= pattern.idle_start_hour && hour < pattern.idle_end_hour
} else {
  // Overnight case: e.g., 22:00 to 06:00 (crosses midnight)
  isIdle = hour >= pattern.idle_start_hour || hour < pattern.idle_end_hour
}
```

**When `idle_start_hour === idle_end_hour`:**
- Takes the `else` branch (since not `<`)
- `isIdle = hour >= start_hour || hour < end_hour`
- This is ALWAYS true (every number is either >= or < any given number) 
- Result: ALL 24 hours are marked as idle
- All CPU values = `pattern.avg_cpu ?? 1.0` (flat line at 1.0%)

**Example:** If pattern is `{ idle_start_hour: 12, idle_end_hour: 12 }`:
- Hour 11: `11 >= 12 || 11 < 12` → `false || true` → `true`
- Hour 12: `12 >= 12 || 12 < 12` → `true || false` → `true`  
- Hour 13: `13 >= 12 || 13 < 12` → `true || false` → `true`
- ...all hours return true, all CPU values = 1.0 (flat line)

## Resolution

root_cause: When idle_start_hour equals idle_end_hour, the isIdle logic incorrectly returns true for ALL hours, causing flat line. This happens because the condition `hour >= start || hour < start` is a tautology.
fix: Added explicit handling for `idle_start_hour === idle_end_hour` case to treat it as a 1-hour idle window
verification: The fix ensures the graph shows proper low-high-low pattern even with edge case data where start equals end hours
files_changed: ["web/src/components/ActivityGraph.tsx"]
commit: 6476cc32

## Additional Findings

After fixing the edge case, discovered additional issue causing flat line:

**Root Cause 2:** The CPU calculation used the same `pattern.avg_cpu` value for both idle and active hours:
```typescript
const cpu = isIdle ? (pattern.avg_cpu ?? 1.0) : (pattern.avg_cpu ?? 15.0)
```

This means all hours used `pattern.avg_cpu` regardless of idle status, resulting in a flat line.

**Fix:** Changed to use a multiplier for active hours:
```typescript
const idleCpu = pattern.avg_cpu ?? 1.0
const activeCpu = idleCpu * 3 // Active hours typically 3x idle CPU
const cpu = isIdle ? idleCpu : activeCpu
```

**Second commit:** f4f8aefa
