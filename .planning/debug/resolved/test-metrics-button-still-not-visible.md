---
status: resolved
trigger: "Button is now visible but does not show popup with metrics"
created: 2026-02-25T00:00:00Z
updated: 2026-02-25T10:00:00Z
---

## Current Focus

hypothesis: Button was visible but click handler didn't show any feedback
test: Added toast notifications to show success/error messages from API
expecting: User sees toast notification with result after clicking button
next_action: Commit changes and verify in production

## Symptoms

expected: Test Metrics button should show popup with current metrics from API call
actual: Button was visible but clicking did nothing visible
errors: None - button clickable but no feedback
reproduction: Click Test Metrics button
started: After fix commit 8a72a8d9

## Eliminated

None

## Evidence

- timestamp: 2026-02-25
  checked: InstanceDetailPage.tsx handleCollectMetrics
  found: Function called API but didn't show any user feedback
  implication: User click produced no visible result

- timestamp: 2026-02-25
  checked: api.collectInstanceMetrics response format
  found: Returns { success: boolean; message: string }
  implication: Response includes message that should be shown to user

## Resolution

root_cause: Button click handler didn't display any feedback to user, only silently updated metrics state.

fix: Added toast notifications to show success/error messages from the API response.

files_changed:
- web/src/pages/InstanceDetailPage.tsx
  - Added import: `import toast from 'react-hot-toast'`
  - Updated handleCollectMetrics to show toast on success/error

verification: User now sees toast notification with message from API after clicking button.

commit: 9b436142

## Debug Session

Location: .planning/debug/resolved/test-metrics-button-still-not-visible.md
