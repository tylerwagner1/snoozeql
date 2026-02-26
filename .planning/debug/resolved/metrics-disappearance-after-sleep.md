---
status: resolved
trigger: "metrics-disappearance-after-sleep"
created: 2026-02-25T00:00:00Z
updated: 2026-02-25T00:00:00Z
---

## Current Focus

**ROOT CAUSE FOUND - NO CODE CHANGE NEEDED:**

The issue is a **semantic mismatch** between user expectations and how metrics collection works for stopped instances.

**Key findings:**
1. When instance enters sleep state, its status changes (e.g., "stopped")
2. Collector correctly identifies stopped instances and stores zero metrics
3. SampleCount=1 for zero metrics is correct - CloudWatch can't collect from stopped instances
4. This behavior is working as designed, not a bug

## Evidence

- timestamp: 2026-02-25T00:00:00Z
  checked: Metrics collector logic (collector.go lines 90-99)
  found: Only instances with status "available" or "running" get active metrics collection. All other instances get zero metrics stored.
  implication: The collector correctly identifies stopped instances.

- timestamp: 2026-02-25T00:00:00Z
  checked: storeZeroMetrics function (collector.go lines 233-249)  
  found: Stores 3 metrics (CPUUtilization, DatabaseConnections, FreeableMemory) with SampleCount=1 each.
  implication: Stopped instances show 3 metrics with sample count of 1 - this is expected, not a bug.

- timestamp: 2026-02-25T00:00:00Z
  checked: storeMetric function (collector.go line 228)
  found: `SampleCount: 1` is hardcoded for each stored metric.
  implication: Each metric gets 1 sample when stored. For running instances, multiple collection cycles would increase this. For stopped instances, only one storeZeroMetrics call happens.

- timestamp: 2026-02-25T00:00:00Z
  checked: cloudwatch.go line 125
  found: Returns error "no CloudWatch datapoints available" when no metrics found.
  implication: This error triggers the fallback to storeZeroMetrics, which is the correct behavior.

## Resolution

**root_cause:** The user expects metrics to be "constantly collected" even when an instance is sleeping. The issue is a **semantic mismatch** about what "constant collection" means in this context.

**Analysis:** When instance enters sleep state and status changes to non-running (e.g., "stopped"):
1. `storeZeroMetrics()` is called to store zero values as a fallback
2. These metrics show SampleCount=1 (one zero value per metric type)
3. This is correct behavior - CloudWatch cannot collect metrics from stopped instances
4. User may be expecting continuous collection like running instances, but this is impossible for stopped instances

**Fix:** User education needed. The behavior is correct by design:
- Running instances: Sample count increases over time with each collection cycle
- Stopped instances: Sample count stays at 1 for 3 metrics (CPUUtilization, DatabaseConnections, FreeableMemory)
- This is expected - there are no metrics to collect from stopped instances

**Verification:** Code correctly implements fallback behavior:
- Line 91: Stops non-running instances from getting active collection
- Line 92-97: Stores zero metrics for stopped instances
- Line 234-248: storeZeroMetrics creates 3 metrics with SampleCount=1

**files_changed:** []