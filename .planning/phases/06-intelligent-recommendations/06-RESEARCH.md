# Phase 6: Intelligent Recommendations - Research

**Researched:** 2026-02-23
**Domain:** Recommendation UI/UX, Pattern-to-Schedule Conversion, Persistence
**Confidence:** HIGH

## Summary

This phase transforms the activity analysis infrastructure from Phase 5 into user-facing schedule recommendations. The core challenge is bridging the gap between detected `IdleWindow` patterns and actionable schedule creation, while providing a UI that clearly communicates confidence, savings potential, and activity patterns.

The existing codebase provides strong foundations: the `Analyzer` service already detects idle windows with confidence scoring, the `Recommendation` model exists in the database schema, and the frontend has established patterns for modals, lists with expand/collapse, and toast notifications.

**Primary recommendation:** Extend the existing `Analyzer.AnalyzeAllInstances()` to generate `Recommendation` records, implement a thin API layer connecting to `RecommendationStore`, and build a list-based UI following the established `FilterPreview` expand/collapse pattern.

## Standard Stack

The phase uses the existing codebase's established patterns - no new dependencies needed.

### Core (Already in Codebase)
| Component | Version | Purpose | Why Standard |
|-----------|---------|---------|--------------|
| Go Chi router | v5 | API endpoints | Already used for all routes |
| PostgreSQL + pgx | v5 | Recommendation persistence | Existing store patterns |
| React | 18.2 | UI components | Already in use |
| Headless UI | 2.2.9 | Modal dialogs | Used in ScheduleModal |
| lucide-react | 0.300 | Icons (ChevronDown, etc.) | Used throughout |
| react-hot-toast | 2.6.0 | Notifications | Used for instance operations |
| Tailwind CSS | 3.4 | Styling | Used throughout |

### Supporting (Already in Codebase)
| Component | Version | Purpose | When to Use |
|-----------|---------|---------|-------------|
| clsx | 2.0.0 | Conditional CSS | Complex class logic |
| recharts | 2.10.0 | Activity graphs | 24-hour CPU visualization |
| cronstrue | 3.12.0 | CRON description | Display suggested schedules |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| recharts for graphs | Custom SVG | More control but more code; recharts already works |
| localStorage for dismissed | Database | Persistence across devices but added complexity |
| Separate recommendations page | Dashboard integration only | Per CONTEXT.md: both dashboard AND dedicated tab |

## Architecture Patterns

### Recommended Project Structure

Extend existing structure with minimal new files:

```
internal/
├── analyzer/
│   ├── analyzer.go          # Extend with GenerateRecommendations()
│   ├── patterns.go          # Already has IdleWindow, ActivityPattern
│   └── recommendation.go    # NEW: Recommendation generation logic
├── api/handlers/
│   └── recommendations.go   # Implement existing stub
├── store/
│   └── postgres.go          # RecommendationStore already exists

web/src/
├── pages/
│   ├── Dashboard.tsx        # Add recommendations list section
│   └── RecommendationsPage.tsx  # Refactor existing stub
├── components/
│   ├── RecommendationCard.tsx   # NEW: List item with expand/collapse
│   ├── RecommendationModal.tsx  # NEW: Details modal with confirm
│   └── ActivityGraph.tsx        # NEW: 24-hour CPU visualization
├── lib/
│   └── api.ts               # Extend Recommendation type
```

### Pattern 1: Recommendation Generation Flow

**What:** Convert `ActivityPattern` to `Recommendation` with schedule suggestion
**When to use:** Manual trigger via dashboard "Generate Recommendations" button

```go
// Source: internal/analyzer/recommendation.go (to be created)
func (a *Analyzer) GenerateRecommendations(ctx context.Context) ([]models.Recommendation, error) {
    patterns, err := a.AnalyzeAllInstances(ctx)
    if err != nil {
        return nil, err
    }

    var recommendations []models.Recommendation
    for instanceID, pattern := range patterns {
        if len(pattern.IdleWindows) == 0 {
            continue
        }

        // Get best idle window (already sorted by confidence)
        window := pattern.IdleWindows[0]
        
        // Skip if already has active recommendation for this instance
        existing, _ := a.store.ListRecommendationsByInstance(instanceID, "pending")
        if len(existing) > 0 {
            continue
        }

        // Generate schedule from pattern
        rec := a.createRecommendationFromPattern(instanceID, window, pattern)
        recommendations = append(recommendations, rec)
    }

    return recommendations, nil
}
```

