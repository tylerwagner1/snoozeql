# Architecture Patterns: Cost Savings Tracking

**Domain:** Infrastructure cost management / Cloud database lifecycle
**Researched:** 2026-02-23
**Overall Confidence:** HIGH (based on existing codebase analysis)

## Executive Summary

SnoozeQL v1.0 already has the foundational infrastructure for cost tracking: an EventStore that logs all start/stop events with timestamps, instances table with `hourly_cost_cents`, and a `savings` table for daily aggregations. The architecture pattern for v1.1 cost tracking is an **Event-Driven Savings Calculator** that hooks into the existing event flow with minimal coupling.

The recommended approach is a **hybrid push/pull model**: push cost data when events occur (calculating savings on stop events), then pull aggregated views on demand with caching for dashboard performance.

---

## 1. Data Model

### Existing Tables (Leverage As-Is)

```sql
-- Already exists in 001_base_schema.sql
-- Events table: The source of truth for all state changes
CREATE TABLE events (
    id UUID PRIMARY KEY,
    instance_id UUID REFERENCES instances(id),
    event_type VARCHAR(50) NOT NULL,      -- 'start', 'stop', 'sleep', 'wake'
    triggered_by VARCHAR(100),             -- 'manual', 'schedule', 'auto'
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    metadata JSONB,                         -- Extensible for cost data
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Instances table: Has hourly_cost_cents
CREATE TABLE instances (
    id UUID PRIMARY KEY,
    hourly_cost_cents INTEGER,             -- Already exists!
    -- ... other fields
);

-- Savings table: Daily aggregations (already defined)
CREATE TABLE savings (
    id UUID PRIMARY KEY,
    instance_id UUID REFERENCES instances(id),
    date DATE NOT NULL,
    stopped_minutes INTEGER NOT NULL,
    estimated_savings_cents INTEGER NOT NULL,
    UNIQUE(instance_id, date)
);
```

### New Migration: Enhanced Events & Savings Views

```sql
-- Migration 006_cost_tracking.sql

-- Add cost fields to events metadata (no schema change needed - use JSONB)
-- Example metadata: {"cost_cents_saved": 125, "hourly_rate_cents": 50, "duration_minutes": 150}

-- Materialized view for dashboard performance
CREATE MATERIALIZED VIEW savings_summary AS
SELECT 
    instance_id,
    DATE_TRUNC('day', date) as day,
    DATE_TRUNC('week', date) as week,
    DATE_TRUNC('month', date) as month,
    SUM(stopped_minutes) as total_stopped_minutes,
    SUM(estimated_savings_cents) as total_savings_cents
FROM savings
GROUP BY instance_id, DATE_TRUNC('day', date), DATE_TRUNC('week', date), DATE_TRUNC('month', date);

CREATE UNIQUE INDEX idx_savings_summary_instance_day 
    ON savings_summary(instance_id, day);

-- Refresh function (called by background job)
CREATE OR REPLACE FUNCTION refresh_savings_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY savings_summary;
END;
$$ LANGUAGE plpgsql;

-- Index for time-range queries on events
CREATE INDEX idx_events_instance_time ON events(instance_id, created_at DESC);

-- Index for finding "start after stop" pairs
CREATE INDEX idx_events_type_time ON events(event_type, created_at);
```

### Data Model Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  instances  │◄────│   events    │────►│   savings   │
│             │     │             │     │             │
│ hourly_cost │     │ event_type  │     │ date        │
│ _cents      │     │ metadata    │     │ stopped_min │
│             │     │ created_at  │     │ savings_$   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           ▼
                 ┌─────────────────────┐
                 │  savings_summary    │
                 │  (materialized)     │
                 │                     │
                 │  daily/weekly/      │
                 │  monthly rollups    │
                 └─────────────────────┘
```

---

## 2. Service Layer

### New Service: SavingsCalculator

```go
// internal/savings/calculator.go

package savings

import (
    "context"
    "time"
    "snoozeql/internal/models"
)

