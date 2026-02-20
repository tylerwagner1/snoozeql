# Feature Research

**Domain:** Database Sleep Scheduling / Instance Lifecycle Management
**Researched:** 2026-02-20
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Manual start/stop instances | Core value proposition - users need immediate control | LOW | AWS RDS API: `StartDBInstance`/`StopDBInstance`, GCP: `PATCH` with `activationPolicy` |
| Instance discovery & listing | Users need to see what they're managing | LOW | Already implemented in existing codebase via discovery service |
| Multi-account/project support | Enterprises have multiple cloud accounts | MEDIUM | Already modeled in existing codebase (CloudAccount entity) |
| Schedule creation with time windows | Users define when instances sleep/wake | MEDIUM | Start time, end time, days of week - standard scheduling UX |
| Instance filtering/search | Can't manage 100+ instances without search | LOW | Filter by name, tags, provider, status |
| Instance status visibility | Must show current state (running/stopped/pending) | LOW | Poll provider APIs for current state |
| Operation history/audit log | Users need to know what happened and when | LOW | Already modeled (Event entity in existing codebase) |
| Error handling & retry | Cloud API calls fail; system must be resilient | MEDIUM | Exponential backoff, dead letter handling |
| Multiple cloud provider support | RDS + Cloud SQL in same dashboard | MEDIUM | Already architected with provider abstraction |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Intelligent schedule recommendations | Analyze usage patterns, suggest optimal sleep windows | HIGH | Key differentiator vs. AWS Instance Scheduler which has no recommendations |
| Regex-based schedule assignment | Flexible matching: `prod-*` or `*-staging-*` patterns | MEDIUM | User explicitly requested this feature |
| Activity-based insights | Show actual usage patterns visually (CloudWatch/Cloud Monitoring) | MEDIUM | Helps justify recommendations; builds user confidence |
| Bulk operations UI | Select multiple instances, apply schedule in one action | LOW | UX efficiency for managing large fleets |
| Unified multi-cloud dashboard | Single pane of glass for AWS + GCP | LOW | Competitors often single-cloud focused |
| Schedule conflict detection | Warn when schedules overlap or conflict | MEDIUM | Prevents user errors |
| Cost savings tracking | Show estimated/actual savings from sleep time | MEDIUM | ROI visibility; justifies tool's value |
| Override support | Temporary "keep awake" for maintenance windows | LOW | Already modeled (Override entity in existing codebase) |
| Tag-based filtering | Use existing AWS/GCP tags for schedule assignment | LOW | Leverages existing organizational patterns |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Real-time wake-on-connect | "Just wake up when someone connects" | Databases take 5-15 minutes to start; connection timeout long gone. Complex DNS/proxy infrastructure needed. | Manual wake with status notification; schedule wake 15 min before expected use |
| Auto-scaling instead of sleep | "Scale down to zero" | RDS/Cloud SQL don't support scale-to-zero; minimum instance size still costs money. Aurora Serverless exists but is different product. | Stop instance completely for true $0 cost |
| Email/SMS notifications | "Tell me when things happen" | Adds infrastructure complexity (SES/SNS/SMTP). POC scope creep. | In-app notification/event log; defer notifications to v2 |
| Multi-user RBAC | "Different team permissions" | Authentication/authorization complexity; session management. Out of scope for POC. | Single-user POC; add auth in later phase |
| Holiday calendar integration | "Skip schedules on holidays" | Timezone complexity; regional holiday databases; edge cases. | Manual override for holidays; simple date exclusion list |
| Complex recurrence patterns | "Every 3rd Tuesday" or "First Monday of month" | cron complexity; testing nightmare; user confusion | Simple weekly patterns; daily on/off times |
| Billing integration | "Show actual cost savings from billing data" | Requires billing API access (Cost Explorer, BigQuery export); additional auth complexity | Estimate savings from instance specs + uptime hours |
| Auto-applying recommendations | "Just do it automatically" | Risk of production outages; user needs confirmation | Show recommendations, require explicit user action |
| Terraform/IaC export | "Export schedules as code" | Scope creep; different users have different IaC tools | API-first design; IaC integration can be built later |
| Database query analysis | "Find truly idle databases by query volume" | Requires DB-level access (not just cloud APIs); permission complexity | Use CloudWatch/Cloud Monitoring metrics for activity |

## Feature Dependencies

```
[Instance Discovery]
    └──requires──> [Multi-account Configuration]
                       └──requires──> [Provider Credentials]

[Schedule Assignment]
    └──requires──> [Instance Discovery]
    └──requires──> [Schedule Creation]

[Intelligent Recommendations]
    └──requires──> [Activity Metrics Ingestion]
                       └──requires──> [Instance Discovery]

[Manual Sleep/Wake]
    └──requires──> [Instance Discovery]
    └──requires──> [Provider Stop/Start APIs]

[Cost Savings Display]
    └──requires──> [Instance Discovery] (for specs)
    └──requires──> [Operation History] (for uptime tracking)

[Bulk Operations]
    └──enhances──> [Manual Sleep/Wake]
    └──enhances──> [Schedule Assignment]

[Schedule Recommendations] ──conflicts──> [Auto-applying] (safety)
```

### Dependency Notes

