# Phase 11: Time-Series Visualization - Research

**Researched:** 2026-02-25
**Domain:** React charting with Recharts, Go REST API
**Confidence:** HIGH

## Summary

Phase 11 implements time-series charts for CPU, Memory, and Connections metrics on the Instance Details page. The codebase already has **Recharts 2.15.4** installed and working (see `CostOverTimeChart.tsx`), making this a straightforward extension of existing patterns.

The key components needed are:
1. A new API endpoint `GET /api/v1/instances/{id}/metrics/history?range=24h` returning hourly metrics
2. A `MetricsChart` React component with tabs for CPU/Memory/Connections
3. Integration into the existing `InstanceDetailPage.tsx`

**Primary recommendation:** Use Recharts `LineChart` with `ReferenceLine` at y=0, fixed domain for percentage metrics, and `Tooltip` with crosshair cursor. Follow the exact pattern established in `CostOverTimeChart.tsx`.

## Standard Stack

The established libraries/tools for this domain:

### Core (Already Installed)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| recharts | 2.15.4 | Time-series line charts | Already in use (CostOverTimeChart), React-native, composable |
| @tanstack/react-query | - | Data fetching | Already used throughout app for API calls |
| tailwindcss | 3.4.0 | Styling | Already in use, matches existing UI |

### Supporting (Already Available)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| lucide-react | 0.300.0 | Icons for tabs/loading | Already installed, use for tab icons |
| clsx | 2.0.0 | Conditional classes | Already installed, use for tab active states |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Recharts | Chart.js/react-chartjs-2 | Would require new dependency, different API |
| Recharts | visx | More flexible but steeper learning curve |
| Custom tabs | @headlessui/react Tab | Already have @headlessui/react but simpler custom tabs suffice for 3 options |

**Installation:** None required - all dependencies already installed.

## Architecture Patterns

### Recommended Project Structure
```
web/src/
├── components/
│   └── MetricsChart.tsx          # New: Reusable metrics chart with tabs
├── pages/
│   └── InstanceDetailPage.tsx    # Modified: Add MetricsChart section
└── lib/
    └── api.ts                    # Modified: Add getMetricsHistory()
```

### Pattern 1: Metrics Chart Component Structure
**What:** Single component handling tabs, time range, loading, and chart rendering
**When to use:** For the Instance Details page metrics visualization
**Example:**
```typescript
// Source: Existing CostOverTimeChart.tsx pattern + CONTEXT.md decisions
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'

interface MetricsChartProps {
  instanceId: string
}

type MetricTab = 'cpu' | 'memory' | 'connections'
type TimeRange = '1h' | '6h' | '24h' | '7d'

export function MetricsChart({ instanceId }: MetricsChartProps) {
  const [activeTab, setActiveTab] = useState<MetricTab>('cpu')  // Default: CPU
  const [timeRange, setTimeRange] = useState<TimeRange>('24h')  // Default: 24h
  
  const { data, isLoading, error } = useQuery({
    queryKey: ['metrics-history', instanceId, timeRange],
    queryFn: () => api.getMetricsHistory(instanceId, timeRange),
  })
  
  // Filter data for active metric
  const chartData = useMemo(() => {
    if (!data) return []
    return data
      .filter(m => m.metric_name === getMetricName(activeTab))
      .map(m => ({ hour: m.hour, value: m.avg_value }))
  }, [data, activeTab])
  
  // ... render tabs, time range selector, chart
}
```

### Pattern 2: API Data Format
**What:** Backend returns flat array of HourlyMetric objects, frontend filters by metric_name
**When to use:** Avoids multiple API calls, single fetch returns all metrics for time range
**Example:**
```typescript
// API Response format
interface MetricsHistoryResponse {
  metrics: HourlyMetric[]  // All metrics (CPU, Memory, Connections) for time range
}

// Frontend filters by metric_name
const cpuData = metrics.filter(m => m.metric_name === 'CPUUtilization')
const memoryData = metrics.filter(m => m.metric_name === 'FreeableMemory')
const connectionsData = metrics.filter(m => m.metric_name === 'DatabaseConnections')
```

