# Feature Landscape: Cost Savings Tracking

**Domain:** Cloud cost management / Database cost optimization dashboards
**Researched:** 2026-02-23
**Confidence:** HIGH (verified with AWS, GCP, and FinOps Foundation sources)

## Overview

This research identifies the features required for SnoozeQL v1.1's cost savings tracking capabilities, aligned with requirements SAV-01 through SAV-05. The findings draw from industry leaders (AWS Cost Explorer, GCP Billing Reports, Kubecost, FinOps Foundation best practices) to define what users expect from cost optimization dashboards.

---

## 1. Core Dashboard Features

Must-have components that users expect from any cost savings dashboard.

### Summary Cards (Hero Metrics)

| Feature | Why Expected | Complexity | Implementation Notes |
|---------|--------------|------------|----------------------|
| **Total Savings (Current Period)** | Primary value proposition - shows SnoozeQL's ROI | Low | Sum of `estimated_savings_cents` from savings table |
| **Month-to-Date Savings** | Standard financial reporting period | Low | Filter savings by current month |
| **Forecasted Savings** | Users want to predict future value | Medium | Project based on schedule patterns + historical data |
| **Savings Trend Indicator** | Shows if savings are improving | Low | Compare current period vs previous period (% change) |
| **Stopped Hours (Current Period)** | Validates that sleep is happening | Low | Sum of `stopped_minutes` / 60 |

**AWS Cost Explorer Pattern:** Shows "Month-to-date costs" and "Forecasted month end costs" prominently at top with percentage change indicators.

**Recommendation:** Use 4 summary cards: Total Savings, Month-to-Date, Forecasted, and Savings Trend (% change vs last period).

### Primary Chart

| Feature | Why Expected | Complexity | Implementation Notes |
|---------|--------------|------------|-------|
| **Time-series savings chart** | Visual trend analysis | Medium | Stacked bar chart showing daily/weekly/monthly savings |
| **Actual vs Projected overlay** | Validates estimates | Medium | Two data series on same chart |
| **Configurable time range** | Different analysis needs | Low | 7d, 30d, 90d, custom range selectors |

**Recommendation:** Use Recharts (already in stack) for a stacked bar chart with daily granularity for short ranges, weekly/monthly for longer ranges.

### Data Table

| Feature | Why Expected | Complexity | Implementation Notes |
|---------|--------------|------------|-------|
| **Per-instance savings breakdown** | Attribution and accountability | Low | Table with instance name, savings, stopped hours |
| **Sortable columns** | Find top savers | Low | Sort by savings, hours stopped, instance name |
| **Drill-down to instance detail** | Investigation workflow | Low | Link to instance detail page |

---

## 2. Key Metrics

What to display and why, based on FinOps best practices and industry standards.

### Primary Metrics (Must Display)

| Metric | Definition | Why Important | Calculation |
|--------|------------|---------------|-------------|
| **Total Savings** | Money not spent due to stopped instances | Primary value metric | `hourly_cost_cents * stopped_minutes / 60` |
| **Stopped Hours** | Total time instances were stopped | Validates system is working | Direct from events or savings table |
| **Running Cost** | What would have been spent if always running | Baseline comparison | `hourly_cost_cents * total_hours` |
| **Actual Cost** | What was actually spent | Reality check | `hourly_cost_cents * running_minutes / 60` |

### Secondary Metrics (Should Display)

| Metric | Definition | Why Important | Calculation |
|--------|------------|---------------|-------------|
| **Savings Percentage** | Savings as % of what would have spent | Easy to understand efficiency | `(savings / running_cost) * 100` |
| **Cost per Instance** | Average hourly cost per instance | Identifies expensive instances | Sum costs / instance count |
| **Schedule Effectiveness** | How well schedules match actual usage | Optimization feedback | Compare scheduled stop time vs actual idle time |

### ROI Metrics (Optional but Valuable)

