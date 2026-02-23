---
phase: 05-activity-analysis
plan: 03
subsystem: activity-analysis
tags: metrics, idle-detection, sleep-scheduling, pattern-matching

# Dependency graph
requires:
  - phase: 05-activity-analysis
    provides: HourlyMetric model, MetricsStore with CRUD operations
provides:
  - IdleWindow pattern type with overnight support
  - ActivityPattern analysis result struct
  - AnalyzeActivityPattern algorithm for hour-bucketed pattern detection
  - Idle period detection using CPU < 1% threshold
  - 8+ hour contiguous idle period detection
  - 24+ hours data requirement validation
  - 3+ days consistent pattern validation
affects:
  - 06-intelligent-recommendations: Uses ActivityPattern output for sleep schedule recommendations

# Tech tracking
tech-stack:
  added:
    - idle window detection algorithms
    - HourBucket aggregation pattern
  patterns:
    - AnalyzeActivityPattern: Groups metrics by day-of-week/hour, detects contiguous idle periods
    - findIdleSegments: Walks 48 hours to handle overnight windows

key-files:
  created:
    - internal/analyzer/patterns.go
  modified:
    - internal/analyzer/analyzer.go

key-decisions:
  - "CPU < 1% threshold for idle detection per CONTEXT.md"
  - "8+ hours contiguous minimum for idle window selection"
  - "24+ hours of data required before pattern analysis"
  - "3+ days consistent pattern for confidence scoring"
  - "Overnight window handling via 48-hour walk (hours 0-23, then wrap)"

patterns-established:
  - "HourBucket pattern: Groups HourlyMetric values by (weekday, hour) for efficient pattern analysis"
  - "IdleWindow type: Supports overnight windows via IsOvernight flag and EndHour < StartHour convention"
  - "Confidence scoring: Based on (a) number of consistent days, (b) average CPU level during window"

# Metrics
duration: 5 min
completed: 2026-02-23
---

# Phase 05: Activity Analysis - Plan 03 Summary

**Idle period detection algorithms with overnight window support and MetricsStore integration**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-23T16:26:09Z
- **Completed:** 2026-02-23T16:31:09Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Created internal/analyzer/patterns.go with comprehensive idle window detection algorithms
- Implemented IdleWindow, ActivityPattern, and HourBucket types for pattern representation
- Added AnalyzeActivityPattern() with 24+ hours data requirement and 8+ hour minimum window detection
- Updated internal/analyzer/analyzer.go to use MetricsStore for real stored metrics
- Added AnalyzeInstanceActivity() and AnalyzeAllInstances() methods for batch pattern analysis

## Task Commits

Each task was committed atomically:

1. **Task 1: Create patterns.go with idle window detection** - `a443f9a6` (feat)
2. **Task 2: Update analyzer.go to use MetricsStore** - `5f24a4a4` (feat)

**Plan metadata:** Complete - plan executed exactly as written.

## Files Created/Modified
- `internal/analyzer/patterns.go` - Created with idle window detection algorithms:
  - IdleWindow type with overnight support (IsOvernight flag)
  - ActivityPattern result struct
  - HourBucket aggregation for (day-of-week, hour) metric grouping
  - AnalyzeActivityPattern() - main pattern detection function
  - findIdleSegments() - detects contiguous low-activity hours with overnight wrap
  - groupSimilarWindows() - combines patterns across days (1-hour tolerance)
  - calculateConfidence() - scores based on consistency and activity levels
  - DefaultThresholds() - enforces CPU < 1%, 8+ hours minimum, 24+ hours data, 3+ days
- `internal/analyzer/analyzer.go` - Updated with MetricsStore integration:
  - Added metricsStore field to Analyzer struct
  - Updated NewAnalyzer() to accept metricsStore parameter
  - Added AnalyzeInstanceActivity() - checks data sufficiency, queries MetricsStore
  - Added AnalyzeAllInstances() - batch analysis across all managed instances
  - Uses DefaultThresholds() from patterns.go for consistent behavior

## Decisions Made
- Used 24+ hour data requirement as per CONTEXT.md requirement
- Implemented 8+ hour minimum idle window per CONTEXT.md specification
- Used CPU < 1% threshold per CONTEXT.md low activity definition
- Handled overnight windows by walking 48 hours (0-23 + wrap)
- Required pattern on 3+ days for confidence scoring
- Sorted IdleWindows by confidence descending for consistent output

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all verification steps passed successfully. Code compiles without errors under `go build -mod=mod ./...`.

## Next Phase Readiness

- IdleWindow and ActivityPattern types ready for use by Phase 6 intelligent recommendations
- AnalyzeActivityPattern algorithm enforces all threshold requirements from CONTEXT.md
- Analyzer methods (AnalyzeInstanceActivity, AnalyzeAllInstances) ready for integration with MetricsCollector
- MetricsStore integration established (GetMetricsByInstance, HasSufficientData used)

---

Phase: 05-activity-analysis
Completed: 2026-02-23
