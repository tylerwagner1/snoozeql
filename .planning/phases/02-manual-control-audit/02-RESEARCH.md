# Phase 2: Manual Control & Audit - Research

**Researched:** 2026-02-20
**Domain:** Sleep/wake operations, audit logging, multi-select UI
**Confidence:** HIGH

## Summary

Phase 2 adds manual sleep/wake controls with multi-select, confirmation dialogs, and an audit log. The codebase already has the foundation: Provider interface with StartDatabase/StopDatabase implemented for AWS RDS and GCP Cloud SQL, InstanceStore with UpsertInstance/ListInstances, and react-hot-toast for notifications.

The primary work is: (1) adding checkbox selection to InstancesPage, (2) building confirmation dialogs, (3) creating an EventStore for audit logging, (4) building an events API endpoint, and (5) creating an AuditLogPage. The existing Event model already exists in the schema with all needed fields.

**Primary recommendation:** Use native React state for selection management, Headless UI Dialog for confirmations, and the existing events table schema for audit logging.

## Standard Stack

The established libraries/tools for this domain:

### Core (Already in Project)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| React | 18.2 | UI framework | Already in use |
| react-hot-toast | - | Toast notifications | Already integrated |
| Chi router | - | HTTP routing | Already in use |
| pgx | v5 | PostgreSQL driver | Already in use |

### Supporting (To Add)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @headlessui/react | 2.1+ | Confirmation dialogs | Modal/dialog UI, accessible |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Headless UI Dialog | Custom modal | Headless UI handles accessibility, focus trapping, keyboard nav |
| Native checkboxes | Third-party data table | Keep it simple; native works fine |
| External state management | useState/useReducer | Selection state is local, no need for Redux/Zustand |

**Installation:**
```bash
npm install @headlessui/react
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── store/
│   ├── postgres.go          # Add EventStore methods
│   └── event_store.go       # (optional) Separate event store file
├── api/handlers/
│   ├── instances.go         # Add bulk start/stop handlers
│   └── events.go            # New: audit log handlers
web/src/
├── components/
│   ├── ConfirmDialog.tsx    # Reusable confirmation modal
│   └── SelectableTable.tsx  # (optional) Table with selection
├── pages/
│   ├── InstancesPage.tsx    # Add selection state and bulk actions
│   └── AuditLogPage.tsx     # New: show events history
└── lib/
    └── api.ts               # Add event endpoints
```

### Pattern 1: Multi-Select with React State
**What:** Manage selection using useState with Set<string>
**When to use:** Table row selection for bulk operations
**Example:**
```typescript
// Source: React best practices
const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

const toggleSelect = (id: string) => {
  setSelectedIds(prev => {
    const next = new Set(prev)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    return next
  })
}

const selectAll = () => {
  setSelectedIds(new Set(instances.map(i => i.id)))
}

const clearSelection = () => {
  setSelectedIds(new Set())
}
```

### Pattern 2: Headless UI Dialog for Confirmation
**What:** Accessible modal dialog for confirming destructive actions
**When to use:** Before sleep/wake operations affecting multiple instances
**Example:**
```typescript
// Source: https://headlessui.com/react/dialog
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react'

function ConfirmDialog({ isOpen, onClose, onConfirm, title, message, confirmText, confirmVariant }) {
  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/30" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="max-w-md bg-slate-800 p-6 rounded-xl border border-slate-700">
          <DialogTitle className="text-lg font-semibold text-white">{title}</DialogTitle>
          <p className="mt-2 text-slate-300">{message}</p>
          <div className="mt-4 flex justify-end gap-3">
            <button onClick={onClose}>Cancel</button>
            <button onClick={onConfirm} className={confirmVariant}>{confirmText}</button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  )
}
```

### Pattern 3: Bulk Operation API Design
**What:** POST endpoint accepting array of instance IDs
**When to use:** Multi-instance sleep/wake operations
**Example:**
```go
// Backend handler pattern
type BulkOperationRequest struct {
    InstanceIDs []string `json:"instance_ids"`
}

type BulkOperationResponse struct {
    Success  []string          `json:"success"`
    Failed   []OperationError  `json:"failed"`
}

type OperationError struct {
    InstanceID string `json:"instance_id"`
    Error      string `json:"error"`
}

// POST /api/v1/instances/bulk-stop
// POST /api/v1/instances/bulk-start
```

