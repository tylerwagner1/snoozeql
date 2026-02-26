# Phase 17: Enhanced Metrics & Data Collection Strategy - Research

**Researched:** 2026-02-25
**Domain:** CloudWatch Metrics Collection, Gap Detection, Data Interpolation
**Confidence:** HIGH

## Summary

This phase enhances the existing metrics collection infrastructure to provide higher granularity data (5-minute intervals instead of hourly) and ensure data continuity through intelligent gap detection and interpolation. The current system collects metrics every 15 minutes at hourly granularity (Period=3600s); this phase changes to 5-minute granularity (Period=300s) with 3 datapoints per collection cycle.

The main changes involve:
1. **CloudWatch API Period Change**: Reduce Period from 3600s to 300s in `cloudwatch.go`
2. **Multi-datapoint Collection**: Fetch 3 datapoints (covering 15 minutes) per collection cycle
3. **Gap Detection on Startup**: Query database for missing intervals since last collection
4. **Interpolation for Gaps**: Create interpolated entries for missing periods to maintain timeline continuity

**Primary recommendation:** Modify the existing `getMetric()` function to use Period=300, add a new method to fetch multiple datapoints, implement startup gap detection with linear interpolation backfill.

## Current Implementation Summary

### Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `internal/metrics/cloudwatch.go` | CloudWatch API client, GetRDSMetrics, getMetric | 257 |
| `internal/metrics/collector.go` | MetricsCollector, RunContinuous, BackfillMetrics | 452 |
| `internal/metrics/store.go` | MetricsStore, UpsertHourlyMetric, HourHasData | 156 |
| `internal/metrics/retention.go` | RetentionCleaner (7-day retention, 24h cleanup) | 138 |
| `cmd/server/main.go` | Collection interval set to 15 minutes (line 196) | - |

### Current Collection Flow

```
Every 15 minutes:
  collector.CollectAll()
    → For each running instance:
      → client.GetRDSMetrics() 
        → getMetric() with Period=3600s, StartTime=now-1h, EndTime=now
        → Returns single aggregated datapoint for last hour
      → storeMetric() → UpsertHourlyMetric with hour truncation
```

### Current CloudWatch API Call (cloudwatch.go:179-196)

```go
input := &cloudwatch.GetMetricStatisticsInput{
    Namespace:  aws.String("AWS/RDS"),
    MetricName: aws.String(metricName),
    Dimensions: []types.Dimension{{
        Name:  aws.String("DBInstanceIdentifier"),
        Value: aws.String(dbInstanceID),
    }},
    StartTime:  aws.Time(start),
    EndTime:    aws.Time(end),
    Period:     aws.Int32(3600), // 1 hour ← CHANGE TO 300
    Statistics: []types.Statistic{
        types.StatisticAverage,
        types.StatisticMaximum,
        types.StatisticMinimum,
    },
}
```

### Database Schema (metrics_hourly)

```sql
CREATE TABLE metrics_hourly (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    metric_name VARCHAR(50) NOT NULL,
    hour TIMESTAMPTZ NOT NULL,  -- Currently truncated to hour
    avg_value FLOAT NOT NULL,
    max_value FLOAT NOT NULL,
    min_value FLOAT NOT NULL,
    sample_count INT NOT NULL DEFAULT 1,
    UNIQUE(instance_id, metric_name, hour)
);
```

**Note:** The schema uses `hour` column name but the unique constraint and truncation logic could work with 5-minute intervals without schema changes—just store timestamps truncated to 5 minutes instead of hours.

## CloudWatch API Considerations for 5-Minute Periods

### Period Constraints (from AWS Documentation)

| Start Time Age | Minimum Period | Max Datapoints |
|----------------|----------------|----------------|
| < 3 hours ago | 60 seconds | 1,440 per call |
| 3-15 days ago | 60 seconds (1 min) | 1,440 per call |
| 15-63 days ago | 300 seconds (5 min) | 1,440 per call |
| > 63 days ago | 3600 seconds (1 hour) | 1,440 per call |

**For 5-minute (300s) periods:**
- Data available for 63 days (vs 455 days for hourly)
- Safe to use for any start time within last 63 days
- Matches our 7-day retention policy well

### Datapoints Calculation

For a 15-minute collection window with 300s period:
- `EndTime - StartTime = 15 minutes = 900 seconds`
- `900s / 300s = 3 datapoints`
- Well within the 1,440 max per call

### Rate Limiting Considerations

