# Pitfalls Research

**Domain:** Database Sleep Scheduling (AWS RDS / GCP Cloud SQL)
**Researched:** 2026-02-20
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: AWS RDS 7-Day Auto-Restart Limit

**What goes wrong:**
AWS RDS automatically restarts any stopped instance after 7 consecutive days to ensure maintenance updates are applied. If your scheduling system doesn't account for this, instances will wake up unexpectedly, incurring charges and potentially confusing users who expect them to remain stopped.

**Why it happens:**
AWS enforces this to prevent instances from falling behind on required maintenance patches. This is a hard platform limitation that cannot be disabled.

**How to avoid:**
- Implement a "re-stop" mechanism that monitors for auto-restart events and immediately stops the instance again after AWS restarts it
- Track the 7-day timer and schedule a brief wake-start cycle (e.g., 90 minutes for maintenance) before AWS forces the restart
- Use EventBridge or CloudWatch Events to detect instance state changes and trigger re-stop operations
- Store "intended state" in your database and reconcile actual state periodically

**Warning signs:**
- Instances appearing "available" when schedules say they should be "stopped"
- Unexpected cost spikes on instances meant to be sleeping
- Users reporting databases are "awake" when they shouldn't be

**Phase to address:**
Phase 1 (Core Scheduling) — This is fundamental to the value proposition. Build the re-stop mechanism from day one.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html (AWS Official)
- https://repost.aws/knowledge-center/rds-stop-seven-days-step-functions (AWS re:Post)

---

### Pitfall 2: Instance State Race Conditions

**What goes wrong:**
Multiple operations conflict when the instance is in a transitional state (`stopping`, `starting`, `modifying`). Attempting to stop an instance that's already stopping, or start one that's already starting, causes API errors and potentially corrupted state tracking.

**Why it happens:**
Stop/start operations are asynchronous and take variable time (minutes to hours depending on instance size, workload, and database state). Code that doesn't properly poll for completion before taking action creates race conditions.

**How to avoid:**
- Always check instance state before issuing commands
- Implement proper state machine: only issue `stop` if state is `available`, only issue `start` if state is `stopped`
- Use polling with backoff to wait for operations to complete
- Handle `stopping` and `starting` states as "in-progress" — don't interrupt
- For GCP: Use operation tracking to await completion before issuing new commands

**Warning signs:**
- Frequent API errors like `InvalidDBInstanceState`
- Operations getting "stuck" in transitional states
- Inconsistency between your tracked state and actual cloud state

**Phase to address:**
Phase 1 (Core Scheduling) — State machine must be robust from the start.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html (AWS Official)
- https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance (GCP Official)

---

### Pitfall 3: Instances with Read Replicas Cannot Be Stopped

**What goes wrong:**
AWS RDS prevents stopping an instance that has read replicas, and also prevents stopping read replicas themselves. GCP Cloud SQL has similar constraints. Scheduling systems that don't detect this configuration will repeatedly fail when trying to sleep these instances.

**Why it happens:**
Stopping a primary breaks replication. The cloud providers enforce this to prevent data consistency issues.

**How to avoid:**
- During instance discovery, flag instances that have read replicas as "unsleepable"
- Check `ReadReplicaDBInstanceIdentifiers` (AWS) or equivalent before scheduling sleep
- Provide clear UI indication: "This instance cannot be stopped because it has X read replicas"
- Offer a "cascade" option that first deletes/stops replicas, then stops primary (with heavy warnings)

**Warning signs:**
- Repeated failures for specific instances
- `InvalidDBInstanceState` errors with messages about read replicas
- User confusion about why certain instances "won't sleep"

**Phase to address:**
Phase 1 (Instance Discovery) — Must detect and flag these instances during discovery.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html#USER_StopInstance.Limitations (AWS Official)

---

### Pitfall 4: Variable Start Time Causes Scheduling Misses

**What goes wrong:**
Start times are unpredictable — ranging from minutes to hours depending on instance size, database size, recovery requirements, and cloud provider load. Scheduling a wake at 8:55 AM for a 9:00 AM meeting doesn't work if startup takes 45 minutes.

**Why it happens:**
AWS documentation explicitly warns: "Starting a DB instance requires instance recovery and can take from minutes to hours." Large databases or unclean shutdowns require longer recovery.