### Pattern 2: Dismissed Recommendations Persistence

**What:** Store dismissed recommendation IDs to survive sessions
**When to use:** User clicks "Mark as Not Now"

Per CONTEXT.md: "Dismissed recommendations remembered across sessions". Use database status update:

```go
// Update recommendation status to 'dismissed'
// This is already supported by RecommendationStore.UpdateRecommendation
func (h *RecommendationHandler) DismissRecommendation(w http.ResponseWriter, r *http.Request, id string) {
    rec, err := h.store.GetRecommendation(id)
    if err != nil {
        // handle error
        return
    }
    
    rec.Status = "dismissed"
    if err := h.store.UpdateRecommendation(rec); err != nil {
        // handle error
        return
    }
    
    // Return success
}
```

### Pattern 3: Confirm Flow - Schedule Creation

**What:** When user confirms recommendation, create actual schedule
**When to use:** User clicks "Create Schedule" in confirmation modal

```go
// Source: internal/api/handlers/recommendations.go
func (h *RecommendationHandler) ApproveRecommendation(w http.ResponseWriter, r *http.Request, id string) {
    rec, err := h.store.GetRecommendation(id)
    if err != nil {
        // handle error
        return
    }

    // Parse suggested schedule from JSONB
    var suggestedSchedule struct {
        Timezone  string `json:"timezone"`
        SleepCron string `json:"sleep_cron"`
        WakeCron  string `json:"wake_cron"`
    }
    json.Unmarshal(rec.SuggestedSchedule, &suggestedSchedule)

    // Get instance to create proper selector
    instance, _ := h.instanceStore.GetInstanceByProviderID(ctx, "", rec.InstanceID)

    // Create schedule matching only this instance
    schedule := &models.Schedule{
        Name:        fmt.Sprintf("AI Suggested: %s", instance.Name),
        Description: "Auto-generated from activity pattern analysis",
        Timezone:    suggestedSchedule.Timezone,
        SleepCron:   suggestedSchedule.SleepCron,
        WakeCron:    suggestedSchedule.WakeCron,
        Selectors: []models.Selector{{
            Name: &models.Matcher{
                Pattern: instance.Name,
                Type:    "exact",
            },
        }},
        Enabled: true,
    }

    if err := h.scheduleStore.CreateSchedule(schedule); err != nil {
        // handle error
        return
    }

    // Update recommendation status
    rec.Status = "approved"
    h.store.UpdateRecommendation(rec)

    // Return created schedule ID
}
```

### Pattern 4: List View with Collapsible Cards

**What:** Recommendation list sorted by savings, expandable for details
**When to use:** Dashboard recommendations section and dedicated page

```tsx
// Source: web/src/components/RecommendationCard.tsx (to be created)
// Following FilterPreview pattern for expand/collapse

interface RecommendationCardProps {
  recommendation: Recommendation;
  onConfirm: (id: string) => void;
  onDismiss: (id: string) => void;
}

function RecommendationCard({ recommendation, onConfirm, onDismiss }: RecommendationCardProps) {
  const [expanded, setExpanded] = useState(false);

  const confidenceLabel = 
    recommendation.confidence >= 80 ? 'High' :
    recommendation.confidence >= 50 ? 'Medium' : 'Low';

  const confidenceColor =
    recommendation.confidence >= 80 ? 'text-green-400 bg-green-500/10 border-green-500/30' :
    recommendation.confidence >= 50 ? 'text-yellow-400 bg-yellow-500/10 border-yellow-500/30' :
    'text-orange-400 bg-orange-500/10 border-orange-500/30';

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 overflow-hidden">
      {/* Summary row - always visible */}
      <div 
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-slate-700/30"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          <span className="px-2 py-1 text-xs bg-purple-500/10 text-purple-400 border border-purple-500/30 rounded">
            AI Suggested
          </span>
          <span className="font-medium text-white">{recommendation.instance_name}</span>
          <span className={`px-2 py-1 text-xs rounded border ${confidenceColor}`}>
            {confidenceLabel} confidence
          </span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-green-400 font-semibold">
            ${recommendation.estimated_daily_savings}/day
          </span>
          {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </div>
      </div>

      {/* Expanded details */}
      {expanded && (
        <div className="px-4 pb-4 border-t border-slate-700">
          {/* Activity pattern visualization */}
          {/* Suggested times */}
          {/* Action buttons */}
        </div>
      )}
    </div>
  );
}
```

