# Phase 14: Grouped Recommendations - Research

**Researched:** 2026-02-25
**Domain:** Recommendation grouping, pattern similarity, savings display
**Confidence:** HIGH

## Summary

Phase 14 transforms the individual-instance recommendation view into a grouped view where instances with similar idle patterns are presented together. This builds on Phase 13's improved idle detection (CPU < 5% AND connections = 0) and the existing recommendation infrastructure from Phase 6.

The core challenges are:
1. **Defining pattern similarity** - What criteria make two idle patterns "similar" enough to group?
2. **Calculating and displaying savings** - Current system already has `estimated_daily_savings` but needs prominent display per-recommendation
3. **UI for grouped view** - Showing pattern groups with member instances while maintaining the existing card-based UX

The existing codebase provides strong foundations: `IdleWindow` has `StartHour`, `EndHour`, and `DaysOfWeek` that define patterns, `Instance.HourlyCostCents` enables savings calculation, and the `RecommendationCard` component already shows `estimated_daily_savings`.

**Primary recommendation:** Group recommendations by pattern signature (start hour ± tolerance, end hour ± tolerance, matching days), display each group as an expandable section showing total group savings and member instances, preserve existing per-recommendation savings display.

## Standard Stack

This phase uses existing codebase patterns - no new dependencies needed.

### Core (Already in Codebase)
| Component | Version | Purpose | Why Standard |
|-----------|---------|---------|--------------|
| Go Chi router | v5 | API endpoints | Already used for all routes |
| PostgreSQL + pgx | v5 | Recommendation storage | Existing store patterns |
| React | 18.2 | UI components | Already in use |
| Tailwind CSS | 3.4 | Styling | Used throughout |

### Supporting (Already in Codebase)
| Component | Version | Purpose | When to Use |
|-----------|---------|---------|-------------|
| lucide-react | 0.300 | Icons (ChevronDown, Users, etc.) | Group headers, instance lists |
| clsx | 2.0.0 | Conditional CSS | Complex class logic |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Backend grouping | Frontend grouping | Backend is simpler, avoids data duplication in API response |
| Hour-exact matching | Hour-range matching (±1 hour) | Range matching groups more instances but may be less precise |
| Weekday-exact matching | Weekday-overlap matching | Overlap allows "mostly same" patterns; exact is simpler |

## Architecture Patterns

### Recommended Project Structure

Extend existing structure with minimal changes:

```
internal/
├── analyzer/
│   └── recommendation.go      # Add pattern signature generation
├── api/handlers/
│   └── recommendations.go     # Add grouping logic to GetAllRecommendations

web/src/
├── pages/
│   └── RecommendationsPage.tsx    # Add grouped view rendering
├── components/
│   ├── RecommendationCard.tsx     # Keep for individual recs in groups
│   └── RecommendationGroup.tsx    # NEW: Group header with expand/collapse
├── lib/
│   └── api.ts                     # Extend with group response type
```

### Pattern 1: Pattern Signature for Grouping

**What:** Generate a grouping key from idle pattern attributes
**When to use:** When comparing patterns for similarity
**Why:** Allows backend to efficiently group without comparing every pair

A pattern signature should capture:
- **Start hour bucket** - Group hours into ranges (e.g., 18-20, 20-22, 22-00)
- **End hour bucket** - Same bucketing approach
- **Day pattern type** - "weekdays", "weekends", "daily"

