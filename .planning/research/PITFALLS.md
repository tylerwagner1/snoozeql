# Pitfalls Research: v1.2 Metrics & Recommendations

**Researched:** 2026-02-24
**Focus:** Common mistakes when adding metrics visualization and recommendation features to existing database management app

## Summary

The primary risks for v1.2 are **CloudWatch API throttling** (hitting rate limits during metrics collection for many instances), **misleading idle detection** (simplistic thresholds missing edge cases like batch jobs or replica instances), and **time-series chart performance** (rendering thousands of data points causing UI lag). Integration pitfalls include metric data gaps during instance state transitions and recommendation grouping logic that creates schedules conflicting with existing ones.

---

## Metrics Collection Pitfalls

### Pitfall 1: CloudWatch API Throttling

- **Risk:** CloudWatch GetMetricStatistics has a rate limit of 400 requests per second per account. With 100 instances × 4 metrics × 15-minute polling = 1,600 API calls/hour, which seems safe. But during catchup (e.g., after collector restart or backfill), parallel requests can hit throttling.

- **Warning signs:**
  - `LimitExceededException` errors in logs
  - Incomplete metrics data for some instances
  - Increasing latency in metrics collection jobs
  - Gaps in time-series charts

- **Prevention:**
  ```go
  // Existing code has retry with exponential backoff - good!
  // Add rate limiting for batch operations
  type RateLimitedCollector struct {
      limiter *rate.Limiter  // e.g., 10 req/sec to stay safe
  }
  
  // For catchup/backfill, process sequentially with pauses
  func (c *MetricsCollector) BackfillMetrics(instanceID string, days int) {
      for day := 0; day < days; day++ {
          c.collectDayMetrics(instanceID, day)
          time.Sleep(100 * time.Millisecond)  // Throttle self
      }
  }
  ```

- **Phase to address:** Phase 1 (Metrics Collection) — verify existing retry logic handles throttling gracefully

---

### Pitfall 2: Missing FreeableMemory to Memory% Conversion

- **Risk:** CloudWatch reports `FreeableMemory` in bytes, not memory utilization percentage. The v1.2 spec says "Memory utilization" but CloudWatch doesn't provide that directly for RDS. Teams calculate it wrong or show raw bytes.

- **Warning signs:**
  - Memory metric shows values like "34359738368" instead of "45%"
  - Memory percentage shows 98%+ when instance is healthy
  - Charts have wildly different Y-axis scales for CPU vs Memory

- **Prevention:**
  ```go
  // Get instance memory capacity from RDS DescribeDBInstances
  // or use lookup table by instance class
  var instanceMemoryGB = map[string]float64{
      "db.t3.micro":  1.0,
      "db.t3.small":  2.0,
      "db.t3.medium": 4.0,
      "db.r5.large":  16.0,
      // ... etc
  }
  
  // Calculate percentage
  memoryPercent := (1 - (freeableMemoryBytes / totalMemoryBytes)) * 100
  ```

- **Phase to address:** Phase 1 (Metrics Collection) — critical before storing memory metrics

---

### Pitfall 3: Stopped Instance Metric Gaps

- **Risk:** CloudWatch doesn't emit metrics for stopped instances. The current collector correctly skips stopped instances, but if an instance was stopped during the collection window, there's a gap. Time-series charts show jagged lines or misleading zero values.

- **Warning signs:**
  - Charts drop to zero during stopped periods (instead of showing no data)
  - Analyzer interprets gaps as "no data" vs "instance stopped"
  - Recommendations generated for already-stopped instances

- **Prevention:**
  ```typescript
  // Chart: Don't connect data points across gaps
  <LineChart connectNulls={false}>
  
  // API: Return explicit null for stopped periods
  type MetricPoint struct {
      Timestamp time.Time
      Value     *float64  // nil = no data (stopped)
      State     string    // "running", "stopped", "unknown"
  }
  
  // UI: Show stopped state differently
  {point.state === 'stopped' && (
      <span className="text-slate-500">Instance stopped</span>
  )}
  ```

