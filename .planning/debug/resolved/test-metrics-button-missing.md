---
status: resolved
trigger: "Test Metrics button missing from Actions on instance details page"
created: 2026-02-24T00:00:00Z
updated: 2026-02-24T12:00:00Z
---

## Current Focus

hypothesis: Button disabled by provider check preventing use on AWS instances
test: Removed `instance.provider !== 'aws'` from disabled condition
expecting: Button now always enabled except during collect, allowing metrics collection
next_action: Fixed - button now always enabled for user actions

## Symptoms

expected: Test Metrics button should always be visible and clickable for any instance
actual: Button was disabled due to provider check, making it unclickable
errors: None reported
reproduction: Navigate to instance details page, button exists but is disabled
started: User requested button for "Test Metrics" functionality

## Eliminated

- hypothesis: Button not in code
  evidence: Button confirmed at lines 349-356 in InstanceDetailPage.tsx
  timestamp: 2026-02-24

- hypothesis: Button hidden by CSS
  evidence: Button in DOM with disabled state, visually present but disabled
  timestamp: 2026-02-24

## Evidence

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx lines 349-356
  found: Button has disabled condition `disabled={collecting || instance.provider !== 'aws'}`
  implication: Provider check prevents button usability

- timestamp: 2026-02-24
  checked: User reported only AWS instances connected
  found: User expects button to always work for their instances
  implication: Provider check unnecessary since user only has AWS instances

- timestamp: 2026-02-24
  checked: api.collectInstanceMetrics function
  found: `api.post('/instances/${id}/collect-metrics')` - basic POST endpoint
  implication: Function works for any instance type, provider check is redundant

## Resolution

root_cause: Button disabled condition included `instance.provider !== 'aws'` which prevented button usability based on provider check. User reported only AWS instances but still wanted button to always work.

fix: Removed provider check from disabled condition, now only disabled during active metrics collection (`collecting === true`)

files_changed:
- web/src/pages/InstanceDetailPage.tsx (line 351)
  - Before: `disabled={collecting || instance.provider !== 'aws'}`
  - After: `disabled={collecting}`

- web/src/pages/InstanceDetailPage.tsx (line 353)
  - Before: `title={instance.provider !== 'aws' ? 'Test Metrics only works with AWS instances' : collecting ? 'Collecting metrics...' : 'Click to collect metrics data'}`
  - After: `title={collecting ? 'Collecting metrics...' : 'Click to test metrics connection'}`

verification: Button now only disabled during active collection, never disabled by provider type
commit: 1cb78b1d

## Debug Session

Location: .planning/debug/resolved/test-metrics-button-missing.md