Current backfill uses 100ms sleep between hour iterations. With 5-minute periods:
- 3x more datapoints per hour
- Existing throttling (100ms) should remain sufficient
- `getMetricWithRetry` already handles `LimitExceededException` with exponential backoff

**Confidence:** HIGH - AWS documentation verified

## Gap Detection Approach

### When Gaps Occur

1. **Server downtime/restart**: Application wasn't running to collect metrics
2. **Network failures**: CloudWatch API calls failed
3. **Instance state changes**: Instance was stopped, then restarted

### Gap Detection Algorithm

```go
// On startup, before first collection:
func (c *MetricsCollector) DetectAndFillGaps(ctx context.Context) error {
    for _, instance := range instances {
        // 1. Get last recorded metric timestamp
        lastMetric, err := c.metricsStore.GetLatestMetricTime(ctx, instance.ID)
        
        // 2. Calculate gap duration
        gapStart := lastMetric.Add(5 * time.Minute)
        gapEnd := time.Now().UTC().Truncate(5 * time.Minute)
        
        // 3. If gap > threshold (e.g., 15 minutes), fill with interpolation
        if gapEnd.Sub(gapStart) > 15*time.Minute {
            c.fillGap(ctx, instance, gapStart, gapEnd)
        }
    }
}
```

### New Store Method Required

```go
// GetLatestMetricTime returns the most recent metric timestamp for an instance
func (s *MetricsStore) GetLatestMetricTime(ctx context.Context, instanceID string) (time.Time, error) {
    query := `
        SELECT MAX(hour) FROM metrics_hourly
        WHERE instance_id = $1`
    var maxTime time.Time
    err := s.db.QueryRow(ctx, query, instanceID).Scan(&maxTime)
    return maxTime, err
}
```

### Verification Example (from Phase Description)

> Verify: 11am-4pm data + 7pm restart → interpolated entries for 4pm-7pm gap

- Last metric at 4pm (16:00)
- Server restart at 7pm (19:00)
- Gap = 3 hours = 36 five-minute intervals
- Create 36 interpolated entries between 4pm and 7pm

**Confidence:** HIGH - Straightforward database query pattern

## Interpolation Strategy Recommendation

### Options Analyzed

| Strategy | Pros | Cons | Use Case |
|----------|------|------|----------|
| **Linear interpolation** | Shows gradual transitions, realistic | May show false precision | Smooth metrics (CPU, Memory) |
| **Last-known value** | Simple, preserves last state | Flat line may be misleading | Connection counts |
| **Zero-fill** | Clear indication of gap | Misrepresents actual state | Already used for stopped instances |
| **Mark as interpolated** | Honest representation | Requires schema change | Future enhancement |

### Recommended: Linear Interpolation

For CPU and Memory metrics, linear interpolation provides the best balance:

```go
func linearInterpolate(startValue, endValue float64, totalSteps, currentStep int) float64 {
    if totalSteps <= 1 {
        return startValue
    }
    fraction := float64(currentStep) / float64(totalSteps-1)
    return startValue + (endValue-startValue)*fraction
}
```

**For Connections metric:** Use last-known value (connections are discrete counts, not continuous).

### Edge Cases

| Scenario | Handling |
|----------|----------|
| No previous data | Skip interpolation, wait for first real collection |
| Gap > 7 days | Cap at 7 days (retention limit), older would be deleted anyway |
| Instance was stopped | Check instance status during gap; if stopped, use zeros |
| First run ever | No interpolation needed |

**Confidence:** MEDIUM - Strategy choice involves trade-offs; linear interpolation is common but other approaches are valid.

## Architecture Patterns

### Modified Collection Flow

```
Every 15 minutes:
  collector.CollectAll()
    → For each running instance:
      → client.GetRDSMetricsHighRes(now-15min, now)  // NEW: 3 datapoints
        → getMetricMultiple() with Period=300s
        → Returns 3 datapoints (one per 5-minute interval)
      → For each datapoint:
        → storeMetric() with 5-minute truncated timestamp
```

### Startup Flow with Gap Detection

```
On server start:
  1. collector.DetectAndFillGaps()
     → Query last metric time per instance
     → Calculate gap duration
     → If gap > 15min: interpolate missing intervals
  2. collector.RunContinuous() // existing loop
```

### Recommended Structure Changes

```go
// cloudwatch.go - Add new method
type RDSMetricsMultiple struct {
    InstanceID string
    Datapoints []RDSMetricDatapoint
}

type RDSMetricDatapoint struct {
    Timestamp   time.Time
    CPU         *MetricValue
    Connections *MetricValue
    FreeMemory  *MetricValue
}

func (c *CloudWatchClient) GetRDSMetricsHighRes(ctx context.Context, dbInstanceID string, start, end time.Time) (*RDSMetricsMultiple, error)
```

