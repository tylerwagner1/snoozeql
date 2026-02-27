---
status: resolved
trigger: "recent changes to Recent Activity / Audit Log aren't showing in my UI with hard refreshing and incognito"
created: 2026-02-27T15:00:00Z
updated: 2026-02-27T15:30:00Z
---

## Current Focus

**RESOLUTION:** Docker frontend container had stale build. Fixed TypeScript errors and rebuilt.

## Symptoms

expected: Changes to Recent Activity and Audit Log pages should be visible after hard refresh and incognito
actual: User sees old UI with filter buttons and "by User/schedule" text instead of icons and "Scheduled/Manual"
errors: None
reproduction: User visits http://localhost:3001 (or /audit-log) with hard refresh and incognito - remains unchanged
started: Just now (after code changes were made)

## Evidence

- timestamp: 2026-02-27T15:00:00Z
  checked: Git diff on AuditLogPage.tsx and Dashboard.tsx
  found: Changes confirmed - filter buttons removed, Calendar/Hand imports added, getTriggerBadge() function added
  implication: File changes ARE saved to git

- timestamp: 2026-02-27T15:00:00Z
  checked: Dev server running on port 3001
  found: Server started successfully, serving files
  implication: Server is running and should be serving updated files

- timestamp: 2026-02-27T15:00:00Z
  checked: curl http://localhost:3001/src/pages/Dashboard.tsx
  found: Output contains "View All", "Scheduled", "Manual" text
  implication: Vite dev server IS serving updated files correctly

- timestamp: 2026-02-27T15:00:00Z
  checked: curl http://localhost:3001/src/pages/AuditLogPage.tsx
  found: Output contains "Calendar, Hand" import
  implication: Vite dev server IS serving updated files correctly

- timestamp: 2026-02-27T15:15:00Z
  checked: /Users/tylerwagner/snoozeql/deployments/docker/
  found: Docker frontend container serves from /dist (static build), not from live Vite server
  implication: Vite in dev mode != production Docker build. User sees production build!

- timestamp: 2026-02-27T15:15:00Z
  checked: npm run build in web directory
  found: TS errors preventing rebuild:
    - ActivityGraph.tsx: 'metrics' unused
    - ActivityGraph.tsx: 'cpuRange' unused
    - api.ts: missing 'is_very_low_activity' field
    - RecommendationCard.tsx: undefined check for is_very_low_activity
  implication: Build was failing silently, leaving old assets in /dist

## Resolution

root_cause: 
Docker frontend container serves static files from `/dist` (production build), while the dev server (Vite on port 3001) serves source files. The user was accessing the production Docker build which had stale assets from a previous commit. Additionally, TypeScript compilation errors prevented the build from being updated.

fix: 
1. Fixed TypeScript errors:
   - Added `is_very_low_activity?: boolean` to `detected_pattern` interface in `web/src/lib/api.ts`
   - Fixed unused `metrics` variable and `cpuRange` variable in `ActivityGraph.tsx`
   - Fixed undefined check for `is_very_low_activity` in `RecommendationCard.tsx`
2. Rebuilt Docker container with `npm run build` in web directory
3. Restarted Docker frontend service

verification: 
- Docker now serves new assets: `index-DhcHfYCR.js`, `index-T2cYK12K.css`
- UI should display Updated Recent Activity with icons for Scheduled/Manual
- Audit Log page should show clean list without filter buttons

files_changed:
- web/src/lib/api.ts: Added missing `is_very_low_activity` field
- web/src/components/ActivityGraph.tsx: Fixed unused variable errors
- web/src/components/RecommendationCard.tsx: Fixed undefined property check

