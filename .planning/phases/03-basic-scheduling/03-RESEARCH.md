# Phase 3: Basic Scheduling - Research

**Researched:** 2026-02-23
**Domain:** Schedule UI with weekly heatmap grid, CRON expression handling, React click-drag patterns
**Confidence:** MEDIUM-HIGH

## Summary

This phase requires building a visual weekly heatmap grid (7 days × 24 hours) for schedule time selection, with optional CRON fallback. The existing codebase already has schedule CRUD infrastructure in place (Schedule model, ScheduleStore, API handlers, frontend pages), but the current UI uses simple CRON text inputs rather than the visual grid. 

The key UI challenge is implementing click-drag "painting" behavior in React for the grid, which requires careful mouse event handling. For CRON functionality, well-established libraries exist: `cronstrue` for JavaScript (human-readable descriptions) and `github.com/robfig/cron/v3` for Go (parsing/scheduling) with `github.com/lnquy/cron` for Go human descriptions.

**Primary recommendation:** Build a custom `WeeklyScheduleGrid` component using native React mouse events (mousedown/mousemove/mouseup) for drag painting, store grid state as a simple 2D boolean array, convert to/from CRON expressions for storage.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `cronstrue` | 3.12.0 | JS: CRON → human description | 1.6k stars, same author as .NET version, 30+ locales |
| `github.com/robfig/cron/v3` | 3.0.1 | Go: CRON parsing & scheduling | 14.1k stars, de-facto Go standard, timezone support |
| `github.com/lnquy/cron` | 1.1.1 | Go: CRON → human description | Port of cronstrue to Go, 26 locales |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `clsx` | 2.0.0 | Already in project | Conditional CSS classes for grid cells |
| `@headlessui/react` | 2.2.9 | Already in project | Modal for schedule creation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom drag logic | `react-dnd` | Overkill for simple grid painting; adds 40kb+ |
| 2D boolean array | Bitmap/bitfield | More complex, marginal space savings (21 bytes vs 168) |
| `cronstrue` | Manual descriptions | Much more work, localization issues |

**Installation:**
```bash
# Frontend - already has most deps, just add cronstrue
cd web && npm install cronstrue

# Backend - add CRON libraries
go get github.com/robfig/cron/v3@v3.0.1
go get github.com/lnquy/cron@v1.1.1
```

## Architecture Patterns

### Recommended Project Structure
```
web/src/
├── components/
│   ├── WeeklyScheduleGrid.tsx    # 7×24 grid with drag painting
│   └── ScheduleModal.tsx         # Create/edit modal with grid
├── hooks/
│   └── useGridDrag.ts           # Drag painting logic (optional extraction)
└── lib/
    └── cronUtils.ts              # Grid ↔ CRON conversion utilities

internal/
├── scheduler/
│   └── cron_parser.go           # CRON validation and description
└── api/handlers/
    └── schedules.go             # Already exists, minimal changes
```

### Pattern 1: Click-Drag Grid Painting
**What:** Track mouse state (isDown, paintMode) and update cells on mousemove
**When to use:** Any grid-based selection where users "paint" areas
**Example:**
```typescript
// Source: Custom implementation pattern
const [grid, setGrid] = useState<boolean[][]>(() => 
  Array(7).fill(null).map(() => Array(24).fill(false))
);
const [isDragging, setIsDragging] = useState(false);
const [paintMode, setPaintMode] = useState<boolean | null>(null);

const handleCellMouseDown = (day: number, hour: number) => {
  setIsDragging(true);
  const newPaintMode = !grid[day][hour]; // Toggle opposite of current
  setPaintMode(newPaintMode);
  updateCell(day, hour, newPaintMode);
};

const handleCellMouseEnter = (day: number, hour: number) => {
  if (isDragging && paintMode !== null) {
    updateCell(day, hour, paintMode);
  }
};

const handleMouseUp = () => {
  setIsDragging(false);
  setPaintMode(null);
};

// Attach mouseup to document for reliable release detection
useEffect(() => {
  document.addEventListener('mouseup', handleMouseUp);
  return () => document.removeEventListener('mouseup', handleMouseUp);
}, []);
```

### Pattern 2: Grid ↔ CRON Conversion
**What:** Convert 7×24 boolean grid to/from CRON expressions
**When to use:** Storing/retrieving schedule data
**Example:**
```typescript
// Grid to CRON: Find contiguous sleep periods
function gridToCron(grid: boolean[][]): { sleepCron: string; wakeCron: string } {
  // Simplified: Find first sleep hour and first wake hour
  // Complex schedules may need multiple CRON expressions
  const activeDays = grid.map((dayGrid, dayIdx) => 
    dayGrid.some(hour => hour) ? dayIdx : null
  ).filter(d => d !== null);
  
  // Find sleep start (first true) and wake start (first false after sleep)
  // ... implementation details
  
  return { sleepCron: '0 22 * * 1-5', wakeCron: '0 7 * * 1-5' };
}

// CRON to Grid: Parse and mark hours
function cronToGrid(sleepCron: string, wakeCron: string): boolean[][] {
  // Parse CRON to extract hours and days
  // Mark sleep hours in grid
  // ... implementation details
}
```

