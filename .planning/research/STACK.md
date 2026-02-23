# Technology Stack: Cost Savings Tracking for SnoozeQL v1.1

**Project:** SnoozeQL v1.1 - Cost Savings Tracking
**Researched:** 2026-02-23
**Confidence:** HIGH (leverages existing infrastructure patterns)

## Executive Summary

SnoozeQL v1.0 already has the core data structures needed for cost tracking: `HourlyCostCents` on instances, `Saving` model with `StoppedMinutes` and `EstimatedSavingsCents`, and `Event` records for stop/start operations. The v1.1 cost tracking feature is primarily about **calculating and visualizing** data that the system already captures.

**Key insight:** No new external libraries are required for core cost calculation. The existing Go backend can implement savings calculation using simple arithmetic from Event timestamps and instance costs. The frontend already uses Recharts for visualization.

---

## 1. Cost Estimation Techniques

### Recommended Approach: Event-Based Calculation

**Method:** Calculate savings from actual stop/start events by measuring stopped duration.

```
Savings = (StoppedDuration in hours) × (HourlyCostCents)
```

**Why this approach:**
- **Accurate:** Based on actual stop/start events, not projections
- **Simple:** No external API dependencies or pricing lookups
- **Already supported:** `Event` model captures all start/stop actions with timestamps
- **POC-appropriate:** Avoids complexity of billing API integration

**Data flow:**
1. `Event` records capture `stop` and `start` actions with `created_at` timestamps
2. Calculate duration between stop → start pairs per instance
3. Multiply stopped duration by `instance.HourlyCostCents`
4. Aggregate into `Saving` records (daily rollup)

### Alternative: Real Billing API Integration (NOT recommended for POC)

**AWS Pricing API:**
- `github.com/aws/aws-sdk-go-v2/service/pricing` provides programmatic access
- Requires parsing complex JSON pricing structures
- Pricing varies by region, instance class, storage type, and reservation model
- Adds significant complexity for marginal accuracy improvement

**GCP Cloud Billing API:**
- `cloud.google.com/billing` for programmatic access
- Similar complexity to AWS

**Verdict:** The existing `HourlyCostCents` field on instances is sufficient for POC. Real billing integration can be a future enhancement.

---

## 2. Recommended Libraries/Tools

### Go Backend (No New Dependencies Required)

| Library | Status | Purpose |
|---------|--------|---------|
| **shopspring/decimal** | OPTIONAL | Precise decimal arithmetic for currency |
| **stdlib time** | EXISTING | Duration calculations |
| **pgx/v5** | EXISTING | PostgreSQL queries for aggregation |

**Recommendation:** Use integer cents throughout (already the pattern). `HourlyCostCents` avoids floating-point issues. Only add `shopspring/decimal` if converting to dollars for display causes rounding issues.

### React Frontend (Leverage Existing)

| Library | Version | Status | Purpose |
|---------|---------|--------|---------|
| **Recharts** | 2.10.0 | EXISTING | Charts and visualizations |
| **lucide-react** | 0.300.0 | EXISTING | Icons for dashboard components |
| **React Router** | 6.20.0 | EXISTING | Navigation between savings views |

**Recommendation:** No new frontend dependencies. Recharts already supports all needed chart types (line, bar, area, pie) for cost visualization.

### External Tools Considered (Not Recommended for POC)

| Tool | What It Does | Why NOT Use |
|------|--------------|-------------|
| **Infracost** | Cloud cost estimation for Terraform | Focused on IaC, not runtime cost tracking |
| **OpenCost** | Kubernetes cost monitoring | Focused on K8s workloads, not managed databases |
| **AWS Cost Explorer API** | Actual billing data | Requires additional IAM permissions, complex integration |
| **GCP Cloud Billing API** | Actual billing data | Same complexity concerns |

---

## 3. Common Data Points for Accurate Estimation

### Already Captured in SnoozeQL v1.0

