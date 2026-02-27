---
phase: quick
plan: 007
type: execute
wave: 1
depends_on: []
files_modified:
  - web/src/pages/RecommendationsPage.tsx
  - web/src/components/RecommendationsTable.tsx
autonomous: true

must_haves:
  truths:
    - "User sees recommendations in a table format"
    - "Each row shows one unique schedule pattern"
    - "Number of Instances Affected column displays instance_count"
    - "User can view details, confirm, or dismiss recommendations"
  artifacts:
    - path: "web/src/components/RecommendationsTable.tsx"
      provides: "Table component for grouped recommendations"
    - path: "web/src/pages/RecommendationsPage.tsx"
      provides: "Updated page rendering table instead of cards"
  key_links:
    - from: "RecommendationsPage.tsx"
      to: "RecommendationsTable.tsx"
      via: "component import and rendering"
---

<objective>
Change recommendations display from expandable cards to a table view.

Purpose: Improve recommendations UX by showing schedule patterns in a scannable table format with "Number of Instances Affected" column.
Output: Table-based recommendations display with row-level actions.
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@web/src/pages/RecommendationsPage.tsx
@web/src/components/RecommendationGroup.tsx
@web/src/lib/api.ts (lines 59-94 for types)
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create RecommendationsTable component</name>
  <files>web/src/components/RecommendationsTable.tsx</files>
  <action>
Create new table component that:

1. Accept props: `groups: RecommendationGroup[]`, `onOpenModal`, `onDismiss`

2. Render table with columns:
   - "Schedule Pattern" (pattern_description from group)
   - "Wake/Sleep" (format: wake_cron / sleep_cron from first recommendation)
   - "Instances Affected" (instance_count from group)
   - "Est. Daily Savings" (total_daily_savings, formatted as $X.XX/day in green)
   - "Actions" (View Details, Dismiss buttons)

3. Table styling:
   - Use existing slate-800/50 bg with slate-700 border for dark mode consistency
   - Header row: text-slate-400, uppercase, text-xs
   - Data rows: hover:bg-slate-700/30 transition
   - Alternate row shading optional

4. Row click behavior:
   - Click row or "View Details" button → open modal for first recommendation in group
   - "Dismiss" button → call onDismiss for ALL recommendations in the group (iterate group.recommendations)

5. Handle multi-instance groups:
   - For groups with instance_count > 1, show pattern-level info
   - When user clicks row, show first recommendation details (modal handles individual confirm)
   - Dismiss button should dismiss entire group (all recommendations)

Use existing patterns from project (Tailwind, lucide-react icons).
  </action>
  <verify>
File exists at web/src/components/RecommendationsTable.tsx
TypeScript compiles: `cd web && npx tsc --noEmit`
  </verify>
  <done>
RecommendationsTable component renders groups as table rows with all required columns and actions.
  </done>
</task>

<task type="auto">
  <name>Task 2: Update RecommendationsPage to use table</name>
  <files>web/src/pages/RecommendationsPage.tsx</files>
  <action>
Update RecommendationsPage.tsx:

1. Replace import of RecommendationGroupComponent with RecommendationsTable:
   ```typescript
   import { RecommendationsTable } from '../components/RecommendationsTable'
   ```

2. Replace the groups.map() rendering (lines 141-148) with:
   ```tsx
   <RecommendationsTable
     groups={groups}
     onOpenModal={handleOpenModal}
     onDismiss={handleDismiss}
   />
   ```

3. Update handleDismiss to support batch dismissal:
   - Accept either single id (string) OR array of ids (string[])
   - Iterate and call api.dismissRecommendation for each
   - Update groups state to remove all dismissed recommendations
   - Signature: `handleDismiss: (ids: string | string[]) => Promise<void>`

4. Keep all other functionality unchanged:
   - Generate button
   - Empty state
   - RecommendationModal for details/confirm

Do NOT modify the modal behavior - it still operates on individual recommendations.
  </action>
  <verify>
`cd web && npm run build` succeeds
Navigate to /recommendations in browser - see table view instead of cards
  </verify>
  <done>
Recommendations page displays table with pattern rows, instance counts, and working actions.
  </done>
</task>

</tasks>

<verification>
1. `cd web && npm run build` - no TypeScript or build errors
2. Browser test: Navigate to /recommendations
   - Table displays with correct columns
   - Each row shows pattern, wake/sleep times, instance count, savings
   - Click row → modal opens with recommendation details
   - Dismiss button removes entire group
3. Empty state still displays correctly when no recommendations
</verification>

<success_criteria>
- Recommendations display as table, not expandable cards
- "Instances Affected" column shows count per pattern group
- All existing functionality preserved (view details, confirm, dismiss, generate)
- TypeScript compiles without errors
- UI matches existing dark theme styling
</success_criteria>

<output>
After completion, create `.planning/quick/007-recommendations-tab-display-changes/007-SUMMARY.md`
</output>
