---
quick: 008
type: execute
autonomous: true
files_modified:
  - internal/api/handlers/recommendations.go

must_haves:
  truths:
    - "Instances with existing enabled schedules don't get recommendations generated"
    - "Instances without schedules still get recommendations"
  artifacts:
    - path: "internal/api/handlers/recommendations.go"
      provides: "Schedule check in GenerateRecommendations handler"
      contains: "GetMatchingSchedules"
  key_links:
    - from: "internal/api/handlers/recommendations.go"
      to: "internal/store/postgres.go"
      via: "scheduleStore.GetMatchingSchedules"
      pattern: "scheduleStore\\.GetMatchingSchedules"
---

<objective>
Exclude instances with existing schedules from recommendation generation.

Purpose: Avoid generating duplicate/unnecessary recommendations for instances that already have sleep/wake schedules assigned.
Output: Modified GenerateRecommendations handler that filters out scheduled instances.
</objective>

<context>
@.planning/STATE.md
@internal/api/handlers/recommendations.go
@internal/store/postgres.go (GetMatchingSchedules method at line 946)
</context>

<tasks>

<task type="auto">
  <name>Task 1: Filter scheduled instances in GenerateRecommendations</name>
  <files>internal/api/handlers/recommendations.go</files>
  <action>
In the `GenerateRecommendations` handler (line 324), after getting recommendations from the analyzer (line 339):
1. Before storing recommendations (line 351), filter out instances that already have matching schedules
2. For each recommendation in `recs`:
   - Get the instance using `h.instanceStore.GetInstanceByID(ctx, rec.InstanceID)`
   - Call `schedules, _ := h.scheduleStore.GetMatchingSchedules(*instance)`
   - If `len(schedules) > 0`, skip this recommendation (don't store it)
3. Log skipped instances at DEBUG level: `log.Printf("DEBUG: Skipping recommendation for %s - already has %d schedule(s)", rec.InstanceID, len(schedules))`

The handler already has `scheduleStore` injected (line 21) and `GetMatchingSchedules` already filters for enabled schedules (line 957 in postgres.go).
  </action>
  <verify>
1. `go build ./...` passes
2. Create a schedule that matches an instance, then call POST /api/v1/recommendations/generate
3. The instance should NOT appear in new recommendations
4. Remove the schedule, regenerate - instance should now appear
  </verify>
  <done>
Instances with enabled schedules are excluded from recommendation generation.
  </done>
</task>

</tasks>

<verification>
- [ ] `go build ./...` compiles without errors
- [ ] Instance with matching schedule: no recommendation generated
- [ ] Instance without schedule: recommendation generated normally
- [ ] Existing pending recommendations unaffected (filter only applies to new generation)
</verification>

<success_criteria>
POST /api/v1/recommendations/generate skips instances that already have enabled schedules matching them.
</success_criteria>

<output>
After completion, update `.planning/STATE.md`:
- Add quick-008-01 to Quick Tasks Completed table
- Update Last activity date
</output>
