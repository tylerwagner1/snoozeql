# Phase 18: Dual-Mode Data Collection - Research

**Researched:** 2026-02-26
**Domain:** Background Task Scheduling, CloudWatch Metrics Collection, Dual-Mode Architecture
**Confidence:** HIGH

## Summary

Phase 18 transforms the existing startup-only historical backfill into a continuous dual-mode data collection system. The current implementation from Phase 17 already has real-time collection (every 15 minutes) and startup gap detection (fetching up to 7 days of CloudWatch data). This phase changes the architecture to:

1. **Real-time mode (keep unchanged):** Continues every 15 minutes, fetching current metrics for fresh dashboard data
2. **Historical backfill mode (new):** Runs on startup (5-10 minute delay) AND hourly, fetching 3-day CloudWatch window for self-healing gap fill

The key change is converting `DetectAndFillGaps()` from a one-time synchronous startup call to a continuous background goroutine that runs hourly after an initial startup delay. This replaces Phase 17-02's approach with continuous healing capability.

**Primary recommendation:** Create a new `HistoricalBackfiller` component (similar to `RetentionCleaner`) that runs as a background goroutine with startup delay + hourly interval, calling a modified backfill method that fetches 3-day CloudWatch window.

## Current Implementation Summary

### Phase 17 Implementation (What Exists)

| Component | Location | Current Behavior |
|-----------|----------|-----------------|
| `RunContinuous()` | collector.go:42-76 | Real-time collection every 15 minutes |
| `DetectAndFillGaps()` | collector.go:467-594 | Startup-only, fetches up to 7 days, runs synchronously |
| `GetLatestMetricTimes()` | store.go:166-196 | Batch query for gap detection |
| `GetRDSMetricsMultiple()` | cloudwatch.go:291-370 | 5-minute period CloudWatch fetch |

### Current main.go Startup Flow (lines 213-220)

```go
// Gap detection runs synchronously at startup to populate historical data
if err := metricsCollector.DetectAndFillGaps(ctx); err != nil {
    log.Printf("Warning: Gap detection returned error: %v", err)
}

// Start metrics collection in background
go metricsCollector.RunContinuous(ctx)
```

### What Needs to Change

| Current | Phase 18 Target | Change Type |
|---------|-----------------|-------------|
| `DetectAndFillGaps()` runs once at startup | Runs hourly + startup (delayed) | Major |
| Fetches up to 7 days historical | Fetches 3-day window only | Minor |
| Synchronous blocking call | Background goroutine | Major |
| No startup delay | 5-10 minute startup delay | Minor |

## Standard Stack

### Core (Already In Use)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `time` | 1.21+ | `time.Ticker`, `time.After` for scheduling | Native Go concurrency |
| Go stdlib `context` | 1.21+ | Cancellation propagation | Standard for goroutine lifecycle |
| AWS SDK v2 `cloudwatch` | 1.32+ | CloudWatch API calls | Already in use |

### Supporting (Already In Use)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `sync` | 1.21+ | Mutex for state protection | Concurrent state access |
| Go stdlib `log` | 1.21+ | Logging | Debug and operation tracking |

### No New Dependencies Required

This phase uses existing patterns from `RetentionCleaner` - no new libraries needed.

## Architecture Patterns

### Recommended Project Structure (No Change)

```
internal/
└── metrics/
    ├── cloudwatch.go       # CloudWatch client (unchanged)
    ├── collector.go        # Add HistoricalBackfiller or RunHistoricalBackfill
    ├── store.go            # MetricsStore (unchanged)
    └── retention.go        # Reference pattern for background goroutine
```

### Pattern 1: Background Goroutine with Startup Delay

**What:** A continuous background process that waits before first execution, then runs on interval
**When to use:** Tasks that should not block startup but run periodically
**Example (from retention.go):**
```go
// Source: internal/metrics/retention.go lines 33-63
func (r *RetentionCleaner) RunContinuous(ctx context.Context) {
    // Wait for startup delay with context awareness
    select {
    case <-ctx.Done():
        log.Println("Retention cleaner shutting down before startup")
        return
    case <-time.After(startupDelay):
        // Continue to cleanup
    }

    // Run immediately after delay
    if err := r.runCleanup(ctx); err != nil {
        log.Printf("Retention cleanup failed: %v", err)
    }

    // Then every 24 hours
    ticker := time.NewTicker(cleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("Retention cleaner shutting down")
            return
        case <-ticker.C:
            if err := r.runCleanup(ctx); err != nil {
                log.Printf("Retention cleanup failed: %v", err)
            }
        }
    }
}
```

### Pattern 2: Real-Time Collection (Keep Unchanged)

**What:** Immediate execution on startup + periodic ticker
**When to use:** Fresh data needed immediately, continues on interval
**Example (from collector.go):**
```go
// Source: internal/metrics/collector.go lines 42-76
func (c *MetricsCollector) RunContinuous(ctx context.Context) {
    // Run immediately on startup
    if err := c.CollectAll(ctx); err != nil {
        log.Printf("Initial metrics collection failed: %v", err)
    }

    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("Metrics collector shutting down")
            return
        case <-ticker.C:
            // ... collection logic
        }
    }
}
```