| Metric | Definition | Why Important | Calculation |
|--------|------------|---------------|-------------|
| **Daily Average Savings** | Normalized daily savings | Consistent comparison | Total savings / days in period |
| **Projected Annual Savings** | Extrapolated yearly value | Business justification | Daily average * 365 |

**FinOps Foundation Guidance:** "Context relevant cost reporting data available to all Core Personas" - metrics should be understandable by both technical and finance users.

---

## 3. User Workflows

How users interact with cost data, based on observed patterns in AWS Cost Explorer and GCP Billing.

### Workflow 1: Executive Summary (Daily Check)

**User Goal:** "Is SnoozeQL saving me money?"

**Steps:**
1. View dashboard summary cards
2. Check savings trend indicator (up/down vs last period)
3. Glance at chart for anomalies
4. Done (< 30 seconds)

**Required Features:**
- Summary cards with clear numbers
- Percentage change indicator
- Clean, scannable UI

### Workflow 2: Cost Attribution (Monthly Review)

**User Goal:** "Which instances are saving the most? Which aren't contributing?"

**Steps:**
1. View dashboard for overall picture
2. View per-instance savings table
3. Sort by savings to find top/bottom performers
4. Drill into specific instances for detail
5. Identify instances that should be stopped more

**Required Features:**
- Per-instance savings table (SAV-04)
- Sorting and filtering
- Instance detail page with savings history

### Workflow 3: Trend Analysis (Optimization)

**User Goal:** "Are my schedules optimized? How have savings changed over time?"

**Steps:**
1. Select longer time range (30d, 90d)
2. Review historical chart
3. Identify patterns (weekends vs weekdays, specific periods)
4. Compare actual vs projected costs (SAV-02)
5. Adjust schedules if needed

**Required Features:**
- Historical activity charts (SAV-03)
- Actual vs projected comparison (SAV-02, SAV-05)
- Time range selector

### Workflow 4: Billing Forecast (Planning)

**User Goal:** "What will my bill look like this month?"

**Steps:**
1. View current month's actual costs
2. View projected costs for rest of month
3. Compare to previous months
4. Validate forecast accuracy

**Required Features:**
- Cost projection (SAV-05)
- Comparison view (actual vs expected)
- Historical monthly totals

---

## 4. Historical Visualization Requirements

Time-based charting requirements based on industry patterns.

### Time Range Options

| Range | Granularity | Use Case |
|-------|-------------|----------|
| **Last 7 days** | Daily | Recent activity, troubleshooting |
| **Last 30 days** | Daily | Monthly review, trend analysis |
| **Last 90 days** | Weekly | Quarterly review, pattern identification |
| **Custom range** | Auto (daily/weekly/monthly) | Specific period analysis |

**AWS Cost Explorer Pattern:** Uses daily granularity for ranges up to 62 days, then automatically switches to monthly granularity.

### Chart Types

| Chart Type | When to Use | Data |
|------------|-------------|------|
| **Stacked Bar** | Primary savings view | Daily/weekly savings by instance or schedule |
| **Line Chart** | Trend comparison | Actual vs projected over time |
| **Area Chart** | Cumulative view | Running total of savings |

**Recommendation:** Primary chart should be a stacked bar chart (matches existing Dashboard.tsx pattern and Recharts capabilities).

### Grouping Options

| Group By | What It Shows | Priority |
|----------|---------------|----------|
| **Date** | Daily/weekly/monthly savings | Must have |
| **Instance** | Per-instance attribution | Must have |
| **Schedule** | Which schedules save most | Should have |
| **Region** | Regional cost distribution | Nice to have |
| **Engine** | Database type comparison | Nice to have |

---

## 5. Per-Instance Savings Attribution

How to show which instances saved money (SAV-04).

### Attribution Model

**Calculation:** For each instance, for each day:
```
daily_savings = hourly_cost_cents * stopped_minutes / 60 / 100 (convert to dollars)
```