// Calculator calculates cost savings from stop/start events
type Calculator struct {
    eventStore    EventStore
    instanceStore InstanceStore
    savingsStore  SavingsStore
}

// EventStore interface (existing)
type EventStore interface {
    ListEventsByInstance(ctx context.Context, instanceID string) ([]models.Event, error)
    CreateEvent(ctx context.Context, event *models.Event) error
}

// InstanceStore interface (existing)
type InstanceStore interface {
    GetInstanceByID(ctx context.Context, id string) (*models.Instance, error)
    ListInstances(ctx context.Context) ([]models.Instance, error)
}

// SavingsStore interface (new)
type SavingsStore interface {
    UpsertDailySaving(ctx context.Context, saving *models.Saving) error
    GetSavingsByInstance(ctx context.Context, instanceID string, start, end time.Time) ([]models.Saving, error)
    GetTotalSavings(ctx context.Context, start, end time.Time) (int, error)
    GetSavingsByDateRange(ctx context.Context, start, end time.Time) ([]models.Saving, error)
}

// NewCalculator creates a new savings calculator
func NewCalculator(es EventStore, is InstanceStore, ss SavingsStore) *Calculator {
    return &Calculator{
        eventStore:    es,
        instanceStore: is,
        savingsStore:  ss,
    }
}

// CalculateEventSavings calculates savings for a specific stop->start event pair
func (c *Calculator) CalculateEventSavings(ctx context.Context, stopEvent *models.Event) (*SavingsResult, error) {
    // Get instance for hourly rate
    instance, err := c.instanceStore.GetInstanceByID(ctx, stopEvent.InstanceID)
    if err != nil {
        return nil, err
    }
    
    // Find matching start event (next start after this stop)
    events, err := c.eventStore.ListEventsByInstance(ctx, stopEvent.InstanceID)
    if err != nil {
        return nil, err
    }
    
    var startEvent *models.Event
    for _, e := range events {
        if e.EventType == "start" && e.CreatedAt.After(stopEvent.CreatedAt) {
            startEvent = &e
            break
        }
    }
    
    // Calculate duration
    var endTime time.Time
    if startEvent != nil {
        endTime = startEvent.CreatedAt
    } else {
        endTime = time.Now() // Still stopped
    }
    
    duration := endTime.Sub(stopEvent.CreatedAt)
    durationMinutes := int(duration.Minutes())
    
    // Calculate savings
    hourlyRate := instance.HourlyCostCents
    savingsCents := (durationMinutes * hourlyRate) / 60
    
    return &SavingsResult{
        InstanceID:      stopEvent.InstanceID,
        StopEventID:     stopEvent.ID,
        StartEventID:    startEvent.ID, // may be empty if still stopped
        DurationMinutes: durationMinutes,
        HourlyRateCents: hourlyRate,
        SavingsCents:    savingsCents,
        StillStopped:    startEvent == nil,
    }, nil
}

// SavingsResult represents calculated savings for an event
type SavingsResult struct {
    InstanceID      string
    StopEventID     string
    StartEventID    string
    DurationMinutes int
    HourlyRateCents int
    SavingsCents    int
    StillStopped    bool
}
```

### Extended Service: EventStore Decorator

```go
// internal/savings/event_decorator.go

package savings

import (
    "context"
    "encoding/json"
    "snoozeql/internal/models"
)

// EventStoreWithSavings decorates EventStore to calculate savings on stop events
type EventStoreWithSavings struct {
    inner      EventStore
    calculator *Calculator
}

// NewEventStoreWithSavings wraps an EventStore with savings calculation
func NewEventStoreWithSavings(inner EventStore, calc *Calculator) *EventStoreWithSavings {
    return &EventStoreWithSavings{
        inner:      inner,
        calculator: calc,
    }
}

