---
phase: quick-004
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: [web/src/pages/InstanceDetailPage.tsx]
autonomous: true

must_haves:
  truths:
    - "Database name at top is readable on dark background"
    - "Quick Stats card is removed"
    - "Single-datapoint Metrics card is removed"
    - "Remaining cards are cleanly organized"
  artifacts:
    - path: "web/src/pages/InstanceDetailPage.tsx"
      provides: "Cleaned up Instance Details page"
  key_links: []
---

<objective>
Clean up InstanceDetailPage: fix dark-on-dark title text, remove unnecessary cards, reorganize layout.

Purpose: Improve readability and reduce UI clutter now that time-series metrics are available.
Output: Cleaner Instance Details page with title readable on dark background.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@web/src/pages/InstanceDetailPage.tsx
@web/src/index.css
</context>

<tasks>

<task type="auto">
  <name>Task 1: Fix title text color and remove unused cards</name>
  <files>web/src/pages/InstanceDetailPage.tsx</files>
  <action>
    1. Change the database name h1 from `text-gray-900` to `text-foreground` (line 278) for dark mode compatibility
    
    2. Remove the Quick Stats card entirely (lines 431-451):
       - This is the card showing Status, Current Cost, and Idle Time
       - Remove the entire `<div className="bg-white shadow-sm border rounded-lg p-6">` block with "Quick Stats"
    
    3. Remove the single-datapoint Metrics card (lines 356-419):
       - This is the Metrics card with MetricCard components showing CPU, Memory, Connections, IOPS
       - We now have MetricsChart which provides time-series data
       - Keep the MetricsChart component (the "Metrics History" section)
       - Also remove the MetricCard component definition (lines 9-41) and helper functions (lines 44-62) since they'll be unused
    
    4. Reorganize remaining cards:
       - Left column (lg:col-span-2): Configuration, Tags, Metrics History (chart)
       - Right column: Actions only
       - Ensure clean spacing with space-y-6
    
    5. Fix any remaining hardcoded dark colors in the file:
       - Change all `text-gray-900` to `text-foreground` where it's used for primary text
       - Change `bg-white` to `bg-card` for card backgrounds
       - Change `border-gray-*` to `border-border` for borders
       - Keep status badge colors (green, blue, gray backgrounds for status indicators)
  </action>
  <verify>
    - Run `cd web && npm run build` to verify no TypeScript errors
    - Check that MetricCard and helper functions are removed
    - Check that Quick Stats card is removed
    - Check that single-datapoint Metrics card is removed
    - Check that title uses text-foreground
  </verify>
  <done>
    - Page builds without errors
    - Title uses text-foreground for dark mode compatibility
    - Quick Stats card removed
    - Single-datapoint Metrics card removed (MetricCard component gone)
    - Actions card remains in right column
    - Metrics History (chart) remains in left column
    - Configuration and Tags cards remain
  </done>
</task>

</tasks>

<verification>
- `cd web && npm run build` passes
- InstanceDetailPage.tsx has no MetricCard component
- InstanceDetailPage.tsx has no "Quick Stats" section
- InstanceDetailPage.tsx has no getMetricValue/getMetricMin/getMetricMax/getMetricSamples functions
- Title h1 uses text-foreground class
</verification>

<success_criteria>
- Build passes with no errors
- Page structure: Left column has Config + Tags + Metrics History chart, Right column has Actions
- Dark mode compatible text colors
</success_criteria>

<output>
After completion, create `.planning/quick/004-instance-details-cleanup/004-SUMMARY.md`
</output>
