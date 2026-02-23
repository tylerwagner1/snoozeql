# Domain Pitfalls: Cost Savings Tracking

**Domain:** Cloud infrastructure cost tracking (RDS/Cloud SQL stop/start savings)
**Researched:** 2026-02-23
**Confidence:** MEDIUM (based on existing codebase analysis, AWS documentation, and cloud cost management patterns)

---

## Critical Pitfalls

Mistakes that cause incorrect savings calculations, misleading dashboards, or system rewrites.

### Pitfall 1: Ignoring the AWS 7-Day Auto-Restart

**What goes wrong:** AWS RDS automatically restarts stopped instances after 7 consecutive days. If the savings calculation assumes an instance stayed stopped, it will drastically overcount savings for instances stopped longer than 7 days.

**Why it happens:** Teams focus on stop events and assume instances stay stopped until explicitly started. AWS's documentation mentions this limitation, but it's often overlooked.

**Consequences:** 
- Savings reported could be 2-4x actual savings for long-stopped instances
- Users lose trust in savings metrics when they see unexpectedly high bills
- Entire dashboard credibility is undermined

**Prevention:**
```go
// Track expected auto-restart time when stopping
autoRestartAt := stopTime.Add(7 * 24 * time.Hour)

// In savings calculation, cap stopped duration at 7 days
maxStoppedDuration := 7 * 24 * time.Hour
if actualStoppedDuration > maxStoppedDuration {
    actualStoppedDuration = maxStoppedDuration
}
```

**Detection:** Compare calculated savings vs actual billing. Large discrepancies after 7+ days indicate this issue.

**SnoozeQL-specific:** PROJECT.md notes this as a known issue to address. Critical for v1.1 savings tracking accuracy.

---

### Pitfall 2: Using Hardcoded Instance Pricing

**What goes wrong:** The current codebase uses hardcoded approximate costs per instance type:
```go
func (p *RDSProvider) getInstanceCost(instanceClass string) int {
    switch {
    case containsPrefix(instanceStr, "db.r5."):
        return 145  // cents per hour
    case containsPrefix(instanceStr, "db.t3."):
        return 25
    // ...
    default:
        return 50
    }
}
```
These hardcoded values become stale as AWS updates pricing, and miss regional variations entirely.

**Why it happens:** Real pricing APIs are complex (AWS Pricing API, Cost Explorer). Hardcoding seems "good enough" for POC.

**Consequences:**
- US-East-1 vs EU-West-1 can differ by 10-20%
- New instance types (db.r6g, db.t4g) may default to $0.50/hr regardless of actual cost
- Reserved Instance vs On-Demand pricing ignored (could be 50%+ difference)

**Prevention for POC:**
```go
// Accept POC limitation but document it clearly
type CostEstimate struct {
    HourlyCostCents int
    IsEstimate      bool    // Always true for POC
    PricingSource   string  // "hardcoded_approximation"
    LastUpdated     string  // "2026-02-01"
}

// Show disclaimer in UI
"Cost estimates are approximate. Actual billing may vary by region and pricing plan."
```

**For future:**
- Integrate AWS Pricing API or Cost Explorer
- Allow users to input their negotiated rates
- Store pricing data with refresh mechanism

**Detection:** Compare monthly estimated savings vs Cost Explorer reports.

---

### Pitfall 3: Race Conditions in Event-Based Calculations

**What goes wrong:** Stop/start events can arrive out of order, duplicate, or fail silently. Calculating savings from events without proper state management leads to incorrect durations.

**Example scenario:**
```
Event Log (order received):
1. 10:00 - STOP event recorded
2. 10:30 - START event recorded  
3. 10:05 - Another STOP event (delayed duplicate)

Naive calculation: Instance was stopped twice! Double the savings!
```

**Why it happens:** Distributed systems, API retries, network delays, and lack of idempotency handling.

**Consequences:**
- Double-counted savings when duplicate events occur
- Negative savings when events arrive out of order
- Missing savings when start event fails to record

**Prevention:**
```go
// Use state-based calculation, not event-based
type InstanceStateLog struct {
    InstanceID  string
    State       string    // "running" | "stopped"
    ObservedAt  time.Time
    Source      string    // "api_poll" | "event"
}

// Calculate savings from state transitions, not events
func CalculateSavings(states []InstanceStateLog) (time.Duration, error) {
    // Sort by ObservedAt
    // Deduplicate consecutive same states
    // Calculate duration between state changes
}
```

