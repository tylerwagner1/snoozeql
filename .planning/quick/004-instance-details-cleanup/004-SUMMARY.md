---
quick_task: 004
description: "Clean up Instance Details page - fix title, remove unused cards, reorganize layout"
status: complete
date: 2026-02-26
---

# Quick Task 004: Instance Details Page Cleanup

## What Changed

### 1. Fixed Title Text Color
- Changed `text-gray-900` to `text-foreground` on the h1 database name
- Title is now readable on dark backgrounds

### 2. Removed Quick Stats Card
- Removed the card showing Status, Current Cost (fake random value), and Idle Time
- Status is already visible in the header badge
- Cost was a placeholder with fake data
- Idle time calculation was removed as well

### 3. Removed Single-Datapoint Metrics Card
- Removed the old "Metrics" card with MetricCard components
- Removed MetricCard component definition and helper functions (getMetricValue, getMetricMin, getMetricMax, getMetricSamples)
- Time-series charts in MetricsChart component provide much better visibility

### 4. Updated All Colors for Dark Mode
- Cards: `bg-white` → `bg-card`
- Borders: `border-gray-*` → `border-border`
- Text: `text-gray-900` → `text-foreground`, `text-gray-500` → `text-muted-foreground`
- Status badges: Updated to use opacity-based colors (e.g., `bg-green-500/20 text-green-400`)
- Buttons: Updated to use semantic colors

### 5. Reorganized Layout
- Left column (2/3): Configuration, Tags, Metrics History (chart)
- Right column (1/3): Actions only
- Clean spacing with consistent card styling

### 6. Simplified Modal
- Updated MetricModal to use semantic colors
- Simplified metric display in modal

## Files Modified
- `web/src/pages/InstanceDetailPage.tsx` (reduced from 492 to ~263 lines)

## Verification
- `npm run build` passes with no TypeScript errors
- All cards use semantic color variables for dark mode compatibility