### Pattern 4: Event/Audit Logging
**What:** Record all operations with metadata to events table
**When to use:** Every sleep/wake operation, schedule change
**Example:**
```go
// Source: Existing schema in 001_base_schema.sql
type Event struct {
    ID             string    `json:"id" db:"id"`
    InstanceID     string    `json:"instance_id" db:"instance_id"`
    EventType      string    `json:"event_type" db:"event_type"`      // "sleep", "wake", "schedule_created", "schedule_updated"
    TriggeredBy    string    `json:"triggered_by" db:"triggered_by"`  // "manual", "schedule", "api"
    PreviousStatus string    `json:"previous_status" db:"previous_status"`
    NewStatus      string    `json:"new_status" db:"new_status"`
    Metadata       []byte    `json:"metadata,omitempty" db:"metadata"` // JSONB for extra context
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
```

### Anti-Patterns to Avoid
- **No confirmation for destructive actions:** Always confirm before sleep/wake
- **Silent failures:** Show toast for each failure in bulk operation
- **Blocking UI during operations:** Use optimistic updates with status transition
- **Not logging audit events:** Every operation must create an event record
- **Storing sensitive data in events:** Keep metadata minimal, no credentials

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Modal dialogs | Custom overlay + state | Headless UI Dialog | Focus trap, ESC handling, backdrop click, accessibility |
| Selection state | Complex reducer | useState with Set | Simple enough for this use case |
| Toast notifications | Custom toast system | react-hot-toast | Already integrated |
| UUID generation | Custom IDs | PostgreSQL uuid-ossp | Already configured |

**Key insight:** The existing codebase has a solid foundation. Focus on composing existing patterns rather than inventing new abstractions.

## Common Pitfalls

### Pitfall 1: AWS 7-Day Auto-Restart
**What goes wrong:** AWS automatically restarts RDS instances after 7 consecutive days of being stopped
**Why it happens:** AWS limitation to ensure maintenance updates can be applied
**How to avoid:** For Phase 2, document in UI that AWS has this limit. Future phases should implement re-stop mechanism
**Warning signs:** Instance shows as "available" despite no user action

### Pitfall 2: Race Conditions on Instance State
**What goes wrong:** User clicks stop while instance is already stopping, or cloud status hasn't synced
**Why it happens:** Instance state in database may lag behind actual cloud state
**How to avoid:** 
- Check current status before allowing action
- Disable buttons for instances in transitional states (stopping, starting)
- Refresh discovery after operations
**Warning signs:** Multiple operations queued for same instance

### Pitfall 3: Read Replicas Cannot Be Stopped (AWS)
**What goes wrong:** Stop operation fails on read replicas
**Why it happens:** AWS RDS limitation - read replicas must stay running
**How to avoid:** Flag read replicas during discovery, filter them from bulk operations
**Warning signs:** API error "cannot stop read replica"

### Pitfall 4: GCP Cloud SQL Different State Names
**What goes wrong:** Status mapping confusion between AWS and GCP
**Why it happens:** AWS uses "available"/"stopped", GCP uses "RUNNABLE"/"SUSPENDED"
**How to avoid:** Normalize status in provider layer (already partially done)
**Warning signs:** UI shows raw provider status strings

### Pitfall 5: Confirmation Dialog Skip
**What goes wrong:** User accidentally clicks confirm without reading
**Why it happens:** Modal fatigue, too many confirmations
**How to avoid:** 
- Show instance count prominently
- Use different colors for sleep (yellow/warning) vs wake (green)
- Require explicit count confirmation for large batches (e.g., "Type '5' to confirm stopping 5 instances")
**Warning signs:** Support tickets for accidental operations

