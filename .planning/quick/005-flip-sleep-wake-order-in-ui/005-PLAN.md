---
phase: quick
plan: 005
type: execute
wave: 1
depends_on: []
files_modified:
  - web/src/components/RecommendationCard.tsx
  - web/src/components/RecommendationModal.tsx
  - web/src/components/RecommendationGroup.tsx
  - web/src/components/ScheduleModal.tsx
  - web/src/components/WeeklyScheduleGrid.tsx
  - web/src/pages/SchedulesPage.tsx
  - web/src/pages/ScheduleEditPage.tsx
  - web/src/pages/ScheduleNewPage.tsx
  - web/src/pages/InstancesPage.tsx
autonomous: true

must_haves:
  truths:
    - "Wake appears before Sleep in all paired UI displays"
    - "Button order shows Wake first, Sleep second where both appear"
    - "Form fields show Wake CRON before Sleep CRON"
    - "Summary text shows Wake: ... | Sleep: ... format"
  artifacts:
    - path: "web/src/components/RecommendationCard.tsx"
      provides: "Wake at / Sleep at order in recommendation cards"
    - path: "web/src/components/RecommendationModal.tsx"
      provides: "Wake at / Sleep at order in recommendation modal"
    - path: "web/src/components/RecommendationGroup.tsx"
      provides: "Wake at / Sleep at order in grouped recommendations"
    - path: "web/src/components/ScheduleModal.tsx"
      provides: "Wake CRON / Sleep CRON field order"
    - path: "web/src/pages/ScheduleEditPage.tsx"
      provides: "Wake Cron / Sleep Cron field order"
    - path: "web/src/pages/ScheduleNewPage.tsx"
      provides: "Wake Cron / Sleep Cron field order"
  key_links: []
---

<objective>
Flip Sleep/Wake order to Wake/Sleep throughout the UI

Purpose: Consistent UI ordering where "Wake" appears first, "Sleep" second - more intuitive as databases wake up to work, then sleep when idle.

Output: All paired Sleep/Wake UI elements reordered to Wake/Sleep
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Flip order in recommendation components</name>
  <files>
    - web/src/components/RecommendationCard.tsx
    - web/src/components/RecommendationModal.tsx
    - web/src/components/RecommendationGroup.tsx
  </files>
  <action>
In each file, swap the two grid columns showing Sleep/Wake times:

**RecommendationCard.tsx (lines 122-131):**
- Move the "Wake at" div (currently second) to be first
- Move the "Sleep at" div (currently first) to be second

**RecommendationModal.tsx (lines 93-110):**
- Move the "Wake at" section (currently second) to be first
- Move the "Sleep at" section (currently first) to be second

**RecommendationGroup.tsx (lines 163-172):**
- Move the "Wake at" div (currently second) to be first
- Move the "Sleep at" div (currently first) to be second
  </action>
  <verify>
Visual inspection: Wake appears left, Sleep appears right in all recommendation displays.
`npm run build` passes in web/ directory.
  </verify>
  <done>All recommendation components show "Wake at" before "Sleep at"</done>
</task>

<task type="auto">
  <name>Task 2: Flip order in schedule forms and displays</name>
  <files>
    - web/src/components/ScheduleModal.tsx
    - web/src/components/WeeklyScheduleGrid.tsx
    - web/src/pages/SchedulesPage.tsx
    - web/src/pages/ScheduleEditPage.tsx
    - web/src/pages/ScheduleNewPage.tsx
  </files>
  <action>
**ScheduleModal.tsx:**
- Swap the "Sleep CRON" and "Wake CRON" input sections (lines 218-270)
- Update getSummary function (line 145): change `Sleep: ${sleepDesc} | Wake: ${wakeDesc}` to `Wake: ${wakeDesc} | Sleep: ${sleepDesc}`

**WeeklyScheduleGrid.tsx (lines 159-167):**
- Swap the legend order: show Wake indicator first, Sleep indicator second

**SchedulesPage.tsx (line 77):**
- Change `Sleep: ${sleepDesc} | Wake: ${wakeDesc}` to `Wake: ${wakeDesc} | Sleep: ${sleepDesc}`

**ScheduleEditPage.tsx (lines 218-251):**
- Swap the two form field divs: Wake Cron first, Sleep Cron second

**ScheduleNewPage.tsx (lines 182-215):**
- Swap the two form field divs: Wake Cron first, Sleep Cron second
  </action>
  <verify>
Visual inspection: Wake fields/legends appear before Sleep in all schedule UIs.
`npm run build` passes in web/ directory.
  </verify>
  <done>All schedule forms and displays show Wake before Sleep</done>
</task>

<task type="auto">
  <name>Task 3: Flip bulk action buttons on Instances page</name>
  <files>
    - web/src/pages/InstancesPage.tsx
  </files>
  <action>
**InstancesPage.tsx (lines 223-236):**
- Swap the button order: "Wake Selected" button first, "Sleep Selected" button second
- Keep the Clear button last

This matches the new Wake-first convention for paired actions.
  </action>
  <verify>
Visual inspection: Wake Selected button appears before Sleep Selected button.
`npm run build` passes in web/ directory.
  </verify>
  <done>Bulk action buttons show Wake before Sleep</done>
</task>

</tasks>

<verification>
1. `cd web && npm run build` - TypeScript compiles without errors
2. Visual check: All Wake/Sleep pairs now show Wake first
</verification>

<success_criteria>
- All paired Sleep/Wake UI elements reordered to Wake/Sleep
- Build passes with no TypeScript errors
- No functional changes, only visual ordering
</success_criteria>

<output>
After completion, create `.planning/quick/005-flip-sleep-wake-order-in-ui/005-SUMMARY.md`
</output>
