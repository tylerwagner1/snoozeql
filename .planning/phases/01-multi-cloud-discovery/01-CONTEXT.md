# Phase 1: Multi-Cloud Discovery - Context

**Gathered:** 2026-02-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Display database instances from multiple AWS and GCP accounts in a unified interface. Users can add cloud account connections, see all instances with status, and navigate to detailed views. Manual sleep/wake and scheduling are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Connection management
- Separate forms for AWS (access key/secret) and GCP (service account JSON)
- Dedicated connections tab for managing all cloud connections
- Users can add unlimited accounts per provider

### Instance list layout
- Data table format (not cards)
- Core five fields: name, status, provider, region, engine
- Sorted alphabetically by default, with clickable column headers for reverse sort
- Quick filters via dropdown menus on columns

### Multi-account organization
- Add provider and account name columns to table
- Unified view without visual separation between provider groups
- Instances from all accounts in single seamless table

### Status indicators
- Status chips for connection status (connected/syncing/failed)
- Skeleton screens for loading states (initial sync and refresh)
- Toast notifications for connection failures and errors

### Dashboard home screen
- Combination of stats cards, cost graphs, and activity feed
- Interactive cards: "running instances: 3", "stopped instances", "DBs scheduled to sleep in next 1hr"
- Clicking stat cards navigates to instances tab pre-filtered by that criteria
- Prominent CTAs: "Add AWS Account", "Add GCP Account", "SeeRecommended Schedules"
- Sidebar quick link to instance list from dashboard

### OpenCode's Discretion
- Exact color scheme and theming
- Specific chart library for cost visualization
- Precise icon set for status indicators
- Card design details (spacing, borders, hover effects)
- Toast notification behavior (duration, position, close options)

</decisions>

<specifics>
## Specific Ideas

- "Stats cards should be clickable and pre-filter the instances view"
- "Cost graph: hourly cost projection based on current instance states"
- "Activity feed: recent sleep/wake operations with timestamps"
- "Sidebar quick link: concise navigation without dedicated tab"
- "Status chips: small, pill-shaped badges next to connection names"

</specifics>

<deferred>
## Deferred Ideas

- Instance details modal/actions — Phase 2 (Manual Control)
- Schedule creation ui — Phase 3 (Basic Scheduling)
- Recommendation engine display — Phase 6 (Intelligent Recommendations)
- Multi-select and bulk operations — Phase 2 (Manual Control)

</deferred>

---

*Phase: 01-multi-cloud-discovery*
*Context gathered: 2026-02-20*
