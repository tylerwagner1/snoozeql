# Phase 5: Activity Analysis - Research

**Researched:** 2026-02-23
**Domain:** AWS CloudWatch metrics collection, time-series storage, pattern detection
**Confidence:** HIGH

## Summary

Phase 5 implements metrics collection from AWS CloudWatch for RDS instances and pattern detection to identify low-activity periods suitable for sleep scheduling. The codebase already has a partial `GetMetrics` implementation in the provider interface and an analyzer skeleton.

**Key findings:**
1. AWS SDK v2 `cloudwatch` package provides `GetMetricStatistics` for fetching RDS metrics with configurable periods and statistics
2. The existing codebase has stub implementations for metrics in `internal/provider/aws/rds.go` that return nil - these need to be completed with actual CloudWatch API calls
3. CloudWatch has API rate limits (400 requests/sec for GetMetricStatistics, 500 for GetMetricData) that require batching and backoff strategies
4. Pattern detection requires storing hourly aggregates per instance (as decided in CONTEXT.md) and analyzing for contiguous low-activity windows

**Primary recommendation:** Use `github.com/aws/aws-sdk-go-v2/service/cloudwatch` with `GetMetricStatistics` to fetch RDS metrics (CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS), store hourly aggregates in a new `metrics` table, and implement sliding-window pattern detection for idle periods.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| aws-sdk-go-v2/service/cloudwatch | v1.54.0 | CloudWatch API client | Official AWS SDK, already using v2 for RDS |
| jackc/pgx/v5 | v5.8.0 | PostgreSQL driver | Already in use for other stores |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| time (stdlib) | - | Time calculations | Duration, scheduling windows |
| context (stdlib) | - | Request cancellation | API timeouts, graceful shutdown |
| sync (stdlib) | - | Concurrent access | Background collection ticker |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| GetMetricStatistics | GetMetricData | GetMetricData is more flexible (multiple metrics per call) but more complex; GetMetricStatistics is simpler for single-metric queries |
| Per-poll storage | CloudWatch Metric Streams | Streams provide push-based delivery but add infrastructure complexity (Kinesis/Firehose) |

**Installation:**
```bash
go get github.com/aws/aws-sdk-go-v2/service/cloudwatch@v1.54.0
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── metrics/
│   ├── collector.go       # Background metric collection service
│   ├── cloudwatch.go      # AWS CloudWatch client wrapper
│   └── store.go           # MetricsStore for persistence
├── analyzer/
│   ├── analyzer.go        # Pattern detection (existing, extend)
│   └── patterns.go        # Idle period detection algorithms
├── provider/
│   └── aws/
│       └── rds.go         # Update GetMetrics implementation
└── store/
    └── postgres.go        # Add metrics queries
```

### Pattern 1: Background Collection Service
**What:** A MetricsCollector that runs on a ticker (15-minute interval per CONTEXT.md), iterates instances, and fetches/stores metrics.
**When to use:** For periodic polling of external APIs
**Example:**
```go
// Source: Existing discovery service pattern in internal/discovery/discovery.go
type MetricsCollector struct {
    registry      *provider.Registry
    metricsStore  *MetricsStore
    instanceStore *store.InstanceStore
    interval      time.Duration
}

func (c *MetricsCollector) RunContinuous(ctx context.Context) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := c.collectAll(ctx); err != nil {
                log.Printf("Metrics collection failed: %v", err)
            }
        }
    }
}
```

### Pattern 2: Retry with Exponential Backoff
**What:** Retry failed CloudWatch API calls 3 times (per CONTEXT.md) with increasing delays
**When to use:** For resilient API calls that may experience transient failures or rate limits
**Example:**
```go
// Source: AWS SDK best practices
func (c *CloudWatchClient) getMetricsWithRetry(ctx context.Context, input *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        output, err := c.client.GetMetricStatistics(ctx, input)
        if err == nil {
            return output, nil
        }
        lastErr = err
        
        // Check for throttling
        var throttle *types.LimitExceededException
        if errors.As(err, &throttle) {
            time.Sleep(time.Duration(1<<attempt) * time.Second)
            continue
        }
        return nil, err // Non-retryable error
    }
    return nil, fmt.Errorf("failed after 3 retries: %w", lastErr)
}
```

### Pattern 3: Hourly Aggregation on Insert
**What:** Compute hourly averages when storing raw metrics, using UPSERT to update existing hour buckets
**When to use:** For efficient time-series storage with bounded growth
**Example:**
```go
// Source: Standard time-series aggregation pattern
func (s *MetricsStore) UpsertHourlyMetric(ctx context.Context, m *HourlyMetric) error {
    query := `
        INSERT INTO metrics_hourly (instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count)
        VALUES ($1, $2, date_trunc('hour', $3), $4, $5, $6, $7)
        ON CONFLICT (instance_id, metric_name, hour) DO UPDATE SET
            avg_value = (metrics_hourly.avg_value * metrics_hourly.sample_count + EXCLUDED.avg_value) 
                        / (metrics_hourly.sample_count + 1),
            max_value = GREATEST(metrics_hourly.max_value, EXCLUDED.max_value),
            min_value = LEAST(metrics_hourly.min_value, EXCLUDED.min_value),
            sample_count = metrics_hourly.sample_count + 1
    `
    _, err := s.db.Exec(ctx, query, m.InstanceID, m.MetricName, m.Timestamp, 
                        m.AvgValue, m.MaxValue, m.MinValue, m.SampleCount)
    return err
}
```

