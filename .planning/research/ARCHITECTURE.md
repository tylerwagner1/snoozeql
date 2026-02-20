# Architecture Research

**Domain:** Database Sleep Scheduling System
**Researched:** Feb 20, 2026
**Confidence:** HIGH (based on existing codebase analysis and infrastructure automation patterns)

## System Overview

Database sleep scheduling systems follow a **control plane + worker** pattern common in infrastructure automation. The existing SnoozeQL codebase already implements core components; the architecture recommendation extends these patterns for activity analysis and recommendations.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PRESENTATION LAYER                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                   │
│  │   Dashboard   │  │   Schedule    │  │    Alerts     │                   │
│  │   (React)     │  │   Editor      │  │   & Status    │                   │
│  └───────┬───────┘  └───────┬───────┘  └───────┬───────┘                   │
│          │                  │                  │                            │
├──────────┴──────────────────┴──────────────────┴────────────────────────────┤
│                              API LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     Chi HTTP API (Go)                                │    │
│  │   /instances  /schedules  /recommendations  /savings  /accounts     │    │
│  └──────────────────────────────┬──────────────────────────────────────┘    │
│                                 │                                            │
├─────────────────────────────────┴────────────────────────────────────────────┤
│                           DOMAIN SERVICES                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  Discovery   │  │  Scheduler   │  │  Analyzer    │  │  Savings     │    │
│  │  Service     │  │  Service     │  │  Service     │  │  Calculator  │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│         │                 │                 │                  │             │
│         │    ┌────────────┴─────────────────┴──────────┐       │             │
│         │    │         Provider Registry                │       │             │
│         └────┤     (AWS RDS, GCP CloudSQL, etc.)       ├───────┘             │
│              └──────────────────────────────────────────┘                     │
│                                 │                                            │
├─────────────────────────────────┴────────────────────────────────────────────┤
│                           DATA LAYER                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                       │
│  │  PostgreSQL  │  │   Metrics    │  │   Events     │                       │
│  │  (Primary)   │  │  Time-Series │  │   Queue      │                       │
│  └──────────────┘  └──────────────┘  └──────────────┘                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | Current Status | Dependencies |
|-----------|----------------|----------------|--------------|
| **Discovery Service** | Poll cloud providers, sync instances to DB, detect tag changes | Implemented (basic) | Provider Registry |
| **Provider Registry** | Abstract cloud API operations (start/stop/status/metrics) | Implemented | AWS/GCP providers |
| **Scheduler Service** | Evaluate cron expressions, match instances to schedules, execute actions | Implemented (partial) | Provider Registry, Override Store |
| **Analyzer Service** | Collect metrics, detect inactivity patterns, generate recommendations | Scaffolded | Provider Registry, Metrics Store |
| **Savings Calculator** | Track stop/start events, calculate cost savings per instance/day | Not implemented | Events Store, Instance Store |
| **Override Service** | Manage keep-alive and skip-next overrides | Schema exists, logic pending | Instance Store |
| **API Layer** | HTTP endpoints, authentication, request validation | Implemented (partial) | All services |

## Existing Architecture Analysis

The current codebase follows a **clean service-oriented architecture** that aligns well with standard patterns:

### What's Working Well

1. **Provider Abstraction** - `Provider` interface in `internal/provider/provider.go` cleanly separates cloud-specific logic
2. **Registry Pattern** - Provider registry allows multi-region, multi-provider support
3. **Domain Models** - `internal/models/models.go` captures all necessary entities
4. **Database Schema** - PostgreSQL schema supports the full feature set

### Current Component Structure

```
internal/
├── api/
│   ├── handlers/       # HTTP handlers (partially implemented)
│   └── middleware/     # Chi middleware (CORS, auth, request tracking)
├── analyzer/           # Activity analysis (scaffolded)
├── config/             # Environment configuration
├── discovery/          # Instance discovery (implemented)
├── models/             # Domain models
├── provider/
│   ├── aws/            # AWS RDS provider (implemented)
│   ├── gcp/            # GCP CloudSQL provider (scaffolded)
│   └── registry.go     # Provider registry
├── scheduler/          # Schedule execution (partial)
└── store/              # PostgreSQL data access
```

## Recommended Project Structure

Extensions to support activity analysis and recommendations:

