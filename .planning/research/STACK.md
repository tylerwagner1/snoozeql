# Stack Research: v1.2 Metrics & Recommendations

**Researched:** 2026-02-24
**Focus:** Stack additions for metrics visualization and recommendation improvements

## Summary

The v1.2 milestone requires minimal stack changes. The existing infrastructure—Go 1.24.0 with AWS SDK v2, React 18.2 with Recharts 2.10.0, and PostgreSQL with the `metrics_hourly` table—already supports the core requirements. The primary work is **adding one CloudWatch metric (FreeableMemory), building time-series chart components, and improving threshold-based idle detection**. No new libraries required.

---

## Existing Stack (No Changes Needed)

### Backend - Fully Sufficient

| Component | Version | Current Capability | v1.2 Need |
|-----------|---------|-------------------|-----------|
| **Go** | 1.24.0 | All backend logic | No change |
| **aws-sdk-go-v2/cloudwatch** | 1.54.0 | `GetMetricStatistics` for CPU, Connections, IOPS | Add FreeableMemory metric |
| **pgx/v5** | 5.8.0 | Queries `metrics_hourly` table | Add time-range queries |
| **Chi router** | 5.2.0 | REST API routing | Add metrics endpoint |

**Key finding:** The `MetricsCollector` in `internal/metrics/collector.go` already collects CPUUtilization, DatabaseConnections, ReadIOPS, and WriteIOPS. Adding FreeableMemory is a single-line addition to the `RDSMetrics` struct and `GetRDSMetrics` method.

### Frontend - Fully Sufficient

| Component | Version | Current Capability | v1.2 Need |
|-----------|---------|-------------------|-----------|
| **React** | 18.2.0 | Component framework | No change |
| **Recharts** | 2.10.0 | `AreaChart`, `LineChart`, `XAxis`, `Tooltip` | Use existing components for time-series |
| **Tailwind CSS** | 3.4.0 | Styling | No change |
| **lucide-react** | 0.300.0 | Icons | No change |

**Key finding:** `ActivityGraph.tsx` already demonstrates time-series visualization with Recharts. The `InstanceDetailPage.tsx` displays metric cards but lacks time-series charts—this is the primary frontend addition.

### Database - Fully Sufficient

| Table | Current Schema | v1.2 Need |
|-------|---------------|-----------|
| `metrics_hourly` | `instance_id`, `metric_name`, `hour`, `avg_value`, `max_value`, `min_value`, `sample_count` | Add `FreeableMemory` as new metric_name |

**Key finding:** The schema is already designed for multiple metric types. Adding FreeableMemory requires no schema changes—just storing it with `metric_name = 'FreeableMemory'`.

---

## Stack Additions

### Backend: Add FreeableMemory Metric Collection

**Change Location:** `internal/metrics/cloudwatch.go` and `internal/models/models.go`

| Component | Change | Rationale |
|-----------|--------|-----------|
| `RDSMetrics` struct | Add `FreeableMemory *MetricValue` field | Store memory utilization data |
| `GetRDSMetrics()` | Add `FreeableMemory` fetch call | Collect from CloudWatch |
| `models.go` | Add `MetricFreeableMemory = "FreeableMemory"` constant | Consistent naming |

**AWS CloudWatch FreeableMemory Details (Verified from official docs):**
- **Namespace:** `AWS/RDS`
- **Metric Name:** `FreeableMemory`
- **Unit:** Bytes
- **Applies to:** All RDS engines
- **Description:** Amount of available random access memory. Reports `/proc/meminfo` MemAvailable for MariaDB, MySQL, Oracle, PostgreSQL.

**Implementation pattern (mirrors existing CPU collection):**
```go
// In cloudwatch.go GetRDSMetrics()
freeMem, err := c.getMetricWithRetry(ctx, dbInstanceID, models.MetricFreeableMemory, startTime, endTime)
if err == nil {
    metrics.FreeableMemory = freeMem
}
```

**Memory Utilization Calculation:**
FreeableMemory is in bytes, not percentage. To display as utilization:
```
Memory Utilization % = 100 - (FreeableMemory / TotalMemory * 100)
```
Note: TotalMemory requires instance class lookup. For v1.2, display FreeableMemory directly (in GB/MB) or calculate percentage client-side if instance memory is known.

### Backend: Time-Series Metrics API

**New Endpoint:** `GET /api/v1/instances/{id}/metrics/history`

| Parameter | Type | Description |
|-----------|------|-------------|
| `start` | ISO8601 | Start of time range (default: 7 days ago) |
| `end` | ISO8601 | End of time range (default: now) |
| `metrics` | string | Comma-separated list: `cpu,memory,connections,iops` |

**Response Format:**
```json
{
  "instance_id": "abc123",
  "start": "2026-02-17T00:00:00Z",
  "end": "2026-02-24T00:00:00Z",
  "series": {
    "CPUUtilization": [
      {"hour": "2026-02-17T00:00:00Z", "avg": 2.3, "max": 5.1, "min": 0.8}
    ],
    "FreeableMemory": [...]
  }
}
```

