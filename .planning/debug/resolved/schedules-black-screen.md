## DEBUG COMPLETE

**Debug Session:** .planning/debug/resolved/schedules-black-screen.md

**Root Cause:** `schedule.selectors` can be `null` from the API, causing `TypeError: Cannot read properties of null (reading 'length')` when rendering schedule rows

**Fix Applied:** Modified `SchedulesPage.tsx` useEffect to map schedules and default `selectors` to empty array if null

**Verification:** Applied defensive programming pattern matching `ScheduleEditPage.tsx` which already handles this correctly

**Files Changed:**
- `web/src/pages/SchedulesPage.tsx`: Added `safeSchedules` mapping with selectors default

**Commit:** 37ca8d43
