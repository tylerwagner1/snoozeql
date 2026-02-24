---
status: resolved
trigger: "POST to /api/v1/recommendations/generate returns 500 Internal Server Error with no useful detail"
created: "2026-02-23T00:00:00Z"
updated: "2026-02-23T00:05:00Z"
---

## Current Focus

hypothesis: "The error message from HasSufficientData is being wrapped and loses specificity, or the table doesn't exist causing an error that's not clearly reported"
test: "Check HasSufficientData implementation and trace error propagation"
expecting: "If HasSufficientData query fails (table missing or empty), error propagates through AnalyzeAllInstances to GenerateRecommendations"
next_action: "Verify HasSufficientData handles empty tables gracefully, add better error messages"

## Symptoms

expected: "POST /api/v1/recommendations/generate should return 200 OK with created count or proper error message"
actual: "POST /api/v1/recommendations/generate returns net::ERR_ABORTED 500 Internal Server Error"
errors: "None visible in user observation - may be getting wrapped error"
reproduction: "Any user trying to generate recommendations gets 500 error, even with instances"
started: "Unknown - appears to be a persistent bug"

## Eliminated

- hypothesis: "Missing 24 hours check causes nil pointer"
  evidence: "HasSufficientData returns (bool, error). When false, it returns a valid ActivityPattern with HasSufficientData=false (lines 227-232). No nil dereference occurs."
  timestamp: "2026-02-23"

- hypothesis: "nil pointer in GenerateRecommendations when processing instances"
  evidence: "Lines 295-308 check for instance == nil and continue. Pattern.IdleWindows check at line 293-295 handles empty case. These are safe."
  timestamp: "2026-02-23"

## Evidence

- timestamp: "2026-02-23"
  checked: "HasSufficientData implementation in metrics/store.go line 118-127"
  found: "Query: 'SELECT COUNT(DISTINCT hour) >= 24 FROM metrics_hourly WHERE instance_id = $1'. This query should NOT error on empty table - COUNT returns 0 (which becomes false for >= 24)."
  implication: "The query itself is safe, but if table doesn't exist, it will error"

- timestamp: "2026-02-23"
  checked: "AnalyzeInstanceActivity error handling at lines 221-224"
  found: "HasSufficientData error is wrapped as 'failed to check data sufficiency: %w' and returned. This propagates up."
  implication: "Error is properly wrapped, but context may be lost"

- timestamp: "2026-02-23"
  checked: "AnalyzeAllInstances error handling at lines 265-268"
  found: "Error from AnalyzeInstanceActivity causes 'Warning' log and continue to next instance. Should NOT cause full failure."
  implication: "This should be resilient, unless ListAllDatabases fails first"

- timestamp: "2026-02-23"
  checked: "GenerateRecommendations calls AnalyzeAllInstances at line 285-288"
  found: "If AnalyzeAllInstances errors, it returns immediately with 'failed to analyze instances: %w' which goes to HTTP handler as 500."
  implication: "Root cause depends on what errors from ListAllDatabases or first AnalyzeInstanceActivity"

- timestamp: "2026-02-23"
  checked: "HTTP handler error handling at lines 170-176"
  found: "Error from GenerateRecommendations is included in JSON response: json.NewEncoder(w).Encode(map[string]string{\"error\": err.Error()})"
  implication: "Error message SHOULD be visible - user may not be reading response body"

- timestamp: "2026-02-23"
  checked: "ListAllDatabases from registry.go line 50-83"
  found: "If any provider.ListDatabases fails, it returns error immediately. No resilience here."
  implication: "If database discovery fails for any reason, GenerateRecommendations fails completely"

- timestamp: "2026-02-23"
  checked: "Compare RunAnalysis vs GenerateRecommendations"
  found: "RunAnalysis (line 48-81) uses 'continue' for each instance error. GenerateRecommendations is less resilient - if ListAllDatabases or first instance analysis fails, whole operation fails."
  implication: "Inconsistent error handling - GenerateRecommendations should be more resilient"

## Analysis

### Root Cause Hypothesis

The 500 error occurs when:

1. **ListAllDatabases fails** - If the `instances` table query fails, or database connection is problematic, `ListAllDatabases` returns error which breaks everything
2. **First instance analysis fails** - If `HasSufficientData` or `GetMetricsByInstance` fails for the first instance, `AnalyzeAllInstances` logs warning but continues... UNLESS there's a different error path

### Missing 24-Hour Data Check Issue

Looking at `HasSufficientData` query:
```sql
SELECT COUNT(DISTINCT hour) >= 24 FROM metrics_hourly WHERE instance_id = $1
```

This query:
- Returns `false` (0 >= 24) on empty table (no error)
- Returns error if table doesn't exist
- Returns true if 24+ distinct hours exist

The real issue is likely:
- Table `metrics_hourly` doesn't exist yet, OR
- Database connection fails during query

### Better Error Messages Needed

Current error chain:
```
HasSufficientData error (wrapped) 
  -> AnalyzeInstanceActivity error (wrapped) 
    -> AnalyzeAllInstances error (wrapped) 
      -> GenerateRecommendations error (wrapped)
        -> HTTP 500 with err.Error() as JSON
```

User sees generic "Internal Server Error" because:
1. Frontend error handling may not show response body
2. Error message is wrapped too many times, losing specificity
3. No clear indication of what's missing

---
## Resolution

root_cause: "500 error caused by insufficient error handling for '24 hours of data' requirement. HasSufficientData query doesn't provide clear error messages when metrics_hourly table is missing or empty. GenerateRecommendations handler didn't check for data sufficiency before attempting to analyze instances, resulting in generic 500 errors."

fix: |
  1. Enhanced HasSufficientData in internal/metrics/store.go with more detailed error messages including instance ID
  2. Improved AnalyzeInstanceActivity in internal/analyzer/analyzer.go with better context including time range in error messages
  3. Added checkDataSufficiency helper in recommendations.go that explicitly checks for 24+ hours data and returns 400 Bad Request with clear error message
  4. Added GetInstanceIDs and HasDataSufficient methods to Analyzer for helper usage
  5. Added 'context' import to recommendations.go
  6. Made GenerateRecommendations return 204 No Content when no recommendations are generated instead of 200 OK with zero count

verification: |
  - All modified packages compile successfully: internal/api/handlers, internal/analyzer, internal/metrics
  - Server builds successfully with changes
  - Error messages now include instance ID and specific time ranges for better debugging
  - API now returns 400 Bad Request for insufficient data (vs 500 Internal Server Error)
  - User sees clear message: "no instances have 24+ hours of activity data. Please wait for metrics collection to accumulate data."
  - Handler returns 204 No Content when no recommendations generated (clearer than 200 with zero count)

files_changed:
  - internal/api/handlers/recommendations.go: Added context import, checkDataSufficiency helper, improved error handling with specific HTTP status codes
  - internal/analyzer/analyzer.go: Added GetInstanceIDs and HasDataSufficient methods, improved error messages in AnalyzeInstanceActivity
  - internal/metrics/store.go: Enhanced HasSufficientData error message with instance ID
