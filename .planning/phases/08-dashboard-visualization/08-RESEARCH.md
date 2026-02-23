# Phase 8: Dashboard & Visualization - Research

**Researched:** 2026-02-23
**Domain:** React Dashboard, Data Visualization, Time-Series Charts
**Confidence:** HIGH

## Summary

Phase 8 builds the savings dashboard visualization UI, consuming the Phase 7 backend APIs to display time-series savings charts, per-instance attribution tables, and cost projection comparisons. The existing codebase already has a solid foundation with React 18, Recharts 2.x, Tailwind CSS 3.x, Headless UI 2.x, and Lucide React icons.

The dashboard will extend the existing `Dashboard.tsx` page (or create a dedicated `SavingsDashboardPage.tsx`) with four key components: summary cards with period-over-period trends, a time-series savings chart with configurable date ranges (7d, 30d, 90d, custom), a per-instance savings attribution table sorted by contribution, and an actual vs projected cost comparison section with disclaimers.

**Primary recommendation:** Use the existing Recharts AreaChart/LineChart patterns from `ActivityGraph.tsx` for time-series, add a simple tab-based date range selector using native HTML select or Headless UI Listbox, and follow the existing card styling patterns from Dashboard.tsx.

## Standard Stack

The established libraries/tools for this domain:

### Core (Already Installed)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| recharts | 2.15.4 | Data visualization charts | Already used in codebase, React-first charting |
| react | 18.2.0 | UI framework | Application foundation |
| tailwindcss | 3.4.0 | Styling | Consistent with rest of app |
| @headlessui/react | 2.2.9 | Accessible UI primitives | Already used for dialogs, modals |
| lucide-react | 0.300.0 | Icons | Consistent icon set |

### Supporting (No New Dependencies Needed)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| clsx | 2.0.0 | Class name composition | Already installed for conditional classes |
| tailwind-merge | 2.0.0 | Merge Tailwind classes | Already installed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Recharts | Chart.js / Visx / Nivo | Would require new dependency, learning curve. Recharts already works well |
| Native date range select | react-datepicker | Overkill for preset ranges. Only needed if true custom date picker required |
| Native number formatting | Intl.NumberFormat | Built-in browser API, no dependency |

**No new npm packages required for Phase 8.**

## Architecture Patterns

### Recommended Project Structure
```
web/src/
├── pages/
│   ├── SavingsPage.tsx              # Main savings dashboard page
│   └── Dashboard.tsx                # Existing (may add summary card link)
├── components/
│   ├── savings/                     # New folder for savings components
│   │   ├── SavingsSummaryCards.tsx  # Summary stat cards
│   │   ├── SavingsChart.tsx         # Time-series line/area chart
│   │   ├── InstanceSavingsTable.tsx # Per-instance attribution table
│   │   ├── CostProjection.tsx       # Actual vs projected comparison
│   │   └── DateRangeSelector.tsx    # 7d/30d/90d/custom tabs
│   ├── ActivityGraph.tsx            # Existing - reference for Recharts patterns
│   └── ...existing components
├── lib/
│   ├── api.ts                       # Add savings API methods
│   └── formatters.ts                # NEW: Money/date formatters (optional)
└── hooks/
    └── useSavings.ts                # NEW: Data fetching hook (optional)
```

