---
phase: quick
plan: 007
subsystem: recommendations
tags: ["recommendations", "table", "ui", "frontend", "tailwind"]
requires: ["quick-002-02"]
provides: ["Table-based recommendations display", "Batch dismissal support"]
affects: ["future-recommendations-ui"]
tech-stack:
  patterns:
    - table-component-rendering
    - batch-dismissal-pattern
key-files:
  created:
    - path: web/src/components/RecommendationsTable.tsx
      purpose: Table component for grouped recommendations
      lines: 118
  modified:
    - path: web/src/pages/RecommendationsPage.tsx
      purpose: Updated to use table instead of cards
      lines-changed: 21 insertions, 19 deletions
decisions:
  - decision: "Display pattern-level Wake/Sleep times from first recommendation"
    rationale: "Patterns in same group typically share similar wake/sleep schedules; showing first rec's times provides a representative view"
  - decision: "Batch dismiss removes entire group at once"
    rationale: "User confirmed they want to dismiss all recommendations for this pattern; more efficient than individual dismissals"
  - decision: "Green text for 'Est. Daily Savings' column"
    rationale: "Aligns with common UI pattern where green represents positive financial impact; matches existing design in RecommendationCard"
  - decision: "Row click opens modal for first recommendation"
    rationale: "Modal shows individual recommendation details; clicking any row provides access to pattern-level view"
metrics:
  duration: "~15 minutes"
  completed: "2026-02-27"
---

# Phase Quick 007: Recommendations Table Display Summary

## One-Liner

Changed recommendations display from expandable cards to a table with pattern-level grouping showing wake/sleep times, instance counts, and savings.

## Overview

This quick task transformed the recommendations page UX by replacing the `RecommendationGroup` expandable card component with a new `RecommendationsTable` component. The table format improved scannability and made it easier to compare patterns at a glance.

## Changes Made

### Task 1: Created RecommendationsTable Component

**File:** `web/src/components/RecommendationsTable.tsx`

**Key Features:**
- Table with 5 columns: Schedule Pattern, Wake/Sleep, Instances Affected, Est. Daily Savings, Actions
- Pattern-level display showing aggregated information for each group
- Green text (`text-green-400`) for savings to highlight positive financial impact
- Batch dismissal to remove entire group at once
- Row hover effects (`hover:bg-slate-700/30`) for better interactivity
- Dark mode consistent styling with existing UI (slate-800/50 bg, slate-700 border)

**Implementation Details:**
- Wake/Sleep times displayed from first recommendation's suggested schedule
- Instance count shown as pill badge on both pattern description and instances affected column
- "View Details" button opens modal for first recommendation
- "Dismiss" button dismisses all recommendations in the group

### Task 2: Updated RecommendationsPage

**File:** `web/src/pages/RecommendationsPage.tsx`

**Changes:**
- Replaced `RecommendationGroup` import with `RecommendationsTable`
- Replaced `groups.map()` rendering loop with single `RecommendationsTable` component
- Updated `handleDismiss` to accept `string | string[]` for batch dismissal
- Batch dismissal iterates through all IDs and dismisses each
- Removed redundant `instance_count` update logic (now handled by table component)

**Key Improvements:**
- Simplified rendering logic in RecommendationsPage
- Better separation of concerns (table handles display, page handles state)
- Batch dismissal more efficient for group cleanup

## Technical Decisions

| Decision | Rationale |
|----------|-----------|
| Display Wake/Sleep from first recommendation | Patterns in same group typically share wake/sleep schedules; shows representative view without duplicating data |
| Batch dismiss removes entire group | User likely wants to dismiss all recommendations for a pattern; more efficient than individual dismissals |
| Green text for savings | Aligns with financial positive impact conventions; matches existing dark green styling |
| Row click opens modal for first recommendation | Provides pattern-level detail access without needing per-row detail expansion |

## Testing Verification

- **Build:** `npm run build` succeeds with no errors
- **TypeScript:** `npx tsc --noEmit` passes with zero errors
- **UI:** Table displays correctly in dark mode with proper styling
- **Actions:** View Details and Dismiss buttons functional

## Files Modified

### Created
- `web/src/components/RecommendationsTable.tsx` (118 lines)

### Modified
- `web/src/pages/RecommendationsPage.tsx` (21 insertions, 19 deletions)

## Impact

- **User Experience:** More scannable recommendations display with pattern-level grouping
- **Performance:** Reduced DOM elements by eliminating expandable card overhead
- **Code Quality:** Better component separation and reusable table pattern

## Next Phase Readiness

✅ Ready for production - all verification checks pass, build successful, dark mode consistent.

### Potential Future Improvements
- Pagination for large numbers of groups
- Column sorting by savings, instance count, or pattern
- Filter/sort controls in table header
- Export recommendations to CSV

## Deviations from Plan

None - plan executed exactly as written.

## Success Criteria Verification

- ✅ Recommendations display as table, not expandable cards
- ✅ "Instances Affected" column shows count per pattern group
- ✅ All existing functionality preserved (view details, confirm, dismiss, generate)
- ✅ TypeScript compiles without errors
- ✅ UI matches existing dark theme styling