| Data Point | Model | Field | Notes |
|------------|-------|-------|-------|
| Instance class | `Instance` | `InstanceType` | e.g., `db.t3.micro` |
| Hourly cost | `Instance` | `HourlyCostCents` | Already populated per-instance |
| Region | `Instance` | `Region` | Affects pricing (not used yet) |
| Stop events | `Event` | `event_type='stop'` | With `created_at` timestamp |
| Start events | `Event` | `event_type='start'` | With `created_at` timestamp |
| Instance status | `Instance` | `Status` | Current state |

### Data Points to Add for v1.1

| Data Point | Purpose | Implementation |
|------------|---------|----------------|
| **Stopped start time** | Calculate duration | Query `Event` for last `stop` event per instance |
| **Daily aggregates** | Historical views | Populate `Saving` table with daily rollups |
| **Projected vs actual** | Forecast accuracy | Compare schedule intent vs realized savings |

### Data Points Deferred (Future Enhancement)

| Data Point | Why Defer |
|------------|-----------|
| Storage costs | Typically constant (not affected by stop/start) |
| Backup costs | Minor, constant cost |
| IOPS costs | Provisioned IOPS continue during stop |
| Multi-AZ pricing | Doubles compute cost but same calculation |
| Reserved instance pricing | Would require billing API integration |

---

## 4. Go Ecosystem for Cost Calculation

### Custom Implementation (Recommended)

The calculation logic is simple enough that custom implementation is cleaner than any library:

```go
// internal/savings/calculator.go

// CalculateDailySavings computes savings for a given instance and day
func CalculateDailySavings(events []models.Event, instance models.Instance, date time.Time) int {
    stoppedMinutes := 0
    
    // Find stop/start pairs within the date
    // Calculate total stopped duration
    
    hoursStopped := float64(stoppedMinutes) / 60.0
    savingsCents := int(hoursStopped * float64(instance.HourlyCostCents))
    
    return savingsCents
}
```

### Useful Patterns from Existing Code

The codebase already has good patterns to follow:

1. **Time duration parsing** in `internal/provider/aws/rds.go:parsePeriod()`
2. **Model structures** in `internal/models/models.go` (Saving model ready to use)
3. **Aggregation queries** can follow PostgreSQL patterns in `internal/store/postgres.go`

### Go Libraries If Needed

| Library | When to Use |
|---------|-------------|
| `github.com/shopspring/decimal` | If precise currency math becomes an issue |
| `github.com/dustin/go-humanize` | For human-readable cost formatting ("$1.2K") |

**Current recommendation:** Don't add either for POC. Use integer cents and format in frontend.

---

## 5. Frontend Visualization

### Existing Capability (Recharts 2.10.0)

The Dashboard already demonstrates cost visualization with the "Cost Over Time (7 days)" chart:

```tsx
// Current pattern in Dashboard.tsx (lines 272-301)
<div className="h-64 w-full flex items-end space-x-1 sm:space-x-2">
  {costData.map((d, i) => {
    const heightPercentage = (d.cost / maxCost) * 100
    return (
      <div key={i} className="flex-1 flex flex-col justify-end group relative">
        <div 
          className="bg-gradient-to-t from-blue-600 via-cyan-500 to-cyan-400..."
          style={{ height: `${Math.max(heightPercentage, 0.5)}%` }}
        />
      </div>
    )
  })}
</div>
```

### Recommended Chart Types for v1.1

| Visualization | Recharts Component | Use Case |
|---------------|-------------------|----------|
| Savings over time | `<LineChart>` | Historical trend |
| Daily savings breakdown | `<BarChart>` | Per-day comparison |
| Savings by instance | `<BarChart>` horizontal | Attribution |
| Projected vs actual | `<ComposedChart>` | Line + area overlay |
| Cumulative savings | `<AreaChart>` | Total savings growth |

### UI Components to Create

```
web/src/components/
├── SavingsLineChart.tsx      # Time series savings trend
├── SavingsBreakdown.tsx      # Per-instance savings list
├── SavingsCard.tsx           # Summary stat card (like existing dashboard cards)
└── ProjectionChart.tsx       # Expected vs actual comparison
```

### Dashboard Integration

The existing `Dashboard.tsx` already shows:
- Total savings stat card (line 180-189)
- Cost over time chart (line 270-301)

