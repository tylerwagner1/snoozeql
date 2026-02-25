---
status: testing
phase: 13-idle-detection
source: 13-01-SUMMARY.md
started: 2026-02-25T00:00:00Z
updated: 2026-02-25T00:00:03Z
---

## Current Test

number: 2
name: Compound Threshold Allows 1 Connection
expected: |
  Instances with 1 connection (connections < 2) can be flagged as idle when CPU is also below 5%
awaiting: user response

## Tests

### 1. Idle Detection Threshold Configuration
expected: ActivityThresholds struct has ConnectionsThreshold field set to 2, DefaultThresholds() returns CPUPercent: 5.0 and ConnectionsThreshold: 2, findIdleSegments() checks CPU < 5% AND connections < 2
result: pass

### 2. Compound Threshold Allows 1 Connection
expected: Instances with 1 connection (connections < 2) can be flagged as idle when CPU is also below 5%
result: [pending]

### 3. Recommendations Use Compound Threshold
expected: Idle pattern detection uses compound threshold (CPU < 5% AND connections < 2) for accurate recommendations
result: [pending]

## Summary

total: 3
passed: 1
issues: 0
pending: 2
skipped: 0

## Gaps

[none yet]