// CreateEvent creates an event and triggers savings calculation for stop events
func (e *EventStoreWithSavings) CreateEvent(ctx context.Context, event *models.Event) error {
    // Create the event first
    if err := e.inner.CreateEvent(ctx, event); err != nil {
        return err
    }
    
    // For start events, finalize savings for the previous stop period
    if event.EventType == "start" || event.EventType == "wake" {
        go e.finalizePreviousStopSavings(ctx, event)
    }
    
    return nil
}

// finalizePreviousStopSavings calculates and stores savings when instance starts
func (e *EventStoreWithSavings) finalizePreviousStopSavings(ctx context.Context, startEvent *models.Event) {
    // Find the previous stop event
    events, err := e.inner.ListEventsByInstance(ctx, startEvent.InstanceID)
    if err != nil {
        return
    }
    
    var previousStop *models.Event
    for _, evt := range events {
        if (evt.EventType == "stop" || evt.EventType == "sleep") && 
           evt.CreatedAt.Before(startEvent.CreatedAt) {
            previousStop = &evt
            break
        }
    }
    
    if previousStop == nil {
        return
    }
    
    // Calculate and store savings
    result, err := e.calculator.CalculateEventSavings(ctx, previousStop)
    if err != nil {
        return
    }
    
    // Store daily savings
    e.calculator.StoreDailySavings(ctx, result)
}
```

### Service Layer Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                          API Handlers                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │
│  │ Instances    │  │ Schedules    │  │ Savings (NEW)            │ │
│  │ Handler      │  │ Handler      │  │ Handler                  │ │
│  └──────┬───────┘  └──────┬───────┘  └────────────┬─────────────┘ │
└─────────┼─────────────────┼───────────────────────┼────────────────┘
          │                 │                       │
          ▼                 ▼                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Service Layer                                 │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │              EventStoreWithSavings (Decorator)                │  │
│  │  ┌─────────────────┐    ┌─────────────────────────────────┐  │  │
│  │  │   EventStore    │───►│   SavingsCalculator (NEW)       │  │  │
│  │  │   (existing)    │    │   - CalculateEventSavings()     │  │  │
│  │  └─────────────────┘    │   - RecalculateHistorical()     │  │  │
│  │                         │   - StoreDailySavings()         │  │  │
│  │                         └─────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ InstanceStore│  │ ScheduleStore│  │ SavingsStore (NEW)       │  │
│  │ (existing)   │  │ (existing)   │  │ - UpsertDailySaving()    │  │
│  └──────────────┘  └──────────────┘  │ - GetSavingsByInstance() │  │
│                                       │ - GetTotalSavings()      │  │
│                                       └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
          │                 │                       │
          ▼                 ▼                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        PostgreSQL                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────────────┐ │
│  │instances │  │schedules │  │  events  │  │  savings +          │ │
│  │          │  │          │  │          │  │  savings_summary    │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. Event Integration

### Hook Points (Minimal Coupling)

The cost calculation integrates at **two points**:

1. **On Event Creation** (push model)
   - When a `start`/`wake` event is created, finalize the previous stop period savings
   - Use the decorator pattern to wrap existing EventStore

2. **On Dashboard Load** (pull model)
   - Query aggregated savings from `savings` table
   - Use materialized view for performance

### Integration Code

```go
// cmd/server/main.go changes

// Initialize savings calculator
savingsStore := store.NewSavingsStore(db)
savingsCalculator := savings.NewCalculator(eventStore, instanceStore, savingsStore)

// Wrap event store with savings decorator
eventStoreWithSavings := savings.NewEventStoreWithSavings(eventStore, savingsCalculator)

// Use decorated store in discovery service
discoveryService = discovery.NewDiscoveryService(
    providerRegistry, 
    instanceStore, 
    accountStore, 
    eventStoreWithSavings,  // <-- Use decorated version
    cfg.Discovery_enabled, 
    cfg.Discovery_interval, 
    []string{},
)
```

### Event Flow Diagram

```
User clicks "Stop"
        │
        ▼
┌───────────────────┐
│  Instance Handler │
│  POST /stop       │
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│ DiscoveryService  │
│ .StopDatabase()   │
└────────┬──────────┘
         │
         ▼
