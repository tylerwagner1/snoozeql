# Phase 12: Metrics Retention - Research

**Researched:** 2026-02-25
**Domain:** Background job implementation for data cleanup in Go/PostgreSQL
**Confidence:** HIGH

## Summary

This phase implements automatic cleanup of metrics data older than 7 days. The existing codebase already has the infrastructure for this: the `metrics_hourly` table stores metrics with an indexed `hour` column, and a `DeleteOldMetrics` method exists in `MetricsStore`. The main work is creating a background job that runs on startup (with delay) and every 24 hours, tracking its last run timestamp to avoid duplicate runs.

The project already follows a consistent pattern for background jobs (see `MetricsCollector.RunContinuous()` and `DiscoveryService.RunContinuous()`) - both use `time.Ticker` with context-aware shutdown. The retention job should follow this same pattern but with the added requirement of tracking last run time in the database to prevent duplicate runs across restarts.

**Primary recommendation:** Add a new `RetentionCleaner` service following existing patterns, store last-run timestamp in the `settings` table (already exists), delete in batches using `DELETE ... WHERE ... LIMIT`, and integrate into `main.go` alongside other background services.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `time` (stdlib) | Go 1.26 | Ticker for scheduling, Duration for intervals | Standard library, no external deps needed |
| `context` (stdlib) | Go 1.26 | Graceful shutdown via context cancellation | Already used throughout codebase |
| `database/sql` (stdlib) | Go 1.26 | Database operations | Already used via pgx/stdlib wrapper |
| `github.com/jackc/pgx/v5` | v5.x | PostgreSQL driver | Already in use project-wide |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `log` (stdlib) | Go 1.26 | Error logging | Already used project-wide for consistency |
| `sync` (stdlib) | Go 1.26 | Mutex for state protection (if needed) | Thread-safe state access |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| time.Ticker | robfig/cron | Overkill for fixed 24h interval, adds dependency |
| Settings table | Separate retention_state table | Unnecessary - settings table already exists for key-value storage |
| In-memory tracking | N/A | Would lose state on restart, against requirements |

**Installation:**
```bash
# No new packages needed - all dependencies already present
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── metrics/
│   ├── store.go           # Already has DeleteOldMetrics
│   ├── collector.go       # Existing background job pattern
│   ├── retention.go       # NEW: RetentionCleaner service
│   └── ...
└── store/
    └── postgres.go        # Settings operations for last-run tracking
```

### Pattern 1: Background Service with Delayed Start
**What:** Start background cleanup after a configurable delay, then run on fixed interval
**When to use:** Cleanup jobs that shouldn't compete with startup operations
**Example:**
```go
// Source: Based on existing MetricsCollector.RunContinuous pattern
func (r *RetentionCleaner) RunContinuous(ctx context.Context) {
    // Delayed start (5-10 minutes per CONTEXT.md)
    startupDelay := 7 * time.Minute
    
    select {
    case <-ctx.Done():
        return
    case <-time.After(startupDelay):
        // Continue to main loop
    }
    
    // Run immediately after delay
    if err := r.runCleanup(ctx); err != nil {
        log.Printf("Retention cleanup failed: %v", err)
    }
    
    // Then every 24 hours
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := r.runCleanup(ctx); err != nil {
                log.Printf("Retention cleanup failed: %v", err)
            }
        }
    }
}
```

### Pattern 2: Skip-if-Recent Execution
**What:** Check last run timestamp before executing, skip if already ran within 24 hours
**When to use:** Prevent duplicate cleanup when app restarts multiple times
**Example:**
```go
// Source: Standard idempotency pattern
func (r *RetentionCleaner) runCleanup(ctx context.Context) error {
    // Check if we already ran within 24 hours
    lastRun, err := r.getLastRunTime(ctx)
    if err == nil && time.Since(lastRun) < 24*time.Hour {
        return nil // Skip - already ran recently
    }
    
    // Perform cleanup
    if err := r.deleteOldMetrics(ctx); err != nil {
        return err // Error will be logged by caller, no retry
    }
    
    // Update last run timestamp
    return r.setLastRunTime(ctx, time.Now())
}
```

