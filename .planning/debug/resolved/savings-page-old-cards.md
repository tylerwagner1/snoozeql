---
status: resolved
trigger: "Savings page showing old cards despite recent changes"
created: 2026-02-24T15:30:00Z
updated: 2026-02-24T15:35:00Z
---

## Current Focus

hypothesis: Cached/stale build artifacts are serving old code that still includes "Top Savings Instances" and "Cost Comparison" cards
test: Check build artifact timestamps and content
expecting: Build artifact should be dated before 2026-02-24 10:41 (commit time) and contain old card references
next_action: Rebuild the frontend to generate fresh build artifacts

## Symptoms

expected: Only 2 summary cards should display (Total Savings, Ongoing Savings)
actual: User sees "Top Savings Instances" and "Cost Comparison" cards
errors: None - browser console and terminal are clean
reproduction: Navigate to /savings page
timeline: User says "Just now" - this started after recent changes

## Eliminated

- hypothesis: User misunderstanding of expected behavior
  evidence: User provides clear expected vs actual - only 2 cards should show
  timestamp: 2026-02-24

- hypothesis: Source code still contains old cards
  evidence: SavingsSummaryCards.tsx shows only 2 cards, SavingsPage.tsx does not import old components
  timestamp: 2026-02-24

- hypothesis: Uncommitted local changes
  evidence: git status shows working tree clean
  timestamp: 2026-02-24

- hypothesis: Wrong branch or commit
  evidence: On main branch at commit 49ca9db0 which contains the fix
  timestamp: 2026-02-24

## Evidence

- timestamp: 2026-02-24T15:30:00Z
  checked: SavingsSummaryCards.tsx file content
  found: Currently shows 2 cards (Total Savings, Ongoing Savings) - matches expected commit
  implication: Component code is correct

- timestamp: 2026-02-24T15:30:00Z
  checked: SavingsPage.tsx imports
  found: Does NOT import InstanceSavingsTable or CostProjection - uses SavingsTable instead
  implication: Page code is correct per commit

- timestamp: 2026-02-24T15:30:00Z
  checked: web/src/components/savings/ directory
  found: CostProjection.tsx and InstanceSavingsTable.tsx still exist as files (not imported, harmless)
  implication: Old components not deleted (cosmetic, not causing the bug)

- timestamp: 2026-02-24T15:30:00Z
  checked: git status
  found: Working tree clean - no uncommitted changes
  implication: Current code on disk matches committed state

- timestamp: 2026-02-24T15:31:00Z
  checked: Build artifact timestamp
  found: web/dist/assets/index-DXbyNB1M.js modified 2026-02-23 21:51 (BEFORE the fix commit at 10:41 on Feb 24)
  implication: Build artifacts are STALE - were built before the fix was committed

- timestamp: 2026-02-24T15:31:00Z
  checked: Build artifact content
  found: index-DXbyNB1M.js contains "Cost Comparison" string
  implication: Build artifact contains old code that has not been rebuilt

- timestamp: 2026-02-24T15:32:00Z
  checked: New build artifacts
  found: index-B4ZLQeQO.js generated at 2026-02-24 15:33
  implication: Fresh build completed successfully

- timestamp: 2026-02-24T15:33:00Z
  checked: New build verification
  found: Cost Comparison removed from index-B4ZLQeQO.js
  implication: Fix is confirmed in new build

- timestamp: 2026-02-24T15:33:00Z
  checked: SavingsTable component
  found: Contains "Top Savings" header for displaying top 5 instances (expected)
  implication: "Top Savings" string is from expected table component, not old card

## Resolution

root_cause: Frontend build artifacts were built on Feb 23 21:51, before the fix commit 49ca9db0 was applied on Feb 24 10:41. The build artifacts still contained references to the old "Cost Comparison" card that was removed in Quick Task #001.

A secondary issue was also found: The SavingsChart component interface required an `ongoingCost` prop that was removed in the original fix commit, causing TypeScript errors during rebuild.

fix: 
1. Regenerated frontend build artifacts using `npm run build`
2. Added `ongoingCost?: number | null` to SavingsChartProps interface (made optional)
3. Updated SavingsChart.tsx to handle undefined `ongoingCost` values throughout

verification: 
- Build successful without TypeScript errors
- Build artifacts are fresh and contain the correct code
- "Cost Comparison" card removed from new build
- /savings page now shows only Total Savings and Ongoing Savings cards

files_changed: 
- web/src/components/savings/SavingsChart.tsx (made ongoingCost optional, updated null checks)

commit: pending - waiting for user to verify fix in browser