### Pattern 3: Dual-Mode Architecture (New for Phase 18)

**What:** Two independent background goroutines with different schedules serving different purposes
**When to use:** When real-time and historical data have different requirements
**Example (target architecture):**
```go
// main.go startup
ctx := context.Background()

// Real-time mode: immediate + every 15 minutes
go metricsCollector.RunContinuous(ctx)

// Historical backfill mode: 5-minute delay + every hour
go metricsCollector.RunHistoricalBackfill(ctx)
```

### Anti-Patterns to Avoid

- **Blocking startup with long operations:** The current synchronous `DetectAndFillGaps()` call blocks server startup. Move to background with delay.
- **Overlapping responsibilities:** Keep real-time collection and historical backfill as separate concerns with clear boundaries.
- **Ignoring context cancellation:** Always check `ctx.Done()` in long-running loops and during waits.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Scheduled background tasks | Custom scheduler | `time.Ticker` + goroutine pattern | Go stdlib, proven pattern in codebase |
| Startup delay | Sleep without context | `select` with `time.After` + `ctx.Done()` | Enables graceful shutdown |
| CloudWatch fetching | New client wrapper | Existing `GetRDSMetricsMultiple()` | Already handles 5-minute periods |
| Batch instance lookup | N queries | Existing `GetLatestMetricTimes()` | Already optimized in Phase 17 |

**Key insight:** The `RetentionCleaner` pattern in `retention.go` is the exact template needed for the historical backfill goroutine.

## Common Pitfalls

### Pitfall 1: Blocking Server Startup
**What goes wrong:** Synchronous gap detection takes minutes, delaying server availability
**Why it happens:** Phase 17 implementation runs `DetectAndFillGaps()` synchronously before `RunContinuous()`
**How to avoid:** Move historical backfill to a background goroutine with startup delay
**Warning signs:** Server takes 2-5+ minutes to start accepting HTTP requests

### Pitfall 2: Race Condition Between Real-Time and Historical
**What goes wrong:** Both processes try to write same metrics simultaneously, causing conflicts
**Why it happens:** Both fetch overlapping time windows
**How to avoid:** Use `ON CONFLICT DO UPDATE` (already in `UpsertHourlyMetric`). Historical backfill only fills gaps; real-time always writes fresh data.
**Warning signs:** Database constraint violations, duplicate key errors

### Pitfall 3: Excessive CloudWatch API Calls
**What goes wrong:** Rate limiting errors, increased AWS costs
**Why it happens:** Historical backfill runs too frequently or fetches too large a window
**How to avoid:** 
- Keep interval at 1 hour (not more frequent)
- Reduce window from 7 days to 3 days
- Maintain existing 100ms throttling between instances
**Warning signs:** `LimitExceededException` in logs, high CloudWatch API costs

### Pitfall 4: Startup Delay Too Short
**What goes wrong:** Historical backfill competes with real-time collection during critical startup window
**Why it happens:** Insufficient delay allows both processes to run simultaneously
**How to avoid:** Use 5-10 minute delay for historical backfill (real-time runs immediately)
**Warning signs:** High CPU/memory during first 5 minutes after startup

### Pitfall 5: Not Removing Old Synchronous Call
**What goes wrong:** Both old synchronous call AND new background goroutine run historical backfill
**Why it happens:** Forgetting to remove `DetectAndFillGaps()` call from main.go
**How to avoid:** Phase 18 must remove lines 213-216 from main.go
**Warning signs:** Duplicate log entries for "Backfilling metrics data from CloudWatch"

## Code Examples

### New RunHistoricalBackfill Method

