---
phase: quick
plan: 002
type: execute
wave: 1
depends_on: []
files_modified:
  - web/src/main.tsx
  - web/src/components/Navigation.tsx
  - web/src/pages/SavingsPage.tsx
  - web/src/components/savings/SavingsChart.tsx
  - web/src/components/savings/SavingsSummaryCards.tsx
  - web/src/components/savings/SavingsTable.tsx
  - web/src/components/savings/CostProjection.tsx
  - web/src/components/savings/InstanceSavingsTable.tsx
  - web/src/components/savings/DateRangeSelector.tsx
autonomous: true

must_haves:
  truths:
    - "No /savings route exists in the application"
    - "Navigation bar has no Savings link"
    - "Savings page and components are deleted"
  artifacts:
    - path: "web/src/main.tsx"
      provides: "Routes without /savings"
      contains: "No SavingsPage import or route"
    - path: "web/src/components/Navigation.tsx"
      provides: "Navigation without Savings link"
      contains: "No /savings link or PiggyBank import"
  key_links: []
---

<objective>
Remove the Savings page from the SnoozeQL application

Purpose: Clean up the frontend by removing the Savings page feature entirely
Output: Application without /savings route, navigation link, or related components
</objective>

<execution_context>
@~/.config/opencode/get-shit-done/workflows/execute-plan.md
@~/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@web/src/main.tsx
@web/src/components/Navigation.tsx
@web/src/pages/SavingsPage.tsx
</context>

<tasks>

<task type="auto">
  <name>Task 1: Remove Savings route and navigation</name>
  <files>web/src/main.tsx, web/src/components/Navigation.tsx</files>
  <action>
    In web/src/main.tsx:
    - Remove the SavingsPage import (line 12)
    - Remove the savings Route element (line 43: `<Route path="savings" element={<SavingsPage />} />`)
    
    In web/src/components/Navigation.tsx:
    - Remove PiggyBank from the lucide-react import (line 2)
    - Remove the entire Savings Link block (lines 26-29)
  </action>
  <verify>
    - `grep -r "SavingsPage" web/src/main.tsx` returns nothing
    - `grep -r "savings" web/src/components/Navigation.tsx` returns nothing
    - `npm run build` in web/ completes without errors
  </verify>
  <done>
    - No /savings route in main.tsx
    - No Savings link in Navigation.tsx
    - Build succeeds
  </done>
</task>

<task type="auto">
  <name>Task 2: Delete Savings page and components</name>
  <files>
    web/src/pages/SavingsPage.tsx,
    web/src/components/savings/SavingsChart.tsx,
    web/src/components/savings/SavingsSummaryCards.tsx,
    web/src/components/savings/SavingsTable.tsx,
    web/src/components/savings/CostProjection.tsx,
    web/src/components/savings/InstanceSavingsTable.tsx,
    web/src/components/savings/DateRangeSelector.tsx
  </files>
  <action>
    Delete all savings-related files:
    - rm web/src/pages/SavingsPage.tsx
    - rm -rf web/src/components/savings/
  </action>
  <verify>
    - `ls web/src/pages/SavingsPage.tsx` returns "No such file"
    - `ls web/src/components/savings/` returns "No such file or directory"
    - `npm run build` in web/ completes without errors
  </verify>
  <done>
    - SavingsPage.tsx deleted
    - All components in web/src/components/savings/ deleted
    - Build succeeds with no missing import errors
  </done>
</task>

</tasks>

<verification>
- Application builds successfully: `cd web && npm run build`
- No references to SavingsPage or /savings in codebase: `grep -r "SavingsPage\|/savings" web/src/`
- Navigation renders without Savings link
</verification>

<success_criteria>
- The /savings route is removed from the application
- The Savings navigation link is removed
- All savings-related components are deleted
- Application builds and runs without errors
</success_criteria>

<output>
After completion, create `.planning/quick/002-remove-savings-page/002-SUMMARY.md`
</output>
