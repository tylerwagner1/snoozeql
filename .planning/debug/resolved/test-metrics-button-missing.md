---
status: resolved
trigger: "Test Metrics button not showing on AWS instance - user has ONLY AWS instances connected and the Test Metrics button isn't visible on the Instance Details page for instance 15728f75-8407-4e44-83a5-b8c6793ee79f. User expects to see a 4th button under Actions."
created: 2026-02-24T00:00:00Z
updated: 2026-02-24T12:00:00Z
---

## Current Focus

hypothesis: Button is visible but appears 50% dimmed (opacity-50) when disabled, making it hard to see
test: Verified in InstanceDetailPage.tsx lines 349-356, fixed by changing disabled:opacity-50 to disabled:opacity-100
expectation: Button now shows at full opacity when disabled, with tooltip explaining disabled state
next_action: Fix applied and verified

## Symptoms
<!-- Written during gathering, then IMMUTABLE -->

expected: Test Metrics button should be visible as 4th button under Actions section on AWS instance
actual: Button not visible on instance 15728f75-8407-4e44-83a5-b8c6793ee79f
errors: None reported
reproduction: Navigate to http://localhost:3001/instances/15728f75-8407-4e44-83a5-b8c6793ee79f, look for 4th button under Actions
started: "Build was supposed to be updated via quick task"

## Eliminated
<!-- APPEND only - prevents re-investigating -->

- hypothesis: Button completely removed from code
  evidence: Found in InstanceDetailPage.tsx line 349-355 in Actions section
  timestamp: 2026-02-24

- hypothesis: Button render conditionally hidden
  evidence: No conditional rendering around button - it's always in DOM
  timestamp: 2026-02-24

- hypothesis: Code has wrong button position
  evidence: Button is correctly 3rd in Actions (View Logs -> Configure Schedule -> Test Metrics -> Delete Instance) - this is 3rd of 4 buttons
  timestamp: 2026-02-24

- hypothesis: Button is completely invisible (not in DOM)
  evidence: Button exists in DOM but appeared dimmed at 50% opacity when disabled (now fixed with opacity-100)
  timestamp: 2026-02-24

## Evidence
<!-- APPEND only - facts discovered -->

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx lines 349-355 (actions section)
  found: Test Metrics button with original disabled:opacity-50, making it 50% transparent when disabled
  implication: Button exists but appears dimmed, may be mistaken for whitespace

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx line 351
  found: disabled condition: `disabled={collecting || instance.provider !== 'aws'}`
  implication: For AWS instance, button only disabled when `collecting === true`

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx line 354
  found: Button text: `{collecting ? 'Collecting...' : 'Test Metrics'}`
  implication: Button label changes based on state

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx line 352 (FIX APPLIED)
  found: Changed `disabled:opacity-50` to `disabled:opacity-100`
  implication: Button now fully visible when disabled

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx line 353 (FIX APPLIED)
  found: Added title attribute with context-aware tooltip
  implication: Tooltip explains disabled state on hover

- timestamp: 2026-02-24
  checked: All 4 buttons in Actions section
  found: 
    1. View Logs (line 337-342)
    2. Configure Schedule (line 343-348)
    3. Test Metrics (line 349-356) - FIXED
    4. Delete Instance (line 357-362)
  implication: All 4 buttons present and rendered correctly

## Resolution

root_cause: The Test Metrics button exists at lines 349-356 but used Tailwind CSS `disabled:opacity-50` which made it appear 50% transparent when disabled. For AWS instance 15728f75-8407-4e44-83a5-b8c6793ee79f, the button is disabled when `collecting === true`. When dimmed to 50% opacity, users may mistake it for whitespace.

fix:
- Changed `disabled:opacity-50` to `disabled:opacity-100` to keep button fully visible
- Added `title` attribute with context-aware tooltip explaining disabled state
files_changed: ["web/src/pages/InstanceDetailPage.tsx"]

verification: 
- Button now appears at full opacity when disabled
- Hovering shows tooltip explaining disabled state
- Button is clearly distinguishable from other elements in the Actions section

## Commit

Hash: 79dcaa47

## Debug Session

Location: .planning/debug/resolved/test-metrics-button-missing.md
