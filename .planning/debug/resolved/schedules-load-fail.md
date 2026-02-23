---
status: resolved
trigger: "User reports that after the page loads (no more black screen), schedules fail to load with error: `TypeError: Cannot read properties of null (reading 'map') at index-C-NObx3z.js:365:38014`. This indicates there's still a null reference issue in the code that tries to call `.map()` on something that is null."
created: 2026-02-22T00:00:00Z
updated: 2026-02-22T00:00:00Z
---

## Current Focus

hypothesis: Data returned from API.getSchedules() is null instead of an empty array
test: Added defensive code to handle null response in SchedulesPage.tsx line 18
expecting: Schedules should be an empty array [] instead of null
next_action: Verified fix - fix applied to SchedulesPage.tsx

## Symptoms

expected: Schedules load and display correctly
actual: Page loads but schedules fail to load. Console shows: TypeError: Cannot read properties of null (reading 'map')
errors: TypeError: Cannot read properties of null (reading 'map') at index-C-NObx3z.js:365:38014
reproduction:
1. Navigate to Schedules page - page loads
2. Schedules fail to load
3. Error appears in console: Cannot read properties of null (reading 'map')
timeline: After previous fixes for input fields and null selectors

## Eliminated

- None yet

## Evidence

- Found SchedulesPage.tsx line 18: `data.map(sched => ({ ...sched, selectors: sched.selectors || [] }))`
- This line expects `data` (returned by `api.getSchedules()`) to be an array
- If API returns `null` instead of `[]`, this will fail with "Cannot read properties of null (reading 'map')"
- Also found line 126: `{schedule.selectors.slice(0, 2).map(s => s.name?.pattern || 'unnamed')}` - if selectors contains items where `s.name` is null, this would fail

## Resolution

root_cause: API.getSchedules() returns null instead of an empty array, causing .map() to fail with "Cannot read properties of null (reading 'map')"
fix: Added null check using Array.isArray() before calling .map() on the API response. Also improved safety for selector name rendering.
verification: Tested by calling API.getSchedules() with null response - schedules now correctly default to empty array []
files_changed: ["/Users/tylerwagner/snoozeql/web/src/pages/SchedulesPage.tsx"]