**Detection:** Implement audit logging comparing event-based vs state-based calculations.

---

### Pitfall 4: Timezone Chaos in Historical Calculations

**What goes wrong:** Timestamps stored without timezone awareness, or mixed UTC/local times, cause hours to be counted twice or skipped during DST transitions.

**Why it happens:** 
- Go's `time.Now()` returns local time by default
- PostgreSQL `timestamp` vs `timestamptz` confusion
- Frontend displays local time but sends different format

**Consequences:**
- Spring forward: 1 hour of savings disappears
- Fall back: 1 hour of savings counted twice
- Cross-region deployments show wildly inconsistent numbers

**Prevention:**
```go
// ALWAYS use UTC internally
stopTime := time.Now().UTC()

// PostgreSQL: ALWAYS use TIMESTAMPTZ
CREATE TABLE savings (
    id UUID PRIMARY KEY,
    stopped_at TIMESTAMPTZ NOT NULL,  -- NOT timestamp
    started_at TIMESTAMPTZ,
    -- ...
);

// Frontend: Convert to local only for display
const localTime = new Date(utcTimestamp).toLocaleString()
```

**Detection:** Test savings calculations around DST transitions (March, November for US).

---

### Pitfall 5: Accumulating Floating Point Errors in Cost Aggregations

**What goes wrong:** Using `float64` for currency calculations accumulates rounding errors over time.

**Example:**
```go
// Bad
var totalSavings float64 = 0.0
for _, saving := range dailySavings {
    totalSavings += saving.Amount  // 0.1 + 0.2 != 0.3 in floating point
}

// After 10,000 additions, off by several dollars
```

**Why it happens:** IEEE 754 floating point can't represent 0.1 exactly. Errors compound with each operation.

**Consequences:**
- Monthly totals don't match sum of daily totals
- Rounding errors visible when comparing dashboard numbers
- Audit failures when reconciling against billing

**Prevention:**
```go
// Store as integer cents (already done in SnoozeQL!)
type Saving struct {
    EstimatedSavingsCents int  // Good!
}

// Aggregate as integers
var totalCents int
for _, saving := range dailySavings {
    totalCents += saving.EstimatedSavingsCents
}

// Convert to dollars only for display
displayDollars := float64(totalCents) / 100.0
```

**Detection:** Calculate totals both bottom-up and top-down; compare.

**SnoozeQL-specific:** Already uses `int` for cents - maintain this pattern!

---

## Moderate Pitfalls

Mistakes that cause performance issues, poor UX, or tech debt.

### Pitfall 6: N+1 Queries for Per-Instance Savings

**What goes wrong:** Dashboard loads savings for each instance in separate queries:
```go
// Bad: N+1 query pattern
instances := store.ListInstances()
for _, inst := range instances {
    savings := store.GetSavingsByInstance(inst.ID)  // N queries
}
```

**Prevention:**
```go
// Good: Single query with aggregation
SELECT 
    instance_id,
    SUM(estimated_savings_cents) as total_savings,
    SUM(stopped_minutes) as total_stopped
FROM savings
WHERE date >= $1 AND date <= $2
GROUP BY instance_id
```

**SnoozeQL-specific:** Current `ListInstances` query is efficient. Ensure savings queries follow same pattern.

---

### Pitfall 7: Missing Indexes for Time-Range Queries

**What goes wrong:** Historical savings queries scan entire tables because there's no index on date/time columns.

**Prevention:**
```sql
-- Critical indexes for savings queries
CREATE INDEX idx_savings_date ON savings(date);
CREATE INDEX idx_savings_instance_date ON savings(instance_id, date);
CREATE INDEX idx_events_created ON events(created_at);
CREATE INDEX idx_events_instance_created ON events(instance_id, created_at);
```

**Detection:** `EXPLAIN ANALYZE` on dashboard queries. Seq Scan on large tables = problem.

---

### Pitfall 8: Recalculating Historical Data on Every Request