┌────────────────────────────────┐
│  EventStoreWithSavings         │
│  .CreateEvent({type: "stop"})  │
│                                │
│  1. Call inner.CreateEvent()   │
│  2. Event saved to DB          │
└────────────────────────────────┘
         │
         │ (Later, when instance starts...)
         ▼
┌────────────────────────────────┐
│  EventStoreWithSavings         │
│  .CreateEvent({type: "start"}) │
│                                │
│  1. Call inner.CreateEvent()   │
│  2. Async: Calculate savings   │
│     from previous stop event   │
│  3. Store in savings table     │
└────────────────────────────────┘
```

---

## 4. Processing Strategy

### Recommendation: Hybrid Push/Pull

| Approach | When to Use | In SnoozeQL |
|----------|-------------|-------------|
| **Push (on event)** | Finalize savings when instance starts | Yes - calculate on start event |
| **Pull (on demand)** | Dashboard aggregations | Yes - query savings table |
| **Background job** | Refresh materialized views | Yes - every 15 minutes |

### Push Strategy (Event-Triggered)

**Trigger:** When a `start` or `wake` event is created

**Action:**
1. Find the previous `stop` or `sleep` event for this instance
2. Calculate duration between stop and start
3. Multiply by hourly_cost_cents
4. Upsert to `savings` table (aggregated by day)

**Why push for finalization:**
- Accurate: Uses exact timestamps from events
- Efficient: No need to scan all events on dashboard load
- Consistent: Each stop period calculated exactly once

### Pull Strategy (On-Demand)

**Trigger:** Dashboard loads or API called

**Action:**
1. Query `savings_summary` materialized view
2. Aggregate by requested time period (day/week/month)
3. Return to frontend

**Why pull for display:**
- Fresh: Can include currently-stopped instances
- Flexible: Different time windows without pre-aggregation
- Fast: Materialized view handles heavy lifting

### Still-Stopped Instances

For instances currently stopped (no matching start event):

```go
// CalculateOngoingSavings calculates savings for currently-stopped instances
func (c *Calculator) CalculateOngoingSavings(ctx context.Context) (map[string]int, error) {
    instances, _ := c.instanceStore.ListInstances(ctx)
    ongoing := make(map[string]int)
    
    for _, inst := range instances {
        if inst.Status == "stopped" {
            // Find last stop event
            events, _ := c.eventStore.ListEventsByInstance(ctx, inst.ID)
            for _, e := range events {
                if e.EventType == "stop" || e.EventType == "sleep" {
                    duration := time.Since(e.CreatedAt)
                    minutes := int(duration.Minutes())
                    ongoing[inst.ID] = (minutes * inst.HourlyCostCents) / 60
                    break
                }
            }
        }
    }
    
    return ongoing, nil
}
```

---

## 5. Historical Calculation

### One-Time Backfill

For events that occurred before v1.1 deployment:

```go
// internal/savings/backfill.go