### Pattern 1: Page-Level Data Fetching with useEffect
**What:** Fetch data at page level, pass to child components
**When to use:** Standard pattern for dashboard pages (matches existing Dashboard.tsx)
**Example:**
```typescript
// Source: Existing pattern from Dashboard.tsx
const SavingsPage = () => {
  const [summary, setSummary] = useState<SavingsSummary | null>(null)
  const [dailySavings, setDailySavings] = useState<DailySavings[]>([])
  const [dateRange, setDateRange] = useState<'7d' | '30d' | '90d'>('30d')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const days = dateRange === '7d' ? 7 : dateRange === '30d' ? 30 : 90
        const [summaryData, dailyData, byInstanceData] = await Promise.all([
          api.getSavingsSummary(days),
          api.getDailySavings(days),
          api.getSavingsByInstance(days)
        ])
        setSummary(summaryData)
        setDailySavings(dailyData.daily_savings)
        // ... etc
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [dateRange])

  return (
    <div className="space-y-6">
      <DateRangeSelector value={dateRange} onChange={setDateRange} />
      <SavingsSummaryCards data={summary} loading={loading} />
      <SavingsChart data={dailySavings} loading={loading} />
      <InstanceSavingsTable data={byInstance} loading={loading} />
      <CostProjection data={summary} loading={loading} />
    </div>
  )
}
```

### Pattern 2: Recharts ResponsiveContainer with AreaChart
**What:** Use ResponsiveContainer for responsive charts
**When to use:** All time-series visualizations
**Example:**
```typescript
// Source: Derived from ActivityGraph.tsx pattern
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'

interface SavingsChartProps {
  data: Array<{ date: string; savings_cents: number }>
}

export function SavingsChart({ data }: SavingsChartProps) {
  // Transform cents to dollars for display
  const chartData = data.map(d => ({
    date: d.date,
    savings: d.savings_cents / 100
  }))

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Savings Over Time</h2>
      <div className="h-64">
        <ResponsiveContainer>
          <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="savingsGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} stroke="#334155" strokeOpacity={0.2} />
            <XAxis
              dataKey="date"
              tick={{ fill: '#94a3b8', fontSize: 10 }}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              tick={{ fill: '#94a3b8', fontSize: 10 }}
              tickLine={false}
              axisLine={false}
              tickFormatter={(v) => `$${v}`}
            />
            <Tooltip
              contentStyle={{
                background: '#1e293b',
                border: '1px solid #334155',
                borderRadius: '0.5rem',
              }}
              formatter={(value: number) => [`$${value.toFixed(2)}`, 'Savings']}
            />
            <Area
              type="monotone"
              dataKey="savings"
              stroke="#10b981"
              fill="url(#savingsGradient)"
              strokeWidth={2}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}
```

### Pattern 3: Summary Cards with Gradient Icons
**What:** Stat cards matching existing Dashboard.tsx styling
**When to use:** Summary statistics display
**Example:**
```typescript
// Source: Existing Dashboard.tsx summary cards
<div className="bg-slate-800/50 rounded-xl p-5 shadow-lg border border-slate-700 hover:border-green-500/50 transition-all group">
  <div className="flex items-center justify-between mb-3">
    <p className="text-sm text-slate-400 font-medium">Total Savings</p>
    <div className="p-2 bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg group-hover:scale-105 transition-transform shadow-lg shadow-green-500/20">
      <TrendingDown className="h-5 w-5 text-white" />
    </div>
  </div>
  <p className="text-3xl font-bold text-white mb-1">{formatCurrency(totalSavingsCents)}</p>
  <p className="text-sm text-green-400">+{percentChange}% vs previous period</p>
</div>
```

### Pattern 4: Date Range Selector (Tab Style)
**What:** Simple tab-based selector for date ranges
**When to use:** Dashboard time range selection
**Example:**
```typescript
interface DateRangeSelectorProps {
  value: '7d' | '30d' | '90d'
  onChange: (value: '7d' | '30d' | '90d') => void
}

export function DateRangeSelector({ value, onChange }: DateRangeSelectorProps) {
  const options = [
    { value: '7d', label: '7 days' },
    { value: '30d', label: '30 days' },
    { value: '90d', label: '90 days' },
  ] as const

  return (
    <div className="flex items-center gap-2 bg-slate-800/50 rounded-lg p-1">
      {options.map(opt => (
        <button
          key={opt.value}
          onClick={() => onChange(opt.value)}
          className={clsx(
            'px-4 py-2 text-sm font-medium rounded-md transition-all',
            value === opt.value
              ? 'bg-blue-600 text-white'
              : 'text-slate-400 hover:text-white hover:bg-slate-700'
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
```

