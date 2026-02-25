# Phase 11: Time-Series Visualization - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Display metrics history (CPU, Memory, Connections) on the Instance Details page with time-series charts and selectable time ranges (1h, 6h, 24h, 7d). Users can view historical patterns to understand instance utilization. This phase builds on Phase 10's metrics collection.

</domain>

<decisions>
## Implementation Decisions

### Chart layout & density
- Tabbed/toggled view — one metric visible at a time with tab selector
- Short tab labels: CPU | Memory | Connections
- Default tab: CPU
- Compact chart height (~150px)

### Time range behavior
- Single time range selector above charts (applies to all tabs)
- Default time range: 24 hours
- No auto-refresh — static view, user manually refreshes page
- Button style: OpenCode's discretion (match existing UI patterns)

### Data point presentation
- Line chart (not area or bar)
- No dots on data points — clean line only
- Crosshair + tooltip on hover (vertical line to x-axis plus value/timestamp)
- Fixed Y-axis scale: 0-100% for CPU and Memory (Connections auto-scales)

### Loading & empty states
- Loading: Spinner centered in chart area
- Empty: Show chart axes with "No data available" message
- Data gaps (e.g., sleeping instances): Show as zero (line drops to zero)
- Errors: Error message displayed in chart area

### OpenCode's Discretion
- Time range button/selector visual style
- Exact spacing and typography
- Chart library choice
- Tooltip formatting details
- X-axis label density/formatting

</decisions>

<specifics>
## Specific Ideas

- Gaps from sleeping instances should show as zero values, not breaks in the line — makes it clear when the instance was inactive
- Keep it compact (150px) to leave room for other Instance Details content

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 11-time-series-visualization*
*Context gathered: 2026-02-25*
