## DEBUG COMPLETE

**Debug Session:** .planning/debug/resolved/cost-over-time-graph-black-screen.md

**Root Cause:** Two issues caused the black screen:
1. UI components used semi-transparent dark backgrounds (bg-slate-800/50, bg-slate-700/50) that blended with the page's black background, making content invisible when data was empty
2. SavingsSummaryCards.tsx accessing `data.top_savers.length` when `data` was `null`, causing a JavaScript crash that rendered the entire UI black

**Fix Applied:**
- Changed all component backgrounds from semi-transparent to fully opaque (bg-slate-800, bg-slate-700)
- Added explicit dark background to page wrapper (bg-slate-900 min-h-screen p-6)
- Added null check in SavingsSummaryCards.tsx with proper empty state rendering
- Updated SavingsSummaryCards.tsx to return placeholder cards with "No data available" when data is null

**Files Changed:**
- /Users/tylerwagner/snoozeql/web/src/pages/SavingsPage.tsx
- /Users/tylerwagner/snoozeql/web/src/components/savings/SavingsChart.tsx
- /Users/tylerwagner/snoozeql/web/src/components/savings/CostProjection.tsx
- /Users/tylerwagner/snoozeql/web/src/components/savings/SavingsSummaryCards.tsx
