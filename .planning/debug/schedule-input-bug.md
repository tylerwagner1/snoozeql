---
status: resolved
trigger: "User reports they cannot type in the Name pattern field or select from the Pattern type dropdown when creating a schedule. The cursor appears but typing doesn't update the field, and dropdown options can't be selected."
created: 2026-02-22T00:00:00Z
updated: 2026-02-22T00:08:30Z
---

## Current Focus

hypothesis: updateSelector function has wrong property path - sets selector.field instead of selector.name.field
test: Code review reveals bug in computed property usage
expecting: updateSelector sets selector.name[field] but should set selector.name[field]
next_action: Fix updateSelector function in both ScheduleNewPage.tsx and ScheduleEditPage.tsx

## Symptoms

expected: User can type in Name pattern field and change Pattern type dropdown
actual: Name pattern field accepts cursor but typing doesn't update, Pattern type dropdown shows options but selection doesn't work
errors: None visible in UI or console
reproduction: 
1. Navigate to Schedules tab
2. Click "Create Schedule"
3. Click on Name pattern field or Pattern type dropdown
4. Try to type or select - doesn't work
started: After deployment fixes for regex and audit logging

## Eliminated

- hypothesis: API issue with Schedule interface
  evidence: Selector interface properly defined as { name: { pattern: string; type: string } }
  timestamp: 2026-02-22T00:05:30Z

- hypothesis: Backend not accepting data
  evidence: Issue occurs on frontend input, data not yet sent to backend
  timestamp: 2026-02-22T00:05:30Z

- hypothesis: React version or hook issue
  evidence: Other inputs like timezone selector work, only selector inputs affected
  timestamp: 2026-02-22T00:06:00Z

## Evidence

- 2026-02-22T00:05:00Z: Examined ScheduleNewPage.tsx - code looks correct at first glance
- 2026-02-22T00:05:30Z: Examined ScheduleEditPage.tsx - code looks correct at first glance
- 2026-02-22T00:06:00Z: Examined SchedulesPage.tsx - navigation to /schedules/new is correct
- 2026-02-22T00:06:30Z: Checked index.css - no blocking styles found
- 2026-02-22T00:06:30Z: Verified Selector interface: { name: { pattern: string; type: string } }
- 2026-02-22T00:07:30Z: FOUND BUG: updateSelector uses `[field]` which creates selector.pattern (flat) not selector.name.pattern (nested)
- 2026-02-22T00:08:00Z: confirmed: selector.name?.pattern in JSX but updateSelector sets selector[field] not selector.name[field]

## Resolution

root_cause: The updateSelector function incorrectly used computed property `[field]` on the selector object directly, but the selector structure is { name: { pattern: string; type: string } }. When typing 'pattern', it tried to set selector['pattern'] = value, but the selector has no 'pattern' property at the top level - it's nested under 'name'. React detects no state change because selector['pattern'] is undefined (no change from undefined) so the input appears frozen.

fix: Changed updateSelector to properly update selector.name[field] by creating a new name object with spread: name: { ...selector.name, [field]: value }

verification: BUILD VERIFIED - TypeScript compiled successfully, Vite production build succeeded. The fix addresses the root cause by properly traversing the nested selector.name structure when updating pattern or type fields.
files_changed: ["web/src/pages/ScheduleNewPage.tsx", "web/src/pages/ScheduleEditPage.tsx"]