### Pitfall 6: Missing Event Context
**What goes wrong:** Audit log shows what happened but not why or by whom
**Why it happens:** Not capturing sufficient metadata
**How to avoid:** Always include: who initiated, what triggered (manual/schedule), previous state, timestamp
**Warning signs:** Cannot answer "who stopped my database?"

## Code Examples

Verified patterns from official sources and existing codebase:

### Selection State Management
```typescript
// web/src/pages/InstancesPage.tsx
const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
const [showConfirmDialog, setShowConfirmDialog] = useState<'sleep' | 'wake' | null>(null)

// Checkbox column in table
<td className="px-4 py-4">
  <input
    type="checkbox"
    checked={selectedIds.has(instance.id)}
    onChange={() => toggleSelect(instance.id)}
    className="rounded border-slate-600 bg-slate-700 text-blue-500"
  />
</td>

// Select all checkbox in header
<th className="px-4 py-4">
  <input
    type="checkbox"
    checked={selectedIds.size === filteredInstances.length && filteredInstances.length > 0}
    onChange={(e) => e.target.checked ? selectAll() : clearSelection()}
  />
</th>

// Bulk action buttons (shown when selection > 0)
{selectedIds.size > 0 && (
  <div className="flex gap-2 items-center">
    <span className="text-sm text-slate-400">{selectedIds.size} selected</span>
    <button onClick={() => setShowConfirmDialog('sleep')}>Sleep Selected</button>
    <button onClick={() => setShowConfirmDialog('wake')}>Wake Selected</button>
    <button onClick={clearSelection}>Clear</button>
  </div>
)}
```

