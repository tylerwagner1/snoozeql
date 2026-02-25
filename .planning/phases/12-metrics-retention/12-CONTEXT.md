# Phase 12: Metrics Retention - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Automatic cleanup of old metrics data to keep the database manageable. Metrics older than 7 days are automatically deleted without affecting application performance.

</domain>

<decisions>
## Implementation Decisions

### Cleanup timing
- Run on app startup (after 5-10 minute delay) + every 24 hours thereafter
- Track last run timestamp in the database to skip if already ran within 24 hours
- Fixed 24-hour interval (not configurable)
- Delete in batches (e.g., 1000 rows at a time) to avoid database locking
- Quiet logging — only log on errors, no success messages
- On error, wait for next scheduled run (no automatic retry)

### Retention boundary
- Hard 7-day cutoff (no grace period)
- Fixed 7-day retention (not configurable via env var)
- Uniform retention across all metrics tables (same rule for all)
- Delete based on metric timestamp, not insertion time

### OpenCode's Discretion
- Exact batch size for deletions
- Startup delay duration within the 5-10 minute range
- Database table/column for storing last run timestamp
- Specific error handling and logging format

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches for background job implementation.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 12-metrics-retention*
*Context gathered: 2026-02-25*