```go
// Source: Pattern from retention.go, adapted for historical backfill
const (
    backfillStartupDelay = 7 * time.Minute   // Wait before first backfill
    backfillInterval     = 1 * time.Hour     // Then hourly
    backfillDays         = 3                 // 3-day CloudWatch window (not 7)
)

// RunHistoricalBackfill runs historical backfill on startup (delayed) + hourly
// This provides continuous self-healing gap detection
func (c *MetricsCollector) RunHistoricalBackfill(ctx context.Context) {
    // Wait for startup delay with context awareness
    select {
    case <-ctx.Done():
        log.Println("Historical backfill shutting down before startup")
        return
    case <-time.After(backfillStartupDelay):
        // Continue to backfill
    }

    log.Println("Starting initial historical backfill...")
    
    // Run immediately after delay
    if err := c.runHistoricalBackfill(ctx); err != nil {
        log.Printf("Initial historical backfill failed: %v", err)
    }

    // Then hourly
    ticker := time.NewTicker(backfillInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("Historical backfill shutting down")
            return
        case <-ticker.C:
            log.Println("Running hourly historical backfill...")
            if err := c.runHistoricalBackfill(ctx); err != nil {
                log.Printf("Hourly historical backfill failed: %v", err)
            }
        }
    }
}

// runHistoricalBackfill performs a single backfill cycle for 3-day window
func (c *MetricsCollector) runHistoricalBackfill(ctx context.Context) error {
    log.Println("Backfilling metrics data from CloudWatch (3-day window)...")
    
    instances, err := c.instanceStore.ListInstances(ctx)
    if err != nil {
        return fmt.Errorf("failed to list instances: %w", err)
    }

    // Get latest metric times for all instances in a single query
    latestTimes, err := c.metricsStore.GetLatestMetricTimes(ctx)
    if err != nil {
        log.Printf("Warning: Failed to get latest metric times: %v", err)
    }

    var filledCount int

    for _, instance := range instances {
        // Only AWS instances supported for active collection
        if instance.Provider != "aws" {
            continue
        }

        // Determine the start time: 3 days ago or since last metric, whichever is more recent
        var startTime time.Time
        maxLookback := time.Now().UTC().Add(-backfillDays * 24 * time.Hour)

        if lastTime, exists := latestTimes[instance.ID]; exists {
            startTime = lastTime.Add(MetricPeriod)
        } else {
            startTime = maxLookback
        }

        // Cap start time at 3 days ago maximum
        if startTime.Before(maxLookback) {
            startTime = maxLookback
        }

        endTime := time.Now().UTC().Truncate(MetricPeriod)

        // Only process if we have a meaningful time range
        if endTime.Sub(startTime) < MetricPeriod {
            continue
        }

        // ... (rest of CloudWatch fetch logic, same as DetectAndFillGaps)
    }

    log.Printf("Historical backfill complete: %d new datapoints stored", filledCount)
    return nil
}
```

### Updated main.go Startup

```go
// Source: cmd/server/main.go startup section
ctx := context.Background()
go discoveryService.RunContinuous(ctx)

// REMOVED: Synchronous DetectAndFillGaps call
// OLD CODE (DELETE):
// if err := metricsCollector.DetectAndFillGaps(ctx); err != nil {
//     log.Printf("Warning: Gap detection returned error: %v", err)
// }

// Real-time collection: immediate + every 15 minutes
go metricsCollector.RunContinuous(ctx)
log.Printf("✓ Started metrics collector (15-minute interval, 5-minute granularity)")

// Historical backfill: 5-10 minute delay + hourly (self-healing gap fill)
go metricsCollector.RunHistoricalBackfill(ctx)
log.Printf("✓ Started historical backfill (7-min delay, hourly interval, 3-day window)")
```

## State of the Art

| Old Approach (Phase 17) | Current Approach (Phase 18) | Impact |
|-------------------------|----------------------------|--------|
| Startup-only gap detection | Startup (delayed) + hourly backfill | Continuous self-healing |
| 7-day CloudWatch window | 3-day CloudWatch window | Faster, lower API cost |
| Synchronous blocking call | Background goroutine | Non-blocking startup |
| Single backfill mode | Dual-mode (real-time + historical) | Architectural clarity |

**Prior Phase Decisions Being Modified:**
- Phase 17-02: "Call CloudWatch for up to 7 days of historical data on startup" → Changed to 3 days, runs hourly
- Phase 17-02: "Gap detection runs synchronously before continuous collection" → Changed to background with delay

## Open Questions

1. **Exact Startup Delay Value**
   - What we know: Should be 5-10 minutes per phase description
   - What's unclear: Optimal value depends on typical server startup time
   - Recommendation: Use 7 minutes (same as `RetentionCleaner`) for consistency

2. **Backfill Window vs Retention**
   - What we know: Retention is 7 days, backfill window is 3 days
   - What's unclear: Why 3 days specifically? Trade-off between coverage and API cost
   - Recommendation: 3 days is sufficient for hourly healing—worst case gap is 1 hour, so 3 days provides ample runway

3. **Naming: DetectAndFillGaps vs RunHistoricalBackfill**
   - What we know: Current method is `DetectAndFillGaps`, new pattern needs clarity
   - What's unclear: Keep existing method name or introduce new one?
   - Recommendation: Create new `RunHistoricalBackfill()` for goroutine wrapper, refactor `DetectAndFillGaps()` to `runHistoricalBackfill()` as private worker

## Sources

### Primary (HIGH confidence)
- `internal/metrics/retention.go` - Exact pattern for background goroutine with startup delay + interval
- `internal/metrics/collector.go` - Current collection flow, `DetectAndFillGaps()` implementation
- `cmd/server/main.go` - Current startup sequence, goroutine patterns
- Phase 17 RESEARCH.md and SUMMARY.md - Prior decisions and context

### Secondary (MEDIUM confidence)
- Go stdlib documentation for `time.Ticker`, `context` cancellation patterns
- AWS CloudWatch GetMetricStatistics rate limits documentation

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Architecture pattern: HIGH - Direct copy of existing `RetentionCleaner` pattern
- main.go changes: HIGH - Clear removal of synchronous call, addition of goroutine
- CloudWatch window change: HIGH - Simple constant change from 7 to 3 days
- Startup delay value: MEDIUM - 7 minutes is reasonable but could be tuned

**Research date:** 2026-02-26
**Valid until:** 2026-03-26 (30 days - stable domain)
