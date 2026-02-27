# Plan 006: Implement Scheduler Daemon

## Objective

Implement a background scheduler daemon that evaluates CRON expressions and triggers wake/sleep operations on matching instances **once per CRON match**.

## Context

### Current State
- `internal/scheduler/scheduler.go` exists with `Run()` method but:
  - `determineAction()` is stubbed (always returns `"stop"`)
  - No `RunContinuous()` method exists
  - Never wired up in `main.go`
- Schedules can be created/enabled in the UI but never execute

### Target State
- Scheduler daemon runs every 1 minute
- Evaluates all enabled schedules against current time
- Triggers wake/sleep actions exactly once per CRON match
- Logs events with `triggered_by: "schedule"`

## Dependencies

**Add to go.mod:**
```
github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
```

## Tasks

### Task 1: Add cronexpr dependency

**File:** `go.mod`

**Changes:**
```bash
go get github.com/gorhill/cronexpr
```

**Verification:**
```bash
go mod tidy && go build ./...
```

---

### Task 2: Update Scheduler struct with new dependencies

**File:** `internal/scheduler/scheduler.go`

**Add imports:**
```go
import (
    "context"
    "fmt"
    "log"
    "regexp"
    "sync"
    "time"

    "github.com/gorhill/cronexpr"
    "snoozeql/internal/models"
    "snoozeql/internal/provider"
    "snoozeql/internal/store"
)
```

**Replace Scheduler struct (lines 13-17):**
```go
// Scheduler manages database schedules
type Scheduler struct {
    store         Store
    registry      *provider.Registry
    instanceStore *store.InstanceStore
    eventStore    *store.EventStore
    lastExecuted  map[string]time.Time // "scheduleID_wake" or "scheduleID_sleep" -> last execution time
    mu            sync.Mutex
}
```

**Replace NewScheduler function (lines 29-35):**
```go
// NewScheduler creates a new scheduler
func NewScheduler(store Store, registry *provider.Registry, instanceStore *store.InstanceStore, eventStore *store.EventStore) *Scheduler {
    return &Scheduler{
        store:         store,
        registry:      registry,
        instanceStore: instanceStore,
        eventStore:    eventStore,
        lastExecuted:  make(map[string]time.Time),
    }
}
```

---

### Task 3: Implement RunContinuous method

**File:** `internal/scheduler/scheduler.go`

**Add after NewScheduler (after line 35):**
```go
// RunContinuous runs the scheduler evaluation on a 1-minute interval
func (s *Scheduler) RunContinuous(ctx context.Context) {
    log.Printf("Scheduler daemon starting (1-minute interval)")

    // Run immediately on startup to catch any missed schedules
    if err := s.Run(ctx); err != nil {
        log.Printf("Initial scheduler run error: %v", err)
    }

    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Printf("Scheduler daemon stopping")
            return
        case <-ticker.C:
            if err := s.Run(ctx); err != nil {
                log.Printf("Scheduler run error: %v", err)
            }
        }
    }
}
```

---

### Task 4: Add shouldExecute method for one-time execution tracking

**File:** `internal/scheduler/scheduler.go`

**Add after RunContinuous:**
```go
// shouldExecute checks if we should execute this schedule's action
// Returns true only if we haven't already executed in the current minute
func (s *Scheduler) shouldExecute(scheduleID, action string, now time.Time) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    key := scheduleID + "_" + action
    lastExec, exists := s.lastExecuted[key]

    // If executed within the same minute, skip
    currentMinute := now.Truncate(time.Minute)
    if exists && lastExec.Truncate(time.Minute).Equal(currentMinute) {
        return false
    }

    // Mark as executed
    s.lastExecuted[key] = now
    return true
}
```

---

### Task 5: Implement determineAction with proper CRON parsing

**File:** `internal/scheduler/scheduler.go`

**Replace determineAction function (lines 180-184):**
```go
func (s *Scheduler) determineAction(schedule models.Schedule, now time.Time) string {
    // Load timezone
    loc, err := time.LoadLocation(schedule.Timezone)
    if err != nil {
        log.Printf("Warning: Invalid timezone %s for schedule %s, using UTC", schedule.Timezone, schedule.Name)
        loc = time.UTC
    }
    nowLocal := now.In(loc)
    currentMinute := nowLocal.Truncate(time.Minute)

    // Parse wake CRON expression
    if schedule.WakeCron != "" {
        wakeCron, err := cronexpr.Parse(schedule.WakeCron)
        if err == nil {
            // Get the next occurrence after 1 minute ago
            wakeNext := wakeCron.Next(currentMinute.Add(-time.Minute))
            if wakeNext.Equal(currentMinute) {
                return "start"
            }
        } else {
            log.Printf("Warning: Invalid wake_cron '%s' for schedule %s: %v", schedule.WakeCron, schedule.Name, err)
        }
    }

    // Parse sleep CRON expression
    if schedule.SleepCron != "" {
        sleepCron, err := cronexpr.Parse(schedule.SleepCron)
        if err == nil {
            // Get the next occurrence after 1 minute ago
            sleepNext := sleepCron.Next(currentMinute.Add(-time.Minute))
            if sleepNext.Equal(currentMinute) {
                return "stop"
            }
        } else {
            log.Printf("Warning: Invalid sleep_cron '%s' for schedule %s: %v", schedule.SleepCron, schedule.Name, err)
        }
    }

    return "none"
}
```

---

### Task 6: Update Run method with instance status checks and event logging

**File:** `internal/scheduler/scheduler.go`

