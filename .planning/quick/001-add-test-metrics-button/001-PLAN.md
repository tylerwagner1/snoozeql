---
phase: quick-001
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - web/src/pages/InstanceDetailPage.tsx
  - web/src/lib/api.ts
  - cmd/server/main.go
autonomous: true

must_haves:
  truths:
    - "User can click 'Test Metrics' button on Instance Details page"
    - "Clicking button triggers immediate metrics collection for that instance"
    - "User sees loading state while metrics are being collected"
    - "User sees success/error feedback after collection completes"
  artifacts:
    - path: "web/src/pages/InstanceDetailPage.tsx"
      provides: "Test Metrics button in Actions section"
      contains: "collectMetrics"
    - path: "web/src/lib/api.ts"
      provides: "collectInstanceMetrics API function"
      exports: ["collectInstanceMetrics"]
    - path: "cmd/server/main.go"
      provides: "POST /instances/:id/collect-metrics endpoint"
      contains: "collect-metrics"
  key_links:
    - from: "InstanceDetailPage.tsx"
      to: "/api/v1/instances/:id/collect-metrics"
      via: "api.collectInstanceMetrics()"
      pattern: "collectInstanceMetrics"
---

<objective>
Add a "Test Metrics" button to the Instance Details page that triggers immediate metrics collection for a single instance.

Purpose: Allow users to manually trigger metrics collection for testing/debugging purposes, especially useful when waiting for the 15-minute interval is impractical.

Output: Working button that calls backend endpoint to collect metrics on-demand.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@web/src/pages/InstanceDetailPage.tsx
@web/src/lib/api.ts
@cmd/server/main.go
@internal/metrics/collector.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add backend endpoint for single-instance metrics collection</name>
  <files>cmd/server/main.go</files>
  <action>
Add POST endpoint `/instances/{id}/collect-metrics` in the instances route group.

Handler logic:
1. Get instance by ID from instanceStore
2. Check if instance is AWS provider (GCP not supported yet)
3. Call metricsCollector.CollectInstance() for that single instance
4. Return JSON: `{success: true, message: "Metrics collected"}` on success
5. Return 400 if instance not found or provider not supported
6. Return 500 if collection fails

Add route after existing instance routes (near line 330):
```go
r.Post("/{id}/collect-metrics", func(w http.ResponseWriter, r *http.Request) {
    // ... handler code
})
```

Note: Need to add a public CollectInstance method to metricsCollector that wraps collectInstance().
  </action>
  <verify>curl -X POST http://localhost:8080/api/v1/instances/{test-id}/collect-metrics -H "Authorization: Bearer dev-key" returns 200 with success message</verify>
  <done>POST /instances/:id/collect-metrics endpoint returns success for valid AWS instance</done>
</task>

<task type="auto">
  <name>Task 2: Add API client method and UI button</name>
  <files>web/src/lib/api.ts, web/src/pages/InstanceDetailPage.tsx</files>
  <action>
**api.ts:** Add collectInstanceMetrics method after getInstanceMetrics (around line 252):
```typescript
collectInstanceMetrics: (instanceId: string) =>
  api.post<{ success: boolean; message: string }>(`/instances/${instanceId}/collect-metrics`),
```

**InstanceDetailPage.tsx:**
1. Add state for collection: `const [collecting, setCollecting] = useState(false)`
2. Add handler function:
```typescript
const handleCollectMetrics = async () => {
  if (!id) return
  setCollecting(true)
  try {
    await api.collectInstanceMetrics(id)
    // Refresh metrics after collection
    const metricsData = await api.getInstanceMetrics(id)
    setMetrics(metricsData)
  } catch (err) {
    console.error('Failed to collect metrics:', err)
  } finally {
    setCollecting(false)
  }
}
```
3. Add button in Actions section (after "Configure Schedule" button, around line 330):
```tsx
<button
  onClick={handleCollectMetrics}
  disabled={collecting || instance.provider !== 'aws'}
  className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
>
  {collecting ? 'Collecting...' : 'Test Metrics'}
</button>
```
Button disabled for non-AWS instances (GCP metrics not implemented).
  </action>
  <verify>Run `npm run build --prefix web` completes without TypeScript errors</verify>
  <done>Test Metrics button visible in Actions, shows loading state, refreshes metrics on success</done>
</task>

<task type="auto">
  <name>Task 3: Expose CollectInstance method on metrics collector</name>
  <files>internal/metrics/collector.go</files>
  <action>
Add public method CollectInstance that allows collecting metrics for a single instance on-demand:

```go
// CollectInstance collects metrics for a single instance on-demand
// This is the public API for manual/triggered collection
func (c *MetricsCollector) CollectInstance(ctx context.Context, instance models.Instance) error {
    // For stopped instances, store zeros
    if instance.Status != "available" && instance.Status != "running" {
        return c.storeZeroMetrics(ctx, instance)
    }
    
    // Only AWS supported for active collection
    if instance.Provider != "aws" {
        return fmt.Errorf("metrics collection not supported for provider: %s", instance.Provider)
    }
    
    client, err := c.getClient(ctx, instance)
    if err != nil {
        return err
    }
    
    return c.collectInstance(ctx, client, instance)
}
```

Add this method around line 128, before the existing collectInstance method.
  </action>
  <verify>go build ./... completes without errors</verify>
  <done>MetricsCollector.CollectInstance() public method available for on-demand collection</done>
</task>

</tasks>

<verification>
1. Backend builds: `go build ./...`
2. Frontend builds: `npm run build --prefix web`
3. Manual test: Start server, navigate to instance details, click "Test Metrics" button
4. Verify metrics refresh after button click completes
</verification>

<success_criteria>
- Test Metrics button appears in Instance Details Actions section
- Button shows "Collecting..." loading state during request
- Button is disabled for non-AWS instances
- Clicking triggers immediate metrics collection
- Metrics section updates after successful collection
- No TypeScript or Go build errors
</success_criteria>

<output>
After completion, create `.planning/quick/001-add-test-metrics-button/001-SUMMARY.md`
</output>
