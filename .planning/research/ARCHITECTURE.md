# Architecture Research: v1.2 Metrics & Recommendations

**Researched:** 2026-02-24
**Focus:** Integration architecture for time-series charts and improved recommendations

## Summary

SnoozeQL v1.2 integrates cleanly with the existing architecture. The metrics system already collects CPU, Connections, Read IOPS, and Write IOPS every 15 minutes. Adding Memory utilization requires only a small extension to the CloudWatch client. Time-series visualization needs a new API endpoint (`GetMetricsByInstance` already exists in store) and a React charting component. Improved threshold detection and grouped recommendations require analyzer logic changes with no schema modifications.

The architecture follows the existing patterns: extend existing components, add minimal new code, leverage the existing `metrics_hourly` table structure.

---

## Integration Points

### Backend Changes

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `internal/models/models.go` | Modify | Add `MetricFreeableMemory` constant for memory metric |
| `internal/metrics/cloudwatch.go` | Modify | Add `FreeableMemory` to `RDSMetrics` struct, fetch in `GetRDSMetrics()` |
| `internal/metrics/collector.go` | Modify | Store memory metric alongside existing metrics |
| `internal/metrics/store.go` | Modify | Update `DeleteOldMetrics` to use 7-day retention instead of 14-day |
| `cmd/server/main.go` | Modify | Add new `/instances/{id}/metrics/history` endpoint |
| `internal/analyzer/patterns.go` | Modify | Update `ActivityThresholds` for CPU < 5% AND connections = 0 logic |
| `internal/analyzer/analyzer.go` | Modify | Implement improved idle detection using compound thresholds |
| `internal/api/handlers/recommendations.go` | Modify | Add grouping logic to `GetAllRecommendations` response |

### Frontend Changes

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `web/src/lib/api.ts` | Modify | Add `getMetricsHistory()` function with time range params |
| `web/src/pages/InstanceDetailPage.tsx` | Modify | Add time-series chart component, toggle between latest/history views |
| `web/src/components/MetricsChart.tsx` | **New** | Recharts line chart for time-series visualization |
| `web/src/components/MetricRangeSelector.tsx` | **New** | Time range selector (24h, 7d) |

### Database Changes

| Change | Description |
|--------|-------------|
| No schema migration needed | `metrics_hourly` table already supports any metric name |
| Retention config change | Update cleanup job from 14-day to 7-day |

---

## Data Flow

### Current Flow (Latest Metrics Only)

```
CloudWatch API                     PostgreSQL                    Frontend
    │                                 │                            │
    │  GetMetricStatistics()          │                            │
    │  (every 15 min)                 │                            │
    │                                 │                            │
    ▼                                 │                            │
┌──────────────────┐                  │                            │
│ MetricsCollector │                  │                            │
│ .CollectAll()    │                  │                            │
└────────┬─────────┘                  │                            │
         │                            │                            │
         │  UpsertHourlyMetric()      │                            │
         │                            │                            │
         ▼                            ▼                            │
    ┌────────────────────────────────────┐                         │
    │         metrics_hourly             │                         │
    │  (id, instance_id, metric_name,    │                         │
    │   hour, avg_value, max_value,      │                         │
    │   min_value, sample_count)         │                         │
    └────────────────────────────────────┘                         │
                        │                                          │
                        │  GetLatestMetrics()                      │
                        │  (DISTINCT ON metric_name)               │
                        ▼                                          │
                  API: GET /instances/{id}/metrics                 │
                        │                                          │
                        └──────────────────────────────────────────►
                                              MetricCard display
```

### v1.2 Flow (Time-Series Charts)

```
CloudWatch API                     PostgreSQL                    Frontend
    │                                 │                            │
    │  GetMetricStatistics()          │                            │
    │  + FreeableMemory (NEW)         │                            │
    │                                 │                            │
    ▼                                 │                            │
┌──────────────────┐                  │                            │
│ MetricsCollector │                  │                            │
│ .CollectAll()    │                  │                            │
│ + memory metric  │                  │                            │
└────────┬─────────┘                  │                            │
         │                            │                            │
         │  UpsertHourlyMetric()      │                            │
         │                            │                            │
         ▼                            ▼                            │
    ┌────────────────────────────────────┐                         │
    │         metrics_hourly             │                         │
    │  (existing schema unchanged)       │                         │
    │  + rows with metric_name =         │                         │
    │    "FreeableMemory"                │                         │
    └────────────────────────────────────┘                         │
         │                  │                                      │
         │ GetLatestMetrics │  GetMetricsByInstance (NEW API)      │
         │ (existing)       │  (with time range: last 7 days)      │
         ▼                  ▼                                      │
   GET /instances/{id}/   GET /instances/{id}/                     │
   metrics                metrics/history?start=&end=              │
         │                  │                                      │
         │                  └──────────────────────────────────────►
         │                              MetricsChart.tsx (Recharts)
         └──────────────────────────────────────────────────────────►
                                              MetricCard (unchanged)
```