### Anti-Patterns to Avoid
- **Fetching data in child components:** Keep data fetching at page level for easier loading states
- **Floating point money calculations:** Always use integer cents from API, only format to dollars in display
- **Custom date picker for preset ranges:** Native select/tabs are simpler for 7d/30d/90d/custom
- **Inline chart configuration:** Extract Recharts config to avoid duplication
- **Missing loading skeletons:** Show loading state while fetching, not blank screen

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Money formatting | String concatenation | `Intl.NumberFormat` | Handles locale, decimals, currency symbols |
| Date formatting | Manual date math | Date parsing in chart | Recharts handles date axis formatting |
| Responsive charts | Manual resize handlers | `ResponsiveContainer` | Recharts handles resize automatically |
| Accessible tabs | DIV with onClick | `<select>` or Headless UI | Keyboard navigation, screen readers |
| Tooltip positioning | Manual absolute positioning | Recharts `<Tooltip>` | Auto-positions, handles edges |
| Gradient backgrounds | Multiple divs | Tailwind gradients | `bg-gradient-to-br from-x to-y` |

**Key insight:** Recharts handles most visualization complexity internally. Focus on data transformation and styling, not chart mechanics.

## Common Pitfalls

### Pitfall 1: Cents vs Dollars Confusion
**What goes wrong:** Displaying raw cents values as if they were dollars
**Why it happens:** API returns `savings_cents` but UI shows `$1523.45` not `$15.23`
**How to avoid:** Create a formatCurrency helper used everywhere
**Warning signs:** Numbers seem 100x too large

```typescript
// Helper function
export function formatCurrency(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100)
}
```

### Pitfall 2: Empty State Confusion
**What goes wrong:** Blank chart when no savings data exists
**Why it happens:** New users have no historical data yet
**How to avoid:** Show explicit empty state with explanation
**Warning signs:** Dashboard looks broken for new users

```typescript
if (dailySavings.length === 0) {
  return (
    <div className="bg-slate-800/50 rounded-xl p-8 text-center">
      <p className="text-slate-400">No savings data yet.</p>
      <p className="text-sm text-slate-500 mt-2">
        Savings are calculated when instances are stopped via schedules or manual actions.
      </p>
    </div>
  )
}
```

### Pitfall 3: Loading State Flicker
**What goes wrong:** Content jumps around when switching date ranges
**Why it happens:** Not maintaining previous data while loading new data
**How to avoid:** Show loading indicator without clearing existing data
**Warning signs:** Content disappears then reappears on range change

### Pitfall 4: Tooltip Number Formatting
**What goes wrong:** Recharts tooltip shows raw number without formatting
**Why it happens:** Default formatter doesn't know about currency
**How to avoid:** Always provide custom formatter prop
**Warning signs:** Tooltip shows `1523.45` instead of `$1,523.45`

### Pitfall 5: Missing Disclaimers on Projections
**What goes wrong:** User misunderstands projected costs as actual billing data
**Why it happens:** SAV-05 requires clear disclaimers but they're forgotten
**How to avoid:** Always show disclaimer text on projection section
**Warning signs:** Users report billing doesn't match dashboard

```typescript
<div className="bg-yellow-900/20 border border-yellow-500/30 rounded-lg p-4 mt-4">
  <p className="text-yellow-400 text-sm">
    <strong>Disclaimer:</strong> Cost projections are estimates based on instance hourly rates.
    Actual cloud billing may differ due to taxes, data transfer, and other charges.
  </p>
</div>
```

### Pitfall 6: Chart Y-Axis Starting at Zero
**What goes wrong:** Small savings changes appear flat on chart
**Why it happens:** Y-axis defaults to starting at 0
**How to avoid:** Consider using `domain={['auto', 'auto']}` for better visualization
**Warning signs:** Chart looks flat even with meaningful data variation