### Anti-Patterns to Avoid

- **Don't auto-apply recommendations:** User must explicitly confirm via modal
- **Don't show raw confidence percentages:** Use High/Medium/Low labels per CONTEXT.md
- **Don't auto-refresh recommendations:** Manual trigger only per CONTEXT.md
- **Don't duplicate schedules:** Check if recommendation already approved before showing

## Don't Hand-Roll

Problems that have existing solutions in the codebase:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Modal dialogs | Custom overlay logic | Headless UI Dialog | Already used in ScheduleModal, ConfirmDialog |
| Expand/collapse | Custom state management | useState + conditional render | Pattern in FilterPreview |
| Toast notifications | Alert components | react-hot-toast | Already integrated |
| Schedule CRON generation | Custom CRON builder | Existing cronUtils.ts | Already has gridToCron |
| CRON description | Custom parser | cronstrue library | Already used in ScheduleModal |
| API client structure | Raw fetch | Existing api.ts pattern | Consistent error handling |

**Key insight:** The codebase already has mature patterns for every UI element needed. Reuse ConfirmDialog for confirmation flow, ScheduleModal patterns for the details modal, and FilterPreview patterns for expandable list items.

## Common Pitfalls

### Pitfall 1: Instance ID vs ProviderID Confusion

**What goes wrong:** Using wrong ID field when creating selectors for schedules
**Why it happens:** `Recommendation.InstanceID` stores different values than `Instance.ID`
**How to avoid:** Always fetch full Instance record before creating schedule; use `instance.Name` for selector matching (since instance IDs can change)
**Warning signs:** Schedule created but doesn't match any instances

### Pitfall 2: JSONB Field Handling in Go

**What goes wrong:** DetectedPattern and SuggestedSchedule are `[]byte` but need struct access
**Why it happens:** PostgreSQL JSONB stored as raw bytes in Go models
**How to avoid:** Always marshal/unmarshal when reading/writing these fields
**Warning signs:** JSON parse errors, nil pointer dereferences

```go
// Correct pattern:
var pattern struct {
    IdleStart int `json:"idle_start"`
    IdleEnd   int `json:"idle_end"`
}
json.Unmarshal(rec.DetectedPattern, &pattern)
```

### Pitfall 3: Timezone Handling for Schedule Times

**What goes wrong:** Suggested sleep/wake times are in wrong timezone
**Why it happens:** IdleWindow hours are in UTC but user expects local time
**How to avoid:** Store suggested timezone with recommendation; display times converted to that timezone
**Warning signs:** Schedule triggers at unexpected hours

### Pitfall 4: Stale Recommendations

**What goes wrong:** Showing recommendations for instances that no longer exist or have changed
**Why it happens:** Recommendations generated once but instances updated later
**How to avoid:** Join recommendations with instances table; filter out recommendations where instance is deleted or already has schedule
**Warning signs:** Click on recommendation shows "Instance not found"

### Pitfall 5: Empty State Edge Cases

**What goes wrong:** No clear user guidance when recommendations unavailable
**Why it happens:** Multiple reasons for empty: no data, no idle patterns, or all dismissed
**How to avoid:** Per CONTEXT.md: show reason for empty state (insufficient data vs. no patterns vs. all dismissed)
**Warning signs:** User confusion about why no recommendations shown

## Code Examples

### API Endpoint Structure