### Pattern 3: Recharts Line Configuration (from CONTEXT.md)
**What:** Line chart with crosshair tooltip, fixed Y-axis for percentages
**When to use:** For all three metrics charts
**Example:**
```typescript
// Source: CONTEXT.md decisions + Recharts docs
<LineChart data={chartData}>
  <CartesianGrid strokeDasharray="3 3" vertical={false} />
  <XAxis 
    dataKey="hour" 
    tickFormatter={(hour) => formatXAxis(hour, timeRange)}
  />
  <YAxis 
    domain={isPercentage ? [0, 100] : ['auto', 'auto']}  // Fixed 0-100 for CPU/Memory
    tickFormatter={(v) => isPercentage ? `${v}%` : v}
  />
  <Tooltip 
    cursor={{ strokeDasharray: '3 3' }}  // Crosshair effect
    content={<CustomTooltip />}
  />
  <Line 
    type="monotone" 
    dataKey="value" 
    stroke="#06b6d4" 
    dot={false}  // No dots per CONTEXT.md
    strokeWidth={2}
  />
</LineChart>
```

### Anti-Patterns to Avoid
- **Multiple API calls per metric:** Don't fetch CPU, Memory, Connections separately - fetch once and filter
- **Area charts for metrics:** User decided line charts, not area
- **Dots on data points:** Explicitly excluded in CONTEXT.md
- **Auto-scaling Y-axis for percentages:** CPU and Memory should always be 0-100%
- **Breaking lines for gaps:** Show zero values for sleeping instances, not gaps

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Tab component | Custom div-based tabs | Simple button group pattern from CostOverTimeChart | Already proven pattern in codebase |
| Time formatting | Custom date logic | `toLocaleTimeString()` with options | Browser handles timezone correctly |
| Loading spinner | Custom CSS animation | Existing Tailwind `animate-spin` | Consistent with rest of app |
| Responsive chart sizing | Manual width/height | Recharts `ResponsiveContainer` | Handles resize events automatically |

**Key insight:** The codebase already has all the patterns needed in `CostOverTimeChart.tsx`. Copy that structure, adapt for line charts.

## Common Pitfalls

### Pitfall 1: Timezone Handling
**What goes wrong:** X-axis shows wrong times because backend stores UTC, frontend displays local
**Why it happens:** `metrics_hourly.hour` is stored as `TIMESTAMPTZ` in UTC
**How to avoid:** 
- Backend returns ISO 8601 strings (e.g., `2026-02-25T14:00:00Z`)
- Frontend uses `new Date(hour).toLocaleTimeString()` for display
- Never manually parse/format - let JavaScript Date handle it
**Warning signs:** Times look shifted by your timezone offset

### Pitfall 2: Empty Time Range Handling  
**What goes wrong:** Chart shows nothing for new instances with no data
**Why it happens:** Query returns empty array, chart renders blank
**How to avoid:**
- Per CONTEXT.md: "Show chart axes with 'No data available' message"
- Render chart grid/axes even with empty data, overlay message
**Warning signs:** Completely blank chart area instead of empty state

### Pitfall 3: Missing Data Points (Gaps)
**What goes wrong:** Line has gaps where instance was sleeping
**Why it happens:** Sleeping instances have no CloudWatch data
**How to avoid:**
- Per CONTEXT.md: "Show as zero (line drops to zero)"
- Backend should NOT return missing hours, but rather zero values
- Frontend fills gaps with zero if backend doesn't
**Warning signs:** Disconnected line segments

### Pitfall 4: Y-Axis Auto-scaling Confusion
**What goes wrong:** CPU chart with 5% max value looks "high" because Y-axis is 0-5
**Why it happens:** Recharts auto-scales by default
**How to avoid:**
- CPU and Memory: `domain={[0, 100]}` (fixed)
- Connections: `domain={[0, 'auto']}` (auto-scale, but min is 0)
**Warning signs:** Misleading visual representation of low values

### Pitfall 5: Re-fetching on Tab Change
**What goes wrong:** Network request every time user switches CPU→Memory→Connections
**Why it happens:** Separate queries per metric
**How to avoid:**
- Single query fetches all metrics for time range
- Filter client-side by `metric_name`
- Cache with React Query
**Warning signs:** Loading spinner on every tab change

## Code Examples

Verified patterns from the codebase:

### Time Range Button Group (from CostOverTimeChart.tsx)
```typescript
// Source: web/src/components/CostOverTimeChart.tsx lines 130-151
const timeRanges: TimeRange[] = ['1h', '6h', '24h', '7d']

<div className="flex bg-gray-100 rounded-lg p-1">
  {timeRanges.map((range) => (
    <button
      key={range}
      onClick={() => setTimeRange(range)}
      className={`px-3 py-1 text-sm rounded-md transition-colors ${
        timeRange === range
          ? 'bg-white text-gray-900 shadow-sm'
          : 'text-gray-600 hover:text-gray-900'
      }`}
    >
      {range}
    </button>
  ))}
</div>
```