**How to avoid:**
- Track historical startup times per instance and use as prediction baseline
- Add configurable "warm-up buffer" (default 30-60 minutes before needed)
- Show estimated wake time in UI: "Will be ready by approximately X:XX"
- For recommendations, factor in startup time when suggesting sleep windows

**Warning signs:**
- Users complaining databases aren't ready when expected
- Application connection errors during supposed "up" periods
- Instances in `starting` state during scheduled "available" windows

**Phase to address:**
Phase 2 (Schedule Enhancement) — Track metrics and improve predictions over time.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StartInstance.html (AWS Official)

---

### Pitfall 5: Storage Charges Continue While Stopped

**What goes wrong:**
Users expect $0 cost when sleeping databases. In reality, they're still charged for:
- Provisioned storage (including Provisioned IOPS)
- Backup storage (manual snapshots + automated backups within retention window)
- Reserved IP addresses (varies by provider)

This leads to user confusion, disappointment, and support requests.

**Why it happens:**
Cloud providers only waive compute (instance hours) charges. Storage is "attached" to the stopped instance and continues to incur costs.

**How to avoid:**
- Clearly communicate in UI: "Sleeping saves compute costs. Storage charges continue."
- Show estimated savings vs. full running cost (not vs. $0)
- In recommendations, calculate actual savings based on instance type pricing, not storage
- Consider adding storage size to the cost calculation: larger storage = less relative savings

**Warning signs:**
- Users expecting bigger savings than delivered
- Support requests about "why am I still being charged?"
- Confusion in savings reports

**Phase to address:**
Phase 1 (UI) — Set expectations correctly from the start.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html (AWS Official: "While your DB instance is stopped, you are charged for provisioned storage")
- https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance (GCP Official: "Stopping an instance suspends instance charges. Charges for storage and IP addresses continue.")

---

### Pitfall 6: Public IP Address Changes After Restart (AWS)

**What goes wrong:**
AWS RDS releases the public IP address when an instance stops and assigns a new one on restart. Applications hardcoded to IP addresses (instead of DNS endpoints) break after sleep/wake cycles.

**Why it happens:**
AWS reclaims IP resources while instances are stopped.

**How to avoid:**
- During onboarding, warn users to use DNS endpoints, not IP addresses
- Consider adding a "connection string" validator that flags IP-based connections
- Document this limitation prominently
- For GCP: Private IPs are retained but public IPs may also change

**Warning signs:**
- Applications failing to connect after wake
- Users reporting "it was working before we started using SnoozeQL"
- Connection timeout errors with IP addresses in logs

**Phase to address:**
Phase 1 (Documentation/Onboarding) — Warn users during setup.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html#USER_StopInstance.PublicIPAddress (AWS Official)

---

### Pitfall 7: Multi-AZ SQL Server Cannot Be Stopped

**What goes wrong:**
RDS for SQL Server in Multi-AZ deployment cannot be stopped at all. Scheduling systems that don't detect this will fail repeatedly for these instances.

**Why it happens:**
SQL Server Multi-AZ has architectural constraints that prevent the stop operation.

**How to avoid:**
- During discovery, check engine type and Multi-AZ flag
- Mark SQL Server Multi-AZ instances as "unsleepable" with clear explanation
- Filter these out of recommendation candidates

**Warning signs:**
- Repeated API failures for SQL Server instances
- Error messages about unsupported operations

**Phase to address:**
Phase 1 (Instance Discovery) — Detect during discovery and flag appropriately.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html#USER_StopInstance.MAZ (AWS Official)

---

### Pitfall 8: Activity Metrics Lag and Granularity Limitations

**What goes wrong:**
CloudWatch and Cloud Monitoring metrics have delays (typically 1-5 minutes) and granularity limitations (often 1-minute or 5-minute periods). Recommendations based on "zero activity" might miss brief activity spikes, leading to databases being slept while actually in use.

**Why it happens:**
Cloud metrics are aggregated and not real-time. A 30-second burst of activity might not register as significant in a 5-minute aggregation period.

**How to avoid:**
- Use conservative thresholds: "low activity" not "zero activity"
- Require sustained low activity over multiple sample periods (e.g., 30 minutes of low activity, not just one sample)
- Provide manual override capability for users who know their patterns
- Consider combining multiple metrics (connections, CPU, I/O) rather than relying on one