**Data Source:** The `Saving` model already exists:
```go
type Saving struct {
    ID                    string
    InstanceID            string
    Date                  string
    StoppedMinutes        int
    EstimatedSavingsCents int
}
```

### Per-Instance View Requirements

| Feature | Description | Priority |
|---------|-------------|----------|
| **Instance Name** | Identifier | Must have |
| **Total Savings** | Sum for selected period | Must have |
| **Stopped Hours** | Sum for selected period | Must have |
| **Hourly Rate** | Instance cost per hour | Should have |
| **Savings Percentage** | % of potential cost saved | Should have |
| **Savings Trend** | Sparkline or change indicator | Nice to have |

### Attribution Accuracy Considerations

1. **Start/Stop Event Tracking:** Use Event table to track exact times
2. **Partial Hour Billing:** AWS bills per-second with 10-minute minimum; our calculation should align
3. **Status Changes:** Account for instance status during calculations
4. **Timezone Handling:** Store times in UTC, display in user's timezone

---

## 6. Actual vs Projected Comparison (SAV-02, SAV-05)

How existing tools show cost comparisons.

### AWS Cost Explorer Pattern

| Metric | Definition |
|--------|------------|
| **Actual Cost** | What you spent (from billing data) |
| **Forecasted Cost** | ML-based projection of month-end cost |
| **Comparison** | % difference from last period |

### For SnoozeQL Context

| Metric | Definition | Calculation |
|--------|------------|-------------|
| **Projected Cost (Without SnoozeQL)** | What costs would be if instances ran 24/7 | `hourly_cost * 24 * days` |
| **Projected Cost (With Schedules)** | Expected cost based on schedule patterns | `hourly_cost * scheduled_running_hours` |
| **Actual Cost** | What was actually spent based on events | `hourly_cost * actual_running_hours` |
| **Savings** | Difference | `Projected - Actual` |

### Visualization Approach

**Recommendation:** Use a dual-series chart showing:
1. **Projected without schedules** (light gray/dotted line) - baseline
2. **Actual cost** (solid color) - reality
3. **Savings** shown as the gap between the two

This clearly demonstrates SnoozeQL's value.

---

## 7. Export and Reporting Requirements

Common requirements for reports, based on industry patterns.

### Export Formats

| Format | Use Case | Priority |
|--------|----------|----------|
| **CSV** | Spreadsheet analysis, import to other tools | Should have |
| **PDF** | Management reports, documentation | Nice to have |
| **API** | Integration with other systems | Already exists |

### Export Contents

| Data | Include | Notes |
|------|---------|-------|
| Instance ID | Yes | For joining with other data |
| Instance Name | Yes | Human readable |
| Date | Yes | Time series |
| Stopped Minutes | Yes | Raw data |
| Estimated Savings | Yes | Calculated value |
| Hourly Cost | Yes | For verification |

### Reporting Workflows

| Report Type | Audience | Frequency | Content |
|-------------|----------|-----------|---------|
| **Daily Summary** | Ops | Daily | Top savers, anomalies |
| **Weekly Digest** | Team | Weekly | Trend, total savings |
| **Monthly Report** | Management | Monthly | ROI, total savings, comparison |

**For POC scope:** CSV export is sufficient. PDF and automated reports are out of scope.

---

## 8. Filtering and Sorting Requirements

What filtering capabilities users need.

### Filter Dimensions

| Dimension | Priority | Already in SnoozeQL |
|-----------|----------|---------------------|
| **Time Range** | Must have | No (needs implementation) |
| **Instance** | Must have | Yes (instance list exists) |
| **Provider (AWS/GCP)** | Should have | Yes (in Instance model) |
| **Region** | Should have | Yes (in Instance model) |
| **Engine** | Nice to have | Yes (in Instance model) |
| **Schedule** | Nice to have | Yes (in Schedule model) |
| **Tags** | Nice to have | Yes (in Instance model) |

