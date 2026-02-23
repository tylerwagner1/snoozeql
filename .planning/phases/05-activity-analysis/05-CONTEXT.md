# Phase 5: Activity Analysis - Context

**Gathered:** 2026-02-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Collect and analyze CloudWatch metrics for AWS RDS instances and Cloud Monitoring metrics for GCP Cloud SQL instances to identify periods of low/zero activity suitable for sleep scheduling. This phase delivers the data foundation for Phase 6 intelligent recommendations.

</domain>

<decisions>
## Implementation Decisions

### Metric collection sources
- AWS CloudWatch first: Implement AWS CloudWatch collection first, add GCP Cloud Monitoring in a later iteration
- Global collection frequency: 15-minute polling for all instances (no per-instance override)
- Retries: Retry failed requests 3 times before marking as failed
- Collection source: Use the existing discovery service's provider registry to determine provider type
- Fallback: If metrics aren't available, fall back to using the instance's existing schedule

### Inactivity detection criteria
- Define "low activity": Near-zero activity (CPU < 1%, queries < 5/min)
- Activity types: Track CPU, queries, connections, and network separately
- Minimum data: 24+ hours of data required (accept brief gaps in data)
- Duration: Flexible duration with 8+ hours total low activity in a day
- Weekend handling: Same inactivity criteria regardless of day (no special weekend treatment)
- Always-low instances: Include all instances with consistently low activity

### Schedule candidate identification
- Idle periods: Suggest overnight (e.g., 10PM-7AM) and weekend sleep schedules
- Ranking: By savings potential (projected cost savings)
- Confidence: Rank by how consistently low the activity was during the period

### Data handling scope
- Retention: 14 days for raw metrics (enough to detect patterns, save storage)
- Aggregation: Per-instance hourly averages (store one hourly average per instance)

### OpenCode's Discretion
- Exact thresholds for "low activity" (CPU percentages, query counts)
- Database schema for storing collected metrics
- Specific API calls for CloudWatch/Cloud Monitoring
- Aggregation algorithms for hourly summaries
- Error handling details for rate limiting

</decisions>

<specifics>
## Specific Ideas

- AWS CloudWatch metrics to collect: CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS, ReadLatency, WriteLatency
- GCP Cloud Monitoring metrics: cpu_utilization, database_connection_count, read_io_rate, write_io_rate
- Activity detection: Look for periods where activity drops significantly compared to the instance's typical pattern
- Schedule suggestions: Present as cards on dashboard with estimated savings
- Historical fallback: If no metrics available, use the instance's existing sleep/wake schedule

</specifics>

<deferred>
## Deferred Ideas

- GCP Cloud Monitoring collection — add in a later phase after AWS is validated
- User feedback on suggestion accuracy — Phase 7

</deferred>

---

*Phase: 05-activity-analysis*
*Context gathered: 2026-02-23*
