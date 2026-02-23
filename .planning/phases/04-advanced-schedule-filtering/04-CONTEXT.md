# Phase 4: Advanced Schedule Filtering - Context

**Gathered:** 2026-02-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can assign schedules to instances using flexible regex-based filters. This phase delivers filter creation UI, filter logic, live preview, and schedule-instance assignment. Schedule TIME creation (the visual grid) was completed in Phase 3.

**Success Criteria:**
1. User can create schedule filters based on instance name using regex patterns
2. User can create schedule filters based on instance tags using regex patterns
3. User can create schedule filters based on cloud provider (AWS/GCP)
4. User can combine filters with AND/OR operators
5. User can preview which instances will match a filter before applying
6. User can view all created schedules in a dedicated schedules tab (already exists from Phase 3)

</domain>

<decisions>
## Implementation Decisions

### Filter Builder UI
- **Filter builder component**: Visual component for building filter rules
- **Rule types**: Instance name (regex), tags (key:value regex), cloud provider (dropdown)
- **Combination operators**: AND/OR toggle for combining multiple rules
- **Inline rule editing**: Add/remove rules dynamically, no separate modal for each rule
- **Visual feedback**: Rules display as pill/chip elements for easy scanning

### Filter Assignment to Schedules
- **ScheduleModal extension**: Add "Instance Filters" section below schedule time selection
- **Filters part of schedule data**: Selectors array stored with schedule (already supported in backend)
- **No separate filter entity**: Filters are properties of schedules, not independent objects

### Preview Functionality
- **Live preview panel**: Shows matching instances as user builds filters
- **Instance count badge**: Quick count of matched instances
- **Instance preview list**: Shows first N matching instances with expandable full list
- **API endpoint for matching**: POST /api/v1/schedules/preview-filter to test filters server-side

### Implementation Details
- **Backend regex matching**: Use Go regexp package for consistent server-side matching
- **Client-side preview**: Fetch all instances, filter client-side for instant preview
- **Selector structure**: Use existing Selector model with Name, Provider, Region, Engine, Tags matchers

### OpenCode's Discretion
- Exact filter builder visual design and layout
- Number of instances to show in preview (suggest 10 with "show all" option)
- Validation error messaging for invalid regex patterns
- Animation/transitions for filter rule addition/removal

</decisions>

<specifics>
## Specific Ideas

- The filter builder should feel like a query builder - visual and intuitive
- Preview should update in real-time as filters are modified
- AND/OR operators should be clearly visible between rules
- Invalid regex patterns should show immediate inline validation errors
- Preview panel should indicate when no instances match (empty state)
- Consider showing which filter rule caused each instance to match/not match

</specifics>

<deferred>
## Deferred Ideas

None - this phase completes the core scheduling workflow:
- Phase 3: Create schedules (when to sleep/wake)
- Phase 4: Assign schedules (which instances)

Future phases (5, 6) focus on activity-based recommendations which is a separate concern.

</deferred>

---

*Phase: 04-advanced-schedule-filtering*
*Context gathered: 2026-02-23*