### Pattern 3: Batched Deletes to Avoid Locking
**What:** Delete in chunks rather than all at once
**When to use:** Large deletes that could lock the table
**Example:**
```go
// Source: PostgreSQL best practices for bulk deletes
func (r *RetentionCleaner) deleteOldMetrics(ctx context.Context) error {
    cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
    batchSize := 1000
    
    for {
        // Delete in batches using subquery with LIMIT
        query := `
            DELETE FROM metrics_hourly 
            WHERE id IN (
                SELECT id FROM metrics_hourly 
                WHERE hour < $1 
                LIMIT $2
            )`
        
        deleted, err := r.db.Exec(ctx, query, cutoff, batchSize)
        if err != nil {
            return err
        }
        
        if deleted < int64(batchSize) {
            break // No more rows to delete
        }
        
        // Small pause between batches to reduce contention
        time.Sleep(100 * time.Millisecond)
    }
    
    return nil
}
```

### Anti-Patterns to Avoid
- **Unbounded deletes:** Never `DELETE FROM ... WHERE hour < ?` without LIMIT - can lock table for minutes
- **Logging success:** Per CONTEXT.md, don't log on success - only on errors
- **Automatic retry:** Per CONTEXT.md, don't retry on error - wait for next scheduled run
- **Configurable retention:** Per CONTEXT.md, hard-code 7 days - not configurable

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Background scheduling | Custom scheduler | time.Ticker | Built into stdlib, well-tested, used elsewhere in codebase |
| Persistent state | Custom state file | settings table | Already exists, schema ready for key-value pairs |
| Context cancellation | Manual shutdown flags | context.Context | Already used project-wide, integrates with goroutine lifecycle |
| Batched deletes | Custom pagination | DELETE with LIMIT subquery | PostgreSQL-native, index-friendly |

**Key insight:** This project already has a consistent pattern for background services. The `MetricsCollector` and `DiscoveryService` both use `RunContinuous(ctx)` with tickers. Follow this pattern exactly - don't introduce new abstractions.

## Common Pitfalls

### Pitfall 1: Not Using UTC for Timestamp Comparisons
**What goes wrong:** Timezone mismatch causes incorrect retention window
**Why it happens:** Go's `time.Now()` uses local time by default
**How to avoid:** Always use `time.Now().UTC()` for timestamp comparisons and storage
**Warning signs:** Metrics being deleted at unexpected times, or not being deleted when expected

### Pitfall 2: Missing Index on hour Column
**What goes wrong:** Delete query becomes a full table scan, extremely slow
**Why it happens:** Forgetting to create index during migration
**How to avoid:** The existing `idx_metrics_hourly_hour` index already exists (see migration 005)
**Warning signs:** Cleanup taking >10 seconds, database CPU spikes during cleanup

### Pitfall 3: Running Cleanup During Peak Traffic
**What goes wrong:** Batched deletes still consume I/O, slowing queries
**Why it happens:** Cleanup scheduled at busy times
**How to avoid:** The startup delay (5-10 minutes) naturally offsets from startup traffic; 24h interval means consistent timing
**Warning signs:** Increased query latency during cleanup windows

### Pitfall 4: Not Handling Database Errors Gracefully
**What goes wrong:** Panic or crash on transient database errors
**Why it happens:** Not wrapping database calls in error handling
**How to avoid:** Log the error and return - next scheduled run will retry. Per CONTEXT.md: no automatic retry
**Warning signs:** Application crashes, lost metrics

### Pitfall 5: Deleting Based on created_at Instead of hour
**What goes wrong:** Metrics not deleted when expected
**Why it happens:** Confusion between insertion time and metric timestamp
**How to avoid:** Per CONTEXT.md - delete based on `hour` column (metric timestamp), not `created_at`
**Warning signs:** Old metrics remaining in database beyond 7 days

## Code Examples

Verified patterns from official sources and existing codebase:

### RetentionCleaner Service Structure
```go
// Source: Based on existing internal/metrics/collector.go pattern
package metrics

import (
    "context"
    "log"
    "time"
    
    "snoozeql/internal/store"
)

const (
    retentionDays    = 7
    cleanupBatchSize = 1000
    startupDelay     = 7 * time.Minute  // Within 5-10 min range
    cleanupInterval  = 24 * time.Hour
    settingsKey      = "metrics_retention_last_run"
)

// RetentionCleaner manages automatic cleanup of old metrics
type RetentionCleaner struct {
    metricsStore *MetricsStore
    db           *store.Postgres
}

// NewRetentionCleaner creates a new retention cleaner
func NewRetentionCleaner(metricsStore *MetricsStore, db *store.Postgres) *RetentionCleaner {
    return &RetentionCleaner{
        metricsStore: metricsStore,
        db:           db,
    }
}
```