Following existing handler patterns:

```go
// Source: internal/api/handlers/recommendations.go
type RecommendationHandler struct {
    store         *store.RecommendationStore
    instanceStore *store.InstanceStore
    scheduleStore *store.ScheduleStore
    analyzer      *analyzer.Analyzer
}

func NewRecommendationHandler(
    recStore *store.RecommendationStore,
    instStore *store.InstanceStore,
    schedStore *store.ScheduleStore,
    analyzer *analyzer.Analyzer,
) *RecommendationHandler {
    return &RecommendationHandler{
        store:         recStore,
        instanceStore: instStore,
        scheduleStore: schedStore,
        analyzer:      analyzer,
    }
}

// GET /api/v1/recommendations
func (h *RecommendationHandler) GetAllRecommendations(w http.ResponseWriter, r *http.Request) {
    // Filter by status if provided
    status := r.URL.Query().Get("status")
    if status == "" {
        status = "pending" // Default to showing pending only
    }

    recs, err := h.store.ListRecommendations(status)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list recommendations"})
        return
    }

    // Enrich with instance details
    enriched := make([]map[string]interface{}, 0, len(recs))
    for _, rec := range recs {
        inst, _ := h.instanceStore.GetInstanceByProviderID(r.Context(), "", rec.InstanceID)
        
        // Parse stored pattern and schedule
        var pattern map[string]interface{}
        var schedule map[string]interface{}
        json.Unmarshal(rec.DetectedPattern, &pattern)
        json.Unmarshal(rec.SuggestedSchedule, &schedule)

        enriched = append(enriched, map[string]interface{}{
            "id":                   rec.ID,
            "instance_id":          rec.InstanceID,
            "instance_name":        inst.Name,
            "detected_pattern":     pattern,
            "suggested_schedule":   schedule,
            "confidence_score":     rec.ConfidenceScore,
            "status":               rec.Status,
            "created_at":           rec.CreatedAt,
            // Calculate savings estimate
            "estimated_daily_savings": calculateDailySavings(inst, pattern),
        })
    }

    // Sort by savings potential (highest first) per CONTEXT.md
    sort.Slice(enriched, func(i, j int) bool {
        return enriched[i]["estimated_daily_savings"].(float64) > enriched[j]["estimated_daily_savings"].(float64)
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(enriched)
}

// POST /api/v1/recommendations/generate
func (h *RecommendationHandler) GenerateRecommendations(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    recs, err := h.analyzer.GenerateRecommendations(ctx)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    // Store new recommendations
    created := 0
    for _, rec := range recs {
        if err := h.store.CreateRecommendation(&rec); err == nil {
            created++
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "created": created,
        "message": fmt.Sprintf("Generated %d new recommendations", created),
    })
}
```

### Frontend API Types

Extend existing api.ts:

```typescript
// Source: web/src/lib/api.ts (extensions)
export interface RecommendationEnriched {
  id: string;
  instance_id: string;
  instance_name: string;
  detected_pattern: {
    idle_start_hour: number;
    idle_end_hour: number;
    days_of_week: string[];
    avg_cpu: number;
    confidence: number;
  };
  suggested_schedule: {
    timezone: string;
    sleep_cron: string;
    wake_cron: string;
  };
  confidence_score: number;
  estimated_daily_savings: number;
  status: 'pending' | 'approved' | 'dismissed';
  created_at: string;
}

// Add to api object:
generateRecommendations: () => api.post<{ created: number; message: string }>('/recommendations/generate'),
dismissRecommendation: (id: string) => api.post<void>(`/recommendations/${id}/dismiss`),
confirmRecommendation: (id: string) => api.post<{ schedule_id: string }>(`/recommendations/${id}/confirm`),
```

### Activity Graph Component

Using recharts (already in package.json):