### Sort Options for Tables

| Column | Direction | Default |
|--------|-----------|---------|
| Savings | Desc | **Yes** (show top savers first) |
| Instance Name | Asc/Desc | No |
| Stopped Hours | Desc | No |
| Date | Desc | No |

---

## 9. Table Stakes vs Differentiators

### Table Stakes (Must Have for v1.1)

Features users expect. Missing = product feels incomplete.

| Feature | Requirement ID | Notes |
|---------|----------------|-------|
| Total savings display | SAV-02 | Summary card |
| Per-instance savings | SAV-04 | Attribution table |
| Historical chart | SAV-03 | Time series visualization |
| Actual vs projected | SAV-02 | Comparison view |
| Time range selection | SAV-03 | 7d, 30d, 90d, custom |

### Differentiators (Nice to Have)

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Savings forecasting | Predict future value | High | ML-based projection |
| Schedule effectiveness scoring | Optimization feedback | Medium | Compare schedules |
| Export to CSV | External analysis | Low | Simple implementation |
| Per-schedule savings | Schedule attribution | Medium | Aggregate by schedule |

### Anti-Features (Do NOT Build)

Features to explicitly NOT build for v1.1 POC.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Real-time billing API integration | Complexity, API costs, out of scope | Use instance hourly cost estimation |
| Multi-currency support | Complexity | Single currency (USD, configurable) |
| PDF report generation | Heavyweight for POC | CSV export only |
| Email reports | Requires notification infrastructure (out of scope) | Manual export |
| Cost anomaly detection | ML complexity | Simple trend display |

---

## 10. Implementation Recommendations

### Phase 1: Core Savings Display (SAV-01, SAV-02)

1. **Savings Calculation Service**
   - Calculate savings from Event records (stop/start events)
   - Populate Saving table with daily aggregates
   - API endpoint: `GET /api/v1/savings?from=DATE&to=DATE`

2. **Summary Cards**
   - Total savings (current month)
   - Savings vs last month (% change)
   - Total stopped hours
   - Add to existing Dashboard.tsx

### Phase 2: Historical Charts (SAV-03)

1. **Time Range Selector**
   - 7d, 30d, 90d, custom
   - Store in URL params for sharing

2. **Savings Chart**
   - Stacked bar chart using Recharts
   - Daily granularity for <=30d, weekly for >30d
   - Group by instance initially

### Phase 3: Per-Instance Attribution (SAV-04)

1. **Savings Table**
   - Sortable by savings, instance name, stopped hours
   - Link to instance detail page
   - Filter by provider, region

2. **Instance Detail Enhancement**
   - Add savings history section to InstanceDetailPage
   - Show daily savings for that instance

### Phase 4: Cost Projection (SAV-05)

1. **Projection Calculation**
   - Based on schedule patterns
   - Compare expected vs actual

2. **Comparison View**
   - Overlay on historical chart
   - Clear visual of "what you saved"

---

## Sources

### High Confidence (Official Documentation)

- AWS Cost and Usage Reports: https://docs.aws.amazon.com/cur/latest/userguide/what-is-cur.html
- AWS Cost Explorer: https://docs.aws.amazon.com/cost-management/latest/userguide/ce-exploring-data.html
- GCP Billing Reports: https://cloud.google.com/billing/docs/how-to/reports
- AWS RDS Pricing: https://aws.amazon.com/rds/pricing/

### Medium Confidence (Industry Standards)

- FinOps Foundation - Reporting & Analytics: https://www.finops.org/framework/capabilities/reporting-analytics/
- IBM Kubecost - Cost monitoring patterns: https://www.apptio.com/products/kubecost/

### Existing SnoozeQL Implementation

- Dashboard.tsx - Current dashboard patterns
- models.go - Saving, Event, Instance models
- STACK.md - Recharts already available

---

*Feature landscape research: 2026-02-23*
