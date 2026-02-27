---
phase: quick-010
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/api/handlers/recommendations.go
  - web/src/components/RecommendationsTable.tsx
autonomous: true

must_haves:
  truths:
    - "Pattern description shows only time range (e.g., 'Idle 10PM to 6AM')"
    - "Recommendations table has 4 columns: Schedule Pattern, Instances Affected, Est. Daily Savings, Actions"
  artifacts:
    - path: "internal/api/handlers/recommendations.go"
      provides: "Simplified pattern description"
      contains: 'fmt.Sprintf("Idle %s", timeRange)'
    - path: "web/src/components/RecommendationsTable.tsx"
      provides: "4-column table without Wake/Sleep"
---

<objective>
Simplify recommendations display by removing redundant information

Purpose: Remove ", x days/week" from pattern descriptions and remove the Wake/Sleep column from the recommendations table to reduce visual noise.
Output: Cleaner pattern descriptions and streamlined table layout
</objective>

<context>
@.planning/STATE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Remove days from pattern description</name>
  <files>internal/api/handlers/recommendations.go</files>
  <action>
In the `describePattern` function (around line 270), change the return statement from:
```go
return fmt.Sprintf("Idle %s, %s", timeRange, dayDesc)
```
to:
```go
return fmt.Sprintf("Idle %s", timeRange)
```

This removes the ", weekdays", ", weekends", ", daily", or ", X days/week" suffix from pattern descriptions.

The dayType/dayDesc logic (lines 257-268) can remain for now - it's not hurting anything and may be useful later. The key change is just the return statement.
  </action>
  <verify>
```bash
cd /Users/tylerwagner/snoozeql && go build ./...
```
Pattern description now shows only time range.
  </verify>
  <done>Pattern descriptions show "Idle 10PM to 6AM" instead of "Idle 10PM to 6AM, weekdays"</done>
</task>

<task type="auto">
  <name>Task 2: Remove Wake/Sleep column from table</name>
  <files>web/src/components/RecommendationsTable.tsx</files>
  <action>
1. Remove the Wake/Sleep column header (lines 29-31):
   Delete the `<th>` element for "Wake/Sleep"

2. Remove the Wake/Sleep column data cell (lines 68-78):
   Delete the entire `<td>` element that displays wake/sleep times

3. Remove the unused variables (lines 46-48):
   Delete the `wakeTime` and `sleepTime` variable declarations since they're no longer used

Table should now have 4 columns: Schedule Pattern, Instances Affected, Est. Daily Savings, Actions
  </action>
  <verify>
```bash
cd /Users/tylerwagner/snoozeql/web && npm run build
```
Table renders with 4 columns.
  </verify>
  <done>Recommendations table displays 4 columns without Wake/Sleep column</done>
</task>

</tasks>

<verification>
1. Go build passes: `go build ./...`
2. Frontend build passes: `cd web && npm run build`
3. Visual check: Recommendations table shows 4 columns
4. Pattern descriptions show only time ranges
</verification>

<success_criteria>
- Pattern descriptions no longer include ", x days/week" or day type
- Recommendations table has exactly 4 columns
- Both builds pass without errors
</success_criteria>

<output>
After completion, create `.planning/quick/010-remove-days-week-from-schedule-pattern/010-SUMMARY.md`
</output>