- **Phase to address:** Phase 2 (Time-series Charts) — handle gaps explicitly in chart rendering

---

### Pitfall 4: 7-Day Retention vs 14-Day Analysis Window

- **Risk:** The v1.2 spec says "7-day retention" but the existing analyzer uses 14-day lookback. Reducing retention breaks pattern detection. Also, changing retention silently loses historical data that can't be recovered.

- **Warning signs:**
  - Analyzer returns "insufficient data" after retention change
  - Users complain recommendations stopped working
  - Time-series charts suddenly have shorter history

- **Prevention:**
  ```go
  // Document the tradeoff clearly
  // Option A: Keep 14-day retention (more storage, better analysis)
  // Option B: 7-day retention (less storage, simpler patterns)
  
  // If reducing, warn and migrate:
  func MigrateRetention(oldDays, newDays int) {
      if newDays < oldDays {
          log.Warn("Reducing retention will delete historical data",
              "deleting_days", oldDays - newDays)
      }
      // Update analyzer config to match
      analyzer.SetLookbackDays(newDays)
  }
  ```

- **Phase to address:** Phase 1 (Metrics Collection) — decide retention policy before implementation

---

### Pitfall 5: Hourly Aggregation Masks Spikes

- **Risk:** Current metrics store aggregates to hourly buckets with avg/min/max. But for idle detection, a 10-minute burst activity during an otherwise idle hour could be masked. The instance shows avg CPU 2% but was actually at 80% briefly.

- **Warning signs:**
  - Instances marked idle but users report "it was doing something"
  - Schedules stop instances mid-batch-job
  - Disconnect between user perception and metrics

- **Prevention:**
  ```go
  // Consider storing finer granularity OR additional signals
  type HourlyMetric struct {
      // Existing
      AvgValue    float64
      MaxValue    float64
      MinValue    float64
      SampleCount int
      // Add: detect variability
      StdDev      float64   // High stddev = variable workload
  }
  
  // In idle detection, consider max, not just avg
  isIdle := metrics.AvgCPU < threshold.CPUPercent && 
            metrics.MaxCPU < threshold.CPUPercent * 2  // Allow some headroom
  ```

- **Phase to address:** Phase 4 (Idle Detection) — consider max values in threshold logic

---

## Visualization Pitfalls

### Pitfall 6: Rendering Performance with Large Datasets

- **Risk:** 7 days × 24 hours × 4 metrics = 672 data points per instance. With 50 instances and multi-metric charts, React re-renders become slow. Recharts can struggle with >1000 points.

- **Warning signs:**
  - Chart hover/tooltip lags >200ms
  - Browser devtools show long "Scripting" times
  - Mobile devices freeze when viewing charts
  - `ResponsiveContainer` causes layout thrashing

- **Prevention:**
  ```typescript
  // Downsample for display
  function downsampleMetrics(data: MetricPoint[], targetPoints: number) {
      if (data.length <= targetPoints) return data;
      const step = Math.ceil(data.length / targetPoints);
      return data.filter((_, i) => i % step === 0);
  }
  
  // Use memo to prevent re-renders
  const chartData = useMemo(() => 
      downsampleMetrics(rawData, 100),
      [rawData]
  );
  
  // Disable animations for large datasets
  <AreaChart animationDuration={data.length > 100 ? 0 : 300}>
  ```

- **Phase to address:** Phase 2 (Time-series Charts) — test performance early with realistic data volumes

---

### Pitfall 7: Timezone Confusion in Charts

- **Risk:** Metrics stored in UTC, displayed without timezone conversion. User in PST sees "idle from 22:00 to 06:00" but that's UTC — actually 14:00-22:00 local time, completely wrong.