```go
// collector.go - Add gap detection
func (c *MetricsCollector) DetectAndFillGaps(ctx context.Context) error
func (c *MetricsCollector) fillGap(ctx context.Context, instance models.Instance, gapStart, gapEnd time.Time) error
```

```go
// store.go - Add helper methods
func (s *MetricsStore) GetLatestMetricTime(ctx context.Context, instanceID string) (time.Time, error)
func (s *MetricsStore) UpsertInterpolatedMetric(ctx context.Context, m *models.HourlyMetric) error // marks as interpolated if needed
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CloudWatch API calls | Raw HTTP client | AWS SDK v2 `cloudwatch.Client` | Already in use, handles auth/retry/serialization |
| Timestamp truncation | Manual math | `time.Truncate(5 * time.Minute)` | Go stdlib, avoids edge cases |
| Database upserts | Manual INSERT/UPDATE | Existing `UpsertHourlyMetric` pattern | Already handles ON CONFLICT |

## Common Pitfalls

### Pitfall 1: Schema Column Name Mismatch
**What goes wrong:** `hour` column name implies hourly granularity, confusing when storing 5-minute data
**Why it happens:** Historical naming from Phase 10
**How to avoid:** Document that `hour` column stores timestamp truncated to period, not necessarily hourly. Consider renaming to `timestamp` in future migration.
**Warning signs:** Code comments mentioning "hour" when dealing with 5-minute data

### Pitfall 2: Inconsistent Timestamp Truncation
**What goes wrong:** Some code truncates to hour, some to 5 minutes, causing duplicate or missing data
**Why it happens:** Gradual migration, forgotten call sites
**How to avoid:** Create a constant or helper: `const MetricPeriod = 5 * time.Minute` and use consistently
**Warning signs:** Unique constraint violations, missing expected data

### Pitfall 3: CloudWatch Period vs Local Truncation Mismatch
**What goes wrong:** CloudWatch returns timestamps at period boundaries; local truncation differs
**Why it happens:** CloudWatch uses its own alignment rules
**How to avoid:** Trust CloudWatch-returned timestamps for storage, don't re-truncate
**Warning signs:** Metrics stored at unexpected timestamps

### Pitfall 4: Interpolation Creates False Confidence
**What goes wrong:** Users see smooth lines during outages, assume data is real
**Why it happens:** Linear interpolation looks authentic
**How to avoid:** Consider adding `is_interpolated` flag to schema (deferred), or document behavior
**Warning signs:** User confusion about "why CPU was steady during the outage"

### Pitfall 5: Gap Detection Query Performance
**What goes wrong:** Startup takes too long with many instances
**Why it happens:** N queries for N instances
**How to avoid:** Batch query: `SELECT instance_id, MAX(hour) FROM metrics_hourly GROUP BY instance_id`
**Warning signs:** Slow server startup

## Code Examples

### Change Period to 300 Seconds (cloudwatch.go)

```go
// Source: internal/metrics/cloudwatch.go lines 179-196
// BEFORE:
Period: aws.Int32(3600), // 1 hour

