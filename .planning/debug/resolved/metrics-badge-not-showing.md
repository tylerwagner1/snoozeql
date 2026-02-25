---
status: resolved
trigger: "no, there's no badge. i would guess something is cached or stale"
created: 2026-02-24T00:00:00Z
updated: 2026-02-24T00:00:00Z
---

## Current Focus

resolution: Badge condition had incorrect `instance.status !== 'stopped'` requirement
fix_complete: Removed the status check condition

## Symptoms

expected: Yellow "Metrics unavailable" badge near Metrics section header when no recent metrics exist.
actual: Badge does not appear
errors: None reported
reproduction: Navigate to instance detail page when metrics are stale (empty or >30 min old)
started: User reports "something is cached or stale"

## Eliminated

- Issue is a cache problem: Frontend builds successfully, code changes are present
- Issue is component not rendering: Component renders fully, badge just doesn't appear
- Issue is isMetricsStale() function: Function logic is correct (checks length === 0 or latest < 30 min)

## Evidence

- Line 232-236 badge conditional: `isMetricsStale() && instance.status !== 'stopped'`
- The condition requires BOTH stale metrics AND instance NOT being stopped
- If instance.status === 'stopped', badge will NOT render even if metrics are stale
- This is incorrect behavior - metrics are unavailable regardless of whether instance is stopped or running
- FIX APPLIED: Removed `instance.status !== 'stopped'` condition from badge on line 232

---

## ROOT CAUSE FOUND

**Debug Session:** .planning/debug/resolved/metrics-badge-not-showing.md

**Root Cause:** The badge rendering condition on line 232 included `instance.status !== 'stopped'`, which prevented the badge from appearing when an instance is stopped - even though metrics would still be unavailable in that case.

**Evidence Summary:**
- Badge conditional: `isMetricsStale() && instance.status !== 'stopped' && (...)`
- When instance status is 'stopped', the badge will not render even if metrics are stale/empty
- This is logically incorrect - metrics unavailability should be reported regardless of instance power state

**Files Involved:**
- web/src/pages/InstanceDetailPage.tsx:232-236

**Fix Applied:** Removed `instance.status !== 'stopped'` condition from badge rendering logic

**Verification:** Badge should now appear whenever `isMetricsStale()` returns true, regardless of instance status