### Anti-Patterns to Avoid
- **Storing raw 15-minute samples:** Creates unbounded table growth; use hourly aggregates
- **Fetching all metrics in one goroutine:** Will hit rate limits; parallelize per-account but respect limits
- **Ignoring instance status:** Don't fetch metrics for stopped instances (no metrics available)
- **Hardcoded thresholds:** Make low-activity thresholds configurable via settings table

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CloudWatch API client | HTTP client + auth | aws-sdk-go-v2/service/cloudwatch | Handles signing, retries, pagination |
| Time bucketing | Manual timestamp math | PostgreSQL `date_trunc('hour', timestamp)` | Handles timezones, DST correctly |
| Exponential backoff | Custom sleep loops | SDK retry config or simple 1<<attempt pattern | Standard pattern for AWS rate limits |
| Concurrent collection | Manual goroutine management | errgroup with semaphore | Cleaner error handling, limit control |

**Key insight:** The AWS SDK already handles most of the complexity (auth, serialization, error types). Focus implementation on the domain logic (what metrics, how to aggregate, pattern detection).

## Common Pitfalls

### Pitfall 1: CloudWatch Data Retention Affects Historical Queries
**What goes wrong:** Requesting 14 days of 1-minute data fails silently (returns partial data)
**Why it happens:** CloudWatch retains 1-minute data for only 15 days, 5-minute for 63 days, 1-hour for 455 days
**How to avoid:** Use 5-minute period (300 seconds) for queries beyond 15 days; document in code
**Warning signs:** Datapoint arrays shorter than expected

### Pitfall 2: RDS Metrics Require DBInstanceIdentifier Dimension
**What goes wrong:** Queries return empty results
**Why it happens:** CloudWatch RDS metrics are namespaced with `AWS/RDS` and require the `DBInstanceIdentifier` dimension
**How to avoid:** Always include the dimension in queries:
```go
Dimensions: []types.Dimension{
    {Name: aws.String("DBInstanceIdentifier"), Value: aws.String(instanceName)},
}
```
**Warning signs:** Empty Datapoints array despite instance being active

### Pitfall 3: Metrics Unavailable for Stopped Instances
**What goes wrong:** GetMetricStatistics returns empty for stopped instances
**Why it happens:** CloudWatch only records metrics when the instance is running
**How to avoid:** Skip metrics collection for instances with status != "available"; fall back to existing schedule per CONTEXT.md
**Warning signs:** Consistently empty metrics for some instances

### Pitfall 4: Rate Limit Throttling with Many Instances
**What goes wrong:** Requests start failing with throttling errors
**Why it happens:** 400 requests/second limit shared across account; collecting 6 metrics for 100 instances = 600 requests
**How to avoid:** 
1. Batch requests using GetMetricData (up to 500 metrics per call)
2. Implement per-account rate limiting (e.g., 10 concurrent requests)
3. Add jitter between collection cycles
**Warning signs:** `LimitExceededException` errors in logs

### Pitfall 5: Timezone Confusion in Pattern Detection
**What goes wrong:** Detected "night" periods are offset by several hours
**Why it happens:** CloudWatch returns UTC timestamps; pattern detection needs instance timezone
**How to avoid:** 
1. Store metrics with UTC timestamps
2. Convert to instance timezone (from schedule.Timezone) during pattern analysis
**Warning signs:** Sleep recommendations for business hours

## Code Examples

Verified patterns from official sources:

### CloudWatch GetMetricStatistics Call
```go
// Source: AWS SDK for Go v2 documentation
import (
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func (c *CloudWatchClient) getCPUUtilization(ctx context.Context, instanceID string, start, end time.Time) ([]types.Datapoint, error) {
    input := &cloudwatch.GetMetricStatisticsInput{
        Namespace:  aws.String("AWS/RDS"),
        MetricName: aws.String("CPUUtilization"),
        Dimensions: []types.Dimension{
            {
                Name:  aws.String("DBInstanceIdentifier"),
                Value: aws.String(instanceID),
            },
        },
        StartTime:  aws.Time(start),
        EndTime:    aws.Time(end),
        Period:     aws.Int32(3600), // 1 hour for hourly aggregates
        Statistics: []types.Statistic{types.StatisticAverage, types.StatisticMaximum},
    }
    
    output, err := c.client.GetMetricStatistics(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("GetMetricStatistics failed: %w", err)
    }
    
    return output.Datapoints, nil
}
```