### Recommendation Data Flow (Improved Detection)

```
metrics_hourly                    Analyzer                    Frontend
    │                                │                           │
    │                                │                           │
    ▼                                │                           │
┌─────────────────────────────┐      │                           │
│ GetMetricsByInstance()      │      │                           │
│ (last 7 days of data)       │      │                           │
└────────────┬────────────────┘      │                           │
             │                       │                           │
             ▼                       │                           │
       ┌─────────────────────────────┘                           │
       │                                                         │
       ▼                                                         │
┌──────────────────────────────────────┐                         │
│  AnalyzeActivityPattern()            │                         │
│                                      │                         │
│  OLD: CPU < 1%                       │                         │
│  NEW: CPU < 5% AND connections = 0   │                         │
│       (compound threshold)           │                         │
└────────────┬─────────────────────────┘                         │
             │                                                   │
             ▼                                                   │
┌──────────────────────────────────────┐                         │
│  IdleWindow[]                        │                         │
│  - Same structure as before          │                         │
│  - Better detection accuracy         │                         │
└────────────┬─────────────────────────┘                         │
             │                                                   │
             ▼                                                   │
       recommendations table                                     │
             │                                                   │
             │  GetAllRecommendations()                          │
             │  + grouping by detected_pattern                   │
             ▼                                                   │
       API response with groups ─────────────────────────────────►
                                      Grouped UI display
```

---

## Component Details

### 1. Memory Metric Collection (Backend)

**File:** `internal/metrics/cloudwatch.go`

Add to `RDSMetrics` struct:
```go
type RDSMetrics struct {
    InstanceID  string
    Timestamp   time.Time
    CPU         *MetricValue
    Connections *MetricValue
    ReadIOPS    *MetricValue
    WriteIOPS   *MetricValue
    Memory      *MetricValue  // NEW: FreeableMemory
}
```

Add fetch call in `GetRDSMetrics()`:
```go
memory, err := c.getMetricWithRetry(ctx, dbInstanceID, "FreeableMemory", startTime, endTime)
if err == nil {
    metrics.Memory = memory
}
```

**File:** `internal/models/models.go`

Add constant:
```go
const MetricFreeableMemory = "FreeableMemory"
```

**File:** `internal/metrics/collector.go`

Add storage in `collectInstance()`:
```go
if metrics.Memory != nil {
    if err := c.storeMetric(ctx, instance.ID, models.MetricFreeableMemory, metrics.Timestamp, metrics.Memory); err != nil {
        log.Printf("Failed to store Memory metric for %s: %v", instance.Name, err)
    }
}
```

### 2. Time-Series API Endpoint (Backend)

**File:** `cmd/server/main.go`

Add new endpoint:
```go
r.Get("/instances/{id}/metrics/history", func(w http.ResponseWriter, r *http.Request) {
    instanceID := chi.URLParam(r, "id")
    
    // Parse query params
    startStr := r.URL.Query().Get("start")
    endStr := r.URL.Query().Get("end")
    
    // Default to last 24 hours
    end := time.Now().UTC()
    start := end.Add(-24 * time.Hour)
    
    if startStr != "" {
        if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
            start = parsed
        }
    }
    if endStr != "" {
        if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
            end = parsed
        }
    }
    
    // Cap at 7 days (retention period)
    maxStart := end.Add(-7 * 24 * time.Hour)
    if start.Before(maxStart) {
        start = maxStart
    }
    
    metrics, err := metricsStore.GetMetricsByInstance(ctx, instanceID, start, end)
    if err != nil {
        // error handling
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
})
```

### 3. Retention Update (Backend)

**File:** `internal/metrics/store.go`

The `DeleteOldMetrics` method already exists. The change is in the cleanup job call (in main.go or a background job):

```go
// Change from 14 days to 7 days
retentionDate := time.Now().UTC().Add(-7 * 24 * time.Hour)
deleted, err := metricsStore.DeleteOldMetrics(ctx, retentionDate)
```

### 4. Improved Threshold Detection (Backend)

**File:** `internal/analyzer/patterns.go`

Update `ActivityThresholds`:
```go
type ActivityThresholds struct {
    CPUPercent           float64 // CPU < 5% (changed from 1%)
    ConnectionsThreshold float64 // connections = 0 (NEW)
    MinIdleHours         int
    MinDataHours         int
    MinDaysConsistent    int
}

func DefaultThresholds() ActivityThresholds {
    return ActivityThresholds{
        CPUPercent:           5.0,  // Changed from 1.0
        ConnectionsThreshold: 0.0,  // NEW: require zero connections
        MinIdleHours:         8,
        MinDataHours:         24,
        MinDaysConsistent:    3,
    }
}
```