- **Warning signs:**
  - Recommendations suggest sleeping during business hours
  - Time-series X-axis shows times that don't match user expectations
  - Confusion between "hour 0" meaning midnight local vs midnight UTC

- **Prevention:**
  ```typescript
  // Always convert to user timezone for display
  const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  
  const formatHour = (utcHour: Date) => {
      return new Intl.DateTimeFormat('en-US', {
          hour: 'numeric',
          timeZone: userTimezone,
      }).format(utcHour);
  };
  
  // Label charts clearly
  <XAxis 
      tickFormatter={formatHour}
  />
  <ChartCaption>
      All times shown in {userTimezone}
  </ChartCaption>
  ```

- **Phase to address:** Phase 2 (Time-series Charts) — ensure consistent timezone handling

---

### Pitfall 8: Y-Axis Scale Inconsistency

- **Risk:** Auto-scaling Y-axis makes 2% CPU look as dramatic as 80% CPU depending on the data range. Users misinterpret small variations as significant.

- **Warning signs:**
  - "Spike" in chart is actually 1% → 3% (trivial)
  - Same chart looks different day-to-day
  - Users make wrong decisions based on visual drama

- **Prevention:**
  ```typescript
  // Fix scale for percentage metrics
  <YAxis 
      domain={[0, 100]}  // Always 0-100% for CPU/Memory
      ticks={[0, 25, 50, 75, 100]}
  />
  
  // Or use sensible fixed scale
  <YAxis
      domain={[0, 50]}  // For "mostly idle" instances
      allowDataOverflow={true}  // Clip extreme values
  />
  
  // Add reference lines for context
  <ReferenceLine y={5} stroke="green" label="Idle threshold" />
  ```

- **Phase to address:** Phase 2 (Time-series Charts) — establish consistent scale conventions

---

### Pitfall 9: Chart Empty States

- **Risk:** Instance has no metrics yet (just discovered), chart shows blank area. User doesn't know if it's loading, broken, or just empty.

- **Warning signs:**
  - White/blank chart areas
  - Users refresh repeatedly thinking it's stuck
  - Support tickets about "broken charts"

- **Prevention:**
  ```typescript
  // Explicit states
  {loading && <ChartSkeleton />}
  {!loading && error && <ChartError message={error} />}
  {!loading && !error && data.length === 0 && (
      <ChartEmpty>
          <p>No metrics available yet</p>
          <p className="text-sm text-slate-500">
              Metrics are collected every 15 minutes. 
              Check back soon!
          </p>
      </ChartEmpty>
  )}
  {!loading && !error && data.length > 0 && (
      <ActualChart data={data} />
  )}
  ```

- **Phase to address:** Phase 2 (Time-series Charts) — handle all states explicitly

---

## Recommendation Pitfalls

### Pitfall 10: Overly Simplistic Idle Detection

- **Risk:** Current threshold is "CPU < 1%". But database replicas often show 0% CPU while serving read traffic. Batch processing databases show 0% CPU 23 hours/day then spike. Simple thresholds miss these patterns.

- **Warning signs:**
  - Recommendations to sleep read replicas (breaks failover)
  - Recommendations ignore weekend-only or monthly jobs
  - High false-positive rate for schedule suggestions

- **Prevention:**
  ```go
  // Enhanced detection criteria
  type IdleDetection struct {
      CPUThreshold         float64  // < 5%
      ConnectionsThreshold int      // == 0 (not just low)
      IOPSThreshold        float64  // < 10 combined
      ConsecutiveHours     int      // 8+ hours of all criteria
      ConsistentDays       int      // 3+ days with same pattern
  }
  
  // Exclude replica instances
  func ShouldAnalyze(instance Instance) bool {
      if strings.Contains(instance.Name, "replica") ||
         strings.Contains(instance.Name, "read") {
          return false  // Skip replicas
      }
      return true
  }
  ```