**Warning signs:**
- Users reporting database was stopped "while I was using it"
- Complaints about recommendations during active periods
- Brief connection spikes visible in detailed logs but not in aggregated metrics

**Phase to address:**
Phase 3 (Recommendations) — Use conservative analysis algorithms.

**Sources:**
- https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/rds-metricscollected.html (AWS Official)
- https://cloud.google.com/sql/docs/mysql/monitor-instance (GCP Official)

---

### Pitfall 9: Regex Schedule Assignment Conflicts

**What goes wrong:**
Multiple schedules with overlapping regex patterns match the same instance. Without clear precedence rules, instances get unpredictable behavior — woken by one schedule, stopped by another, creating thrashing.

**Why it happens:**
Flexible regex assignment is powerful but creates ambiguity. `prod-.*` and `.*-mysql` both match `prod-api-mysql`.

**How to avoid:**
- Implement clear precedence rules (e.g., most specific wins, or first match wins)
- Show UI warning when creating schedule with patterns that overlap existing schedules
- Display which schedule applies to each instance in the instance list
- Consider "exclusive assignment" mode where each instance can only belong to one schedule

**Warning signs:**
- Instances rapidly cycling between states
- User confusion about which schedule applies
- Conflicting events in audit log

**Phase to address:**
Phase 2 (Schedule Management) — Build conflict detection into schedule creation.

---

### Pitfall 10: Minimum Billing Increment on Start

**What goes wrong:**
AWS charges a minimum of 10 minutes per start, regardless of how long the instance runs. Frequent start/stop cycles (e.g., stopping and starting every hour) can actually cost more than leaving the instance running.

**Why it happens:**
Cloud providers have minimum billing units to cover provisioning costs.

**How to avoid:**
- Calculate break-even point: how long must instance sleep to justify start cost?
- Warn users if proposed sleep window is too short to be cost-effective
- In recommendations, filter out windows shorter than cost-effective threshold
- Track actual cost savings vs. estimated to validate approach

**Warning signs:**
- Savings reports showing minimal or negative savings
- High number of start events relative to sleep hours
- Users gaming the system with very short windows

**Phase to address:**
Phase 3 (Recommendations) — Factor minimum billing into cost calculations.