**v1.1 enhancement:** Replace mock `totalSavings` calculation with real API data:
```tsx
// Current (mock)
const totalSavings = instances.reduce((sum, inst) => sum + (inst.hourly_cost_cents / 100) * 24 * 7, 0)

// v1.1 (real API)
const [savingsData, setSavingsData] = useState<SavingsSummary | null>(null)
// ... fetch from /api/savings/summary
```

---

## 6. Migration Path: Incremental Implementation

### Phase 1: Backend Savings Calculation (Day 1-2)

1. **Create savings calculator service**
   ```
   internal/savings/
   ├── calculator.go     # Core calculation logic
   ├── aggregator.go     # Daily rollup logic
   └── service.go        # Business logic orchestration
   ```

2. **Implement savings API endpoints** (extend existing `handlers/savings.go`)
   - `GET /api/savings/summary` - Total savings stats
   - `GET /api/savings/history` - Daily savings history
   - `GET /api/savings/by-instance/{id}` - Per-instance breakdown

3. **Populate Saving model** from Event data
   - Background job to calculate daily aggregates
   - Calculate current-day savings on-demand

### Phase 2: Frontend Visualization (Day 2-3)

1. **Create API client methods** in `lib/api.ts`
2. **Build savings components** using existing Recharts patterns
3. **Integrate into Dashboard** - replace mock data with real API

### Phase 3: Historical Analysis (Day 3-4)

1. **Add projected savings** (based on schedules)
2. **Compare projected vs actual** 
3. **Add per-instance savings attribution**

---

## Database Schema (Already Exists)

The `Saving` model in `internal/models/models.go` is ready:

```go
type Saving struct {
    ID                    string `json:"id" db:"id"`
    InstanceID            string `json:"instance_id" db:"instance_id"`
    Date                  string `json:"date" db:"date"`
    StoppedMinutes        int    `json:"stopped_minutes" db:"stopped_minutes"`
    EstimatedSavingsCents int    `json:"estimated_savings_cents" db:"estimated_savings_cents"`
}
```

**Check:** Verify the PostgreSQL `savings` table exists and matches this schema.

---

## API Design

### Savings Summary
```
GET /api/savings/summary?period=7d

Response:
{
  "total_savings_cents": 45230,
  "period_start": "2026-02-16",
  "period_end": "2026-02-23",
  "instance_count": 5,
  "total_stopped_hours": 840
}
```

### Savings History
```
GET /api/savings/history?period=30d

Response:
{
  "data": [
    {"date": "2026-02-22", "savings_cents": 6500, "stopped_hours": 12},
    {"date": "2026-02-23", "savings_cents": 7200, "stopped_hours": 14}
  ]
}
```

### Per-Instance Savings
```
GET /api/savings/by-instance/{id}?period=7d

Response:
{
  "instance_id": "db-prod-1",
  "total_savings_cents": 12500,
  "daily_breakdown": [...]
}
```

---

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| Calculation approach | HIGH | Simple math from existing Event data |
| Go implementation | HIGH | No new dependencies, follows existing patterns |
| Frontend visualization | HIGH | Recharts already in use, proven patterns |
| Data availability | HIGH | Events and instance costs already captured |
| Accuracy | MEDIUM | Estimated costs may differ from actual billing |

---

## Sources

- **Official:** AWS RDS Pricing (https://aws.amazon.com/rds/pricing/) - Verified 2026-02-23
- **Official:** GCP Cloud Billing API (https://cloud.google.com/billing/docs/reference/rest) - Verified 2026-02-23
- **Official:** Recharts documentation - Already in use, v2.10.0
- **Open Source:** Infracost (https://github.com/infracost/infracost) - Reviewed for patterns, not recommended for this use case
- **Open Source:** OpenCost (https://github.com/opencost/opencost) - Reviewed, K8s-focused
- **Go Package:** shopspring/decimal (https://pkg.go.dev/github.com/shopspring/decimal) - Optional, v1.4.0
- **Codebase:** SnoozeQL v1.0 - Reviewed existing models, API patterns, and frontend components
