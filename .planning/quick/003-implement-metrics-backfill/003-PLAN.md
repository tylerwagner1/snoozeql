---
phase: quick-003
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/metrics/collector.go
  - cmd/server/main.go
autonomous: true

must_haves:
  truths:
    - "User can trigger historical metrics backfill for an instance via API"
    - "Backfill collects up to 7 days of CloudWatch metrics at hourly granularity"
    - "Self-throttling prevents CloudWatch rate limit errors during backfill"
  artifacts:
    - path: "internal/metrics/collector.go"
      provides: "BackfillMetrics method"
      contains: "func (c *MetricsCollector) BackfillMetrics"
    - path: "cmd/server/main.go"
      provides: "POST /instances/{id}/metrics/backfill endpoint"
      contains: "metrics/backfill"
  key_links:
    - from: "cmd/server/main.go"
      to: "metricsCollector.BackfillMetrics"
      via: "API handler"
      pattern: "BackfillMetrics"
---

<objective>
Implement metrics backfill to collect historical CloudWatch data for instances

Purpose: New instances or instances that missed collection cycles need historical data for pattern analysis and time-series charts. Currently, only the last hour is collected every 15 minutes.

Output: BackfillMetrics method + API endpoint that collects up to 7 days of historical metrics
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/STATE.md
@internal/metrics/collector.go
@internal/metrics/cloudwatch.go
@cmd/server/main.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add BackfillMetrics method to MetricsCollector</name>
  <files>internal/metrics/collector.go</files>
  <action>
Add a BackfillMetrics method to MetricsCollector that:

1. Accepts context, instance, and days (max 7) parameters
2. Iterates backward hour-by-hour from now to (now - days*24 hours)
3. For each hour, calls CloudWatch with that hour's start/end time
4. Stores each metric using existing storeMetric() helper
5. Self-throttles with 100ms sleep between hours (per PITFALLS.md recommendation)
6. Skips hours that already have data (check metricsStore first)
7. Returns count of hours backfilled and any error

Method signature:
```go
func (c *MetricsCollector) BackfillMetrics(ctx context.Context, instance models.Instance, days int) (int, error)
```

Also add a helper method to CloudWatchClient that fetches metrics for a specific hour:
```go
func (c *CloudWatchClient) GetRDSMetricsForHour(ctx context.Context, dbInstanceID string, hour time.Time) (*RDSMetrics, error)
```

This is similar to GetRDSMetrics but uses the provided hour instead of "last hour".

Important: Cap days at 7 (CloudWatch free tier limitation). Log progress every 24 hours of backfill.
  </action>
  <verify>
`go build ./...` compiles without errors
  </verify>
  <done>
BackfillMetrics method exists and compiles, accepts instance and days, iterates hours with throttling
  </done>
</task>

<task type="auto">
  <name>Task 2: Add backfill API endpoint</name>
  <files>cmd/server/main.go</files>
  <action>
Add POST /api/v1/instances/{id}/metrics/backfill endpoint that:

1. Extracts instance ID from URL params
2. Fetches instance from instanceStore (404 if not found)
3. Only allows AWS instances (400 for non-AWS with helpful message)
4. Reads optional `days` query param (default 7, max 7)
5. Calls metricsCollector.BackfillMetrics()
6. Returns JSON response with hours_backfilled count

Place the route near the existing metrics endpoints (around line 711).

Response format:
```json
{
  "success": true,
  "instance_id": "...",
  "hours_backfilled": 168,
  "days_requested": 7
}
```

Error response:
```json
{
  "error": "Backfill failed: ...",
  "instance_id": "..."
}
```
  </action>
  <verify>
`go build ./...` compiles without errors
  </verify>
  <done>
POST /instances/{id}/metrics/backfill endpoint exists and wired to BackfillMetrics
  </done>
</task>

<task type="auto">
  <name>Task 3: Test backfill end-to-end</name>
  <files>N/A - verification only</files>
  <action>
1. Rebuild and restart server: `go build ./cmd/server && ./server` (or use existing process)
2. Get an AWS instance ID from the dashboard or curl /api/v1/instances
3. Check current metrics count for that instance
4. Call the backfill endpoint:
   ```bash
   curl -X POST "http://localhost:8080/api/v1/instances/{id}/metrics/backfill?days=1" \
     -H "Content-Type: application/json"
   ```
5. Verify response shows hours_backfilled > 0
6. Check metrics history endpoint shows more data points:
   ```bash
   curl "http://localhost:8080/api/v1/instances/{id}/metrics/history?range=24h"
   ```

Note: If no AWS instances available or CloudWatch returns no data, the test should still confirm:
- Endpoint responds with 200
- Response has correct JSON structure
- Logs show backfill progress
  </action>
  <verify>
Backfill endpoint returns 200 with hours_backfilled in response, logs show throttled collection
  </verify>
  <done>
Backfill API works end-to-end: triggers collection, stores metrics, returns count
  </done>
</task>

</tasks>

<verification>
1. `go build ./...` passes
2. POST /instances/{id}/metrics/backfill endpoint exists and responds
3. Non-AWS instances return 400 with helpful message
4. Response includes hours_backfilled count
5. Self-throttling visible in logs (100ms pauses)
</verification>

<success_criteria>
- BackfillMetrics method collects historical CloudWatch data hour-by-hour
- API endpoint triggers backfill with configurable days parameter
- Self-throttling prevents rate limit errors
- Metrics appear in /metrics/history after backfill
</success_criteria>

<output>
After completion, create `.planning/quick/003-implement-metrics-backfill/003-SUMMARY.md`
</output>