- **Phase to address:** Phase 4 (Idle Detection) — refine detection algorithm before wide rollout

---

### Pitfall 11: Recommendation-Schedule Conflicts

- **Risk:** User has existing schedule for instance. New recommendation suggests different schedule. Confirming recommendation creates conflicting schedule, or worse, overrides existing one silently.

- **Warning signs:**
  - Instance has multiple active schedules
  - Schedule conflicts causing unexpected wake/sleep
  - Users lose custom schedules they created

- **Prevention:**
  ```go
  // Check for existing schedules before creating recommendation
  func (a *Analyzer) GenerateRecommendations() ([]Recommendation, error) {
      for _, pattern := range patterns {
          // Check if instance already has active schedule
          existing := a.scheduleStore.GetSchedulesForInstance(pattern.InstanceID)
          if len(existing) > 0 {
              // Skip or flag for review
              log.Info("Skipping instance with existing schedule",
                  "instance", pattern.InstanceID,
                  "existing_schedule", existing[0].Name)
              continue
          }
      }
  }
  
  // In UI, show existing schedule
  {recommendation.hasExistingSchedule && (
      <Warning>
          This instance already has schedule "{existingSchedule.name}".
          Confirming will create a new schedule that may conflict.
      </Warning>
  )}
  ```

- **Phase to address:** Phase 5 (Grouped Recommendations) — check for conflicts in recommendation generation

---

### Pitfall 12: Grouped Recommendations Create Wrong Schedules

- **Risk:** Grouping instances by "similar patterns" sounds good, but instances in same group may need slightly different schedules. Group schedule picks one time, suboptimal for others.

- **Warning signs:**
  - Grouped schedule wakes instance 2 hours before it's needed
  - Some instances in group have different timezone users
  - Users want to modify just one instance in group

- **Prevention:**
  ```go
  // Allow per-instance overrides in group
  type GroupedRecommendation struct {
      GroupName     string
      DefaultSchedule Schedule
      Instances     []struct {
          ID       string
          Override *Schedule  // nil = use default
      }
  }
  
  // In UI, show individual instance details
  <GroupRecommendation>
      <h3>Suggested for 5 dev databases</h3>
      <DefaultSchedule schedule={rec.defaultSchedule} />
      <details>
          <summary>View individual instances</summary>
          {rec.instances.map(inst => (
              <InstanceRow 
                  instance={inst}
                  onOverride={(schedule) => handleOverride(inst.id, schedule)}
              />
          ))}
      </details>
  </GroupRecommendation>
  ```

- **Phase to address:** Phase 5 (Grouped Recommendations) — design per-instance override UX

---

### Pitfall 13: Stale Recommendations

- **Risk:** Recommendation generated Monday based on last week's data. User reviews Friday. Data has changed — pattern no longer valid, but recommendation still shown.

- **Warning signs:**
  - Accepted recommendations create schedules that immediately cause issues
  - "Detected pattern" in recommendation doesn't match current behavior
  - Users don't trust recommendations

- **Prevention:**
  ```go
  // Add staleness check
  type Recommendation struct {
      GeneratedAt    time.Time
      DataWindowEnd  time.Time
      IsStale        bool  // Computed field
  }
  
  func (r *Recommendation) CheckStaleness() bool {
      // Stale if generated > 3 days ago
      return time.Since(r.GeneratedAt) > 72 * time.Hour
  }
  
  // In UI, warn about stale recommendations
  {recommendation.isStale && (
      <Warning>
          This recommendation is based on data from {recommendation.dataWindowEnd}.
          Consider regenerating recommendations.
      </Warning>
  )}
  ```

- **Phase to address:** Phase 3 (Recommendation Engine) — add staleness tracking

---

### Pitfall 14: Recommendation Confidence Theater

- **Risk:** Showing "85% confidence" sounds precise but the confidence calculation is arbitrary (existing code adds 0.1 for this, 0.2 for that). Users trust it more than warranted.