### Pattern 3: Modal with Grid Component
**What:** Use Headless UI Dialog for create/edit modal containing the grid
**When to use:** Schedule creation/editing flow
**Example:**
```tsx
// Source: Headless UI Dialog pattern + existing project style
import { Dialog } from '@headlessui/react'

<Dialog open={isOpen} onClose={() => setIsOpen(false)}>
  <div className="fixed inset-0 bg-black/50" />
  <div className="fixed inset-0 flex items-center justify-center p-4">
    <Dialog.Panel className="bg-slate-800 rounded-xl p-6 max-w-4xl w-full">
      <Dialog.Title className="text-xl font-bold text-white">
        Create Schedule
      </Dialog.Title>
      
      <WeeklyScheduleGrid 
        grid={grid} 
        onChange={setGrid}
        showCronFallback={showCronFallback}
      />
      
      {/* Name input, submit button, etc. */}
    </Dialog.Panel>
  </div>
</Dialog>
```

### Anti-Patterns to Avoid
- **Don't use onClick for cells:** Click fires after mouseup, breaks drag painting
- **Don't track selection in CSS only:** Need state for form submission
- **Don't parse CRON on every render:** Memoize expensive conversions
- **Don't store grid in JSONB directly:** Convert to CRON for the scheduler to use

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CRON human descriptions | String concatenation | `cronstrue` / `lnquy/cron` | i18n, edge cases, special chars |
| CRON parsing | Regex matching | `robfig/cron/v3` | Timezone handling, validation, Next() calculation |
| CRON validation | Simple field checking | `robfig/cron/v3.ParseStandard()` | Standard compliance, error messages |
| Modal/Dialog | Custom overlay | `@headlessui/react` Dialog | Focus trapping, accessibility |

