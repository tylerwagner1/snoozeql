# Phase 3: Basic Scheduling - Context

**Gathered:** 2026-02-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can create time-based sleep/wake schedules. This phase delivers schedule CRUD and the visual time selection interface. Instance assignment with regex filters is Phase 4.

**Success Criteria:**
1. User can create a schedule specifying start time, end time, and days of week
2. Created schedules appear in the schedules list

</domain>

<decisions>
## Implementation Decisions

### Schedule Time Selection (Weekly Grid)
- **Visual heatmap grid**: 7 columns (days) × 24 rows (hours) of clickable cells
- **Two states per cell**: Sleep (dark) / Wake (light) — binary toggle
- **Click-drag painting**: Toggle mode — first cell clicked determines action (if Wake, drag paints Sleep and vice versa)
- **CRON fallback**: "Switch to CRON" button swaps grid for text input with description of what the CRON expression achieves
- Grid allows visual "painting" of sleep windows across the week

### Schedule List Display
- **Table layout** with columns: Name, Active Days Summary, Sleep Hours Summary, Status Toggle, Instances Count
- **Actions column** with explicit Edit/Delete buttons (not kebab menu, not row-click)
- **Empty state**: Friendly illustration + "Create your first schedule" CTA button
- Active days shown as summary text (e.g., "Mon-Fri", "Weekdays", "Every day")
- Sleep hours shown as summary (e.g., "10pm-6am", "22:00-06:00")

### Create Schedule Flow
- **Modal dialog** for creation (not dedicated page, not inline expansion)
- **Required fields**: Name (user must provide), at least one sleep hour selected on grid
- **Grid starts empty** — user paints from scratch, no pre-filled suggestions
- **After create**: Close modal, show new schedule in list (no toast, no follow-up prompts)
- Edit uses same modal with pre-populated grid

### OpenCode's Discretion
- Exact grid cell sizing and visual design
- Hour label formatting (12h vs 24h)
- CRON input validation and error messaging
- Table sorting/filtering options
- Responsive behavior for the grid on mobile

</decisions>

<specifics>
## Specific Ideas

- The weekly grid should feel like a calendar heatmap where you can visually "see" when databases will sleep
- Toggle-mode drag painting is intuitive — whatever state you click first, dragging paints the opposite
- CRON mode is for power users who want precise control or copy-paste existing expressions

</specifics>

<deferred>
## Deferred Ideas

- **Instance assignment with regex filters and live preview** — Phase 4 (Advanced Schedule Filtering)
  - User wants "Instances affected" preview when adding selectors/filters
  - Preview should show count and names when filter like "name = oregon*" is applied
  - This is the core of Phase 4's regex-based assignment with preview

</deferred>

---

*Phase: 03-basic-scheduling*
*Context gathered: 2026-02-23*