**Sources:**
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StartInstance.html (AWS Official: "10-minute minimum charge")

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Storing instance state in scheduler memory only | Simpler implementation | State lost on restart, inconsistency after crashes | Never — always persist intended state |
| Polling instance state on fixed interval without jitter | Predictable behavior | API rate limiting, thundering herd on many instances | Only in single-instance POC |
| Assuming sync operations (fire-and-forget) | Simpler code | Race conditions, failed operations not detected | Never for production scheduling |
| Hardcoding retry counts and backoff | Works initially | Fails under load or provider throttling | Only for POC, must be configurable |
| Single provider implementation with provider-specific code | Faster initial development | Expensive refactor when adding second provider | Acceptable for hackathon POC with clear abstraction boundaries |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| AWS RDS API | Not handling `InvalidDBInstanceState` errors | Check state before operation, retry with backoff on transient errors |
| AWS CloudWatch | Requesting metrics for stopped instances | Stopped instances don't emit metrics — handle gracefully |
| GCP Cloud SQL | Ignoring operation completion status | Poll operation until DONE before issuing next command |
| GCP Cloud Monitoring | Missing permissions for metrics.read | Ensure service account has monitoring.viewer role |
| Cross-account AWS | Using single set of credentials | Per-connection credential management, STS assume role for cross-account |
| Multi-project GCP | Single service account | Per-project service accounts or organization-level access |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Sequential instance polling | Slow discovery cycle | Parallel polling with concurrency limit | 50+ instances |
| Full instance sync on every poll | High API costs, rate limiting | Incremental sync, cache with TTL | 100+ instances |
| Unbounded metrics history storage | Database growth, slow queries | Rolling retention (30-90 days), aggregation | 30+ days of data |
| Real-time schedule evaluation | High CPU on schedule changes | Pre-compute next actions, invalidate on change | 50+ schedules |
| Per-instance CloudWatch queries | Rate limiting, slow recommendations | Batch metrics requests, use GetMetricData for multiple metrics | 20+ instances |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Storing cloud credentials in database | Credential theft if DB compromised | Use secrets manager, environment variables, or IAM roles |
| Overly broad IAM permissions | Attacker could delete instances | Minimum viable: rds:StopDBInstance, rds:StartDBInstance, rds:DescribeDBInstances |
| Single credential for all accounts | Breach affects all accounts | Separate credentials per cloud account/project |
| Logging instance identifiers without sanitization | Info leak in logs | Structured logging with controlled fields |
| No audit trail for sleep/wake operations | Cannot investigate incidents | Log all operations with timestamp, user, instance, result |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Stopping without confirmation | Accidental production database sleep | Require confirmation, especially for non-development instances |
| Silent schedule conflicts | Confusion about why instance is in wrong state | Show active schedule on instance detail, warn on conflict creation |
| "Savings" without context | Misleading expectations | Show "Estimated savings: $X/month (compute only, storage charges continue)" |
| Unclear wake time | Users arrive to find database still starting | Show "Starting... estimated ready: 8:47 AM" with progress |
| No emergency wake button | Panic when need database NOW | Prominent "Wake Now" button that bypasses schedule |
| Hiding recommendation confidence | Users accept weak recommendations | Show LOW/MEDIUM/HIGH confidence, explain why |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Sleep operation:** Often missing proper state validation — verify instance is actually `available` before stopping
- [ ] **Wake operation:** Often missing wait-for-ready — verify instance reaches `available` state, not just `starting`
- [ ] **Schedule execution:** Often missing retry on transient failures — verify failed operations are retried with backoff
- [ ] **Discovery sync:** Often missing deleted instance handling — verify instances removed from cloud are removed from local state
- [ ] **Multi-account:** Often missing per-account error isolation — verify one account's API error doesn't block other accounts
- [ ] **Recommendations:** Often missing activity spike detection — verify recommendations exclude instances with sporadic high-activity bursts
- [ ] **Cost savings:** Often missing storage cost caveat — verify UI shows that storage charges continue

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Instance stuck in wrong state | LOW | Reconcile actual cloud state, update local tracking, clear pending operations |
| Schedule conflict causing thrashing | MEDIUM | Pause all schedules, identify conflicts, resolve precedence, resume one-by-one |
| Credentials expired/revoked | LOW | Update credentials, re-validate, resume operations |
| 7-day auto-restart missed | LOW | Detect restart, trigger re-stop, log event for audit |
| Recommendation stopped active instance | HIGH | Implement manual override flag, contact user, add to exclusion list |
| API rate limiting | MEDIUM | Increase backoff, reduce polling frequency, batch operations |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 7-day auto-restart | Phase 1 (Core Scheduling) | Test by stopping instance for 8+ days in non-prod |
| State race conditions | Phase 1 (Core Scheduling) | Load test with concurrent operations on same instance |
| Read replica constraint | Phase 1 (Discovery) | Discovery UI shows "unsleepable" flag with reason |
| Variable start time | Phase 2 (Schedule Enhancement) | Track actual vs. estimated times, alert on large variance |
| Storage charge confusion | Phase 1 (UI) | UI review confirms savings messaging includes storage caveat |
| Public IP changes | Phase 1 (Documentation) | Onboarding flow includes DNS usage warning |
| Multi-AZ SQL Server | Phase 1 (Discovery) | Discovery correctly identifies and flags these instances |
| Metrics lag | Phase 3 (Recommendations) | Recommendations require sustained low activity, not single sample |
| Regex conflicts | Phase 2 (Schedule Management) | Schedule creation shows warning on overlap |
| Minimum billing | Phase 3 (Recommendations) | Recommendations filter out windows below cost-effective threshold |

## Sources

- AWS RDS Stop Instance Documentation: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html
- AWS RDS Start Instance Documentation: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StartInstance.html
- AWS re:Post - Stop RDS for 7+ Days: https://repost.aws/knowledge-center/rds-stop-seven-days-step-functions
- GCP Cloud SQL Start/Stop Documentation: https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance
- GCP Cloud SQL Monitoring: https://cloud.google.com/sql/docs/mysql/monitor-instance
- AWS CloudWatch Metrics Overview: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/rds-metricscollected.html

---
*Pitfalls research for: Database Sleep Scheduling (SnoozeQL)*
*Researched: 2026-02-20*