- **Warning signs:**
  - All recommendations show 70-90% confidence (no real differentiation)
  - Users skip review because "85% is good enough"
  - Confidence doesn't correlate with actual success rate

- **Prevention:**
  ```go
  // Either make confidence meaningful or simplify
  // Option A: Calibrate to actual outcomes (requires tracking)
  // Option B: Simple categories
  type ConfidenceLevel string
  const (
      ConfidenceHigh   = "high"    // 5+ consistent days, clear pattern
      ConfidenceMedium = "medium"  // 3-4 days, some variance
      ConfidenceLow    = "low"     // Detectable pattern but uncertain
  )
  
  // In UI, explain what confidence means
  <ConfidenceBadge level="high">
      <Tooltip>
          Pattern detected consistently for 5+ days. 
          High likelihood this schedule matches actual usage.
      </Tooltip>
  </ConfidenceBadge>
  ```

- **Phase to address:** Phase 3 (Recommendation Engine) — calibrate or simplify confidence display

---

## Integration Pitfalls

### Pitfall 15: Existing 7-Day Auto-Restart Not Addressed

- **Risk:** PROJECT.md notes "AWS 7-day auto-restart: implement re-stop mechanism" as deferred. If recommendations create 7-day sleep schedules, instances auto-restart and recommendation value is undermined.

- **Warning signs:**
  - Users report instances waking unexpectedly on day 7
  - Savings less than projected
  - Audit log shows no wake event but instance is running

- **Prevention:**
  ```go
  // In schedule creation, warn about 7-day limit
  func ValidateSchedule(schedule Schedule) error {
      maxSleepDuration := calculateMaxSleepDuration(schedule)
      if maxSleepDuration > 7 * 24 * time.Hour {
          return fmt.Errorf(
              "schedule would keep instance stopped for %v, "+
              "but AWS auto-restarts after 7 days", 
              maxSleepDuration)
      }
      return nil
  }
  
  // Add disclaimer in recommendations
  <RecommendationNote>
      Note: AWS automatically restarts RDS instances after 7 consecutive days stopped.
      This schedule includes wake cycles to prevent unexpected restarts.
  </RecommendationNote>
  ```

- **Phase to address:** Phase 3 or Phase 5 — ensure schedules respect 7-day limit

---

### Pitfall 16: Instance State Race Conditions

- **Risk:** PROJECT.md notes "Instance state race conditions: implement proper state machine" as deferred. Recommendations that trigger rapid stop/start could hit these race conditions.

- **Warning signs:**
  - API calls fail with "instance is not in valid state"
  - Instance stuck in "stopping" for extended periods
  - Duplicate events in audit log

- **Prevention:**
  ```go
  // Before creating schedule from recommendation, verify instance state
  func ConfirmRecommendation(rec Recommendation) error {
      instance := instanceStore.Get(rec.InstanceID)
      if instance.Status != "available" && instance.Status != "running" {
          return fmt.Errorf(
              "instance %s is in state %s, cannot create schedule",
              rec.InstanceID, instance.Status)
      }
      // Proceed with schedule creation
  }
  
  // In scheduler, add state verification
  func (s *Scheduler) ExecuteStop(instanceID string) error {
      instance := s.store.Get(instanceID)
      if instance.Status != "available" && instance.Status != "running" {
          log.Warn("Skipping stop, instance not in stoppable state",
              "instance", instanceID, "state", instance.Status)
          return nil  // Skip gracefully, don't error
      }
      return s.provider.Stop(instanceID)
  }
  ```

- **Phase to address:** All phases — defensive coding around instance state

---

### Pitfall 17: API Endpoint Bloat

- **Risk:** Adding metrics + recommendations features could add 6-10 new API endpoints. Without planning, naming becomes inconsistent and frontend has complex integration.

