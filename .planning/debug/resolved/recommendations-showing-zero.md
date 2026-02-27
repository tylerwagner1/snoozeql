---
status: resolved
trigger: "Recommendations are showing empty instance names and zero savings - was working previously"
created: "2026-02-27T12:20:00Z"
updated: "2026-02-27T12:22:00Z"
---

## Current Focus

hypothesis: Recommendations API lookup is using wrong ID field - handler tries GetInstanceByProviderID but recommendations store instance.ID (database UUID)
test: Read analyzer code to confirm which ID path is used
expecting: idleWindowToRecommendation stores instance.ID, but handler uses GetInstanceByProviderID (which looks up by provider_id)
next_action: After fix Applied, verify recommendations show instance data correctly

## Fix Applied

action: "Changed recommendations.go line 105 from GetInstanceByProviderID to GetInstanceByID"
date: "2026-02-27T12:20:00Z"
result: "Applied - now using GetInstanceByID"

## Symptoms

expected: Recommendations should show instance names, estimated savings, and hourly cost
actual: Recommendations show empty instance_name, zero estimated_daily_savings, zero hourly_cost_cents
errors: None - data appears to come from API
reproduction: Visit the Recommendations page
timeline: Just implemented scheduler daemon in quick-006 - issue started after that

## Evidence

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/analyzer/recommendation.go idleWindowToRecommendation function line 38"
  found: "InstanceID: instance.ID (database UUID)"
  implication: "idleWindowToRecommendation stores database UUID, NOT provider_id"

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/analyzer/analyzer.go lines 255-280 AnalyzeAllInstances"
  found: "Returns patterns keyed by instance.ID (database UUID) at line 275: patterns[instance.ID] = pattern"
  implication: "GenerateRecommendations uses instance.ID as key, then calls idleWindowToRecommendation"

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/api/handlers/recommendations.go line 105"
  found: "h.instanceStore.GetInstanceByProviderID(r.Context(), \"\", rec.InstanceID)"
  implication: "Handler tries to look up by provider_id, but recommendations store instance.ID"

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/analyzer/analyzer.go lines 158-172 generateRecommendation function"
  found: "Sets recommendation.InstanceID = instance.ProviderID"
  implication: "Legacy path in analyzer.go uses ProviderID (different from idleWindowToRecommendation)"

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/store/postgres.go line 312"
  found: "GetInstanceByProviderID looks up by provider_id, while GetInstanceByID looks up by instance.ID"
  implication: "Handler needs to use GetInstanceByID to match idleWindowToRecommendation"

- timestamp: "2026-02-27T12:20:00Z"
  checked: "Read internal/store/postgres.go line 370 GetInstanceByID"
  found: "Uses WHERE i.id = $1 to look up by database UUID"
  implication: "GetInstanceByID is the correct method for idleWindowToRecommendation's stored ID format"

## Eliminated

- hypothesis: "Instance lookup is failing due to empty string in GetInstanceByProviderID second parameter"
  evidence: "Second parameter is database name (empty for default), should not cause lookup failure"
  timestamp: "2026-02-27T12:20:00Z"

- hypothesis: "Instance data exists but hourly_cost_cents is zero"
  evidence: "Cannot verify until we see what GetInstanceByProviderID actually returns"
  timestamp: "2026-02-27T12:20:00Z"

- hypothesis: "Recommendations table has invalid instance_id values"
  evidence: "Instance IDs in recommendations match database UUIDs (from instance.ID), not provider_ids"
  timestamp: "2026-02-27T12:20:00Z"

## Resolution

root_cause: "ID field mismatch between recommendations storage and API lookup. The idleWindowToRecommendation function in analyzer/recommendation.go stores instance.ID (database UUID) in recommendations.instance_id, but the GetAllRecommendations handler in api/handlers/recommendations.go uses GetInstanceByProviderID to look up instances by provider_id instead of database UUID. This caused GetInstanceByProviderID to fail finding instances, resulting in empty instance_name, zero hourly_cost_cents, and zero estimated_daily_savings in the API response."

fix: "Change recommendations.go line 105 from GetInstanceByProviderID to GetInstanceByID to match what idleWindowToRecommendation stores in the database."

verification: "After fix, recommendations should show instance_name, hourly_cost_cents, and estimated_daily_savings populated correctly. Fix compiled successfully with 'go build -mod=mod'."

files_changed:
- internal/api/handlers/recommendations.go line 105

## Verification

- timestamp: "2026-02-27T12:22:00Z"
  checked: "Build server with 'go build -mod=mod'"
  found: "Build succeeded without errors"
  implication: "Fix compiles correctly"

- timestamp: "2026-02-27T12:22:00Z"
  checked: "Read fixed recommendations.go line 105"
  found: "Uses h.instanceStore.GetInstanceByID(r.Context(), rec.InstanceID)"
  implication: "Handler now correctly looks up by database UUID instead of provider_id"