```tsx
// Source: web/src/components/ActivityGraph.tsx (to be created)
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';

interface ActivityGraphProps {
  pattern: {
    idle_start_hour: number;
    idle_end_hour: number;
    hourly_cpu?: number[]; // Optional: actual CPU values if available
  };
}

export function ActivityGraph({ pattern }: ActivityGraphProps) {
  // Generate 24-hour data
  const data = Array.from({ length: 24 }, (_, hour) => {
    const isIdle = pattern.idle_end_hour > pattern.idle_start_hour
      ? hour >= pattern.idle_start_hour && hour < pattern.idle_end_hour
      : hour >= pattern.idle_start_hour || hour < pattern.idle_end_hour;

    return {
      hour: `${hour}:00`,
      cpu: pattern.hourly_cpu?.[hour] ?? (isIdle ? 0.5 : 15), // Estimate if no data
      isIdle,
    };
  });

  return (
    <div className="h-32 w-full">
      <ResponsiveContainer>
        <AreaChart data={data}>
          <XAxis 
            dataKey="hour" 
            tick={{ fill: '#94a3b8', fontSize: 10 }}
            tickLine={false}
          />
          <YAxis 
            tick={{ fill: '#94a3b8', fontSize: 10 }}
            tickLine={false}
            domain={[0, 20]}
            tickFormatter={(v) => `${v}%`}
          />
          <Tooltip
            contentStyle={{ background: '#1e293b', border: 'none' }}
            labelStyle={{ color: '#f8fafc' }}
          />
          <Area
            type="monotone"
            dataKey="cpu"
            stroke="#3b82f6"
            fill="url(#cpuGradient)"
          />
          {/* Highlight idle window */}
          <ReferenceLine
            x={`${pattern.idle_start_hour}:00`}
            stroke="#10b981"
            strokeDasharray="3 3"
            label={{ value: 'Sleep', fill: '#10b981', fontSize: 10 }}
          />
          <ReferenceLine
            x={`${pattern.idle_end_hour}:00`}
            stroke="#f59e0b"
            strokeDasharray="3 3"
            label={{ value: 'Wake', fill: '#f59e0b', fontSize: 10 }}
          />
          <defs>
            <linearGradient id="cpuGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.3} />
              <stop offset="100%" stopColor="#3b82f6" stopOpacity={0} />
            </linearGradient>
          </defs>
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
```

### Confirmation Modal

Following ConfirmDialog and ScheduleModal patterns:

