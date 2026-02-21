---
phase: 02-manual-control-audit
plan: 04
subsystem: ui
tags:
  - multi-select
  - bulk-actions
  - ui-components
  - confirmation-dialogs
requires:
  - 02-02
  - 02-03
provides:
  - Multi-select table UI with checkbox column
  - Bulk action buttons (Sleep Selected, Wake Selected)
  - Confirmation dialogs for bulk operations
affects:
  - 02-05
---

# Phase 02 Plan 04: Multi-select and Bulk Actions Summary

Multi-select table with bulk action buttons in InstancesPage - users can select multiple instances and perform bulk sleep/wake operations with confirmation dialogs.

**One-liner:** Multi-select table with bulk action buttons, confirmation dialogs, and toast notifications for bulk sleep/wake operations

## Key Features

- **Checkbox column** in table header (selects all filtered instances) and rows (individual selection)
- **Bulk action buttons** appear when instances are selected, showing count of actionable instances
- **Confirmation dialogs** using Headless UI before executing bulk operations
- **Toast notifications** using react-hot-toast for operation success/failure
- **Partial success handling** - continues processing remaining instances if one fails
- **Optimistic UI updates** - instance status updates immediately after confirmation

## Decisions Made

- Used `Set<string>` for `selectedIds` state for efficient O(1) lookups
- Bulk action buttons show count of actionable vs total selected (e.g., "Sleep Selected (3)")
- Only "available" and "running" instances can be stopped; only "stopped" instances can be started
- Confirmation dialog shows exact count of instances being operated on
- Bulk operations return `BulkOperationResponse` with `success` and `failed` arrays
- Toast messages differentiate between success and failure cases

## Tech Tracking

### Added Dependencies
- `react-hot-toast@^2.4.1` - notifications library (already installed)

### New Patterns
- Selection state management with Set for O(1) lookups
- Computed properties based on filteredAndSortedInstances
- Bulk operation API integration with optimistic updates

### Files Created/Modified
- **Modified:** `web/src/pages/InstancesPage.tsx` (+157 lines)
  - Added selection state: `selectedIds`, `showConfirmDialog`, `bulkLoading`
  - Added handlers: `toggleSelect`, `selectAll`, `clearSelection`
  - Added computed: `allSelected`, `selectedInstances`, `stoppableSelected`, `startableSelected`
  - Added handlers: `handleBulkSleep`, `handleBulkWake`
  - Added UI: checkbox column, bulk action buttons, confirmation dialogs
- **Modified:** `web/src/components/ConfirmDialog.tsx` (already existed)
- **Modified:** `web/src/lib/api.ts` (BulkOperationResponse, bulkStopInstances, bulkStartInstances already existed)

## Metrics

- **Duration:** ~3 minutes
- **Tasks:** 2/2 complete
- **Lines added:** 157
- **Commits:** 1

## Verification

1. **Build verified:** `cd web && npm run build` compiles without errors
2. **Visual verification:** Table has checkbox column, bulk buttons appear when selection > 0
3. **Confirmation dialog:** Opens when clicking bulk action buttons
4. **Toast notifications:** Success/failure messages appear after operations
5. **Partial success:** If some instances fail, others still process

## Next Steps

- Plan 02-05: Create audit log page to view all events
- Human verification checkpoint after audit log page implementation
- Verify end-to-end workflow: select → confirm → operation → audit log
