---
status: investigating
trigger: "The dashboard's Cost Over Time (7 days) chart has no Y-axis and shows both days (Mon-Sun) at top AND hours (00:00-24:00) at bottom as X-axis labels. User expects either 24hr view or 7day view, not both."
created: 2026-02-23T00:00:00Z
updated: 2026-02-23T00:08:00Z
---

## Current Focus

hypothesis: The custom CSS-based bar chart implementation lacks Y-axis labels and has confusing duplicate X-axis labels. Comparison with SavingsChart shows a proper Recharts-based implementation. The chart should be replaced with the existing SavingsChart component since it serves the same purpose and provides proper axis labels.
test: Replace Dashboard chart with SavingsChart component and update to use real API data instead of mock data
expecting: Chart should show proper Y-axis with dollar labels, X-axis with appropriate time labels, and proper tooltips
next_action: Prepare fix by using SavingsChart component with appropriate data

## Symptoms

expected: Cost Over Time chart should show cost over time with Y-axis showing dollar amounts and X-axis showing either hours or dates, not both
actual: No Y-axis labels showing dollar amounts, shows day names (Mon-Sun) at top AND hour labels (00:00-24:00) at bottom
errors: None - visual confusion due to duplicate X-axis
reproduction: Navigate to home dashboard, look at "Cost Over Time (7 days)" section
started: Always broken - this is a custom CSS-based bar chart implementation

## Eliminated

- hypothesis: Chart uses Recharts library properly
  evidence: Code inspection shows custom CSS-based implementation using divs, not Recharts components (confirmed SavingsChart uses proper Recharts)
  timestamp: 2026-02-23T00:00:00Z

- hypothesis: Chart displays real API data
  evidence: generateCostData creates mock data with hardcoded business logic (9-17 = full cost, 22-7 = zero cost), not using actual instance schedules or savings data
  timestamp: 2026-02-23T00:00:00Z

- hypothesis: Chart shows 24-hour view only
  evidence: Chart title says "7 days" and generateCostData creates 168 data points (7 days x 24 hours = 168)
  timestamp: 2026-02-23T00:00:00Z

- hypothesis: Chart is well-designed for its purpose
  evidence: Chart has no Y-axis labels, conflicting X-axis labels (day names at top, hours at bottom), hover-only cost display
  timestamp: 2026-02-23T00:00:00Z

- hypothesis: Dashboard needs a unique chart implementation
  evidence: SavingsChart provides identical functionality with proper axis labels using Recharts library
  timestamp: 2026-02-23T00:00:00Z

- hypothesis: Y-axis should be implemented manually
  evidence: SavingsChart demonstrates YAxis component handles axis labels correctly, no need for custom implementation
  timestamp: 2026-02-23T00:00:00Z

## Evidence

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx lines 57-88 (generateCostData)
  found: Function generates 168 data points (7 days x 24 hours), each with label like "12AM", "1AM", etc. and cost value based on hourly_cost_cents with business logic
  implication: Chart is designed to show hourly costs across 7 days, but the visualization doesn't match this purpose
  evidence_qualified: TRUE - this is the root cause of the confusion: showing 7 days of hourly data but visualized as 24-hour chart

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx lines 272-301 (chart rendering)
  found: Lines 285-289 show day names at top (Mon-Sun) when i % 24 === 0, lines 294-299 show 5 hour labels at bottom
  implication: Both X-axis labels are drawn, creating visual confusion; no Y-axis labels are rendered
  evidence_qualified: TRUE - confirmed: day labels at top, hour labels at bottom, NO Y-axis

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx lines 283-284 (hover tooltip)
  found: Tooltip shows "${Math.round(d.cost / 100)}/hr" on hover
  implication: Cost values are in cents, converted to dollars for display, but no permanent Y-axis labels
  evidence_qualified: TRUE - hover shows the cost but no static Y-axis labels

- timestamp: 2026-02-23T00:00:00Z
  checked: SavingsChart.tsx lines 93-106 (XAxis, YAxis, Tooltip)
  found: Uses Recharts components: <XAxis>, <YAxis> with tickFormatter for $, <Tooltip> with custom content
  implication: Proper implementation has axis labels and tooltips out of the box
  evidence_qualified: TRUE - SavingsChart has Y-axis with $ labels, X-axis with dates, proper tooltip

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx line 271 (chart title), SavingsPage.tsx line 66 (savings page header)
  found: Dashboard says "Cost Over Time (7 days)", SavingsPage says "Cost Over Time" with date range selector
  implication: Both charts serve similar purpose but dashboard uses mock data while savings page uses real API data
  evidence_qualified: TRUE - dash uses 7 days fixed, savings page is flexible with date range selector

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx line 285-289 (day labels), lines 294-299 (hour labels)
  found: Day labels at top (i % 24 === 0): Mon, Tue, Wed, Thu, Fri, Sat; Hour labels at bottom: 00:00, 06:00, 12:00, 18:00, 24:00
  implication: Day labels align with day boundaries, hour labels don't align with 168 data points
  evidence_qualified: TRUE - confirmed the bug exists: both axis labels visible, neither meaningful