```
internal/
├── api/
│   ├── handlers/
│   │   ├── instances.go       # Instance CRUD + start/stop
│   │   ├── schedules.go       # Schedule CRUD
│   │   ├── recommendations.go # Recommendation review/apply
│   │   ├── savings.go         # Savings reports
│   │   └── settings.go        # Configuration
│   └── middleware/
├── analyzer/
│   ├── analyzer.go            # Main analysis service
│   ├── patterns.go            # Pattern detection algorithms (NEW)
│   ├── metrics_collector.go   # CloudWatch/StackDriver metrics (NEW)
│   └── recommendation.go      # Recommendation generation (NEW)
├── scheduler/
│   ├── scheduler.go           # Schedule execution
│   ├── cron.go                # Cron expression parsing (NEW)
│   └── overrides.go           # Override checking (NEW)
├── calculator/                 # (NEW) Savings calculation
│   ├── calculator.go          # Main service
│   └── pricing.go             # Instance pricing lookup
├── events/                     # (NEW) Event handling
│   ├── emitter.go             # Event emission
│   └── processor.go           # Event processing/aggregation
└── store/
    ├── postgres.go            # Base connection
    ├── instance_store.go      # Instance CRUD (NEW)
    ├── schedule_store.go      # Schedule CRUD (NEW)
    ├── metrics_store.go       # Metrics persistence (NEW)
    └── savings_store.go       # Savings aggregation (NEW)
```

### Structure Rationale