### EventStore Implementation
```go
// internal/store/postgres.go - Add EventStore

type EventStore struct {
    db *Postgres
}

func NewEventStore(db *Postgres) *EventStore {
    return &EventStore{db: db}
}

func (s *EventStore) CreateEvent(ctx context.Context, event *models.Event) error {
    query := `
        INSERT INTO events (instance_id, event_type, triggered_by, previous_status, new_status, metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at`
    return s.db.QueryRowContext(ctx, query,
        event.InstanceID, event.EventType, event.TriggeredBy,
        event.PreviousStatus, event.NewStatus, event.Metadata,
    ).Scan(&event.ID, &event.CreatedAt)
}

func (s *EventStore) ListEvents(ctx context.Context, limit int, offset int) ([]models.Event, error) {
    query := `
        SELECT id, instance_id, event_type, triggered_by, previous_status, new_status, metadata, created_at
        FROM events ORDER BY created_at DESC LIMIT $1 OFFSET $2`
    rows, err := s.db.Query(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var events []models.Event
    for rows.Next() {
        var e models.Event
        err := rows.Scan(&e.ID, &e.InstanceID, &e.EventType, &e.TriggeredBy,
            &e.PreviousStatus, &e.NewStatus, &e.Metadata, &e.CreatedAt)
        if err != nil {
            return nil, err
        }
        events = append(events, e)
    }
    return events, rows.Err()
}

func (s *EventStore) ListEventsByInstance(ctx context.Context, instanceID string) ([]models.Event, error) {
    query := `
        SELECT id, instance_id, event_type, triggered_by, previous_status, new_status, metadata, created_at
        FROM events WHERE instance_id = $1 ORDER BY created_at DESC`
    rows, err := s.db.Query(ctx, query, instanceID)
    // ... same pattern as above
}
```

### Bulk Operation Handler
```go
// POST /api/v1/instances/bulk-stop
func (h *InstanceHandler) BulkStop(w http.ResponseWriter, r *http.Request) {
    var req BulkOperationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    ctx := r.Context()
    var success []string
    var failed []OperationError
    
    for _, instanceID := range req.InstanceIDs {
        // Get instance to find provider
        instance, err := h.instanceStore.GetInstanceByID(ctx, instanceID)
        if err != nil {
            failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not found"})
            continue
        }
        
        // Check if instance can be stopped
        if instance.Status != "available" && instance.Status != "running" {
            failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not in stoppable state"})
            continue
        }
        
        // Stop the instance
        if err := h.discoveryService.StopDatabase(ctx, instance.Provider, instance.Name); err != nil {
            failed = append(failed, OperationError{InstanceID: instanceID, Error: err.Error()})
            continue
        }
        
        // Log the event
        event := &models.Event{
            InstanceID:     instanceID,
            EventType:      "sleep",
            TriggeredBy:    "manual",
            PreviousStatus: instance.Status,
            NewStatus:      "stopping",
        }
        h.eventStore.CreateEvent(ctx, event)
        
        success = append(success, instanceID)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(BulkOperationResponse{Success: success, Failed: failed})
}
```

### Confirmation Dialog Component
```typescript
// web/src/components/ConfirmDialog.tsx
import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react'

interface ConfirmDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmText: string
  confirmVariant: 'danger' | 'warning' | 'success'
  loading?: boolean
}

export function ConfirmDialog({ 
  isOpen, onClose, onConfirm, title, message, confirmText, confirmVariant, loading 
}: ConfirmDialogProps) {
  const variantClasses = {
    danger: 'bg-red-600 hover:bg-red-500',
    warning: 'bg-yellow-600 hover:bg-yellow-500',
    success: 'bg-green-600 hover:bg-green-500',
  }
  
  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop 
        transition
        className="fixed inset-0 bg-black/50 duration-200 ease-out data-closed:opacity-0" 
      />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel 
          transition
          className="max-w-md w-full bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl duration-200 ease-out data-closed:scale-95 data-closed:opacity-0"
        >
          <DialogTitle className="text-lg font-semibold text-white">{title}</DialogTitle>
          <p className="mt-3 text-slate-300">{message}</p>
          <div className="mt-6 flex justify-end gap-3">
            <button 
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg"
            >
              Cancel
            </button>
            <button 
              onClick={onConfirm}
              disabled={loading}
              className={`px-4 py-2 text-white rounded-lg ${variantClasses[confirmVariant]} disabled:opacity-50`}
            >
              {loading ? 'Processing...' : confirmText}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  )
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Bootstrap modals | Headless UI | 2022+ | Better accessibility, unstyled composability |
| Redux for local state | useState/useReducer | React 16.8+ | Simpler for component-local state |
| Manual focus trapping | Dialog component | Native HTML dialog / Headless UI | Built-in accessibility |

**Deprecated/outdated:**
- Using `alert()` or `confirm()` for confirmations - poor UX, blocking
- Managing dialog state outside component - keep state local

## Open Questions

Things that couldn't be fully resolved:

1. **Schedule CRUD for AUDIT-02**
   - What we know: Need to log schedule creation/updates
   - What's unclear: Schedule CRUD UI is Phase 3, but AUDIT-02 requires logging
   - Recommendation: Add event logging to schedule endpoints in Phase 3, or add schedule endpoints now without full UI

2. **Bulk operation concurrency limits**
   - What we know: Provider APIs have rate limits
   - What's unclear: Exact limits for AWS RDS and GCP Cloud SQL batch operations
   - Recommendation: Start with sequential operations (simple), add parallel with rate limiting in future if needed

3. **User identification for audit**
   - What we know: `triggered_by` field exists
   - What's unclear: Is there user auth? Who is "the user"?
   - Recommendation: Use "manual" for now, add user ID when auth is implemented

## Sources

### Primary (HIGH confidence)
- Headless UI Dialog documentation: https://headlessui.com/react/dialog - Dialog patterns, accessibility
- AWS RDS StopDBInstance: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html - 7-day restart, limitations
- GCP Cloud SQL activation policy: https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance - ALWAYS/NEVER policy
- React useState documentation: https://react.dev/reference/react/useState - Selection state patterns

### Secondary (MEDIUM confidence)
- Existing codebase analysis - Provider implementations, store patterns
- Existing schema in 001_base_schema.sql - Events table structure

### Tertiary (LOW confidence)
- None - all findings verified with official sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing libraries, Headless UI is well-documented
- Architecture: HIGH - Building on existing patterns in codebase
- Pitfalls: HIGH - AWS/GCP limitations well-documented in official docs

**Research date:** 2026-02-20
**Valid until:** 2026-03-20 (30 days - stable patterns)