### Tab Component for Metrics
```typescript
// Pattern based on CONTEXT.md decisions
const tabs = [
  { key: 'cpu', label: 'CPU' },
  { key: 'memory', label: 'Memory' },
  { key: 'connections', label: 'Connections' },
] as const

<div className="flex border-b border-gray-200">
  {tabs.map((tab) => (
    <button
      key={tab.key}
      onClick={() => setActiveTab(tab.key)}
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        activeTab === tab.key
          ? 'border-blue-500 text-blue-600'
          : 'border-transparent text-gray-500 hover:text-gray-700'
      }`}
    >
      {tab.label}
    </button>
  ))}
</div>
```

### Loading State (from existing patterns)
```typescript
// Loading spinner centered in chart area (per CONTEXT.md)
{isLoading && (
  <div className="h-[150px] flex items-center justify-center">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
  </div>
)}
```

### Empty State with Axes
```typescript
// Show axes with message (per CONTEXT.md)
{!isLoading && chartData.length === 0 && (
  <div className="h-[150px] relative">
    <ResponsiveContainer width="100%" height="100%">
      <LineChart data={[{ hour: '', value: 0 }]}>
        <XAxis dataKey="hour" />
        <YAxis domain={[0, 100]} />
        <CartesianGrid strokeDasharray="3 3" />
      </LineChart>
    </ResponsiveContainer>
    <div className="absolute inset-0 flex items-center justify-center">
      <p className="text-gray-500 text-sm">No data available</p>
    </div>
  </div>
)}
```

### Backend Endpoint Structure
```go
// Source: Existing pattern from cmd/server/main.go lines 652-670
r.Get("/instances/{id}/metrics/history", func(w http.ResponseWriter, r *http.Request) {
    instanceID := chi.URLParam(r, "id")
    rangeParam := r.URL.Query().Get("range")  // 1h, 6h, 24h, 7d
    
    // Calculate time range
    var duration time.Duration
    switch rangeParam {
    case "1h":
        duration = time.Hour
    case "6h":
        duration = 6 * time.Hour
    case "7d":
        duration = 7 * 24 * time.Hour
    default:
        duration = 24 * time.Hour  // Default 24h
    }
    
    start := time.Now().Add(-duration)
    end := time.Now()
    
    metrics, err := metricsStore.GetMetricsByInstance(ctx, instanceID, start, end)
    // ... return JSON
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Class components | Functional + hooks | React 16.8 (2019) | All new code uses hooks |
| fetch + useEffect | TanStack Query | Current standard | Caching, loading states handled |
| Manual chart sizing | ResponsiveContainer | Recharts standard | Automatic resize handling |

**Deprecated/outdated:**
- None relevant - Recharts 2.x is current, patterns align with React 18

## Open Questions

Things that couldn't be fully resolved:

1. **X-axis label density**
   - What we know: Need to show timestamps on x-axis
   - What's unclear: How many labels to show for each time range (every hour? every 6 hours for 7d?)
   - Recommendation: Use Recharts `interval` prop to auto-calculate, or fixed: 1h=every 15min, 6h=every hour, 24h=every 4h, 7d=every day

2. **Memory metric name mapping**
   - What we know: Backend stores as "FreeableMemory" (CloudWatch name)
   - What's unclear: Whether to display as "Memory" or "Memory Available"
   - Recommendation: Display as "Memory" in tab (per CONTEXT.md), show "Memory Available %" in tooltip

## Sources

### Primary (HIGH confidence)
- `/Users/tylerwagner/snoozeql/web/package.json` - Confirmed Recharts 2.15.4 installed
- `/Users/tylerwagner/snoozeql/web/src/components/CostOverTimeChart.tsx` - Existing Recharts pattern
- `/Users/tylerwagner/snoozeql/internal/metrics/store.go` - GetMetricsByInstance() already exists
- `/Users/tylerwagner/snoozeql/deployments/docker/migrations/005_metrics_hourly.sql` - DB schema verified
- `/Users/tylerwagner/snoozeql/.planning/phases/11-time-series-visualization/11-CONTEXT.md` - User decisions

### Secondary (MEDIUM confidence)
- Recharts documentation (verified against installed version)
- Existing codebase patterns

### Tertiary (LOW confidence)
- None - all findings verified against codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already installed and verified
- Architecture: HIGH - Following existing patterns in codebase
- Pitfalls: HIGH - Based on CONTEXT.md decisions and common Recharts issues

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (30 days - stable stack)