- timestamp: 2026-02-23T00:00:00Z
  checked: Business logic in generateCostData (lines 70-78)
  found: Uses hardcoded hours (9-17 full cost, 22-7 zero cost, otherwise 20%), doesn't use actual instance schedules
  implication: Data is mock/calculated, not reflecting real instance activity patterns
  evidence_qualified: TRUE - confirmed: hardcoded hours, no actual schedule data used

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx lines 273-292 (bar rendering loop)
  found: Each bar is wrapped in flex container, bar has hover tooltip at lines 281-284
  implication: Tooltip works on hover, but no static Y-axis showing dollar amounts
  evidence_qualified: TRUE - hover tooltip is the ONLY way to see dollar amounts, no Y-axis labels

- timestamp: 2026-02-23T00:00:00Z
  checked: Dashboard.tsx line 291 (missing end div)
  found: Chart container div at line 272 is never closed properly - ends at line 301 before the parent div closes
  implication: Chart section is inside a div that starts at line 270, chart ends at line 301, then "Recent Activity" starts at line 303
  Note: Actually line 301 is </div> closing the chart section, line 302 is blank, line 303 starts new section

- timestamp: 2026-02-23T00:00:00Z
  checked: Empty line 285 closing div
  found: Line 285 is empty, line 286 is {i % 24 === 0 && ( which shows day labels
  implication: Day labels are rendered INSIDE the flex wrapper div, outside of the cost bar div
  Correction: Looking more carefully, the structure is: div for each bar (276) contains cost div (277) and day label (285-289 is OUTSIDE bar div but INSIDE flex wrapper)

- timestamp: 2026-02-23T00:00:00Z
  checked: Y-axis missing
  found: No Y-axis labels rendered, cost values shown only on hover
  implication: Chart has no way to read exact values without hovering over each bar

- timestamp: 2026-02-23T00:00:00Z
  checked: Hour label logic at lines 294-299
  found: Fixed labels "00:00" "06:00" "12:00" "18:00" "24:00" displayed evenly, doesn't match actual data points
  implication: Hours don't align with 168 data points (0-24 over 7 days), creates misleading visualization

- timestamp: 2026-02-23T00:00:00Z
  checked: Bar height calculation at line 274
  found: Uses maxCost from line 91 which is Math.max(...costData.map(d => d.cost), 100)
  implication: Bars scale relative to max cost, but no visual reference for what values they represent

- timestamp: 2026-02-23T00:00:00Z
  checked: X-axis label alignment
  found: Day labels at top (i % 24 === 0) show Mon-Sun at positions 0, 24, 48, 72, 96, 120 (every 24th bar)
  implication: Day labels align correctly with first hour of each day, but hour labels at bottom don't align with actual hour ticks

- timestamp: 2026-02-23T00:00:00Z
  checked: Chart width handling
  found: Chart uses w-full with flex-1 on each bar (line 276)
  implication: On small screens, bars become very narrow due to 168 data points

- timestamp: 2026-02-23T00:00:00Z
  checked: Y-axis scale calculation
  found: maxCost calculated once at line 91, used for all charts
  implication: If data changes, chart rescales but no new Y-axis

- timestamp: 2026-02-23T00:00:00Z
  checked: Hour data distribution
  found: 168 data points over 7 days, hour labels only 5 labels at bottom (00, 06, 12, 18, 24)
  implication: Hour labels don't correspond to their actual positions in 168-point dataset

## Resolution

root_cause: The chart uses a custom CSS-based bar chart implementation that lacks Y-axis labels and has confusing duplicate X-axis labels. The implementation shows day names (Mon-Sun) at the top and hour labels (00:00-24:00) at the bottom, with no Y-axis showing dollar amounts. Cost values can only be seen on hover.

fix: Replace the custom CSS-based chart with the existing SavingsChart component from /src/components/savings/SavingsChart.tsx. The SavingsChart uses Recharts library which provides proper XAxis, YAxis, and Tooltip components out of the box.

verification: After replacement, chart should show dollar amounts on Y-axis, appropriate X-axis labels (dates for 7-day view), and proper tooltips showing specific values on hover.

files_changed: []