// AFTER:
Period: aws.Int32(300), // 5 minutes
```

### New Multi-Datapoint Fetch Method

```go
// Source: Pattern from existing GetRDSMetricsForHour
func (c *CloudWatchClient) GetRDSMetricsMultiple(ctx context.Context, dbInstanceID string, start, end time.Time) ([]RDSMetricDatapoint, error) {
    // Fetch with Period=300, returns multiple datapoints
    input := &cloudwatch.GetMetricStatisticsInput{
        Namespace:  aws.String("AWS/RDS"),
        MetricName: aws.String(metricName),
        Dimensions: []types.Dimension{{
            Name:  aws.String("DBInstanceIdentifier"),
            Value: aws.String(dbInstanceID),
        }},
        StartTime:  aws.Time(start),
        EndTime:    aws.Time(end),
        Period:     aws.Int32(300), // 5 minutes
        Statistics: []types.Statistic{
            types.StatisticAverage,
            types.StatisticMaximum,
            types.StatisticMinimum,
        },
    }
    
    output, err := c.client.GetMetricStatistics(ctx, input)
    // ... process all datapoints, not just the most recent one
}
```

### Gap Detection on Startup

```go
// Source: Pattern for collector.go
func (c *MetricsCollector) DetectAndFillGaps(ctx context.Context) error {
    log.Println("Checking for metric gaps...")
    
    instances, err := c.instanceStore.ListInstances(ctx)
    if err != nil {
        return err
    }
    
    for _, instance := range instances {
        if instance.Provider != "aws" {
            continue
        }
        
        lastTime, err := c.metricsStore.GetLatestMetricTime(ctx, instance.ID)
        if err != nil {
            // No previous data, skip gap detection
            continue
        }
        
        gapStart := lastTime.Add(5 * time.Minute)
        gapEnd := time.Now().UTC().Truncate(5 * time.Minute)
        
        gapDuration := gapEnd.Sub(gapStart)
        if gapDuration > 15*time.Minute {
            log.Printf("Gap detected for %s: %s to %s (%v)", 
                instance.Name, gapStart, gapEnd, gapDuration)
            if err := c.fillGap(ctx, instance, gapStart, gapEnd); err != nil {
                log.Printf("Failed to fill gap for %s: %v", instance.Name, err)
            }
        }
    }
    return nil
}
```

### Linear Interpolation Fill

```go
func (c *MetricsCollector) fillGap(ctx context.Context, instance models.Instance, gapStart, gapEnd time.Time) error {
    // Get boundary values for interpolation
    startMetrics, _ := c.metricsStore.GetMetricsAtTime(ctx, instance.ID, gapStart.Add(-5*time.Minute))
    endMetrics, _ := c.metricsStore.GetMetricsAtTime(ctx, instance.ID, gapEnd)
    
    intervals := int(gapEnd.Sub(gapStart) / (5 * time.Minute))
    
    for i := 0; i < intervals; i++ {
        timestamp := gapStart.Add(time.Duration(i) * 5 * time.Minute)
        
        for _, metricName := range []string{models.MetricCPUUtilization, models.MetricFreeableMemory} {
            startVal := getMetricValue(startMetrics, metricName)
            endVal := getMetricValue(endMetrics, metricName)
            
            interpolated := linearInterpolate(startVal, endVal, intervals, i)
            
            m := &models.HourlyMetric{
                InstanceID:  instance.ID,
                MetricName:  metricName,
                Hour:        timestamp,
                AvgValue:    interpolated,
                MaxValue:    interpolated,
                MinValue:    interpolated,
                SampleCount: 0, // 0 indicates interpolated
            }
            c.metricsStore.UpsertHourlyMetric(ctx, m)
        }
        
        // For connections, use last-known value
        connVal := getMetricValue(startMetrics, models.MetricDatabaseConnections)
        // ... store connection metric with connVal
    }
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Period=3600 (hourly) | Period=300 (5-min) | This phase | 12x more granular data |
| Single datapoint per collection | 3 datapoints per collection | This phase | Better coverage |
| No gap handling | Gap detection + interpolation | This phase | Continuous timeline |

**Note:** No schema changes required—`hour` column name is misleading but functional for 5-minute timestamps.

## Open Questions

1. **SampleCount for Interpolated Data**
   - What we know: Current code uses `SampleCount: 1` for real data
   - What's unclear: Best value for interpolated data (0? negative? separate flag?)
   - Recommendation: Use `SampleCount: 0` to indicate interpolated; document this convention

2. **Backfill Method Update**
   - What we know: `BackfillMetrics` exists for on-demand historical collection
   - What's unclear: Should it also use 5-minute periods?
   - Recommendation: Yes, update to match—but prioritize main collection first

3. **Frontend Display Impact**
   - What we know: Charts currently show hourly data
   - What's unclear: Do charts need updates for 5-minute granularity?
   - Recommendation: Charts should auto-adapt (more datapoints = smoother line), no changes needed

## Sources

### Primary (HIGH confidence)
- `internal/metrics/cloudwatch.go` - Current CloudWatch API usage patterns
- `internal/metrics/collector.go` - Current collection flow and BackfillMetrics pattern
- AWS CloudWatch GetMetricStatistics API documentation - Period constraints, retention rules
- `internal/metrics/store.go` - Current database patterns

### Secondary (MEDIUM confidence)
- AWS CloudWatch Metrics Concepts documentation - Data retention by period

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- CloudWatch API changes: HIGH - AWS documentation verified, existing patterns established
- Gap detection approach: HIGH - Standard database query patterns
- Interpolation strategy: MEDIUM - Trade-off decision, recommendation is sound but alternatives exist
- Schema compatibility: HIGH - Verified existing schema works without changes

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (30 days - stable domain)
