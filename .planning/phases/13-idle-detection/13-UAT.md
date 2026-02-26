---
status: testing
phase: 13-idle-detection
source: 13-01-SUMMARY.md
started: 2026-02-26T15:00:00Z
updated: 2026-02-26T15:05:00Z
---

## Current Test

number: 4
name: Verify recommendation generation for oregon-database
expected: |
  When GenerateRecommendations is called:
  1. Analyzer analyzes idle patterns from metrics data
  2. For oregon-database (or any running database with idle patterns), creates a recommendation
  3. Recommendation includes SuggestedSchedule with timezone, sleep_cron, wake_cron
  4. Recommendation stored with status "pending" in database
  
  Expected recommendation structure:
  - InstanceID: oregon-database's ID
  - DetectedPattern: {idle_start_hour, idle_end_hour, days_of_week, avg_cpu, confidence}
  - SuggestedSchedule: {timezone: "UTC", sleep_cron: "cron expression", wake_cron: "cron expression"}
  - ConfidenceScore: 0-100 scale
  - Status: "pending"
awaiting: user response

## Tests

### 1. Idle Detection Threshold Configuration
expected: ActivityThresholds struct has ConnectionsThreshold field set to 2, DefaultThresholds() returns CPUPercent: 5.0 and ConnectionsThreshold: 2, findIdleSegments() checks CPU < 5% AND connections < 2
result: pass

### 2. Compound Threshold Allows 1 Connection
expected: Instances with 1 connection (connections < 2) can be flagged as idle when CPU is also below 5%
result: pass

### 3. Recommendations Use Compound Threshold
expected: Idle pattern detection uses compound threshold (CPU < 5% AND connections < 2) for accurate recommendations
result: pass

### 4. Verify recommendation generation for oregon-database
expected: |
  When GenerateRecommendations is called:
  1. Analyzer analyzes idle patterns from metrics data
  2. For oregon-database (or any running database with idle patterns), creates a recommendation
  3. Recommendation includes SuggestedSchedule with timezone, sleep_cron, wake_cron
  4. Recommendation stored with status "pending" in database
  
  Expected recommendation structure:
  - InstanceID: oregon-database's ID
  - DetectedPattern: {idle_start_hour, idle_end_hour, days_of_week, avg_cpu, confidence}
  - SuggestedSchedule: {timezone: "UTC", sleep_cron: "cron expression", wake_cron: "cron expression"}
  - ConfidenceScore: 0-100 scale
  - Status: "pending"
result: issue
reported: "400 Bad Request: ERROR: invalid input syntax for type uuid: \"dbo-dev-master-mysql\""
severity: blocker
root_cause: Instance ID mismatch - analyzer's GetInstanceIDs() uses provider's ListAllDatabases which returns AWS DB instance names instead of database UUIDs. HasDataSufficient query expected UUID but received name.
artifacts:
  - path: "internal/analyzer/analyzer.go"
    issue: "GetInstanceIDs() returns AWS DB instance identifiers instead of database UUIDs"
  - path: "internal/metrics/store.go"
    issue: "Missing UpsertMinuteMetric and GetInstanceIDs methods"
missing:
  - "Fix GetInstanceIDs to use correct UUIDs from database"
  - "Add missing UpsertMinuteMetric method to MetricsStore"
  - "Add missing GetInstanceIDs method to MetricsStore"
debug_session: ".planning/debug/recommendation-generation-400.md"

## Summary

total: 4
passed: 3
issues: 1
pending: 0
skipped: 0

---