### Settings-Based Last Run Tracking
```go
// Source: Uses existing settings table pattern from internal/models/models.go
func (r *RetentionCleaner) getLastRunTime(ctx context.Context) (time.Time, error) {
    var lastRun time.Time
    query := `SELECT value->>'timestamp' FROM settings WHERE key = $1`
    err := r.db.QueryRow(ctx, query, settingsKey).Scan(&lastRun)
    return lastRun, err
}

func (r *RetentionCleaner) setLastRunTime(ctx context.Context, t time.Time) error {
    query := `
        INSERT INTO settings (key, value, scope, updated_at)
        VALUES ($1, jsonb_build_object('timestamp', $2::text), 'global', NOW())
        ON CONFLICT (key) DO UPDATE SET 
            value = jsonb_build_object('timestamp', $2::text),
            updated_at = NOW()
    `
    _, err := r.db.Exec(ctx, query, settingsKey, t.UTC().Format(time.RFC3339))
    return err
}
```

### Batched Delete Implementation
```go
// Source: PostgreSQL best practices + existing store pattern
func (r *RetentionCleaner) deleteInBatches(ctx context.Context, cutoff time.Time) error {
    for {
        query := `
            DELETE FROM metrics_hourly 
            WHERE id IN (
                SELECT id FROM metrics_hourly 
                WHERE hour < $1 
                LIMIT $2
            )`
        
        deleted, err := r.db.Exec(ctx, query, cutoff, cleanupBatchSize)
        if err != nil {
            return err
        }
        
        if deleted == 0 {
            break
        }
        
        // Brief pause between batches
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(100 * time.Millisecond):
        }
    }
    
    return nil
}
```

### Main.go Integration
```go
// Source: Based on existing main.go startup pattern
// In main.go, after metricsCollector initialization:

// Initialize retention cleaner
retentionCleaner := metrics.NewRetentionCleaner(metricsStore, db)

// Start retention cleanup in background
go retentionCleaner.RunContinuous(ctx)
log.Printf("✓ Started metrics retention cleaner (7-day retention, 24h interval)")
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| DELETE without LIMIT | Batched deletes with LIMIT subquery | Always recommended | Prevents table locks |
| Cron-based external jobs | In-process goroutines | Go 1.0+ | Simpler deployment, no external deps |
| Manual state files | Database-persisted state | Current best practice | Survives container restarts |

**Deprecated/outdated:**
- External cron jobs for Go apps: Modern Go services should handle their own scheduling
- Polling-based scheduling: Use `time.Ticker` which is interrupt-based and efficient

## Open Questions

Things that couldn't be fully resolved:

1. **Optimal batch size**
   - What we know: 1000 rows is a reasonable default for PostgreSQL
   - What's unclear: Exact optimal size depends on row size and server specs
   - Recommendation: Start with 1000, adjust if monitoring shows issues. Per CONTEXT.md, this is OpenCode's discretion.

2. **Pause duration between batches**
   - What we know: Some pause helps reduce contention
   - What's unclear: Optimal pause duration
   - Recommendation: 100ms is reasonable; longer if issues observed

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/metrics/collector.go` - Background job pattern
- Existing codebase: `internal/metrics/store.go` - DeleteOldMetrics method exists
- Existing codebase: `cmd/server/main.go` - Service initialization pattern
- Existing migration: `deployments/docker/migrations/005_metrics_hourly.sql` - Table structure with index
- Go official docs: `time.Ticker` - Scheduling API

### Secondary (MEDIUM confidence)
- PostgreSQL documentation: Batched deletes best practices
- Go official docs: `context.Context` for cancellation

### Tertiary (LOW confidence)
- None - all patterns verified against existing codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies, existing patterns
- Architecture: HIGH - Following established project patterns exactly
- Pitfalls: HIGH - Common PostgreSQL/Go patterns well-documented

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (30 days - stable domain, mature patterns)