## Code Examples

Verified patterns from official sources:

### API Types for Savings Endpoints
```typescript
// Source: Phase 7 API implementation - add to api.ts
export interface SavingsSummary {
  total_savings_cents: number
  ongoing_savings_cents: number
  period: {
    start: string
    end: string
  }
  top_savers: Array<{
    instance_id: string
    savings_cents: number
    stopped_hours: number
  }>
}

export interface DailySavingsResponse {
  daily_savings: Array<{
    date: string
    savings_cents: number
    stopped_minutes: number
    hourly_rate_cents?: number
  }>
}

export interface InstanceSavings {
  instance_id: string
  name: string
  provider: string
  region: string
  savings_cents: number
  stopped_hours: number
}
```

### API Methods to Add
```typescript
// Source: Matches Phase 7 backend endpoints - add to api.ts
// Savings endpoints
getSavingsSummary: (days: number = 30) =>
  api.get<SavingsSummary>(`/savings?days=${days}`),

getDailySavings: (days: number = 30) =>
  api.get<DailySavingsResponse>(`/savings/daily?days=${days}`),

getSavingsByInstance: (days: number = 30, limit: number = 20) =>
  api.get<InstanceSavings[]>(`/savings/by-instance?days=${days}&limit=${limit}`),

getInstanceSavings: (instanceId: string, days: number = 30) =>
  api.get<{
    instance_id: string
    total_savings_cents: number
    ongoing_savings_cents: number
    savings: Array<{
      date: string
      stopped_minutes: number
      savings_cents: number
      hourly_rate_cents: number
    }>
  }>(`/instances/${instanceId}/savings?days=${days}`),
```

### Per-Instance Savings Table
```typescript
// Source: Follows existing table patterns in codebase
interface InstanceSavingsTableProps {
  data: InstanceSavings[]
  loading: boolean
}

export function InstanceSavingsTable({ data, loading }: InstanceSavingsTableProps) {
  if (loading) {
    return <div className="animate-pulse bg-slate-800/50 rounded-xl h-64" />
  }

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Top Saving Instances</h2>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="text-left text-sm text-slate-400 border-b border-slate-700">
              <th className="pb-3 font-medium">Instance</th>
              <th className="pb-3 font-medium">Provider</th>
              <th className="pb-3 font-medium">Region</th>
              <th className="pb-3 font-medium text-right">Hours Stopped</th>
              <th className="pb-3 font-medium text-right">Savings</th>
            </tr>
          </thead>
          <tbody>
            {data.map((instance) => (
              <tr key={instance.instance_id} className="border-b border-slate-700/50 hover:bg-slate-700/30">
                <td className="py-3 text-sm text-white font-medium">{instance.name}</td>
                <td className="py-3 text-sm text-slate-300">{instance.provider.toUpperCase()}</td>
                <td className="py-3 text-sm text-slate-300">{instance.region}</td>
                <td className="py-3 text-sm text-slate-300 text-right">{instance.stopped_hours.toFixed(1)}h</td>
                <td className="py-3 text-sm text-green-400 text-right font-medium">
                  {formatCurrency(instance.savings_cents)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
```