```tsx
// Source: web/src/components/RecommendationModal.tsx (to be created)
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react';
import { X, Clock, TrendingDown, Calendar } from 'lucide-react';
import { describeCron } from '../lib/cronUtils';
import { ActivityGraph } from './ActivityGraph';

interface RecommendationModalProps {
  isOpen: boolean;
  onClose: () => void;
  recommendation: RecommendationEnriched | null;
  onConfirm: (id: string) => Promise<void>;
  loading?: boolean;
}

export function RecommendationModal({
  isOpen,
  onClose,
  recommendation,
  onConfirm,
  loading,
}: RecommendationModalProps) {
  if (!recommendation) return null;

  const { detected_pattern, suggested_schedule, confidence_score, estimated_daily_savings } = recommendation;

  const confidenceLabel = 
    confidence_score >= 80 ? 'High' :
    confidence_score >= 50 ? 'Medium' : 'Low';

  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/50 backdrop-blur-sm" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="max-w-lg w-full bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl">
          <div className="flex items-center justify-between mb-6">
            <DialogTitle className="text-xl font-bold text-white">
              Confirm Schedule Recommendation
            </DialogTitle>
            <button onClick={onClose} className="text-slate-400 hover:text-white">
              <X className="w-6 h-6" />
            </button>
          </div>

          {/* Instance info */}
          <div className="mb-4">
            <span className="px-2 py-1 text-xs bg-purple-500/10 text-purple-400 border border-purple-500/30 rounded mr-2">
              AI Suggested
            </span>
            <span className="text-lg font-semibold text-white">{recommendation.instance_name}</span>
          </div>

          {/* Activity pattern graph */}
          <div className="mb-6 p-4 bg-slate-900/50 rounded-lg border border-slate-700">
            <h3 className="text-sm font-medium text-slate-300 mb-2">Detected Activity Pattern</h3>
            <ActivityGraph pattern={detected_pattern} />
          </div>

          {/* Suggested schedule */}
          <div className="mb-6 space-y-3">
            <h3 className="text-sm font-medium text-slate-300">Suggested Schedule</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
                <div className="flex items-center gap-2 text-slate-400 mb-1">
                  <Clock className="h-4 w-4" />
                  <span className="text-xs uppercase">Sleep at</span>
                </div>
                <p className="text-white font-mono">{describeCron(suggested_schedule.sleep_cron)}</p>
              </div>
              <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700">
                <div className="flex items-center gap-2 text-slate-400 mb-1">
                  <Clock className="h-4 w-4" />
                  <span className="text-xs uppercase">Wake at</span>
                </div>
                <p className="text-white font-mono">{describeCron(suggested_schedule.wake_cron)}</p>
              </div>
            </div>
          </div>

          {/* Summary stats */}
          <div className="mb-6 flex items-center justify-between p-4 bg-green-500/10 rounded-lg border border-green-500/30">
            <div className="flex items-center gap-2">
              <TrendingDown className="h-5 w-5 text-green-400" />
              <span className="text-green-400">Estimated savings</span>
            </div>
            <span className="text-2xl font-bold text-green-400">
              ${estimated_daily_savings.toFixed(2)}/day
            </span>
          </div>

          {/* Confidence */}
          <div className="mb-6 text-sm text-slate-400">
            <span className="font-medium">{confidenceLabel} confidence</span> - based on {detected_pattern.days_of_week?.length || 0} days of consistent activity patterns
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3">
            <button
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              onClick={() => onConfirm(recommendation.id)}
              disabled={loading}
              className="px-4 py-2 bg-green-600 hover:bg-green-500 text-white rounded-lg shadow-lg shadow-green-500/20 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create Schedule'}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Auto-apply schedules | User confirmation required | Phase 6 design | Users maintain control |
| Percentage confidence | Label-based (High/Med/Low) | CONTEXT.md decision | Clearer interpretation |
| Real-time updates | Manual refresh | CONTEXT.md decision | Predictable UX, less server load |
| Card grid layout | List view with collapse | CONTEXT.md decision | Better information density |

**Not applicable (already implemented):**
- IdleWindow detection logic (Phase 5)
- Confidence scoring algorithm (Phase 5)
- CloudWatch metrics collection (Phase 5)

## Open Questions

Things that couldn't be fully resolved:

1. **Hourly CPU data storage**
   - What we know: MetricsStore has hourly aggregates, HourlyMetric model exists
   - What's unclear: Whether to fetch historical data for graph or use summary only
   - Recommendation: Start with summary-only graph using idle window hours; add detailed graphs in future phase if needed

2. **Savings calculation formula**
   - What we know: Instance.HourlyCostCents is available
   - What's unclear: Exact formula for daily savings (assume full idle window stopped?)
   - Recommendation: Calculate as `(idle_hours * hourly_cost_cents / 100)` for daily estimate

3. **Recommendation cleanup/expiry**
   - What we know: Recommendations stay in database indefinitely
   - What's unclear: When should old/stale recommendations be cleaned up?
   - Recommendation: Add cleanup logic in future phase; for now, dismissed stays dismissed

## Sources

### Primary (HIGH confidence)
- Existing codebase patterns: postgres.go, handlers/, models/
- Existing UI patterns: ScheduleModal, ConfirmDialog, FilterPreview
- Phase 5 patterns.go: IdleWindow, ActivityPattern types

### Secondary (MEDIUM confidence)
- CONTEXT.md Phase 6 decisions: Locked decisions guiding implementation
- Headless UI Dialog: https://headlessui.com/react/dialog
- recharts documentation: https://recharts.org/en-US/api

### Tertiary (LOW confidence)
- Cost calculation estimates (needs validation with real data)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All dependencies already in codebase
- Architecture: HIGH - Follows established patterns
- UI patterns: HIGH - Reuses existing components
- Pitfalls: MEDIUM - Based on codebase analysis, not production experience

**Research date:** 2026-02-23
**Valid until:** 90 days (stable patterns, internal codebase)