- **Schedule Assignment requires Instance Discovery:** Can't assign schedules to instances you haven't discovered
- **Recommendations require Activity Metrics:** Need historical data to identify patterns
- **Cost Savings requires Operation History:** Need to track actual start/stop times to calculate savings
- **Recommendations conflict with Auto-apply:** Require explicit user confirmation for safety

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] **Instance Discovery** — Foundation; users must see their databases
- [x] **Multi-account Support** — Enterprises need this from day one
- [ ] **Manual Sleep/Wake** — Core value; immediate user control
- [ ] **Instance Status Display** — Users need to see current state
- [ ] **Basic Schedule Creation** — Define start/end times for sleep
- [ ] **Regex-based Schedule Assignment** — User's requested flexible matching
- [ ] **Operation History** — Know what happened and when

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] **Activity Metrics Ingestion** — Trigger: users want to see usage patterns
- [ ] **Intelligent Recommendations** — Trigger: enough activity data collected (1+ week)
- [ ] **Cost Savings Display** — Trigger: users asking "how much did I save?"
- [ ] **Bulk Operations** — Trigger: users managing 10+ instances
- [ ] **Schedule Conflict Detection** — Trigger: user errors with overlapping schedules
- [ ] **Override Support** — Trigger: users need temporary exceptions

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **Holiday calendar** — Complexity outweighs value for POC
- [ ] **Email notifications** — Infrastructure overhead for POC
- [ ] **Multi-user auth** — Out of scope per PROJECT.md
- [ ] **Billing API integration** — Estimation sufficient for now
- [ ] **Complex recurrence** — Weekly patterns sufficient for most cases

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Instance discovery | HIGH | LOW (exists) | P1 |
| Manual sleep/wake | HIGH | LOW | P1 |
| Instance status | HIGH | LOW | P1 |
| Basic scheduling | HIGH | MEDIUM | P1 |
| Regex assignment | HIGH | MEDIUM | P1 |
| Operation history | MEDIUM | LOW (exists) | P1 |
| Activity metrics | HIGH | MEDIUM | P2 |
| Recommendations | HIGH | HIGH | P2 |
| Cost savings display | MEDIUM | LOW | P2 |
| Bulk operations | MEDIUM | LOW | P2 |
| Override support | MEDIUM | LOW | P2 |
| Schedule conflicts | LOW | MEDIUM | P3 |
| Tag-based filtering | MEDIUM | LOW | P3 |

**Priority key:**
- P1: Must have for launch (validates core concept)
- P2: Should have, add when core is stable
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | AWS Instance Scheduler | GCP Cloud Scheduler | IBM Cloudability | SnoozeQL |
|---------|------------------------|---------------------|------------------|----------|
| Start/stop scheduling | ✓ Tag-based | ✓ HTTP/Pub-Sub triggers | ✗ Cost visibility only | ✓ Direct API |
| RDS support | ✓ Full | ✗ (needs custom Cloud Function) | ✗ | ✓ Full |
| Cloud SQL support | ✗ AWS only | ✓ Indirect | ✗ | ✓ Full |
| Multi-cloud | ✗ | ✗ | ✓ | ✓ |
| Recommendations | ✗ | ✗ | ✓ (optimization) | ✓ |
| Activity analysis | ✗ | ✗ | ✓ (spend) | ✓ (metrics) |
| Regex matching | ✗ (exact tags) | ✗ | N/A | ✓ |
| UI dashboard | ✗ (CLI/DynamoDB) | ✗ (Console) | ✓ | ✓ |
| Setup complexity | HIGH (CloudFormation) | MEDIUM | LOW (SaaS) | LOW |
| Cross-account | ✓ (with remote stacks) | ✗ | ✓ | ✓ |

### Our Differentiating Approach

**AWS Instance Scheduler** is infrastructure-focused (CloudFormation, DynamoDB, Lambda). No UI, no recommendations, AWS-only. Complex setup requiring tagging discipline.

**GCP Cloud Scheduler** is a generic job scheduler, not purpose-built for databases. Requires custom Cloud Functions to actually stop/start instances.

**IBM Cloudability** and similar FinOps tools focus on cost visibility and optimization recommendations, but don't provide direct instance control or scheduling.

**SnoozeQL's approach:**
1. **Purpose-built for database sleep** — Not generic compute scheduling
2. **Multi-cloud first** — AWS + GCP in unified interface
3. **Intelligent by default** — Recommendations, not just scheduling
4. **Low setup friction** — Connect accounts, see instances, create schedules
5. **Regex-based flexibility** — Match naming conventions, not just exact tags

## Sources

**AWS Instance Scheduler (HIGH confidence - official):**
- https://aws.amazon.com/solutions/implementations/instance-scheduler-on-aws/
- https://github.com/aws-solutions/instance-scheduler-on-aws
- https://docs.aws.amazon.com/solutions/latest/instance-scheduler-on-aws/operator-guide.html

**GCP Cloud SQL (HIGH confidence - official):**
- https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance
- https://cloud.google.com/scheduler/docs/overview

**GCP Cloud Scheduler (HIGH confidence - official):**
- https://cloud.google.com/scheduler/docs/overview

**FinOps/Cost Management (MEDIUM confidence - vendor sites):**
- https://www.apptio.com/products/cloudability/
- https://www.flexera.com/
- https://www.infracost.io/

---
*Feature research for: Database Sleep Scheduling*
*Researched: 2026-02-20*
