---
phase: quick-009
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/api/handlers/recommendations.go
autonomous: true

must_haves:
  truths:
    - "GetAllRecommendations excludes instances with existing enabled schedules"
    - "Recommendations for scheduled instances are filtered before grouping"
  artifacts:
    - path: "internal/api/handlers/recommendations.go"
      provides: "Filtered recommendations in GetAllRecommendations"
      contains: "GetMatchingSchedules"
  key_links:
    - from: "GetAllRecommendations"
      to: "scheduleStore.GetMatchingSchedules"
      via: "filter loop before enrichment"
      pattern: "GetMatchingSchedules.*instance"
---

<objective>
Filter recommendations in GetAllRecommendations to exclude instances with existing schedules

Purpose: Ensure the recommendations list only shows instances that don't already have schedules, matching the filtering logic already applied in GenerateRecommendations.
Output: Updated GetAllRecommendations handler that filters out scheduled instances
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@internal/api/handlers/recommendations.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add schedule filtering to GetAllRecommendations</name>
  <files>internal/api/handlers/recommendations.go</files>
  <action>
In GetAllRecommendations, after the enrichment loop (around line 155), add filtering to exclude recommendations for instances that already have enabled schedules:

1. After enriching recommendations into `enriched` slice, create a new slice `filtered []enrichedRec`
2. Loop through `enriched` and for each recommendation:
   - Get the instance using `h.instanceStore.GetInstanceByID(r.Context(), rec.InstanceID)`
   - If instance found, call `h.scheduleStore.GetMatchingSchedules(*instance)`
   - If schedules exist (len > 0), skip this recommendation (log debug message)
   - Otherwise, append to `filtered`
3. Use `filtered` instead of `enriched` when calling `groupRecommendations`

Pattern follows GenerateRecommendations (lines 350-373) but uses enrichedRec type.
  </action>
  <verify>
- `go build ./...` compiles successfully
- Review the code change to confirm filtering logic matches GenerateRecommendations pattern
  </verify>
  <done>GetAllRecommendations returns only recommendations for instances without existing schedules</done>
</task>

</tasks>

<verification>
- `go build ./...` passes
- Code review confirms filtering occurs before groupRecommendations call
- Log messages match pattern from GenerateRecommendations
</verification>

<success_criteria>
- GetAllRecommendations excludes recommendations for instances with existing enabled schedules
- Filtering logic mirrors GenerateRecommendations pattern
- Build passes without errors
</success_criteria>

<output>
After completion, create `.planning/quick/009-filter-existing-correctly/009-SUMMARY.md`
</output>