// BackfillHistoricalSavings processes all historical stop/start pairs
func (c *Calculator) BackfillHistoricalSavings(ctx context.Context) error {
    // Get all instances
    instances, err := c.instanceStore.ListInstances(ctx)
    if err != nil {
        return err
    }
    
    for _, instance := range instances {
        events, err := c.eventStore.ListEventsByInstance(ctx, instance.ID)
        if err != nil {
            continue
        }
        
        // Process events in chronological order
        sort.Slice(events, func(i, j int) bool {
            return events[i].CreatedAt.Before(events[j].CreatedAt)
        })
        
        var currentStop *models.Event
        for _, event := range events {
            switch event.EventType {
            case "stop", "sleep":
                currentStop = &event
            case "start", "wake":
                if currentStop != nil {
                    result, err := c.CalculateEventSavings(ctx, currentStop)
                    if err == nil {
                        c.StoreDailySavings(ctx, result)
                    }
                    currentStop = nil
                }
            }
        }
        
        // Handle still-stopped instances
        if currentStop != nil && instance.Status == "stopped" {
            result, _ := c.CalculateEventSavings(ctx, currentStop)
            if result != nil {
                c.StoreDailySavings(ctx, result)
            }
        }
    }
    
    return nil
}
```

### Backfill Endpoint

```go
// POST /api/v1/savings/backfill (admin only)
r.Post("/savings/backfill", func(w http.ResponseWriter, r *http.Request) {
    go savingsCalculator.BackfillHistoricalSavings(r.Context())
    w.WriteHeader(http.StatusAccepted)
    w.Write([]byte(`{"status":"backfill_started"}`))
})
```

### Idempotency

The `UNIQUE(instance_id, date)` constraint on `savings` table ensures:
- Running backfill multiple times is safe
- Daily aggregations are updated, not duplicated
- Use `ON CONFLICT DO UPDATE` for upsert semantics

---

## 6. API Design

### New Endpoints

```
GET  /api/v1/savings                     # Overall savings summary
GET  /api/v1/savings/daily?days=30       # Daily breakdown
GET  /api/v1/savings/by-instance         # Per-instance breakdown
GET  /api/v1/instances/{id}/savings      # Single instance savings
POST /api/v1/savings/backfill            # Trigger historical calculation
```

### Response Schemas

```json
// GET /api/v1/savings
{
  "total_savings_cents": 125000,
  "period": "30d",
  "savings_by_period": {
    "today": 2500,
    "this_week": 15000,
    "this_month": 45000,
    "all_time": 125000
  },
  "top_savers": [
    {
      "instance_id": "uuid",
      "instance_name": "dev-postgres",
      "savings_cents": 45000,
      "stopped_hours": 720
    }
  ],
  "ongoing_savings": {
    "currently_stopped": 3,
    "projected_daily_cents": 3600
  }
}

// GET /api/v1/savings/daily?days=30
{
  "daily_savings": [
    {"date": "2026-02-22", "savings_cents": 2500, "stopped_minutes": 1440},
    {"date": "2026-02-21", "savings_cents": 2100, "stopped_minutes": 1200}
  ]
}

// GET /api/v1/instances/{id}/savings
{
  "instance_id": "uuid",
  "instance_name": "dev-postgres",
  "hourly_cost_cents": 50,
  "total_savings_cents": 45000,
  "total_stopped_minutes": 54000,
  "events": [
    {
      "stop_at": "2026-02-20T18:00:00Z",
      "start_at": "2026-02-21T08:00:00Z",
      "duration_minutes": 840,
      "savings_cents": 700
    }
  ]
}
```

### Handler Implementation

```go
// internal/api/handlers/savings.go (enhanced)

type SavingsHandler struct {
    calculator    *savings.Calculator
    instanceStore InstanceStore
}