```go
// Source: internal/analyzer/recommendation.go
type PatternSignature struct {
    StartBucket string   // e.g., "evening" (18-22), "night" (22-06)
    EndBucket   string   // e.g., "morning" (06-10)
    DayType     string   // "weekdays", "weekends", "daily"
}

func generatePatternSignature(pattern map[string]interface{}) PatternSignature {
    startHour := int(pattern["idle_start_hour"].(float64))
    endHour := int(pattern["idle_end_hour"].(float64))
    daysOfWeek := pattern["days_of_week"].([]interface{})
    
    return PatternSignature{
        StartBucket: hourToBucket(startHour),
        EndBucket:   hourToBucket(endHour),
        DayType:     daysToType(daysOfWeek),
    }
}

func hourToBucket(hour int) string {
    switch {
    case hour >= 6 && hour < 10:
        return "early-morning"  // 6-10 AM
    case hour >= 10 && hour < 14:
        return "midday"         // 10 AM - 2 PM
    case hour >= 14 && hour < 18:
        return "afternoon"      // 2-6 PM
    case hour >= 18 && hour < 22:
        return "evening"        // 6-10 PM
    default:
        return "night"          // 10 PM - 6 AM
    }
}

func daysToType(days []interface{}) string {
    if len(days) >= 7 {
        return "daily"
    }
    weekdays := 0
    weekends := 0
    for _, d := range days {
        day := d.(string)
        if day == "Saturday" || day == "Sunday" {
            weekends++
        } else {
            weekdays++
        }
    }
    if weekdays >= 4 && weekends == 0 {
        return "weekdays"
    }
    if weekends >= 2 && weekdays == 0 {
        return "weekends"
    }
    return "mixed"
}
```

### Pattern 2: API Response with Groups

**What:** Return recommendations organized into groups
**When to use:** In `GetAllRecommendations` endpoint

```go
// Source: internal/api/handlers/recommendations.go

type RecommendationGroup struct {
    PatternDescription   string                 `json:"pattern_description"`
    TotalDailySavings    float64                `json:"total_daily_savings"`
    InstanceCount        int                    `json:"instance_count"`
    Recommendations      []enrichedRec          `json:"recommendations"`
}

// In GetAllRecommendations, after enriching recommendations:
groups := groupByPatternSignature(enriched)

// Sort groups by total savings (highest first)
sort.Slice(groups, func(i, j int) bool {
    return groups[i].TotalDailySavings > groups[j].TotalDailySavings
})
```

### Pattern 3: Frontend Grouped Display

**What:** Render recommendation groups with expandable instance lists
**When to use:** In RecommendationsPage when displaying grouped view

```tsx
// Source: web/src/components/RecommendationGroup.tsx
interface RecommendationGroupProps {
    group: {
        pattern_description: string
        total_daily_savings: number
        instance_count: number
        recommendations: RecommendationEnriched[]
    }
    onOpenModal: (rec: RecommendationEnriched) => void
    onDismiss: (id: string) => void
}

function RecommendationGroup({ group, onOpenModal, onDismiss }: RecommendationGroupProps) {
    const [expanded, setExpanded] = useState(true) // Start expanded

    return (
        <div className="bg-slate-800/30 rounded-xl border border-slate-700 overflow-hidden">
            {/* Group header */}
            <div 
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-slate-700/30"
                onClick={() => setExpanded(!expanded)}
            >
                <div className="flex items-center gap-3">
                    <Users className="h-5 w-5 text-slate-400" />
                    <span className="font-medium text-white">{group.pattern_description}</span>
                    <span className="text-sm text-slate-400">
                        {group.instance_count} instance{group.instance_count !== 1 ? 's' : ''}
                    </span>
                </div>
                <div className="flex items-center gap-4">
                    <span className="text-green-400 font-semibold">
                        ${group.total_daily_savings.toFixed(2)}/day total
                    </span>
                    {expanded ? <ChevronUp /> : <ChevronDown />}
                </div>
            </div>

            {/* Instance list */}
            {expanded && (
                <div className="px-4 pb-4 space-y-2">
                    {group.recommendations.map(rec => (
                        <RecommendationCard 
                            key={rec.id}
                            recommendation={rec}
                            onOpenModal={onOpenModal}
                            onDismiss={onDismiss}
                            compact // Optional: more compact display within groups
                        />
                    ))}
                </div>
            )}
        </div>
    )
}
```

### Anti-Patterns to Avoid

- **Don't over-group:** Too-broad buckets (e.g., all evening patterns) reduce usefulness
- **Don't hide savings:** Per-instance savings must still be visible, not just group totals
- **Don't break existing flow:** Individual recommendation actions (view, dismiss, confirm) must still work
- **Don't force grouping:** Single-instance groups are valid - display them normally

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Pattern matching | Complex algorithm | Simple bucketing | Buckets are interpretable, algorithms are not |
| Savings calculation | New formula | Existing `calculateDailySavings` | Already implemented and tested |
| Group display | Complex nested lists | Simple expand/collapse | Matches existing UI patterns |
| API structure | New endpoint | Extend existing `/recommendations` | Backward compatible |