- **analyzer/** - Self-contained analysis domain. Pattern detection logic should be isolatable for testing.
- **calculator/** - Separate from analyzer because savings calculation uses different data (events) than pattern detection (metrics).
- **events/** - Event-driven architecture enables loose coupling between action execution and savings tracking.
- **store/** - Split from single postgres.go to domain-specific stores for maintainability.

## Architectural Patterns

### Pattern 1: Provider Interface Abstraction

**What:** Abstract cloud-specific operations behind a common interface
**When to use:** Multi-cloud support required
**Trade-offs:** Adds indirection but enables provider-agnostic business logic

**Example (already implemented):**
```go
// internal/provider/provider.go
type Provider interface {
    ListDatabases(ctx context.Context) ([]models.Instance, error)
    StartDatabase(ctx context.Context, id string) error
    StopDatabase(ctx context.Context, id string) error
    GetDatabaseStatus(ctx context.Context, id string) (string, error)
    GetMetrics(ctx context.Context, providerName, id, period string) (map[string]any, error)
}
```

### Pattern 2: Event Sourcing for Audit Trail

**What:** Store all state changes as immutable events
**When to use:** Need audit trail, savings calculation, or debugging history
**Trade-offs:** More storage, but enables replay and analysis

**Example:**
```go
// internal/events/emitter.go
type EventEmitter struct {
    store EventStore
}

func (e *EventEmitter) EmitStateChange(ctx context.Context, instanceID string, event StateChangeEvent) error {
    return e.store.Insert(ctx, Event{
        InstanceID:     instanceID,
        EventType:      event.Type,      // "scheduled_stop", "scheduled_start", "manual_stop"
        TriggeredBy:    event.Source,    // "schedule:dev-hours", "user:admin@corp.com"
        PreviousStatus: event.OldStatus, // "available"
        NewStatus:      event.NewStatus, // "stopped"
        Metadata:       event.Metadata,
        CreatedAt:      time.Now(),
    })
}
```

### Pattern 3: Selector-Based Instance Matching

**What:** Use selectors (like Kubernetes label selectors) for dynamic instance grouping
**When to use:** Schedules should apply to dynamically changing instance sets
**Trade-offs:** More complex matching logic but scales better than explicit assignment

**Example (already in models):**
```go
// Selector defines matching criteria for dynamic schedule assignment
type Selector struct {
    Name     *Matcher            `json:"name,omitempty"`
    Provider *string             `json:"provider,omitempty"`
    Region   *Matcher            `json:"region,omitempty"`
    Engine   *Matcher            `json:"engine,omitempty"`
    Tags     map[string]*Matcher `json:"tags,omitempty"`
}
```

### Pattern 4: Background Worker with Tick Loop

**What:** Long-running goroutine that executes on intervals
**When to use:** Discovery polling, schedule evaluation, metrics collection
**Trade-offs:** Simple but limited scaling; adequate for single-instance deployment

**Example (already implemented in discovery):**
```go
// RunContinuous runs discovery on the configured interval
func (d *DiscoveryService) RunContinuous(ctx context.Context) {
    ticker := time.NewTicker(d.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := d.Run(ctx); err != nil {
                log.Printf("Discovery run failed: %v", err)
            }
        }
    }
}
```

## Data Flow

### Discovery Flow

```
Cloud Provider API
       ↓
   Provider.ListDatabases()
       ↓
   Discovery Service
       ↓ (upsert)
   PostgreSQL (instances table)
       ↓
   API Response
```

### Scheduling Flow

```
Cron Tick (every minute)
       ↓
   Scheduler.Run()
       ↓
   Load enabled schedules
       ↓
   For each schedule:
       ├── Match instances via selectors
       ├── Check override status
       ├── Determine action (start/stop/none)
       └── Execute via Provider
              ↓
          Event Emitter
              ↓
          PostgreSQL (events table)
```

### Analysis Flow (NEW)

```
Analysis Tick (configurable, e.g., every 6 hours)
       ↓
   Analyzer.RunAnalysis()
       ↓
   For each managed instance:
       ├── Fetch metrics (Provider.GetMetrics)
       ├── Detect inactivity patterns
       ├── Calculate confidence score
       └── Generate recommendation if confidence > threshold
              ↓
          PostgreSQL (recommendations table)
              ↓
          UI shows pending recommendations
              ↓
          User approves → Schedule created
```

### Savings Calculation Flow (NEW)

```
Event Stored (stop/start)
       ↓
   Events Processor (async)
       ↓
   Calculate duration between stop→start
       ↓
   Look up instance hourly cost
       ↓
   Aggregate to savings table
       ↓
   Dashboard queries savings
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-100 instances | Current architecture is fine. Single Go binary, PostgreSQL. |
| 100-1000 instances | Consider rate limiting cloud API calls (AWS RDS has low API limits). Queue actions. |
| 1000+ instances | Split scheduler into workers. Use Redis for job queuing. Consider TimescaleDB for metrics. |

### Scaling Priorities

1. **First bottleneck: Cloud API rate limits** - AWS RDS DescribeDBInstances has ~10 TPS limit. Implement exponential backoff and batch operations. Cache instance state.

2. **Second bottleneck: Scheduler contention** - If many schedules fire at same cron minute, execute in parallel with bounded concurrency pool.

3. **Third bottleneck: Metrics storage** - If collecting high-resolution metrics, PostgreSQL may struggle. Consider TimescaleDB extension or separate time-series DB.

## Anti-Patterns

### Anti-Pattern 1: Synchronous API to Cloud Operations

**What people do:** Block HTTP request while waiting for start/stop to complete
**Why it's wrong:** Cloud operations take 5-15 minutes. HTTP timeout, poor UX.
**Do this instead:** Return immediately with "status: starting/stopping", poll for completion

### Anti-Pattern 2: Storing Credentials in Instance Metadata

**What people do:** Store AWS credentials in instance tags or database columns
**Why it's wrong:** Security risk, credentials exposed in UI/logs
**Do this instead:** Store credentials at account level with encryption, reference by account ID

### Anti-Pattern 3: Per-Instance Schedule Assignment

**What people do:** Store schedule_id on each instance record
**Why it's wrong:** Doesn't scale. Adding instance requires manual schedule assignment.
**Do this instead:** Use selectors. Schedule matches instances dynamically.

### Anti-Pattern 4: Hardcoded Cron Evaluation

**What people do:** Evaluate cron every second, comparing to current time
**Why it's wrong:** Inefficient, race conditions, time drift issues
**Do this instead:** Use gorhill/cronexpr or robfig/cron to calculate next run time

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| AWS RDS | SDK (aws-sdk-go-v2) | Already implemented. Add CloudWatch for metrics. |
| GCP CloudSQL | SDK (google.golang.org/api) | Scaffolded. Needs metrics via Cloud Monitoring. |
| Azure SQL | SDK (azure-sdk-for-go) | Not started. Similar pattern to AWS/GCP. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| API → Services | Direct function calls | Single binary, no need for RPC |
| Services → Providers | Interface abstraction | Enables testing with mocks |
| Services → Store | Repository pattern | Each store is a single-table repository |
| Scheduler → Events | Event emission | Decoupled logging/savings tracking |

## Build Order Implications

Based on component dependencies, recommended build order:

### Phase 1: Core Enhancement
Build **Event System** first - it's a dependency for savings calculation and audit logging.
- `events/emitter.go`
- Integrate into existing start/stop operations

### Phase 2: Metrics Infrastructure  
Build **Metrics Collection** before analysis - analyzer depends on having metrics data.
- Add CloudWatch integration to AWS provider
- Create `store/metrics_store.go`
- Implement `analyzer/metrics_collector.go`

### Phase 3: Activity Analysis
Build **Pattern Detection** after metrics are available.
- `analyzer/patterns.go` - inactivity detection algorithms
- `analyzer/recommendation.go` - generate schedule suggestions
- Full analyzer pipeline

### Phase 4: Savings & Reporting
Build **Savings Calculator** after events are flowing.
- `calculator/calculator.go`
- `store/savings_store.go`
- Dashboard integration

### Phase 5: Enhanced Scheduling
Improve **Scheduler** with better cron handling and override logic.
- `scheduler/cron.go` - proper cron library integration
- `scheduler/overrides.go` - keep-alive, skip-next handling

### Critical Dependencies

```
Events ─────┬─────→ Savings Calculator
            │
            └─────→ Audit Trail (existing events table)

Metrics ────┬─────→ Analyzer ─────→ Recommendations
Collection  │
            └─────→ Dashboard Charts

Provider ───┬─────→ Discovery
Interface   │
            └─────→ Scheduler ─────→ Events
```

## Sources

- Existing SnoozeQL codebase analysis (HIGH confidence)
- AWS RDS start/stop patterns: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html (HIGH confidence)
- Go project structure patterns: https://github.com/golang-standards/project-layout (MEDIUM confidence - community convention)
- Cron parsing: github.com/robfig/cron (HIGH confidence - widely used library)

---
*Architecture research for: Database Sleep Scheduling System*
*Researched: Feb 20, 2026*