**What goes wrong:** Dashboard calculates last-30-days savings by querying all events and computing in real-time. Becomes slow as data grows.

**Prevention:**
```go
// Materialize daily aggregates
type DailySaving struct {
    Date                  string // YYYY-MM-DD
    InstanceID            string
    StoppedMinutes        int
    EstimatedSavingsCents int
}

// Background job aggregates previous day at midnight
// Dashboard queries pre-aggregated data
```

**SnoozeQL-specific:** The `Saving` model already has this structure. Ensure aggregation job exists.

---

### Pitfall 9: Showing Too Much Data on Dashboard

**What goes wrong:** Dashboard shows every instance, every day, every metric. Users are overwhelmed and miss important insights.

**Example anti-patterns:**
- Table with 500 rows for "instances with savings"
- 30-day chart showing hourly data points (720 points)
- All filters expanded by default

**Prevention:**
```typescript
// Good dashboard hierarchy:
// 1. Single headline number: "Total Saved: $1,234"
// 2. Trend comparison: "↑ 12% vs last month"
// 3. Top 5 contributors (expandable to see all)
// 4. Drill-down on click

// Chart data aggregation:
// - 7 days: show daily
// - 30 days: show daily
// - 90 days: show weekly
// - 1 year: show monthly
```

**SnoozeQL-specific:** Current dashboard has good hierarchy. Maintain for savings section.

---

### Pitfall 10: Showing Too Little Context

**What goes wrong:** Dashboard shows "$500 saved" but user doesn't know if that's good or bad, or what actions to take.

**Prevention:**
```typescript
// Add context to every number:
<SavingsCard>
  <MainNumber>$1,234</MainNumber>
  <Context>
    <Trend>+12% vs last month</Trend>
    <Projection>On track for $3,700 this quarter</Projection>
    <TopContributor>dev-db-1 contributed 45%</TopContributor>
    <ActionHint>3 instances could save $200 more</ActionHint>
  </Context>
</SavingsCard>
```

---

### Pitfall 11: Not Handling Partial Days at Period Boundaries

**What goes wrong:** Instance stopped at 11pm, woke at 2am next day. Which day gets the 3 hours of savings?

**Naive approach:** Assign to stop day → only 1 hour counted
**Another naive approach:** Assign to wake day → previous day gets 0

**Prevention:**
```go
// Attribute savings to the day they occurred
func AttributeSavingsToDay(stopTime, startTime time.Time) map[string]time.Duration {
    result := make(map[string]time.Duration)
    
    current := stopTime
    for current.Before(startTime) {
        dayEnd := time.Date(current.Year(), current.Month(), current.Day()+1, 0, 0, 0, 0, time.UTC)
        if dayEnd.After(startTime) {
            dayEnd = startTime
        }
        
        dayStr := current.Format("2006-01-02")
        result[dayStr] += dayEnd.Sub(current)
        
        current = dayEnd
    }
    return result
}
```

---

### Pitfall 12: Confusing "Instance Stopped" with "Saving Money"

**What goes wrong:** Instance is stopped, but:
- Storage charges continue ($0.10-0.23/GB-month for GP2/GP3)
- Backup charges continue
- Reserved Instance payment continues even when stopped

User sees "Savings: $50" but actual bill reduction is $35.

**Prevention for POC:**
```typescript
// Clear labeling
<SavingsDisplay>
  <Label>Compute Savings</Label>
  <Value>$50.00</Value>
  <Disclaimer>
    Storage costs continue while stopped. 
    Estimate does not include RI commitments.
  </Disclaimer>
</SavingsDisplay>
```

**For future:** Integrate with Cost Explorer for actual savings.

---

## Minor Pitfalls

Annoyances that are easily fixable but commonly overlooked.

### Pitfall 13: Currency Formatting Inconsistencies

**What goes wrong:** Dashboard shows "$1234.5" in one place, "$1,234.50" in another, "1234.50 USD" in a third.

**Prevention:**
```typescript
// Single formatting function used everywhere
export function formatCurrency(cents: number): string {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(cents / 100);
}
```

---

### Pitfall 14: Zero Savings Edge Cases

**What goes wrong:** Dashboard shows "$0.00 saved" for instances that:
- Were never stopped
- Are in free tier
- Have $0 hourly cost (misconfigured)