**Key insight:** The grouping is primarily a UI/presentation concern. The backend just needs to add a `group_key` field; frontend can handle the actual grouping and display logic.

## Common Pitfalls

### Pitfall 1: Over-Engineering Pattern Similarity

**What goes wrong:** Complex similarity metrics (Jaccard distance, fuzzy matching) that are hard to explain
**Why it happens:** Pattern "similarity" seems like a machine learning problem
**How to avoid:** Use simple, explainable buckets: "Evening to morning, weekdays" is understandable
**Warning signs:** Users can't predict which instances will group together

### Pitfall 2: Hiding Individual Savings

**What goes wrong:** Only showing group total, users can't see which instances save most
**Why it happens:** Focusing on "grouped" view loses per-instance detail
**How to avoid:** Keep per-recommendation savings display; group totals are additive
**Warning signs:** Users click into groups to find individual savings

### Pitfall 3: Breaking Single-Instance UX

**What goes wrong:** Groups of 1 instance look awkward or require extra clicks
**Why it happens:** Assuming all patterns will have multiple matches
**How to avoid:** Render single-instance groups as regular recommendation cards (no group wrapper)
**Warning signs:** Lots of "Group: 1 instance" headers cluttering the UI

### Pitfall 4: Inconsistent Sort Order

**What goes wrong:** Groups sorted by savings but instances within sorted differently
**Why it happens:** Different sorting logic in different places
**How to avoid:** Sort both levels by daily savings descending
**Warning signs:** High-savings instances buried in low-savings groups

### Pitfall 5: Pattern Description Mismatch

**What goes wrong:** Pattern description says "weekdays" but some instances show weekend days
**Why it happens:** Bucket assignment uses different logic than description generation
**How to avoid:** Generate description from the same signature used for grouping
**Warning signs:** User confusion about why instances are grouped together

## Code Examples

### Savings Calculation (Existing - Verify)

The existing calculation in `recommendations.go` is correct:

```go
// Source: internal/api/handlers/recommendations.go (lines 114-119)
// Calculate idle hours (handling overnight windows)
idleHours := idleEndHour - idleStartHour + 1
if idleEndHour <= idleStartHour {
    idleHours = (24 - idleStartHour) + idleEndHour + 1
}
dailySavings := float64(idleHours*instance.HourlyCostCents) / 100.0
```

This correctly:
- Handles overnight windows (e.g., 22:00 to 06:00 = 9 hours)
- Converts cents to dollars
- Multiplies idle hours by hourly cost

### Pattern Description Generation

```go
// Source: internal/api/handlers/recommendations.go (new function)
func describePattern(pattern map[string]interface{}) string {
    startHour := int(pattern["idle_start_hour"].(float64))
    endHour := int(pattern["idle_end_hour"].(float64))
    daysOfWeek := pattern["days_of_week"].([]interface{})
    
    // Format time range
    startTime := formatHour(startHour)
    endTime := formatHour(endHour)
    timeRange := fmt.Sprintf("%s to %s", startTime, endTime)
    
    // Format days
    dayType := daysToType(daysOfWeek)
    var dayDesc string
    switch dayType {
    case "weekdays":
        dayDesc = "weekdays"
    case "weekends":
        dayDesc = "weekends"
    case "daily":
        dayDesc = "daily"
    default:
        dayDesc = fmt.Sprintf("%d days/week", len(daysOfWeek))
    }
    
    return fmt.Sprintf("Idle %s, %s", timeRange, dayDesc)
}

func formatHour(hour int) string {
    if hour == 0 {
        return "midnight"
    }
    if hour == 12 {
        return "noon"
    }
    if hour < 12 {
        return fmt.Sprintf("%dam", hour)
    }
    return fmt.Sprintf("%dpm", hour-12)
}
```

### Grouped API Response Structure

