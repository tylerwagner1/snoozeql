---
quick: 009
phase: quick
plan: 009
subsystem: recommendations
tags:
  - recommendations
  - filtering
  - schedule-check
  - GetAllRecommendations
requires:
  - Phase quick 008: Exclude instances with existing schedules
provides:
  - Filtered recommendations in GetAllRecommendations: Instances with enabled schedules no longer appear in recommendations list
affects:
  - UI recommendations display: Only unscheduled instances shown
tech-stack:
  added:
    - Filter logic in GetAllRecommendations handler
  patterns:
    - Schedule filtering before grouping recommendations
---

# Phase quick 009: Filter recommendations for instances with existing schedules in GetAllRecommendations Summary

## Overview

Successfully implemented filtering to exclude instances with existing enabled schedules from the `GetAllRecommendations` handler. This ensures the recommendations list only shows instances that don't already have schedules, matching the filtering logic already applied in `GenerateRecommendations`.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Filter after enrichment, before grouping | Maintain consistency with `GenerateRecommendations` pattern; instance details still available for display even if recommendation will be filtered |
| Include recommendations if instance/schedule fetch fails | Conservative approach: if we can't verify schedule status, include the recommendation to avoid false negatives |
| Log DEBUG messages when skipping | Allows tracking of how many recommendations are being filtered, useful for debugging and monitoring |

## Deviations from Plan

None - plan executed exactly as written.

## Files Modified

| File | Changes |
|------|---------|
| `internal/api/handlers/recommendations.go` | Added schedule filtering loop in `GetAllRecommendations` after enrichment (lines 157-184) |

## Commits

- `7b423518`: feat(quick-009-01): filter recommendations for instances with existing schedules

## Verification

- ✅ `go build -mod=mod ./...` compiles without errors
- ✅ Filter logic applied before `groupRecommendations` call
- ✅ DEBUG logging for skipped recommendations (tracks how many are filtered)
- ✅ Pattern matches `GenerateRecommendations` filtering (lines 350-373)

## Code Review

The filtering loop follows the same pattern as `GenerateRecommendations`:

1. Get instance by ID
2. Call `scheduleStore.GetMatchingSchedules(*instance)`
3. If schedules exist, skip with debug log
4. Otherwise, include in filtered slice

The `filtered` slice is used instead of `enriched` when calling `groupRecommendations` (line 301 in original).

## Testing Recommendations

1. **Test with existing schedule**: Create a schedule matching an instance ID appearing in recommendations, then call `GET /api/v1/recommendations`. That instance should NOT appear in the returned recommendations.
2. **Test without schedule**: Remove the schedule, call `GET /api/v1/recommendations`. The instance should now appear if it has a valid pattern.
3. **Verify debug logs**: Check application logs for `DEBUG: Skipping recommendation for ...` messages when filtering occurs.
4. **Verify build**: Run `go build ./...` to confirm no compilation errors.

## Metrics

- **Duration**: Quick execution ~3 minutes
- **Lines changed**: 28 insertions, 1 deletion
- **Tests added**: None (manual verification required - no test framework in place)

## Next Phase Readiness

**Ready for:** Next quick task or full phase.

**Conclusions:** Filtering successfully integrated into `GetAllRecommendations`. The handler now properly excludes instances with existing enabled schedules consistent with `GenerateRecommendations` behavior. No blockers or concerns carried forward.