**Key insight:** CRON expressions have many edge cases (L, W, #, ranges, steps). Even simple-looking validation logic will miss cases that established libraries handle.

## Common Pitfalls

### Pitfall 1: Mouse Release Outside Grid
**What goes wrong:** User starts dragging, moves mouse outside grid, releases - grid stays in drag mode
**Why it happens:** mouseup only fires if released over the same element
**How to avoid:** Attach mouseup listener to `document`, not grid cells
**Warning signs:** Grid keeps painting when mouse re-enters after leaving

### Pitfall 2: CRON Timezone Confusion
**What goes wrong:** Schedule runs at wrong time; 10pm EST runs at 10pm UTC
**Why it happens:** CRON expressions don't inherently include timezone; must be tracked separately
**How to avoid:** Store timezone in Schedule model (already present), use `robfig/cron` with `cron.WithLocation()`
**Warning signs:** Tests pass locally but schedules fire at wrong times in production

### Pitfall 3: Complex Schedules Can't Be Represented
**What goes wrong:** User paints non-contiguous sleep periods (e.g., 1am-3am AND 2pm-4pm); can't store as single CRON
**Why it happens:** Standard CRON has limited expressiveness for multiple ranges
**How to avoid:** Option A: Store as JSON array in addition to CRON. Option B: Create multiple schedule records. Option C: Use extended CRON syntax
**Warning signs:** User creates schedule in grid, reopens edit modal, sees different pattern

### Pitfall 4: Grid State Lost on Mode Switch
**What goes wrong:** User paints grid, clicks "Switch to CRON", CRON input shows, can't get back to grid with same data
**Why it happens:** One-way conversion, no state synchronization
**How to avoid:** Always maintain grid state as source of truth; CRON is display/fallback only
**Warning signs:** Mode toggle loses user's work

## Code Examples

Verified patterns from official sources:

### cronstrue Usage (JavaScript)
```typescript
// Source: https://github.com/bradymholt/cRonstrue
import cronstrue from 'cronstrue';

cronstrue.toString("*/5 * * * *");
// "Every 5 minutes"

cronstrue.toString("0 23 ? * MON-FRI");
// "At 11:00 PM, Monday through Friday"

cronstrue.toString("0 22 * * *", { use24HourTimeFormat: true });
// "At 22:00"
```

### robfig/cron Usage (Go) - Parsing
```go
// Source: https://pkg.go.dev/github.com/robfig/cron/v3
import "github.com/robfig/cron/v3"

// Validate CRON expression
parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
schedule, err := parser.Parse("0 22 * * 1-5")
if err != nil {
    // Invalid CRON expression
}

// Get next occurrence
nextRun := schedule.Next(time.Now())
```

### lnquy/cron Usage (Go) - Human Description
```go
// Source: https://github.com/lnquy/cron
import "github.com/lnquy/cron"

exprDesc, _ := cron.NewDescriptor()
desc, _ := exprDesc.ToDescription("0 23 ? * MON-FRI", cron.Locale_en)
// "At 11:00 PM, Monday through Friday"
```

### WeeklyScheduleGrid Component Structure
```tsx
// Source: Custom implementation based on common patterns
interface WeeklyScheduleGridProps {
  grid: boolean[][];
  onChange: (grid: boolean[][]) => void;
  disabled?: boolean;
}

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

export function WeeklyScheduleGrid({ grid, onChange, disabled }: WeeklyScheduleGridProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [paintMode, setPaintMode] = useState<boolean | null>(null);
  
  const updateCell = (day: number, hour: number, value: boolean) => {
    const newGrid = grid.map((d, di) => 
      di === day ? d.map((h, hi) => hi === hour ? value : h) : d
    );
    onChange(newGrid);
  };
  
  return (
    <div className="grid grid-cols-[auto_repeat(24,1fr)] gap-px bg-slate-700">
      {/* Header row with hours */}
      <div className="bg-slate-800 p-1" />
      {HOURS.map(hour => (
        <div key={hour} className="bg-slate-800 text-xs text-center py-1 text-slate-400">
          {hour}
        </div>
      ))}
      
      {/* Day rows */}
      {DAYS.map((day, dayIdx) => (
        <>
          <div key={`label-${day}`} className="bg-slate-800 px-2 py-2 text-sm text-slate-300">
            {day}
          </div>
          {HOURS.map(hour => (
            <div
              key={`${day}-${hour}`}
              className={clsx(
                'h-6 cursor-pointer transition-colors',
                grid[dayIdx][hour] 
                  ? 'bg-indigo-600 hover:bg-indigo-500' 
                  : 'bg-slate-900 hover:bg-slate-800'
              )}
              onMouseDown={() => {
                if (disabled) return;
                setIsDragging(true);
                const newMode = !grid[dayIdx][hour];
                setPaintMode(newMode);
                updateCell(dayIdx, hour, newMode);
              }}
              onMouseEnter={() => {
                if (disabled || !isDragging || paintMode === null) return;
                updateCell(dayIdx, hour, paintMode);
              }}
            />
          ))}
        </>
      ))}
    </div>
  );
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CRON text input only | Visual grid + CRON fallback | Current project decision | Better UX for non-technical users |
| Single CRON field | Separate sleep/wake CRONs | Existing model | Explicit start/stop times |
| Inline form | Modal dialog | Project decision | Cleaner list page, focused editing |

**Deprecated/outdated:**
- `robfig/cron/v2`: Use v3 (has breaking changes but better API)
- `gopkg.in/robfig/cron.v2`: Use module path `github.com/robfig/cron/v3`

## Open Questions

Things that couldn't be fully resolved:

1. **Complex schedule representation**
   - What we know: Standard CRON can't represent all grid patterns (e.g., non-contiguous hours)
   - What's unclear: How to handle when user paints 1-3am AND 8-10pm on same day
   - Recommendation: For Phase 3, simplify to single contiguous sleep window per day. Flag complex patterns with validation warning. Full support deferred.

2. **Grid data persistence format**
   - What we know: Current model uses `sleep_cron` and `wake_cron` fields
   - What's unclear: Whether to add a `grid_data` JSONB field for exact grid state
   - Recommendation: Keep using CRON fields for scheduler compatibility. Accept that some grid patterns can't round-trip perfectly.

3. **Mobile touch support**
   - What we know: Mouse events work, but touch requires different handling
   - What's unclear: How important is mobile editing for this app
   - Recommendation: Focus on desktop for Phase 3. Add touch support in future if needed.

## Sources

### Primary (HIGH confidence)
- `github.com/robfig/cron/v3` - Official Go documentation, v3.0.1 API reference
- `cronstrue` GitHub README - JavaScript library v3.12.0 usage examples
- `github.com/lnquy/cron` GitHub README - Go CRON description library v1.1.1

### Secondary (MEDIUM confidence)
- Existing codebase analysis - Schedule model, handlers, frontend pages
- React mouse event patterns - Common implementation approach

### Tertiary (LOW confidence)
- Grid painting UX patterns - No authoritative source, based on common implementations

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Well-established libraries with clear usage
- Architecture: MEDIUM - Custom grid component, proven patterns but no canonical React implementation
- Pitfalls: HIGH - Based on codebase analysis and common issues

**Research date:** 2026-02-23
**Valid until:** 60 days (stable libraries, no fast-moving APIs)
