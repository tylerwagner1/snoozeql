---
phase: 14-grouped-recommendations
plan: 02
tasks_completed: 3
commits:
  - 6672980f: API types update
  - 8098e9bc: Full implementation commit
---

# Phase 14 Plan 02: Frontend Grouped Recommendations Summary

## Overview

Successfully implemented grouped recommendations view on the frontend. The recommendations page now displays recommendations organized by similar idle patterns, with expandable groups showing member instances and total savings.

## What Was Built

### API Type Updates
- **RecommendationGroup interface**: pattern_description, pattern_key, total_daily_savings, instance_count, recommendations array
- **GroupedRecommendationsResponse interface**: groups array
- Updated getRecommendations() to return GroupedRecommendationsResponse

### RecommendationGroup Component
- Pattern description header (e.g., "Idle 10pm to 6am, weekdays")
- Instance count display
- Total daily savings per group
- Expandable/collapse functionality
- Single-instance groups render as regular cards (no group wrapper)
- Nested RecommendationCard for each member

### Page Updates
**RecommendationsPage.tsx:**
- State: `groups: RecommendationGroup[]` instead of `recommendations: RecommendationEnriched[]`
- Fetches grouped data from API
- Displays groups with RecommendationGroup components
- Handles dismiss/confirm by removing from groups and updating totals
- Shows pending count from groups
- Empty state when no pending groups

**Dashboard.tsx:**
- Updated to handle grouped recommendations response
- Pending actions count calculated from groups
- Recommendations section shows first group's recommendations (up to 3)
- Dismiss/confirm handlers update groups state

## Files Modified

| File | Changes |
|------|---------|
| `web/src/lib/api.ts` | Added RecommendationGroup, GroupedRecommendationsResponse types; updated getRecommendations return type |
| `web/src/components/RecommendationGroup.tsx` | NEW: Main component for grouped display |
| `web/src/pages/RecommendationsPage.tsx` | Updated to render groups |
| `web/src/pages/Dashboard.tsx` | Updated to handle grouped response |

## Requirements Coverage

- ✅ **REC-02**: Recommendations are grouped by similar idle patterns
- ✅ **REC-03**: Each recommendation displays estimated daily savings (both per-group total and per-instance)

## Verification

```bash
# Build frontend check
cd web && npm run build  # Success

# API response structure (with running backend)
curl -H "Authorization: Bearer dev-key" http://localhost:8080/api/v1/recommendations | jq '.groups[0]'
# Should return: { pattern_description, pattern_key, total_daily_savings, instance_count, recommendations: [...] }

# UI verification
# Visit http://localhost:5173/recommendations
# - Expect grouped recommendations with expandable sections
# - Single-instance groups render as regular cards
# - Click group header to expand/collapse
```

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Single-instance groups as regular cards | Avoids awkward "Group: 1 instance" UI overhead |
| Always start expanded | Ensures visibility of all patterns on first load |
| Show first 3 recommendations in Dashboard | Preview without overwhelming dashboard |
| Update dismiss/confirm handlers | Maintain correct totals after removal |

## Success Criteria Met

- ✅ User sees recommendations organized into pattern groups
- ✅ Each group shows pattern description and total savings
- ✅ User can expand/collapse groups to see member instances
- ✅ Single-instance groups render as regular cards without group wrapper
- ✅ Per-instance savings still visible within groups

## Next Steps

Phase 14 complete. Ready to proceed to Phase 15 (if any) or release v1.2.

## Known Limitations

1. Group collapse state does not persist across page reloads
2. Dashboard only shows first 3 recommendations (from first group)
3. No pattern filtering or search within groups
