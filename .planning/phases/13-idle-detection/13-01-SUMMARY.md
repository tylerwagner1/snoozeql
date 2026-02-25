---
phase: 13-idle-detection
plan: 01
subsystem: analyzer
tags: [idle-detection, compound-threshold, cpu, connections]

# Dependency graph
requires:
  - phase: 10-metrics-collection
    provides: FreeableMemory metric for complete picture
provides:
  - Compound idle threshold: CPU < 5% AND connections == 0
affects: idle-pattern-detection, recommendations

# Tech tracking
tech-stack:
  added: []
  patterns: [compound-threshold, cpu-connections-conjunction]

key-files:
  created: []
  modified:
    - internal/analyzer/patterns.go

key-decisions:
  - "CPUPercent: 5.0" - Requirement REC-01 specifies CPU < 5% for idle flag
  - "ConnectionsThreshold: 0" - Requirement REC-01 specifies connections = 0 for idle flag
  - "Use <= for connections check" - Ensures负conns <= 0负 handles edge case where conns could be 0 or negative

patterns-established:
  - "Compound idle threshold: idle detection requires both conditions to be true"
  - "Float64 for threshold fields" - Consistent with existing threshold pattern

# Metrics
duration: 5min
completed: 2026-02-25
---

# Phase 13: Idle Detection Summary

**Compound idle threshold using CPU < 5% AND connections = 0 to prevent false positives on active instances**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-25T00:00:00Z
- **Completed:** 2026-02-25T00:05:00Z
- **Tasks:** 3
- **Files modified:** 1

## Accomplishments
- Added ConnectionsThreshold field to ActivityThresholds struct
- Updated DefaultThresholds() with CPUPercent: 5.0, ConnectionsThreshold: 0
- Modified findIdleSegments() to check both CPU < 5% AND connections == 0 before marking hour as idle
- Updated comments to reflect new 5% CPU threshold

## Task Commits

Each task was committed atomically:

1. **Task 1: Update ActivityThresholds struct and defaults** - `cf349402` (feat)
2. **Task 2: Update findIdleSegments to check compound threshold** - `cf349402` (feat, same commit as Task 1)
3. **Task 3: Update struct comment to reflect new threshold** - `cf349402` (feat, same commit as Task 1)

**Plan metadata:** docs(13-01): complete idle detection compound threshold plan

## Files Created/Modified
- `internal/analyzer/patterns.go` - Added ConnectionsThreshold field, updated DefaultThresholds(), updated findIdleSegments() compound check

## Decisions Made
- Used `<=` for connections comparison since threshold is 0 and connections can't be negative
- All changes in single commit since tasks are closely related

## Deviations from Plan

**1. [Rule 1 - Bug] Duplicate field name in struct literal**

- **Found during:** Task 1 (updating DefaultThresholds())
- **Issue:** When adding ConnectionsThreshold field, accidentally added it twice causing compile error "duplicate field name ConnectionsThreshold"
- **Fix:** Removed duplicate line
- **Files modified:** internal/analyzer/patterns.go
- **Verification:** Changed `go build ./...` from failing to passing
- **Committed in:** cf349402 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Bug detected during build; fix was simple (remove duplicate line). No scope creep.

## Issues Encountered

None - all planned verification commands passed:
- `go build ./...` - Compiles without errors
- `grep -B2 -A3 "isIdle = cpu" internal/analyzer/patterns.go` - Shows compound condition
- `grep "CPUPercent:" internal/analyzer/patterns.go` - Shows CPUPercent: 5.0
- `grep -A6 "type ActivityThresholds struct" internal/analyzer/patterns.go` - Shows both CPUPercent and ConnectionsThreshold fields

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Idle detection foundation complete with compound threshold
- REC-01 requirement satisfied: CPU < 5% AND connections = 0
- Ready for Phase 14 (Grouped Recommendations) which will use these improved idle patterns
- No blockers - clean transition

---

*Phase: 13-idle-detection*
*Completed: 2026-02-25*