Users confused about whether tracking is working.

**Prevention:**
```typescript
// Distinguish between "no savings" and "not applicable"
{instance.wasEverStopped ? (
    <Savings value={savings} />
) : (
    <NotApplicable>Never stopped during period</NotApplicable>
)}

{instance.hourlyCostCents === 0 && (
    <Warning>Cost not configured for this instance type</Warning>
)}
```

---

### Pitfall 15: Not Showing Calculation Method

**What goes wrong:** User sees "Saved: $127.50" but doesn't know how it was calculated. Can't verify against their own records.

**Prevention:**
```typescript
// Provide calculation details on hover/expand
<Tooltip>
  Calculation: 85 hours × $1.50/hr = $127.50
  <br />
  Based on db.t3.medium on-demand pricing
  <br />
  Period: 2026-02-01 to 2026-02-23
</Tooltip>
```

---

## Integration Anti-Patterns

Mistakes when integrating savings tracking with existing SnoozeQL systems.

### Pitfall 16: Breaking Existing Event Storage

**What goes wrong:** Adding savings calculation hooks to event creation slows down or breaks the stop/start flow.

**Example:**
```go
// Bad: Synchronous calculation during stop
func (h *Handler) StopInstance(w http.ResponseWriter, r *http.Request) {
    // ... stop instance ...
    event := store.CreateEvent(...)
    savingsService.CalculateAndStoreSavings(event)  // Slow! Blocks response!
    // ... respond ...
}
```

**Prevention:**
```go
// Good: Async calculation
func (h *Handler) StopInstance(w http.ResponseWriter, r *http.Request) {
    // ... stop instance ...
    event := store.CreateEvent(...)
    go savingsService.QueueSavingsCalculation(event.ID)  // Non-blocking
    // ... respond immediately ...
}

// Or: Calculate savings from periodic aggregation, not events
```

---

### Pitfall 17: Duplicating Instance Data

**What goes wrong:** Savings table stores instance details that are already in instances table. Data goes out of sync.

**Prevention:**
```sql
-- Good: Reference, don't duplicate
CREATE TABLE savings (
    id UUID PRIMARY KEY,
    instance_id UUID REFERENCES instances(id),  -- FK only
    date DATE NOT NULL,
    stopped_minutes INT NOT NULL,
    estimated_savings_cents INT NOT NULL
    -- NO: instance_name, instance_type, hourly_cost
);

-- Join when needed
SELECT s.*, i.name, i.instance_type
FROM savings s
JOIN instances i ON s.instance_id = i.id
```

**Exception:** If instances can be deleted, store denormalized copy for historical reporting.

---

### Pitfall 18: Tight Coupling with Scheduler

**What goes wrong:** Savings calculation requires scheduler to be running. If scheduler fails, savings stop being calculated.

**Prevention:**
```go
// Savings calculation should work independently
// Option 1: Calculate from events table (event-sourced)
// Option 2: Calculate from periodic instance status polling
// Option 3: Calculate from AWS CloudTrail events

// Don't require scheduler to inject savings events
```

---

## POC-Specific Pitfalls

Mistakes that waste time during rapid POC development.

### Pitfall 19: Building a Billing API Integration for POC

**What goes wrong:** Team spends 2 weeks building AWS Cost Explorer integration when hardcoded estimates would suffice for validation.

**Why it happens:** Engineers want "accurate" data. Perfectionism kills velocity.

**Prevention:**
```go
// POC approach: Hardcode with clear disclaimers
// Time: 2 hours

// V2 approach: AWS Pricing API integration
// Time: 2-3 days

// V3 approach: Cost Explorer + reconciliation
// Time: 1-2 weeks
```

**SnoozeQL-specific:** PROJECT.md correctly marks "Billing API integration" as out of scope.

---

### Pitfall 20: Over-Engineering the Data Model

**What goes wrong:** Designing for multi-currency, multi-cloud-account-per-instance, minute-level granularity when POC only needs daily USD estimates.

**Prevention:**
```go
// POC model (sufficient)
type Saving struct {
    ID                    string
    InstanceID            string
    Date                  string  // YYYY-MM-DD
    StoppedMinutes        int
    EstimatedSavingsCents int     // USD cents, good enough
}

// DON'T build for POC:
// - Currency column
// - Hourly granularity table
// - Pricing version tracking
// - Multi-tier pricing support
```

