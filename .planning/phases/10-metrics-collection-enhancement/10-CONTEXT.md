# Phase 10: Metrics Collection Enhancement - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

System reliably collects and persists CPU, Memory, and Connections metrics. Adds FreeableMemory to existing CloudWatch collector, ensures 15-minute collection interval, and persists data across restarts.

</domain>

<decisions>
## Implementation Decisions

### Memory representation
- Store memory as **percentage available**, not raw bytes
- Calculate percentage using **hardcoded instance class mapping** (db.t3.*, db.r5.*, db.m5.* — ~20 entries)
- When instance class is unmapped, show **'N/A'** for memory metric (don't skip collection)

### Error handling
- On collection failure for one instance: **skip and continue** with other instances (log the failure)
- Store **NULL** for missing individual metrics (not 0, not skip entire row)
- Surface failures via **instance health badge** showing "Metrics unavailable"
- Badge clears **immediately on next successful collection** (no debounce)

### Collection timing
- **Store zeros explicitly** for sleeping instances (shows "asleep" state in metrics data)
- After instance wake: **wait for next scheduled run** (no immediate collection)
- After app restart: **wait for next scheduled time** (resume normal 15-min cycle)
- Collection interval: **fixed at 15 minutes** (hardcoded, not configurable)

### OpenCode's Discretion
- Exact implementation of instance class memory lookup
- Logging format and verbosity for failures
- Badge UI styling and positioning

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches for the technical implementation.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-metrics-collection-enhancement*
*Context gathered: 2026-02-24*
