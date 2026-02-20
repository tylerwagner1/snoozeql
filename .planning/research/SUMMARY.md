# Project Research Summary

**Project:** SnoozeQL
**Domain:** Database Sleep Scheduling (Multi-Cloud RDS/Cloud SQL Management)
**Researched:** 2026-02-20
**Confidence:** HIGH

## Executive Summary

SnoozeQL is a database sleep scheduling system that automates the stopping and starting of AWS RDS and GCP Cloud SQL instances to reduce cloud costs during inactive periods. The existing codebase provides a strong foundation with provider abstraction, instance discovery, and a well-structured Go backend with React frontend. Research confirms this architecture aligns with best practices for infrastructure automation tools.

The recommended approach is to extend the existing architecture rather than rebuild. Key additions needed are: CloudWatch/Cloud Monitoring integration for activity metrics, robfig/cron for timezone-aware scheduling, and an event system for audit trails and savings calculation. The most critical differentiator is **intelligent schedule recommendations** based on activity analysis — competitors like AWS Instance Scheduler provide only manual configuration with no usage insights.

The primary risks are platform constraints (AWS 7-day auto-restart, instances with read replicas cannot be stopped) and user expectations around cost savings (storage charges continue while stopped). These require proactive handling: the 7-day restart must be addressed in Phase 1 core scheduling, and storage cost communication must be clear in all UI surfaces. Variable startup times (5-45 minutes) also impact user experience and must be accounted for in schedule recommendations.

## Key Findings

### Recommended Stack

The existing stack (Go 1.24.0, React 18.2, PostgreSQL, Chi router) is solid and appropriate. Key additions are CloudWatch SDK for AWS activity monitoring, Cloud Monitoring SDK for GCP, and robfig/cron for schedule execution. Frontend would benefit from TanStack Query for data fetching and date-fns for timezone handling.

**Core technologies:**
- **Go 1.24.0 + Chi v5.2.5**: Already in use — excellent for concurrent cloud API calls, lightweight HTTP routing
- **pgx v5.8.0**: Best-in-class PostgreSQL driver with connection pooling and JSONB support
- **robfig/cron v3.0.1**: Standard Go cron library — timezone-aware, panic recovery, skip-if-running options
- **aws-sdk-go-v2/service/cloudwatch**: NEW — required for DatabaseConnections metrics to detect activity
- **cloud.google.com/go/monitoring v1.24.3**: NEW — Cloud Monitoring API for GCP activity metrics
- **TanStack Query 5.x**: Recommended for frontend — handles caching, refetching, optimistic updates

### Expected Features

**Must have (table stakes):**
- Manual start/stop instances — core value proposition
- Instance discovery & listing — foundation for all features
- Multi-account support — enterprises require this from day one
- Basic schedule creation with time windows — core automation
- Regex-based schedule assignment — user-requested flexible matching
- Instance status visibility — users need current state
- Operation history/audit log — know what happened and when

**Should have (competitive differentiators):**
- Intelligent schedule recommendations — key differentiator vs. AWS Instance Scheduler
- Activity-based insights visualization — builds user confidence in recommendations
- Cost savings tracking — ROI visibility justifies the tool
- Override support — temporary exceptions for maintenance windows

**Defer (v2+):**
- Multi-user RBAC — authentication complexity out of POC scope
- Email/SMS notifications — infrastructure overhead for POC
- Holiday calendar integration — timezone complexity, edge cases
- Complex recurrence patterns (every 3rd Tuesday) — simple weekly patterns sufficient

### Architecture Approach

The system follows a **control plane + worker pattern** common in infrastructure automation. The existing service-oriented architecture with provider abstraction is well-suited. Key patterns: Provider Interface (abstracts cloud operations), Event Sourcing (audit trail + savings calculation), Selector-Based Matching (schedules match instances dynamically via regex/tags), Background Workers (polling loops for discovery and scheduling).

**Major components:**
1. **Discovery Service** — Polls cloud providers, syncs instances to PostgreSQL, detects tag changes
2. **Scheduler Service** — Evaluates cron expressions, matches instances to schedules, executes start/stop via providers
3. **Analyzer Service** — Collects metrics, detects inactivity patterns, generates recommendations (NEW)
4. **Provider Registry** — Abstracts AWS/GCP operations behind common interface (exists, extend for metrics)
5. **Savings Calculator** — Tracks stop/start events, calculates cost savings (NEW)

### Critical Pitfalls

1. **AWS 7-Day Auto-Restart** — AWS forces restart after 7 days stopped. Must implement "re-stop" mechanism that monitors for auto-restart events and immediately stops again. **Address in Phase 1.**

2. **Instance State Race Conditions** — Stop/start operations are async (minutes to hours). Must implement proper state machine: only stop if `available`, only start if `stopped`, handle transitional states as in-progress. **Address in Phase 1.**

3. **Read Replicas Cannot Be Stopped** — Instances with read replicas cannot be stopped (AWS/GCP constraint). Discovery must flag these as "unsleepable" with clear UI explanation. **Address in Phase 1.**

4. **Storage Charges Continue** — Users expect $0 cost when sleeping. In reality: storage, backups, IPs still charged. UI must clearly communicate "saves compute costs, storage charges continue." **Address in Phase 1 UI.**