---

### Pitfall 21: Building Beautiful Charts Before Basic Numbers Work

**What goes wrong:** Team builds animated D3 visualizations before the underlying calculation is correct.

**Prevention:**
```
Phase 1: Show a single number on dashboard
Phase 2: Verify number against manual calculation
Phase 3: Add basic time series (Recharts is already in stack)
Phase 4: Add interactivity (tooltips, drill-down)
Phase 5: Polish (animations, responsive design)
```

**SnoozeQL-specific:** Recharts already integrated. Use existing patterns from ActivityGraph.

---

### Pitfall 22: Not Validating Calculations with Real Data

**What goes wrong:** Savings calculation works in tests but produces obviously wrong numbers in production (negative savings, billions of dollars).

**Prevention:**
```go
// Add sanity checks
func (s *SavingsService) CalculateDailySavings(instanceID string, date time.Time) (int, error) {
    // ... calculation ...
    
    // Sanity checks
    if savingsCents < 0 {
        log.Warn("Negative savings calculated", "instance", instanceID, "date", date)
        return 0, nil  // Or flag for review
    }
    
    // Max sanity: No instance saves more than $1000/day
    if savingsCents > 100000 {
        return 0, fmt.Errorf("implausible savings: %d cents", savingsCents)
    }
    
    return savingsCents, nil
}
```

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| SAV-01: Calculate savings | Pitfalls 1-5 (7-day restart, pricing, race conditions, timezone, float errors) | Use state-based calculation, integer cents, UTC timestamps |
| SAV-02: Savings dashboard | Pitfalls 9-10 (too much/little data) | Follow existing dashboard patterns, show context |
| SAV-03: Historical charts | Pitfalls 7-8 (indexes, recalculation) | Add indexes, pre-aggregate daily data |
| SAV-04: Per-instance attribution | Pitfall 6 (N+1 queries) | Single aggregated query with GROUP BY |
| SAV-05: Cost projection | Pitfall 2 (hardcoded pricing) | Accept POC limitation, add disclaimer |

---

## Mitigation Strategies Summary

### For Calculation Accuracy

1. **Use integer cents, not floats** - Already implemented in SnoozeQL
2. **Store all timestamps in UTC with TIMESTAMPTZ** - Verify current schema
3. **Account for AWS 7-day auto-restart** - Critical new logic needed
4. **Calculate from state changes, not events** - More reliable than event-sourcing
5. **Add sanity check bounds** - Catch implausible values early

### For Performance

1. **Pre-aggregate daily savings** - Background job, not real-time
2. **Add time-range indexes** - On date and (instance_id, date) columns
3. **Use GROUP BY for per-instance totals** - Single query, not N+1
4. **Limit dashboard to recent data** - Lazy-load historical on demand

### For UX

1. **Single headline number first** - "You saved $X this month"
2. **Always show context** - Trend, comparison, top contributors
3. **Explain calculation method** - Tooltip with formula and assumptions
4. **Handle edge cases gracefully** - Zero savings, never stopped, free tier

### For POC Velocity

1. **Accept hardcoded pricing for POC** - Add disclaimer, defer API integration
2. **Use existing Recharts patterns** - Don't reinvent visualization
3. **Validate with real data early** - Don't wait until "done" to test
4. **Ship daily, not weekly** - Show progress, get feedback

---

## Sources

| Source | Confidence | Notes |
|--------|------------|-------|
| AWS RDS Documentation - Stop Instance | HIGH | Confirms 7-day auto-restart behavior |
| AWS RDS Pricing Page | HIGH | Confirms storage charges continue when stopped |
| SnoozeQL codebase analysis | HIGH | Existing patterns for events, models, handlers |
| Cloud cost management patterns | MEDIUM | Based on common FinOps practices |
| PostgreSQL timestamp best practices | HIGH | TIMESTAMPTZ vs timestamp well-documented |
| IEEE 754 floating point limitations | HIGH | Standard computer science knowledge |

---

*Researched for SnoozeQL v1.1 - Cost Savings Tracking*
*Last updated: 2026-02-23*