**Implementation:** Add handler in `internal/api/handlers/instances.go` that queries `MetricsStore.GetMetricsByInstance()`.

### Frontend: Time-Series Chart Component

**New Component:** `web/src/components/MetricsChart.tsx`

Uses existing Recharts components already imported in the codebase:

```tsx
import { 
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid 
} from 'recharts'

interface MetricsChartProps {
  data: Array<{hour: string, value: number}>
  metricName: string
  color: string
  unit: string
}
```

**Integration Point:** `InstanceDetailPage.tsx` currently shows static metric cards. Add collapsible chart section below each card or tabbed view for different metrics.

### Backend: Improved Idle Detection Thresholds

**Current Implementation (patterns.go):**
```go
type ActivityThresholds struct {
    CPUPercent:        1.0,  // CPU < 1%
    QueriesPerMin:     5.0,  // Queries < 5/min
    MinIdleHours:      8,    // 8+ hours of low activity
    MinDataHours:      24,   // 24+ hours of data required
    MinDaysConsistent: 3,    // Pattern on at least 3 days
}
```

**v1.2 Enhancement:** Add memory threshold to idle detection

| Threshold | Current | v1.2 Addition |
|-----------|---------|---------------|
| CPU | < 1% | No change |
| Memory | N/A | FreeableMemory > 90% of total (very idle) |
| Connections | Indirect via QueriesPerMin | < 2 active connections |
| IOPS | Not used in idle calc | < 10 combined ReadIOPS+WriteIOPS |

**Implementation:**
```go
// patterns.go
type ActivityThresholds struct {
    CPUPercent         float64 // CPU < X%
    FreeableMemoryPct  float64 // FreeableMemory > X% of total = idle
    MaxConnections     float64 // Connections < X = idle
    MaxCombinedIOPS    float64 // ReadIOPS + WriteIOPS < X = idle
    MinIdleHours       int
    MinDataHours       int
    MinDaysConsistent  int
}
```

---

## What NOT to Add

| Component | Reason |
|-----------|--------|
| **date-fns** | Use native Date APIs or existing formatters for chart timestamps |
| **moment.js** | Heavy library, unnecessary for simple time formatting |
| **Chart.js** | Already have Recharts, don't mix charting libraries |
| **Recharts upgrade to 3.x** | v2.10.0 is stable and sufficient; major version bump adds risk |
| **Enhanced Monitoring (RDS)** | Would provide more OS-level metrics but requires enabling per-instance, adds cost, overkill for POC |
| **CloudWatch Logs Insights** | For log analysis, not metrics visualization |
| **AWS Cost Explorer API** | Out of scope for metrics milestone |
| **TanStack Query (React Query)** | Would improve caching but adds learning curve; existing fetch patterns work |
| **D3.js directly** | Recharts wraps D3; don't need direct D3 usage |

---

## Confidence

**HIGH** - All additions are incremental extensions of existing patterns:

| Area | Confidence | Reasoning |
|------|------------|-----------|
| FreeableMemory collection | HIGH | Same pattern as existing CPU/Connections collection; verified in AWS docs |
| Time-series API | HIGH | `GetMetricsByInstance()` already exists in MetricsStore |
| Recharts time-series | HIGH | `ActivityGraph.tsx` proves the pattern works |
| Threshold improvements | HIGH | `patterns.go` already has threshold config struct |

**Verification sources:**
- AWS RDS CloudWatch Metrics official docs (fetched 2026-02-24)
- Recharts GitHub repository (v3.7.0 latest, but v2.10.0 in project is stable)
- Codebase review: `internal/metrics/`, `web/src/components/ActivityGraph.tsx`

---

## Implementation Checklist

### Backend (Go)
- [ ] Add `MetricFreeableMemory` constant to `internal/models/models.go`
- [ ] Add `FreeableMemory *MetricValue` to `RDSMetrics` struct in `cloudwatch.go`
- [ ] Fetch FreeableMemory in `GetRDSMetrics()` method
- [ ] Store FreeableMemory in `collectInstance()` method
- [ ] Add `GET /api/v1/instances/{id}/metrics/history` endpoint
- [ ] Add memory threshold to `ActivityThresholds` struct
- [ ] Update `findIdleSegments()` to consider memory in idle detection

### Frontend (React/TypeScript)
- [ ] Add `getInstanceMetricsHistory()` to `web/src/lib/api.ts`
- [ ] Create `MetricsChart.tsx` component using Recharts LineChart
- [ ] Add chart section to `InstanceDetailPage.tsx`
- [ ] Add time range selector (24h, 7d, 14d)
- [ ] Add metric toggle buttons (CPU, Memory, Connections, IOPS)

### No New Dependencies Required
The existing `package.json` and `go.mod` are sufficient for all v1.2 work.