- **Warning signs:**
  - `/api/metrics`, `/api/instance-metrics`, `/api/instances/{id}/metrics` all exist
  - Frontend imports from multiple endpoint files
  - Swagger/OpenAPI doc becomes confusing

- **Prevention:**
  ```go
  // Plan endpoints upfront
  // Metrics:
  //   GET /api/instances/{id}/metrics       - latest metrics
  //   GET /api/instances/{id}/metrics/series?start=&end= - time series
  
  // Recommendations:
  //   GET /api/recommendations              - list (existing)
  //   POST /api/recommendations/generate    - trigger generation (existing)
  //   GET /api/recommendations/{id}         - detail
  //   POST /api/recommendations/{id}/confirm - accept (existing)
  //   DELETE /api/recommendations/{id}      - dismiss (existing)
  
  // Don't add:
  //   GET /api/metrics  (global, unused)
  //   GET /api/recommendations/pending (use query param instead)
  ```

- **Phase to address:** Phase 1 — design API structure before implementation

---

### Pitfall 18: Memory Usage During Pattern Analysis

- **Risk:** Loading 7 days × 24 hours × 4 metrics × 100 instances into memory for analysis = ~67,200 metric points. In Go that's manageable, but careless code can 10x memory usage.

- **Warning signs:**
  - Container OOM kills during recommendation generation
  - Analysis takes >30 seconds
  - GC pauses visible in logs

- **Prevention:**
  ```go
  // Process instances in batches, not all at once
  func (a *Analyzer) GenerateRecommendations() ([]Recommendation, error) {
      instanceIDs := a.GetManagedInstanceIDs()
      
      var results []Recommendation
      batchSize := 10
      
      for i := 0; i < len(instanceIDs); i += batchSize {
          end := min(i + batchSize, len(instanceIDs))
          batch := instanceIDs[i:end]
          
          batchResults := a.analyzeBatch(batch)
          results = append(results, batchResults...)
      }
      return results, nil
  }
  
  // Stream metrics from DB instead of loading all
  func (s *MetricsStore) StreamMetrics(ctx context.Context, instanceID string, cb func(m HourlyMetric)) error {
      rows, _ := s.db.Query(ctx, query, instanceID)
      defer rows.Close()
      for rows.Next() {
          var m HourlyMetric
          rows.Scan(&m)
          cb(m)  // Process immediately, don't accumulate
      }
      return nil
  }
  ```

- **Phase to address:** Phase 3 (Recommendation Engine) — profile memory during analysis

---

## Confidence

**MEDIUM** - Based on:
- Existing codebase analysis (HIGH confidence on current patterns)
- CloudWatch documentation (HIGH confidence on API behavior)
- Common time-series visualization issues (MEDIUM - general patterns, not SnoozeQL-specific)
- Recommendation system anti-patterns (MEDIUM - based on typical FinOps tool challenges)

**Verification needed:**
- CloudWatch rate limits for specific account tier
- Recharts performance thresholds with actual data volume
- User expectations around recommendation freshness

---

## Phase-Specific Summary

| Phase | High-Risk Pitfalls | Mitigation Priority |
|-------|-------------------|---------------------|
| Phase 1: Metrics Collection | #1 (throttling), #2 (memory %), #4 (retention) | Add rate limiting, fix memory calculation |
| Phase 2: Time-series Charts | #6 (performance), #7 (timezone), #9 (empty states) | Test with realistic data, convert to local TZ |
| Phase 3: Recommendation Engine | #13 (staleness), #14 (confidence), #18 (memory) | Add staleness tracking, simplify confidence |
| Phase 4: Idle Detection | #5 (spikes masked), #10 (simplistic detection) | Use max values, require connections=0 |
| Phase 5: Grouped Recommendations | #11 (conflicts), #12 (wrong schedules) | Check existing schedules, allow overrides |

---

*Researched for SnoozeQL v1.2 - Metrics & Recommendations*
*Last updated: 2026-02-24*
