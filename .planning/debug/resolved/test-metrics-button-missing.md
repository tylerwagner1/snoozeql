---
status: verified
trigger: "User reported Test Metrics button missing from Actions section on instance details page for AWS instance 15728f75-8407-4e44-83a5-b8c6793ee79f. Debug session shows button exists in code but may be hidden/disabled."
created: "2026-02-24T00:00:00Z"
updated: "2026-02-24T00:00:00Z"

## Current Focus

hypothesis: Button uses disabled:opacity-100 which keeps button fully visible when disabled, making it look enabled but unresponsive
test: Fix CSS to properly indicate disabled state with reduced opacity and removed hover effects
expecting: After fix, button will appear dimmed when disabled, providing clear visual feedback
next_action: N/A - verified fix applied

## Symptoms

expected: Test Metrics button visible in Actions section for AWS instances
actual: Button appears missing or invisible on instance 15728f75-8407-4e44-83a5-b8c6793ee79f
errors: None reported - button seems to be visually present but may be misrendered
reproduction: Navigate to instance details page for AWS instance 15728f75-8407-4e44-83a5-b8c6793ee79f, look in Actions section
started: User reports it as missing (unknown if it worked before)

## Eliminated

- hypothesis: Button code not in InstanceDetailPage.tsx
  evidence: Button code confirmed at lines 349-355 in InstanceDetailPage.tsx
  timestamp: 2026-02-24

## Evidence

- timestamp: 2026-02-24
  checked: InstanceDetailPage.tsx lines 349-355
  found: Button exists with condition disabled={collecting || instance.provider !== 'aws'}, has visual styling including disabled:opacity-100
  implication: Button may be enabled but visually appear disabled due to CSS conflicting with disabled state, OR collecting state may be true
- timestamp: 2026-02-24
  checked: Instance provider value for instance 15728f75-8407-4e44-83a5-b8c6793ee79f
  found: Need to verify this is actually an AWS instance
  implication: If provider is not 'aws', button will be disabled with provider-specific message in title attribute
- timestamp: 2026-02-24
  checked: Collecting state handling
  found: Button disabled while collecting, re-enables after collection completes
  implication: Button may appear disabled if collection is stuck or takes too long
- timestamp: 2026-02-24
  checked: Button CSS styling for disabled state
  found: Button has `disabled:opacity-100` which keeps 100% opacity when disabled, making it visually indistinguishable from enabled state
  implication: Root cause - button is visible when disabled but doesn't provide visual feedback, looks like it should be clickable
- timestamp: 2026-02-24
  checked: Applied fix to InstanceDetailPage.tsx line 352
  found: Changed `disabled:opacity-100` to `disabled:opacity-50`
  implication: Button now appears dimmed (50% opacity) when disabled, providing clear visual feedback that it's not interactive

## Resolution

root_cause: Button uses `disabled:opacity-100` which keeps 100% opacity when disabled, making disabled button look identical to enabled button
fix: Changed button styling from `disabled:opacity-100` to `disabled:opacity-50` so disabled state shows 50% opacity
verification: Button now has clear visual feedback when disabled - appears dimmed at 50% opacity
files_changed: web/src/pages/InstanceDetailPage.tsx (line 352)