func (h *SavingsHandler) GetSavings(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get time range from query params (default 30 days)
    days := 30
    if d := r.URL.Query().Get("days"); d != "" {
        if parsed, err := strconv.Atoi(d); err == nil {
            days = parsed
        }
    }
    
    end := time.Now()
    start := end.AddDate(0, 0, -days)
    
    // Get historical savings
    totalSavings, _ := h.calculator.GetTotalSavings(ctx, start, end)
    
    // Get ongoing savings for currently-stopped instances
    ongoing, _ := h.calculator.CalculateOngoingSavings(ctx)
    
    // Get top savers
    topSavers, _ := h.calculator.GetTopSavers(ctx, start, end, 5)
    
    response := map[string]interface{}{
        "total_savings_cents": totalSavings,
        "period":              fmt.Sprintf("%dd", days),
        "top_savers":          topSavers,
        "ongoing_savings": map[string]interface{}{
            "currently_stopped":    len(ongoing),
            "projected_daily_cents": sumValues(ongoing) * 24 * 60 / sumMinutes(ongoing),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

---

## 7. Common Pitfalls

### Anti-Pattern 1: Calculating Savings on Every Dashboard Load

**What goes wrong:** Query all events, join with instances, calculate durations on every page load. Becomes O(n*m) where n=events, m=instances.

**Why it happens:** Seems simpler than maintaining aggregated table.

**Consequences:** Dashboard becomes slow as event count grows. 10,000 events = 2-3 second load times.

**Prevention:** Use the `savings` table for pre-aggregated data. Calculate once on event, read many times.

### Anti-Pattern 2: Tight Coupling to Event Creation

**What goes wrong:** Modify EventStore.CreateEvent() directly to add savings calculation inline.

**Why it happens:** Path of least resistance.

**Consequences:** 
- Harder to test EventStore in isolation
- Savings errors can fail event creation
- Cannot disable savings calculation easily

**Prevention:** Use decorator pattern. EventStoreWithSavings wraps EventStore, keeping responsibilities separate.

### Anti-Pattern 3: Ignoring Currently-Stopped Instances

**What goes wrong:** Only calculate savings when instance starts. Dashboard shows $0 for stopped instance that's been off for 3 days.

**Why it happens:** Natural to calculate only when period is "complete."

**Consequences:** Users don't see real-time value. Stopped instance appears to have no savings.

**Prevention:** Calculate "ongoing savings" separately using `time.Now() - last_stop_event`. Add to dashboard totals.

### Anti-Pattern 4: Floating Point for Money

**What goes wrong:** Store savings as `float64` or `decimal`.

**Why it happens:** Seems natural for "dollars.cents" format.

**Consequences:** Rounding errors accumulate. $99.99 + $0.02 might equal $100.00999999.

**Prevention:** Store everything in cents as `INTEGER`. Convert to dollars only in frontend. Already done correctly in existing schema.

### Anti-Pattern 5: Not Handling Instance Cost Changes

**What goes wrong:** Instance is stopped, user changes instance type (cost changes), instance starts. Savings calculated with wrong rate.

**Why it happens:** Assume hourly_cost is static.

**Consequences:** Inaccurate savings if instance was resized while stopped.

**Prevention:** Store `hourly_rate_cents` in event metadata at time of stop. Use that rate for calculation, not current instance rate.

```go
// When creating stop event
event.Metadata = json.Marshal(map[string]interface{}{
    "hourly_rate_cents": instance.HourlyCostCents,
})
```

### Anti-Pattern 6: Background Job Only (No Push)

**What goes wrong:** Rely entirely on periodic background job to calculate all savings.

**Why it happens:** Simpler architecture initially.

**Consequences:**
- Dashboard shows stale data (up to 15 minutes old)
- Background job becomes heavy over time
- Harder to trace which calculation was wrong

**Prevention:** Push on event for accuracy, background job only for view refresh.

---

## Implementation Checklist

### Phase 1: Data Model
- [ ] Create migration 006_cost_tracking.sql
- [ ] Add savings_summary materialized view
- [ ] Add indexes for event time-range queries

### Phase 2: Service Layer
- [ ] Create `internal/savings/calculator.go`
- [ ] Create `internal/savings/event_decorator.go`  
- [ ] Create `internal/store/savings_store.go`

### Phase 3: Integration
- [ ] Wire up EventStoreWithSavings in main.go
- [ ] Add savings calculation to event flow
- [ ] Implement ongoing savings calculation

### Phase 4: API
- [ ] Create savings handler endpoints
- [ ] Add backfill endpoint
- [ ] Wire routes in main.go

### Phase 5: Background Jobs
- [ ] Add materialized view refresh job
- [ ] Run every 15 minutes with metrics collector

---

## Sources

- **HIGH confidence:** Existing codebase analysis (`internal/store/postgres.go`, `internal/discovery/discovery.go`, `cmd/server/main.go`)
- **HIGH confidence:** Existing schema (`deployments/docker/migrations/001_base_schema.sql`)
- **HIGH confidence:** Existing models (`internal/models/models.go`)
- **MEDIUM confidence:** PostgreSQL materialized view patterns (standard PostgreSQL documentation)
