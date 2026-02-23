# Phase 4: Advanced Schedule Filtering - Research

**Researched:** 2026-02-23
**Domain:** Regex-based filtering, filter builder UI patterns, instance matching
**Confidence:** HIGH

## Summary

This phase extends the existing schedule system with filter-based instance assignment. The backend already supports the `Selector` model with `Matcher` types (exact, contains, prefix, suffix, regex), so the main work is building the frontend filter builder UI and the live preview functionality.

The key UI challenge is creating an intuitive filter builder that non-technical users can understand while supporting power-user regex patterns. For the preview, we can leverage the existing `/api/v1/instances` endpoint and filter client-side for instant feedback, with optional server-side validation for complex regex.

**Primary recommendation:** Build a `FilterBuilder` component that renders selector rules as visual chips, with AND/OR toggle between rules. Preview fetches all instances and applies JavaScript regex matching for instant client-side filtering. Backend handles actual schedule execution with Go regexp.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go `regexp` | stdlib | Server-side regex matching | Built-in, RE2 syntax (safe, no catastrophic backtracking) |
| JavaScript `RegExp` | native | Client-side preview filtering | Built-in, mostly compatible with Go RE2 |
| `@headlessui/react` | 2.2.9 | Already in project | Filter builder dropdowns, transitions |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `clsx` | 2.0.0 | Already in project | Conditional CSS classes for filter chips |
| `lucide-react` | existing | Already in project | Icons for add/remove/filter actions |

### No New Dependencies Needed

All functionality can be built with existing project dependencies:
- Headless UI for dropdowns and transitions
- clsx for conditional styling
- lucide-react for icons
- Existing API client for data fetching

## Architecture Patterns

### Recommended Project Structure
```
web/src/
├── components/
│   ├── FilterBuilder.tsx       # Main filter builder component
│   ├── FilterRule.tsx          # Single filter rule component
│   ├── FilterPreview.tsx       # Instance preview panel
│   └── ScheduleModal.tsx       # Extended with filter section
└── lib/
    └── filterUtils.ts          # Filter matching utilities

internal/
├── api/handlers/
│   └── schedules.go            # Extended with preview endpoint
└── scheduler/
    └── matcher.go              # Instance matching logic (new)
```

### Pattern 1: Filter Builder as Controlled Component
**What:** FilterBuilder receives selectors array and onChange callback
**When to use:** When parent component needs to manage filter state
**Example:**
```typescript
// Source: Custom implementation based on common React patterns
interface FilterBuilderProps {
  selectors: Selector[];
  onChange: (selectors: Selector[]) => void;
  instances?: Instance[];  // Optional for preview
}

export function FilterBuilder({ selectors, onChange, instances }: FilterBuilderProps) {
  const [combineOperator, setCombineOperator] = useState<'and' | 'or'>('and');
  
  const addRule = () => {
    onChange([...selectors, { name: { pattern: '', type: 'contains' } }]);
  };
  
  const updateRule = (index: number, selector: Selector) => {
    const updated = [...selectors];
    updated[index] = selector;
    onChange(updated);
  };
  
  const removeRule = (index: number) => {
    onChange(selectors.filter((_, i) => i !== index));
  };
  
  // ...render logic
}
```

### Pattern 2: Client-Side Instance Matching
**What:** Filter instances in JavaScript for instant preview
**When to use:** When dataset is small enough (< 1000 instances) for client-side filtering
**Example:**
```typescript
// Source: Custom implementation for SnoozeQL
function matchInstance(instance: Instance, selectors: Selector[], operator: 'and' | 'or'): boolean {
  if (selectors.length === 0) return true;
  
  const matches = selectors.map(selector => matchSelector(instance, selector));
  
  return operator === 'and' 
    ? matches.every(m => m) 
    : matches.some(m => m);
}

function matchSelector(instance: Instance, selector: Selector): boolean {
  // Match name
  if (selector.name) {
    if (!matchField(instance.name, selector.name)) return false;
  }
  
  // Match provider
  if (selector.provider) {
    const instanceProvider = instance.provider.startsWith('aws') ? 'aws' : 'gcp';
    if (instanceProvider !== selector.provider) return false;
  }
  
  // Match tags
  if (selector.tags) {
    for (const [key, matcher] of Object.entries(selector.tags)) {
      const tagValue = instance.tags?.[key];
      if (!tagValue || !matchField(tagValue, matcher)) return false;
    }
  }
  
  return true;
}

function matchField(value: string, matcher: Matcher): boolean {
  switch (matcher.type) {
    case 'exact': return value === matcher.pattern;
    case 'contains': return value.includes(matcher.pattern);
    case 'prefix': return value.startsWith(matcher.pattern);
    case 'suffix': return value.endsWith(matcher.pattern);
    case 'regex': {
      try {
        return new RegExp(matcher.pattern).test(value);
      } catch {
        return false;
      }
    }
    default: return false;
  }
}
```