Update idle detection in `findIdleSegments()`:
```go
// Check if this hour is "idle" per v1.2 thresholds
// COMPOUND CONDITION: CPU < 5% AND connections = 0
isIdle := false
if bucket != nil && len(bucket.CPUValues) > 0 {
    cpu = average(bucket.CPUValues)
    if len(bucket.ConnValues) > 0 {
        conns = average(bucket.ConnValues)
    }
    
    // NEW: Compound threshold check
    isIdle = cpu < thresholds.CPUPercent && conns <= thresholds.ConnectionsThreshold
}
```

### 5. Grouped Recommendations (Backend)

**File:** `internal/api/handlers/recommendations.go`

Add grouping to `GetAllRecommendations()`:

```go
// After enriching all recommendations, group by pattern
type groupedResponse struct {
    Groups []recommendationGroup `json:"groups"`
}

type recommendationGroup struct {
    Pattern         string         `json:"pattern"`        // e.g., "weeknight_idle"
    Count           int            `json:"count"`
    TotalSavings    float64        `json:"total_savings"`
    Recommendations []enrichedRec  `json:"recommendations"`
}

// Group recommendations by similar patterns
groups := make(map[string][]enrichedRec)
for _, rec := range enriched {
    // Extract pattern key from detected_pattern
    patternKey := extractPatternKey(rec.DetectedPattern)
    groups[patternKey] = append(groups[patternKey], rec)
}

// Convert to response format
var response groupedResponse
for pattern, recs := range groups {
    totalSavings := 0.0
    for _, r := range recs {
        totalSavings += r.EstimatedDailySavings
    }
    response.Groups = append(response.Groups, recommendationGroup{
        Pattern:         pattern,
        Count:           len(recs),
        TotalSavings:    totalSavings,
        Recommendations: recs,
    })
}
```

### 6. Frontend Chart Component (New)

**File:** `web/src/components/MetricsChart.tsx`

```tsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import type { HourlyMetric } from '../lib/api';

interface MetricsChartProps {
    metrics: HourlyMetric[];
    metricName: string;
    unit: string;
    color: string;
}

const MetricsChart = ({ metrics, metricName, unit, color }: MetricsChartProps) => {
    // Filter and transform data for this metric
    const data = metrics
        .filter(m => m.metric_name.toLowerCase() === metricName.toLowerCase())
        .map(m => ({
            hour: new Date(m.hour).toLocaleString(),
            value: m.avg_value,
            min: m.min_value,
            max: m.max_value,
        }));
    
    return (
        <ResponsiveContainer width="100%" height={200}>
            <LineChart data={data}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="hour" tick={{ fontSize: 10 }} />
                <YAxis unit={unit} />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="value" stroke={color} name={metricName} />
            </LineChart>
        </ResponsiveContainer>
    );
};

export default MetricsChart;
```

### 7. Frontend API Extension

**File:** `web/src/lib/api.ts`

Add new function:
```typescript
getMetricsHistory: (instanceId: string, start?: string, end?: string) => {
    const params = new URLSearchParams();
    if (start) params.set('start', start);
    if (end) params.set('end', end);
    const query = params.toString();
    return api.get<HourlyMetric[]>(`/instances/${instanceId}/metrics/history${query ? `?${query}` : ''}`);
}
```

---

## Suggested Build Order

1. **Memory metric collection** - Extend CloudWatch client and collector
   - Why first: Foundation for all other changes, minimal risk
   - Dependencies: None
   - Testing: Can verify metrics appear in metrics_hourly table

2. **Time-series API endpoint** - Add `/instances/{id}/metrics/history`
   - Why second: Backend must exist before frontend
   - Dependencies: Step 1 (want memory in history)
   - Testing: Can test with curl before UI exists

3. **Frontend MetricsChart component** - Create Recharts wrapper
   - Why third: Can develop against new API
   - Dependencies: Step 2
   - Testing: Visual testing on Instance Detail page

4. **Retention update** - Change from 14-day to 7-day
   - Why fourth: Non-breaking, config change
   - Dependencies: None (but do after chart works to avoid data loss during dev)
   - Testing: Verify cleanup job deletes old data

5. **Improved threshold detection** - CPU < 5% AND connections = 0
   - Why fifth: Isolated change to analyzer
   - Dependencies: None
   - Testing: Unit tests, then manual with real data

6. **Grouped recommendations** - Update API response format
   - Why sixth: After detection improvements so groups reflect new thresholds
   - Dependencies: Step 5
   - Testing: API response inspection, then UI update

---

## Confidence

**HIGH** - All integration points verified against existing codebase:
- `GetMetricsByInstance()` already exists in MetricsStore
- `metrics_hourly` table supports arbitrary metric names
- CloudWatch `FreeableMemory` is a standard RDS metric
- Analyzer pattern detection is well-isolated for modification
- RecommendationHandler already enriches responses, adding grouping is incremental

**Risk areas (LOW):**
- Recharts dependency: Standard library, well-documented
- API response format change: Grouped recommendations changes response shape; may need frontend compatibility check