5. **Variable Start Time** — Startup ranges from minutes to hours. Cannot schedule wake at 8:55 AM for 9:00 AM meeting. Track historical startup times, add warm-up buffer (30-60 min). **Address in Phase 2.**

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Core Scheduling Foundation
**Rationale:** Must establish reliable start/stop operations before building intelligence on top. 7-day auto-restart and state machine are fundamental — broken core undermines everything.
**Delivers:** Working sleep/wake operations, schedule creation, instance management
**Addresses:** Manual sleep/wake, instance status, basic scheduling, regex assignment, operation history
**Avoids:** 7-day auto-restart (implement re-stop), state race conditions (proper state machine), read replica constraint (flag in discovery), storage charge confusion (clear UI messaging)

### Phase 2: Schedule Enhancement & Conflict Detection
**Rationale:** Once core scheduling works, add robustness features. Variable start times and regex conflicts are user-impacting issues that need addressing before recommendations.
**Delivers:** Improved scheduling reliability, conflict detection, startup time tracking
**Uses:** robfig/cron for proper cron evaluation, event system for tracking
**Implements:** Schedule conflict detection, startup time estimation, override system

### Phase 3: Activity Analysis & Metrics
**Rationale:** Metrics collection must precede recommendations — analyzer depends on having historical data. This phase builds the intelligence foundation.
**Delivers:** CloudWatch/Cloud Monitoring integration, activity dashboards, metrics storage
**Uses:** aws-sdk-go-v2/cloudwatch, cloud.google.com/go/monitoring
**Implements:** Metrics collector, activity visualization, pattern detection groundwork
**Avoids:** Metrics lag issues (use conservative thresholds, require sustained low activity)

### Phase 4: Intelligent Recommendations
**Rationale:** Key differentiator but requires metrics data accumulated in Phase 3. This is where SnoozeQL exceeds competitors.
**Delivers:** Schedule recommendations based on usage patterns, confidence scoring
**Uses:** Accumulated metrics, pattern detection algorithms
**Implements:** Analyzer service, recommendation generation, user confirmation workflow
**Avoids:** Minimum billing increment trap (filter windows below cost-effective threshold)

### Phase 5: Savings & Reporting
**Rationale:** Depends on events flowing from Phase 1-2 and metrics from Phase 3. ROI reporting validates the tool's value.
**Delivers:** Cost savings dashboard, estimated vs. actual savings, per-instance reporting
**Uses:** Event history, instance specs, pricing data
**Implements:** Savings calculator, aggregation queries, dashboard charts

### Phase Ordering Rationale

- **Dependency chain:** Discovery → Scheduling → Events → Metrics → Analysis → Recommendations → Savings
- **Risk mitigation:** Phase 1 addresses 4 of 5 critical pitfalls — get these right before adding complexity
- **Value delivery:** Each phase delivers usable functionality. Phase 1 alone provides value (manual control + basic scheduling). Recommendations (Phase 4) are the differentiator but require foundation.
- **Architecture alignment:** Follows build order from ARCHITECTURE.md — Events first (dependency for savings), Metrics before Analysis, Scheduler enhancement after core works

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3 (Activity Analysis):** CloudWatch/Cloud Monitoring API specifics, metrics granularity, optimal collection intervals
- **Phase 4 (Recommendations):** Pattern detection algorithms, confidence scoring methodology, threshold tuning

Phases with standard patterns (skip research-phase):
- **Phase 1 (Core Scheduling):** Well-documented AWS/GCP APIs, existing codebase provides patterns
- **Phase 2 (Schedule Enhancement):** robfig/cron is well-documented, standard conflict detection patterns
- **Phase 5 (Savings):** Straightforward calculation from events + pricing, established patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Versions verified via pkg.go.dev, existing codebase validates core stack |
| Features | HIGH | Competitor analysis from official sources, clear table stakes identified |
| Architecture | HIGH | Based on existing codebase analysis + infrastructure automation best practices |
| Pitfalls | HIGH | All critical pitfalls sourced from official AWS/GCP documentation |

**Overall confidence:** HIGH

### Gaps to Address

- **GCP Cloud SQL scheduling details:** Provider scaffolded but not fully implemented — needs completion in Phase 1
- **Pricing data source:** Savings calculation needs instance pricing lookup — consider AWS Pricing API or static lookup table
- **Multi-instance deployment:** Research focused on single-instance POC; distributed locking needed for HA deployment (v2 consideration)
- **Startup time prediction:** No existing data — must collect empirically before recommendations can account for it

## Sources

### Primary (HIGH confidence)
- AWS RDS Stop/Start Documentation — https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_StopInstance.html
- AWS CloudWatch RDS Metrics — https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/rds-metricscollected.html
- GCP Cloud SQL Start/Stop — https://cloud.google.com/sql/docs/mysql/start-stop-restart-instance
- GCP Cloud Monitoring — https://cloud.google.com/sql/docs/mysql/monitor-instance
- pkg.go.dev — robfig/cron v3.0.1, aws-sdk-go-v2 v1.41.1, cloud.google.com/go/monitoring v1.24.3

### Secondary (MEDIUM confidence)
- AWS Instance Scheduler solution — https://aws.amazon.com/solutions/implementations/instance-scheduler-on-aws/
- Go project layout conventions — https://github.com/golang-standards/project-layout
- TanStack Query documentation — https://tanstack.com/query/latest

### Tertiary (LOW confidence)
- Startup time estimates (5-45 minutes) — based on AWS documentation language ("minutes to hours") rather than empirical data

---
*Research completed: 2026-02-20*
*Ready for roadmap: yes*