**Replace Run function (lines 37-80):**
```go
// Run executes all active schedules for all instances
func (s *Scheduler) Run(ctx context.Context) error {
    schedules, err := s.store.ListSchedules()
    if err != nil {
        return fmt.Errorf("failed to list schedules: %w", err)
    }

    now := time.Now()

    for _, schedule := range schedules {
        if !schedule.Enabled {
            continue
        }

        action := s.determineAction(schedule, now)
        if action == "none" {
            continue
        }

        // Check if we already executed this action for this schedule in this minute
        if !s.shouldExecute(schedule.ID, action, now) {
            continue
        }

        log.Printf("Schedule '%s' triggered action: %s", schedule.Name, action)

        // Get matching instances from database (not from provider registry)
        instances, err := s.instanceStore.ListInstances(ctx)
        if err != nil {
            log.Printf("Warning: Failed to list instances for schedule %s: %v", schedule.Name, err)
            continue
        }

        // Filter to matching instances
        var matchingInstances []models.Instance
        for _, instance := range instances {
            if matchesSelector(instance, schedule.Selectors) {
                matchingInstances = append(matchingInstances, instance)
            }
        }

        if len(matchingInstances) == 0 {
            log.Printf("Schedule '%s': No matching instances found", schedule.Name)
            continue
        }

        log.Printf("Schedule '%s': Found %d matching instances", schedule.Name, len(matchingInstances))

        for _, instance := range matchingInstances {
            // Skip if instance is already in target state
            if action == "start" && (instance.Status == "available" || instance.Status == "starting" || instance.Status == "running") {
                log.Printf("Skipping start for %s - already %s", instance.Name, instance.Status)
                continue
            }
            if action == "stop" && (instance.Status == "stopped" || instance.Status == "stopping") {
                log.Printf("Skipping stop for %s - already %s", instance.Name, instance.Status)
                continue
            }

            // Check for active override
            if hasActiveOverride(instance) {
                log.Printf("Skipping %s - active override exists", instance.Name)
                continue
            }

            // Determine new status for event logging
            var newStatus string
            var eventType string
            if action == "start" {
                newStatus = "starting"
                eventType = "wake"
            } else {
                newStatus = "stopping"
                eventType = "sleep"
            }

            // Log event BEFORE attempting action
            if s.eventStore != nil {
                event := &models.Event{
                    InstanceID:     instance.ID,
                    EventType:      eventType,
                    TriggeredBy:    "schedule",
                    PreviousStatus: instance.Status,
                    NewStatus:      newStatus,
                }
                if err := s.eventStore.CreateEvent(ctx, event); err != nil {
                    log.Printf("Warning: Failed to create event for %s: %v", instance.Name, err)
                }
            }

            // Execute action using ProviderName (e.g., "aws_uuid_us-west-2")
            switch action {
            case "stop":
                if err := s.registry.StopDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
                    log.Printf("Failed to stop %s: %v", instance.Name, err)
                } else {
                    log.Printf("Stopped %s (schedule: %s)", instance.Name, schedule.Name)
                }
            case "start":
                if err := s.registry.StartDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
                    log.Printf("Failed to start %s: %v", instance.Name, err)
                } else {
                    log.Printf("Started %s (schedule: %s)", instance.Name, schedule.Name)
                }
            }
        }
    }

    return nil
}
```

---

### Task 7: Remove unused getMatchingInstances method

**File:** `internal/scheduler/scheduler.go`

**Delete lines 82-97** (the `getMatchingInstances` function that uses registry.ListAllDatabases - we now use instanceStore.ListInstances directly in Run)

---

### Task 8: Wire up scheduler daemon in main.go

**File:** `cmd/server/main.go`

**Add import:**
```go
import (
    // ... existing imports ...
    "snoozeql/internal/scheduler"
)
```

**Add after metrics retention cleaner startup (after line 227, after the `log.Printf("✓ Started metrics retention cleaner...")`):**
```go
// Start scheduler daemon in background
schedulerService := scheduler.NewScheduler(scheduleStore, providerRegistry, instanceStore, eventStore)
go schedulerService.RunContinuous(ctx)
log.Printf("✓ Started scheduler daemon (1-minute interval)")
```

---

## File Summary

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Modify | Add `github.com/gorhill/cronexpr` dependency |
| `internal/scheduler/scheduler.go` | Modify | Add RunContinuous, shouldExecute, fix determineAction, update Run with status checks and event logging |
| `cmd/server/main.go` | Modify | Wire up scheduler daemon as background goroutine |

---

## Verification

### Build Check
```bash
cd /Users/tylerwagner/snoozeql && go build ./...
```

### Manual Test
1. Start server: `go run cmd/server/main.go`
2. Look for log: `✓ Started scheduler daemon (1-minute interval)`
3. Create/update "secondary sleeper" schedule with CRON times ~2-3 minutes in the future
4. Watch logs for:
   - `Schedule 'secondary sleeper' triggered action: start` (or stop)
   - `Started oregon-secondary-database (schedule: secondary sleeper)`
5. Check events table for `triggered_by: "schedule"` entries

---

## Success Criteria

- [ ] `go build ./...` passes
- [ ] Server starts with `✓ Started scheduler daemon` log message
- [ ] When CRON time matches, action triggers exactly once
- [ ] Instance status is checked before executing (skip if already in target state)
- [ ] Events are logged with `triggered_by: "schedule"`
- [ ] Timezone handling works correctly

---

## Estimated Duration

~15-20 minutes implementation + build verification
