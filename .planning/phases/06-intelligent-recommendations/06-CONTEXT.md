# Phase 6: Intelligent Recommendations - Context

**Gathered:** 2026-02-23
**Status:** Ready for planning

<domain>
## Phase Boundary

System generates schedule recommendations based on activity analysis that users can review and apply. This phase takes the idle period detection from Phase 5 and presents actionable recommendations to users for schedule creation.

</domain>

<decisions>
## Implementation Decisions

### Recommendation format
- List view layout (not card grid or tabbed interface)
- Collapsible details in cards (core info visible, expand for details)
- "AI Suggested" badge for visual distinction
- Sorted by savings potential (highest first)
- Manual refresh only with dashboard button
- Clicking recommendation opens details modal
- Dismissed recommendations stored across sessions

### Recommendation behavior
- Simple confirm flow: show suggested times, user confirms to create
- Allow users to dismiss individual recommendations ("Mark as Not Now")
- Dismissed recommendations remembered across sessions
- Manual trigger button on dashboard to generate recommendations

### Content presentation
- Activity pattern summary displayed in recommendation details (24-hour CPU graph, connection patterns)
- Confidence scores shown as simple labels (High/Medium/Low) not percentages or stars
- Confidence thresholds: High (80-100%), Medium (50-79%), Low (0-49%)
- Projected cost savings displayed in USD

### Empty states
- Call to action on empty dashboard (prompt to refresh or wait for data)
- Minimum data requirement clearly stated ("Need 24+ hours of data")
- Dismissed recommendations count shown in empty state message
- Manual "Generate Recommendations" button visible

### OpenCode's Discretion
- Exact UI component library choices
- Animation/transitions for expand/collapse
- Error handling for recommendation failures
- Specific color palette for confidence labels

</decisions>

<specifics>
## Specific Ideas

- "Recommendations should feel like smart suggestions from a system that understands my usage patterns"
- Confidence labels should be simple and interpretable: "High confidence" rather than "85%"
- Savings estimate should be prominently displayed (users care about cost)
- Dismissed recommendations should stay gone until manually re-enabled

</specifics>

<deferred>
## Deferred Ideas

- User feedback on recommendation accuracy — future phase
- Recommendation history/audit log — future phase
- Bulk dismiss multiple recommendations at once — future phase

</deferred>

---

*Phase: 06-intelligent-recommendations*
*Context gathered: 2026-02-23*
