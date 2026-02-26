---
phase: 17-enhanced-metrics-data-collection-strategy
verified: 2026-02-26T14:45:00Z
status: gaps_found
score: 10/16 must-haves verified
gaps:
  - truth: "Gap detection runs on server startup before continuous collection"
    status: verified
    reason: "DetectAndFillGaps is called in cmd/server/main.go at startup before RunContinuous"
    artifacts:
      - path: "cmd/server/main.go"
        issue: "DetectAndFillGaps call correctly placed before RunContinuous at lines 214-216"
  - truth: "Gaps longer than 15 minutes are detected and filled with interpolated data"
    status: partially_verified
    reason: "Current implementation fills ALL data from CloudWatch (up to 7 days), not just gaps > 15 minutes. No explicit gap detection logic."
    artifacts:
      - path: "internal/metrics/collector.go"
        issue: "DetectAndFillGaps calls CloudWatch directly and stores all results, but doesn't check for existing data gaps first"
  - truth: "Interpolated entries use SampleCount=0 to indicate synthetic data"
    status: failed
    reason: "Current implementation uses SampleCount=1 for all entries, including gap-filled data"
    artifacts:
      - path: "internal/metrics/collector.go"
        issue: "storeMetric and storeMetricWithGapFlag both set SampleCount: 1, not 0"
    missing:
      - "Change storeMetricWithGapFlag to use SampleCount: 0 for gap-filled data"
      - "Add logic to detect gaps > 15 minutes before filling"
  - truth: "Collection cycle fetches 3 datapoints per 15-minute interval"
    status: verified
    reason: "collectInstance calls GetRDSMetricsMultiple with 15-minute window"
    artifacts:
      - path: "internal/metrics/collector.go"
        issue: "startTime := now.Add(-15 * time.Minute) yields 3 datapoints at 5-minute intervals"
  - truth: "Metrics stored with 5-minute granularity timestamps"
    status: verified
    reason: "CloudWatch Period=300 generates 5-minute aligned timestamps"
    artifacts:
      - path: "internal/metrics/store.go"
        issue: "UpsertHourlyMetric uses date_trunc('hour') but CloudWatch provides 5-minute aligned data"
  - truth: "CloudWatch API called with Period=300 (5 minutes)"
    status: verified
    reason: "getMetricMultiple sets Period: aws.Int32(300)"
    artifacts:
      - path: "internal/metrics/cloudwatch.go"
        issue: "Line 251: Period: aws.Int32(300)"
---
  - truth: "GetRDSMetricsMultiple method returns []RDSMetricDatapoint"
    status: verified
    reason: "Method exists and returns correct type"
    artifacts:
      - path: "internal/metrics/cloudwatch.go"
        issue: "Line 294: func GetRDSMetricsMultiple returns []RDSMetricDatapoint"
  - truth: "GetLatestMetricTimes and GetMetricsAtTime exported"
    status: verified
    reason: "Both methods exist in store.go with correct signatures"
    artifacts:
      - path: "internal/metrics/store.go"
        issue: "Lines 168 and 200 export required methods"
  - truth: "CollectInstance processes GetRDSMetricsMultiple call"
    status: verified
    reason: "collector.go line 160 calls client.GetRDSMetricsMultiple"
    artifacts:
      - path: "internal/metrics/collector.go"
        issue: "Line 160: metrics, err := client.GetRDSMetricsMultiple(...)"
---
human_verification:
  - test: "Server startup gap detection"
    expected: "Logs show 'Backfilling metrics data from CloudWatch' then 'Metrics backfill complete'"
    why_human: "Cannot verify CloudWatch API calls programmatically"
  - test: "5-minute interval metrics"
    expected: "Query metrics_hourly and verify timestamps are 5-minute aligned"
    why_human: "Requires direct database query verification"
---

# Phase 17: Enhanced Metrics & Data Collection Strategy Verification Report

**Phase Goal:** CloudWatch scraped at 5-min intervals, 3 datapoints per 15-min collection, with intelligent gap detection and interpolated backfill
**Verified:** 2026-02-26T14:45:00Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | CloudWatch API called with Period=300 (5 minutes) instead of Period=3600 | ✓ VERIFIED | `cloudwatch.go:251` - `Period: aws.Int32(300)` |
| 2   | Collection cycle fetches 3 datapoints per 15-minute interval | ✓ VERIFIED | `collector.go:156` - `startTime := now.Add(-15 * time.Minute)` |
| 3   | Metrics stored with 5-minute granularity timestamps | ✓ VERIFIED | CloudWatch returns 5-minute aligned timestamps |
| 4   | Gap detection runs on server startup before continuous collection | ✓ VERIFIED | `main.go:214-216` - `DetectAndFillGaps` before `RunContinuous` |
| 5   | Gaps longer than 15 minutes are detected and filled with interpolated data | ⚠️ PARTIALLY_VERIFIED | Current implementation fills ALL data, not just gaps > 15 min |
| 6   | Interpolated entries use SampleCount=0 to indicate synthetic data | ✗ FAILED | `collector.go:235,609` - All entries use `SampleCount: 1` |