```typescript
// Source: web/src/lib/api.ts (extend)
export interface RecommendationGroup {
    pattern_description: string
    pattern_key: string           // For stable sorting/keying
    total_daily_savings: number
    instance_count: number
    recommendations: RecommendationEnriched[]
}

export interface GroupedRecommendationsResponse {
    groups: RecommendationGroup[]
    ungrouped_count: number       // Recommendations not fitting any group
}

// Update getRecommendations to return grouped response
getRecommendations: (status?: string) => {
    const params = status ? `?status=${status}` : ''
    return api.get<GroupedRecommendationsResponse>(`/recommendations${params}`)
}
```

### Backend Grouping Logic

```go
// Source: internal/api/handlers/recommendations.go (new function)
func groupRecommendations(recs []enrichedRec) []RecommendationGroup {
    // Group by pattern signature
    groupMap := make(map[string]*RecommendationGroup)
    
    for _, rec := range recs {
        sig := generatePatternSignature(rec.DetectedPattern)
        key := fmt.Sprintf("%s-%s-%s", sig.StartBucket, sig.EndBucket, sig.DayType)
        
        if groupMap[key] == nil {
            groupMap[key] = &RecommendationGroup{
                PatternDescription: describePattern(rec.DetectedPattern),
                PatternKey:         key,
                Recommendations:    []enrichedRec{},
            }
        }
        
        groupMap[key].Recommendations = append(groupMap[key].Recommendations, rec)
        groupMap[key].TotalDailySavings += rec.EstimatedDailySavings
        groupMap[key].InstanceCount++
    }
    
    // Convert to slice
    var groups []RecommendationGroup
    for _, g := range groupMap {
        // Sort recommendations within group by savings
        sort.Slice(g.Recommendations, func(i, j int) bool {
            return g.Recommendations[i].EstimatedDailySavings > g.Recommendations[j].EstimatedDailySavings
        })
        groups = append(groups, *g)
    }
    
    // Sort groups by total savings
    sort.Slice(groups, func(i, j int) bool {
        return groups[i].TotalDailySavings > groups[j].TotalDailySavings
    })
    
    return groups
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Flat recommendation list | Grouped by pattern | Phase 14 | Better pattern visibility |
| Per-instance savings only | Group totals + individual | Phase 14 | Clearer aggregate impact |
| No pattern comparison | Pattern signatures | Phase 14 | Enables grouping logic |

**Existing and reusable:**
- IdleWindow detection (Phase 5, updated Phase 13)
- Savings calculation formula (Phase 6)
- RecommendationCard component (Phase 6)
- expand/collapse UI pattern (existing in FilterPreview)

## Open Questions

Things that couldn't be fully resolved:

1. **Bucket granularity**
   - What we know: 5 time buckets (early-morning, midday, afternoon, evening, night) seem reasonable
   - What's unclear: Whether finer buckets (2-hour instead of 4-hour) would be better
   - Recommendation: Start with 5 buckets; refine based on real data distribution

2. **Mixed day patterns**
   - What we know: Some patterns may have 4 weekdays + 1 weekend day
   - What's unclear: Should these group with "weekdays" or "mixed"?
   - Recommendation: Use 80% threshold (4 of 5 weekdays = weekdays, else mixed)

3. **Group collapse state persistence**
   - What we know: Groups start expanded
   - What's unclear: Should collapse state persist across page reloads?
   - Recommendation: Don't persist; always start expanded for visibility

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/analyzer/recommendation.go`, `internal/api/handlers/recommendations.go`
- Existing UI: `web/src/components/RecommendationCard.tsx`, `web/src/pages/RecommendationsPage.tsx`
- Phase 13 implementation: Compound idle threshold (CPU < 5% AND connections = 0)

### Secondary (MEDIUM confidence)
- Phase 6 RESEARCH.md: UI patterns, API structure
- Phase 6 CONTEXT.md: User expectations for recommendations

### Tertiary (LOW confidence)
- Bucket boundaries (chosen based on typical business hours; may need adjustment)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All dependencies already in codebase
- Architecture: HIGH - Extends existing patterns
- Pattern similarity: MEDIUM - Bucket approach is reasonable but not validated with real data
- UI patterns: HIGH - Follows established expand/collapse pattern

**Research date:** 2026-02-25
**Valid until:** 60 days (stable patterns, internal codebase)