### Pattern 3: Filter Rule as Visual Chip
**What:** Each filter rule rendered as an editable pill/chip
**When to use:** For visual filter builders
**Example:**
```tsx
// Source: Common UI pattern for filter/tag builders
interface FilterRuleProps {
  selector: Selector;
  onChange: (selector: Selector) => void;
  onRemove: () => void;
}

export function FilterRule({ selector, onChange, onRemove }: FilterRuleProps) {
  const [fieldType, setFieldType] = useState<'name' | 'provider' | 'tag'>('name');
  
  return (
    <div className="flex items-center gap-2 bg-slate-700 rounded-lg p-2">
      {/* Field type dropdown */}
      <select 
        value={fieldType}
        onChange={(e) => setFieldType(e.target.value as any)}
        className="bg-slate-800 text-sm rounded px-2 py-1"
      >
        <option value="name">Name</option>
        <option value="provider">Provider</option>
        <option value="tag">Tag</option>
      </select>
      
      {/* Match type dropdown (for name/tag) */}
      {fieldType !== 'provider' && (
        <select 
          value={selector.name?.type || 'contains'}
          onChange={(e) => onChange({
            ...selector,
            name: { ...selector.name!, type: e.target.value as any }
          })}
          className="bg-slate-800 text-sm rounded px-2 py-1"
        >
          <option value="contains">contains</option>
          <option value="exact">equals</option>
          <option value="prefix">starts with</option>
          <option value="suffix">ends with</option>
          <option value="regex">matches regex</option>
        </select>
      )}
      
      {/* Value input */}
      <input
        type="text"
        value={selector.name?.pattern || ''}
        onChange={(e) => onChange({
          ...selector,
          name: { ...selector.name!, pattern: e.target.value }
        })}
        className="bg-slate-900 text-sm rounded px-2 py-1 flex-1"
        placeholder={fieldType === 'provider' ? undefined : 'pattern...'}
      />
      
      {/* Remove button */}
      <button onClick={onRemove} className="text-slate-400 hover:text-red-400">
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}
```

### Anti-Patterns to Avoid
- **Don't validate regex on every keystroke:** Debounce or validate on blur to avoid performance issues
- **Don't fetch instances on every filter change:** Cache instances and filter client-side
- **Don't require all filter fields:** Each selector should have optional fields (name OR provider OR tags)
- **Don't mix client/server regex dialects:** Test regex patterns match between JS and Go RE2

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dropdown menus | Custom dropdown | `@headlessui/react` Listbox | Focus management, keyboard navigation |
| Regex validation | Manual parsing | Try-catch around `new RegExp()` | Comprehensive syntax checking |
| Debounced input | setTimeout chains | Custom hook or existing | Memory leak prevention |

## Common Pitfalls

### Pitfall 1: JavaScript vs Go Regex Mismatch
**What goes wrong:** Pattern works in preview but fails on server, or vice versa
**Why it happens:** JavaScript uses PCRE-style regex, Go uses RE2 (no backreferences, lookaheads)
**How to avoid:** Test regex patterns on both sides; document RE2 limitations in UI help text
**Warning signs:** Preview shows matches but scheduled actions don't target expected instances

### Pitfall 2: Empty Selectors Array Means "All Instances"
**What goes wrong:** Schedule with no filters runs against ALL instances unexpectedly
**Why it happens:** Common logic treats empty selectors as "no filter = match all"
**How to avoid:** Require at least one selector, or show explicit "All instances" warning
**Warning signs:** User creates schedule, doesn't add filters, wakes up to find all DBs sleeping

### Pitfall 3: Case Sensitivity Confusion
**What goes wrong:** "Production" doesn't match "production"
**Why it happens:** Regex is case-sensitive by default
**How to avoid:** Add case-insensitive option (Go: `(?i)` prefix, JS: `i` flag), or always lowercase compare
**Warning signs:** Users report filters don't work; they're using wrong case

### Pitfall 4: Tag Key vs Tag Value Matching
**What goes wrong:** User wants to match instances with tag "env=prod" but filter only checks value
**Why it happens:** UI doesn't clearly separate tag key from tag value matching
**How to avoid:** Explicit key and value fields in tag filter, with clear labels
**Warning signs:** Users confused about how to filter by specific tags

## Code Examples

Verified patterns from official sources:

### Go RE2 Regex Matching
```go
// Source: Go standard library regexp package
import "regexp"

func matchRegex(pattern, value string) bool {
    re, err := regexp.Compile(pattern)
    if err != nil {
        return false
    }
    return re.MatchString(value)
}

// Case-insensitive matching
func matchRegexInsensitive(pattern, value string) bool {
    re, err := regexp.Compile("(?i)" + pattern)
    if err != nil {
        return false
    }
    return re.MatchString(value)
}
```

### Instance Matcher in Go
```go
// Source: Custom implementation for SnoozeQL
package scheduler

import (
    "regexp"
    "strings"
    "snoozeql/internal/models"
)

// MatchInstance checks if an instance matches a set of selectors
func MatchInstance(instance *models.Instance, selectors []models.Selector, operator string) bool {
    if len(selectors) == 0 {
        return false // Require explicit selection
    }
    
    for _, sel := range selectors {
        matches := matchSelector(instance, sel)
        if operator == "or" && matches {
            return true
        }
        if operator == "and" && !matches {
            return false
        }
    }
    
    return operator == "and"
}

func matchSelector(instance *models.Instance, sel models.Selector) bool {
    // Check name
    if sel.Name != nil {
        if !matchMatcher(instance.Name, sel.Name) {
            return false
        }
    }
    
    // Check provider
    if sel.Provider != nil {
        provider := "gcp"
        if strings.HasPrefix(instance.Provider, "aws") {
            provider = "aws"
        }
        if *sel.Provider != provider {
            return false
        }
    }
    
    // Check tags
    for tagKey, matcher := range sel.Tags {
        tagValue, ok := instance.Tags[tagKey]
        if !ok || !matchMatcher(tagValue, matcher) {
            return false
        }
    }
    
    return true
}

func matchMatcher(value string, matcher *models.Matcher) bool {
    switch matcher.Type {
    case models.MatchExact:
        return value == matcher.Pattern
    case models.MatchContains:
        return strings.Contains(value, matcher.Pattern)
    case models.MatchPrefix:
        return strings.HasPrefix(value, matcher.Pattern)
    case models.MatchSuffix:
        return strings.HasSuffix(value, matcher.Pattern)
    case models.MatchRegex:
        re, err := regexp.Compile(matcher.Pattern)
        if err != nil {
            return false
        }
        return re.MatchString(value)
    default:
        return false
    }
}
```

### Filter Preview API Endpoint
```go
// Source: Custom implementation for SnoozeQL
// POST /api/v1/schedules/preview-filter
func (h *ScheduleHandler) PreviewFilter(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Selectors []models.Selector `json:"selectors"`
        Operator  string            `json:"operator"` // "and" or "or"
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    instances, err := h.instanceStore.ListInstances()
    if err != nil {
        http.Error(w, "Failed to list instances", http.StatusInternalServerError)
        return
    }
    
    var matched []models.Instance
    for _, inst := range instances {
        if scheduler.MatchInstance(&inst, req.Selectors, req.Operator) {
            matched = append(matched, inst)
        }
    }
    
    json.NewEncoder(w).Encode(map[string]any{
        "matched_count": len(matched),
        "instances":     matched,
    })
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hard-coded instance lists | Dynamic regex filters | This project | Flexible, maintainable assignment |
| Server-only preview | Client-side preview | Common practice | Instant feedback, better UX |
| Single match type | Multiple match types (exact, contains, regex) | This project | Power users can use regex, others use simpler options |

## Open Questions

Things that couldn't be fully resolved:

1. **AND/OR operator scope**
   - What we know: Need to combine multiple selectors
   - What's unclear: Is it AND between all selectors, or group-level operators?
   - Recommendation: Single operator for all selectors (simpler). Complex boolean logic deferred.

2. **Maximum instances in preview**
   - What we know: Need to show matched instances
   - What's unclear: Performance at scale (100s of instances)
   - Recommendation: Show count + first 10 instances, "View all" button for full list

3. **Regex syntax help**
   - What we know: Users may not know regex
   - What's unclear: How much help text/examples to provide
   - Recommendation: Simple preset options (contains, starts with) + "regex" for power users, with tooltip explaining basics

## Sources

### Primary (HIGH confidence)
- Go regexp package documentation - RE2 syntax, MatchString behavior
- JavaScript RegExp MDN documentation - Browser regex implementation
- Existing SnoozeQL models.go - Selector and Matcher type definitions

### Secondary (MEDIUM confidence)
- Existing SnoozeQL frontend patterns - Component structure, API client usage
- Headless UI documentation - Listbox, Menu components

### Tertiary (LOW confidence)
- Filter builder UX patterns - No authoritative source, based on common implementations

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing libraries, no new dependencies
- Architecture: HIGH - Extending existing patterns, clear data model
- Pitfalls: HIGH - Based on codebase analysis and common regex issues

**Research date:** 2026-02-23
**Valid until:** 90 days (stable patterns, no external API changes expected)