### Database Schema for Metrics Storage
```sql
-- Source: PostgreSQL time-series best practices
CREATE TABLE metrics_hourly (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    metric_name VARCHAR(50) NOT NULL, -- CPUUtilization, DatabaseConnections, etc.
    hour TIMESTAMPTZ NOT NULL,        -- Truncated to hour in UTC
    avg_value FLOAT NOT NULL,
    max_value FLOAT NOT NULL,
    min_value FLOAT NOT NULL,
    sample_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(instance_id, metric_name, hour)
);

-- Index for efficient time-range queries
CREATE INDEX idx_metrics_hourly_instance_time ON metrics_hourly(instance_id, hour DESC);
CREATE INDEX idx_metrics_hourly_hour ON metrics_hourly(hour DESC);

-- Retention cleanup (run daily)
-- DELETE FROM metrics_hourly WHERE hour < NOW() - INTERVAL '14 days';
```

### Pattern Detection: Find Low Activity Windows
```go
// Source: Standard sliding window algorithm
type ActivityWindow struct {
    StartHour int  // 0-23
    EndHour   int  // 0-23 (can be < start for overnight)
    DayOfWeek []time.Weekday
    AvgCPU    float64
    AvgConns  float64
}

func (a *Analyzer) findIdleWindows(metrics []HourlyMetric, threshold ThresholdConfig) []ActivityWindow {
    // Group by day of week and hour
    buckets := make(map[time.Weekday]map[int][]float64) // weekday -> hour -> cpuValues
    
    for _, m := range metrics {
        dow := m.Hour.Weekday()
        hour := m.Hour.Hour()
        if buckets[dow] == nil {
            buckets[dow] = make(map[int][]float64)
        }
        if m.MetricName == "CPUUtilization" {
            buckets[dow][hour] = append(buckets[dow][hour], m.AvgValue)
        }
    }
    
    // Find contiguous hours where avg CPU < threshold (1% per CONTEXT.md)
    var windows []ActivityWindow
    for dow, hours := range buckets {
        inWindow := false
        var window ActivityWindow
        
        for h := 0; h < 24; h++ {
            values := hours[h]
            avg := average(values)
            
            if avg < threshold.CPUPercent && len(values) > 0 {
                if !inWindow {
                    window = ActivityWindow{StartHour: h, DayOfWeek: []time.Weekday{dow}}
                    inWindow = true
                }
                window.AvgCPU = (window.AvgCPU*float64(h-window.StartHour) + avg) / float64(h-window.StartHour+1)
            } else if inWindow {
                window.EndHour = h
                if h-window.StartHour >= 8 { // 8+ hours per CONTEXT.md
                    windows = append(windows, window)
                }
                inWindow = false
            }
        }
    }
    
    return windows
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| aws-sdk-go v1 | aws-sdk-go-v2 | 2021 | Better modularity, cleaner API |
| GetMetricStatistics only | GetMetricData preferred | 2018 | More flexible, supports expressions |
| Manual pagination | SDK paginator types | v2 | Cleaner iteration over results |

**Deprecated/outdated:**
- `aws-sdk-go` (v1): Still works but v2 is recommended for new projects
- `GetMetricStatistics` for bulk queries: Use `GetMetricData` for fetching multiple metrics efficiently

## Open Questions

Things that couldn't be fully resolved:

1. **GetMetricData vs GetMetricStatistics for this use case**
   - What we know: GetMetricData supports multiple metrics per call (up to 500), better for batch operations
   - What's unclear: Exact performance difference for our 6-metric-per-instance use case
   - Recommendation: Start with GetMetricStatistics (simpler); refactor to GetMetricData if rate limiting becomes an issue

2. **Optimal collection window alignment**
   - What we know: Collecting every 15 minutes as decided
   - What's unclear: Whether to align to clock time (00, 15, 30, 45) or use offset for distributed load
   - Recommendation: Use clock alignment for predictable aggregation, add jitter between instances

3. **Pattern detection sensitivity tuning**
   - What we know: 8+ hours of low activity, CPU < 1% thresholds from CONTEXT.md
   - What's unclear: How to handle instances with variable workloads (some days active, some not)
   - Recommendation: Require pattern consistency across multiple days (e.g., 3/7 days)

## Sources

### Primary (HIGH confidence)
- AWS SDK for Go v2 cloudwatch package - pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/cloudwatch
- AWS CloudWatch RDS Metrics documentation - docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-metrics.html
- AWS CloudWatch GetMetricStatistics API - docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricStatistics.html
- AWS CloudWatch Service Quotas - docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html

### Secondary (MEDIUM confidence)
- Existing codebase patterns in internal/provider/aws/rds.go, internal/discovery/discovery.go
- PostgreSQL time-series patterns (standard practice)

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using already-integrated AWS SDK, following existing codebase patterns
- Architecture: HIGH - Mirrors existing discovery service pattern
- API usage: HIGH - Official AWS documentation verified
- Pitfalls: HIGH - Documented in official AWS sources
- Pattern detection: MEDIUM - Domain-specific algorithms not from authoritative source

**Research date:** 2026-02-23
**Valid until:** 2026-03-23 (30 days - AWS SDK and APIs are stable)