### Cost Projection with Disclaimer
```typescript
// Source: SAV-05 requirement for disclaimers
interface CostProjectionProps {
  actualCostCents: number
  projectedAlwaysOnCents: number
}

export function CostProjection({ actualCostCents, projectedAlwaysOnCents }: CostProjectionProps) {
  const savingsCents = projectedAlwaysOnCents - actualCostCents
  const savingsPercent = projectedAlwaysOnCents > 0 
    ? ((savingsCents / projectedAlwaysOnCents) * 100).toFixed(1)
    : '0'

  return (
    <div className="bg-slate-800/50 rounded-xl p-6 border border-slate-700">
      <h2 className="text-lg font-semibold text-white mb-4">Cost Comparison</h2>
      
      <div className="grid grid-cols-2 gap-6 mb-6">
        <div className="p-4 bg-slate-700/30 rounded-lg">
          <p className="text-sm text-slate-400 mb-1">If Always Running</p>
          <p className="text-2xl font-bold text-slate-300">
            {formatCurrency(projectedAlwaysOnCents)}
          </p>
        </div>
        <div className="p-4 bg-green-900/20 rounded-lg border border-green-500/30">
          <p className="text-sm text-slate-400 mb-1">Actual Cost</p>
          <p className="text-2xl font-bold text-green-400">
            {formatCurrency(actualCostCents)}
          </p>
        </div>
      </div>

      <div className="p-4 bg-green-900/30 rounded-lg mb-4">
        <p className="text-sm text-slate-400">You saved</p>
        <p className="text-3xl font-bold text-green-400">
          {formatCurrency(savingsCents)} <span className="text-lg">({savingsPercent}%)</span>
        </p>
      </div>

      {/* SAV-05 Required Disclaimer */}
      <div className="bg-yellow-900/20 border border-yellow-500/30 rounded-lg p-4">
        <p className="text-yellow-400 text-sm">
          <strong>Disclaimer:</strong> Cost projections are estimates based on instance hourly rates 
          at the time of stop/start events. Actual cloud provider billing may differ due to:
        </p>
        <ul className="text-yellow-400/80 text-sm mt-2 ml-4 list-disc">
          <li>Reserved instance pricing or savings plans</li>
          <li>Data transfer and storage charges</li>
          <li>Taxes and regional pricing variations</li>
        </ul>
      </div>
    </div>
  )
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Class components | Functional + hooks | React 16.8 (2019) | All new code uses hooks |
| Moment.js | date-fns or native Intl | 2020+ | Moment is deprecated, not needed for this phase |
| CSS-in-JS | Tailwind utility classes | 2022+ | Project already uses Tailwind |
| Redux for simple state | useState/useEffect | 2020+ | Local state sufficient for dashboard |

**Deprecated/outdated:**
- Moment.js: Not needed - use native `Intl.DateTimeFormat` for display
- Chart.js: Would require migration - stick with Recharts
- Material UI: Project uses Headless UI + Tailwind

## Open Questions

Things that couldn't be fully resolved:

1. **Custom Date Range Implementation**
   - What we know: 7d/30d/90d presets are straightforward
   - What's unclear: Does "custom" need a full date picker or just start/end inputs?
   - Recommendation: Start with presets only. Add custom date picker in future iteration if users request it.

2. **Period-over-Period Calculation**
   - What we know: API returns current period totals
   - What's unclear: Backend may need enhancement to return previous period for comparison
   - Recommendation: Calculate client-side by fetching 2x the range, or note as future enhancement

3. **Real-time Ongoing Savings**
   - What we know: API returns `ongoing_savings_cents` for currently-stopped instances
   - What's unclear: Should this update live on the dashboard or on page refresh?
   - Recommendation: Start with page-refresh only. Add polling (30s interval) if users need real-time.

## Sources

### Primary (HIGH confidence)
- Existing codebase: `Dashboard.tsx`, `ActivityGraph.tsx`, `api.ts` - Established patterns
- Phase 7 SUMMARY: Savings API endpoints and response formats
- Recharts GitHub README (v2.15.4 installed) - Chart component API

### Secondary (MEDIUM confidence)
- Tailwind CSS documentation - Utility class patterns
- Headless UI documentation - Accessible component patterns
- LogRocket blog: Dashboard best practices

### Tertiary (LOW confidence)
- None - all patterns derived from existing codebase and official docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing dependencies only
- Architecture: HIGH - Following established codebase patterns
- Pitfalls: HIGH - Derived from common React/charting issues and SAV requirements
- Code examples: HIGH - Based on existing codebase patterns

**Research date:** 2026-02-23
**Valid until:** 2026-03-23 (30 days - stable patterns)