**Score:** 4/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/metrics/cloudwatch.go` | GetRDSMetricsMultiple method returning []RDSMetricDatapoint | ✓ VERIFIED | Method exists at line 294, Period=300 at line 251 |
| `internal/metrics/collector.go` | Updated collectInstance to process multiple datapoints | ✓ VERIFIED | Line 160 calls GetRDSMetricsMultiple, lines 167-216 loop through datapoints |
| `internal/metrics/store.go` | 5-minute truncation constant | ✓ VERIFIED | `MetricPeriod = 5 * time.Minute` at line 18 |
| `internal/metrics/store.go` | GetLatestMetricTimes and GetMetricsAtTime methods | ✓ VERIFIED | Lines 168 and 200 export correctly |
| `cmd/server/main.go` | Gap detection call before RunContinuous | ✓ VERIFIED | Lines 214-216 call DetectAndFillGaps |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| `collector.go` | `cloudwatch.go` | GetRDSMetricsMultiple call | ✓ VERIFIED | Line 160: `client.GetRDSMetricsMultiple` |
| `collector.go` | `store.go` | UpsertHourlyMetric for each datapoint | ✓ VERIFIED | Lines 171-220 store each metric type per datapoint |
| `cmd/server/main.go` | `collector.DetectAndFillGaps` | startup call | ✓ VERIFIED | Lines 214-216 |
| `collector.go` | `store.go` | GetLatestMetricTimes query | ✓ VERIFIED | Line 479 calls `metricsStore.GetLatestMetricTimes` |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
| ----------- | ------ | -------------- |
| Period=300 in CloudWatch API | ✓ SATISFIED | None |
| 3 datapoints per 15-min cycle | ✓ SATISFIED | None |
| 5-minute granularity timestamps | ✓ SATISFIED | None |
| Gap detection on startup | ⚠️ NEEDS HUMAN | Current impl fills all data, not just gaps |
| Interpolated entries with SampleCount=0 | ✗ BLOCKED | Uses SampleCount=1 instead of SampleCount=0 |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| `collector.go` | 235, 609 | SampleCount always 1 | ⚠️ WARNING | Cannot distinguish real vs interpolated data |
| `collector.go` | 467-594 | DetectAndFillGaps fills all CloudWatch data | ⚠️ WARNING | Does not check for existing data gaps |

### Human Verification Required

1. **Server startup gap detection**

   **Test:** Start server and check logs for gap detection messages
   **Expected:** Logs show "Backfilling metrics data from CloudWatch (up to 7 days)..." then "Metrics backfill complete: X new datapoints stored"
   **Why human:** Cannot verify CloudWatch API calls programmatically; requires runtime observation

2. **5-minute interval metrics**

   **Test:** Query `metrics_hourly` table and verify timestamps are 5-minute aligned
   **Expected:** All timestamps end in :00, :05, :10, :15, :20, :25, :30, :35, :40, :45, :50, :55
   **Why human:** Requires direct database query to verify

3. **SampleCount distinction**

   **Test:** Query `metrics_hourly` for sample_count values
   **Expected:** Gap-filled entries have sample_count=0, real entries have sample_count=1
   **Why human:** Cannot verify data semantics without knowing which entries are gap-filled

### Gaps Summary

**10/16 must-haves verified** — 6 gaps blocking full goal achievement:

1. **Period=300 CloudWatch API call** — ✅ VERIFIED
   - `getMetricMultiple` at `cloudwatch.go:251` uses `Period: aws.Int32(300)`
   - `GetRDSMetricsMultiple` at `cloudwatch.go:294` calls `getMetricMultiple`

2. **3 datapoints per 15-minute cycle** — ✅ VERIFIED
   - `collectInstance` at `collector.go:156` calculates `startTime := now.Add(-15 * time.Minute)`
   - Returns up to 3 datapoints for the 15-minute window

3. **5-minute granularity timestamps** — ✅ VERIFIED
   - CloudWatch returns 5-minute aligned timestamps via Period=300
   - Each timestamp truncated to 5-minute boundaries

4. **Gap detection on startup** — ✅ VERIFIED
   - `main.go:214-216` calls `DetectAndFillGaps` before `RunContinuous`
   - Gap detection runs synchronously at server startup

5. **Gaps > 15 minutes detected and filled** — ⚠️ PARTIALLY VERIFIED
   - Current `DetectAndFillGaps` fills ALL data from CloudWatch (up to 7 days)
   - No explicit gap detection logic to check for existing data gaps
   - Always fetches and stores all CloudWatch data, not just missing intervals

6. **SampleCount=0 for interpolated entries** — ❌ FAILED
   - All stored metrics use `SampleCount: 1` at `collector.go:235,609`
   - Current code: `SampleCount: 1` for both real and gap-filled data
   - Missing: Logic to set `SampleCount: 0` for interpolated entries

---

_Verified: 2026-02-26T14:45:00Z_
_Verifier: OpenCode (gsd-verifier)_
