# Stack Research

**Domain:** Database sleep scheduling (multi-cloud RDS/Cloud SQL management)
**Researched:** 2026-02-20
**Confidence:** HIGH

## Executive Summary

This research validates and extends the existing SnoozeQL technology stack for a database sleep scheduling application. The current stack (Go 1.24.0, React 18.2, PostgreSQL, Chi router) is solid and appropriate for this domain. Key additions needed are: CloudWatch metrics SDK for AWS activity monitoring, GCP Cloud Monitoring SDK for GCP activity, and robust cron scheduling for sleep/wake automation.

## Recommended Stack

### Core Backend Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.24.0 | Backend runtime | Already in use, excellent for concurrent cloud API calls. Modern Go with improved tooling. |
| Chi | v5.2.5 | HTTP router | Already at v5.2.0, minor update available. Lightweight, composable middleware, standard library compatible. |
| pgx | v5.8.0 | PostgreSQL driver | Already in use. Best-in-class Go PostgreSQL driver with connection pooling and JSONB support. |
| robfig/cron | v3.0.1 | Cron scheduling | Standard Go cron library with 5,150+ importers. Supports timezone-aware scheduling, job wrappers for recovery/logging. |

**Confidence:** HIGH - Verified via pkg.go.dev and GitHub releases on 2026-02-20.

### Cloud Provider SDKs

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| aws-sdk-go-v2 | v1.41.1 | AWS core SDK | Already in use. Official AWS SDK v2 with modular design. |
| aws-sdk-go-v2/service/rds | v1.116.0 | RDS operations | Already in use. Provides StartDBInstance, StopDBInstance APIs. |
| aws-sdk-go-v2/service/cloudwatch | v1.54.0 | Activity metrics | **NEW** - Required for GetMetricData to fetch DatabaseConnections, CPUUtilization metrics for activity detection. |
| google.golang.org/api | v0.267.0 | GCP core API | Already in use for Cloud SQL admin operations. |
| cloud.google.com/go/monitoring | v1.24.3 | GCP metrics | **NEW** - Cloud Monitoring API for fetching cloudsql.googleapis.com/database/network/connections and CPU metrics. |

**Confidence:** HIGH - Versions verified via pkg.go.dev on 2026-02-20.

### Frontend Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| React | 18.2.0 | UI framework | Already in use. Stable, widely adopted. React 19.x available but 18.2 is mature and well-tested. |
| Vite | 5.0.8 → 7.3.1 | Build tool | Current version outdated. **Consider upgrade** to Vite 7.x for faster builds and improved DX. |
| TypeScript | 5.3.3 → 5.9.3 | Type safety | Current version works but 5.9 has improved inference. Optional upgrade. |
| React Router DOM | 6.20.0 → 7.13.0 | Routing | Major version available with data loading improvements. **Consider upgrade** for better async support. |
| Tailwind CSS | 3.4.0 → 4.2.0 | Styling | Tailwind v4 is production-ready with improved performance. **Optional upgrade**. |

**Confidence:** MEDIUM - React 18.2 is stable; upgrades are optional improvements.

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| TanStack Query | 5.90.21 | Server state management | **NEW** - Recommended for frontend API calls. Handles caching, refetching, optimistic updates. |
| react-hook-form | 7.71.1 | Form handling | For schedule creation forms with validation. |
| zod | 4.3.6 | Schema validation | Frontend/backend validation schemas. Already using JSON for API, Zod provides type-safe parsing. |
| date-fns | 4.1.0 | Date manipulation | Schedule time calculations, timezone handling. Smaller than moment.js, tree-shakeable. |
| Recharts | 2.10.0 | Charting | Already in use for dashboard. Continue using for activity visualization. |

**Confidence:** HIGH for TanStack Query (standard React data fetching); MEDIUM for others (optional but recommended).

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| golangci-lint | Go linting | Already configured in Makefile. Keep using for code quality. |
| ESLint 8.x | JS/TS linting | Already configured. Note: ESLint 9 has breaking config format. |
| Docker Compose | Local development | Already in use for PostgreSQL and multi-service setup. |
| Air | Hot reload | Consider adding for Go hot reload during development. |

## Installation

### Backend Dependencies (Go)

```bash
# Core (already present)
go get github.com/go-chi/chi/v5@v5.2.5
go get github.com/jackc/pgx/v5@v5.8.0

# AWS SDK (add CloudWatch)
go get github.com/aws/aws-sdk-go-v2/service/cloudwatch@latest

# GCP SDK (add Cloud Monitoring)
go get cloud.google.com/go/monitoring@latest

# Cron scheduling (new)
go get github.com/robfig/cron/v3@v3.0.1
```

### Frontend Dependencies (npm)

```bash
# Recommended additions
npm install @tanstack/react-query@^5
npm install date-fns@^4
npm install zod@^4
npm install react-hook-form@^7

# Optional upgrades (major version changes - test thoroughly)
npm install vite@^7 react-router-dom@^7
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| robfig/cron | go-co-op/gocron v2.19 | gocron is more feature-rich (distributed locks, event listeners) but adds complexity. Use for multi-instance deployments needing job coordination. |
| pgx | database/sql + lib/pq | Use lib/pq only if you need standard library compatibility. pgx is faster and more feature-rich. |
| TanStack Query | SWR | SWR is simpler but TanStack Query has better devtools and mutation support. Use SWR for simpler read-only data fetching. |
| date-fns | Day.js | Day.js has smaller bundle size but date-fns is more feature-complete for timezone handling. |
| React 18.2 | React 19.x | Upgrade to React 19 if you need the new `use()` hook or improved suspense. Not required for this project. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| aws-sdk-go v1 | Deprecated, worse performance, no module support | aws-sdk-go-v2 (already using) |
| moment.js | Large bundle size, deprecated | date-fns or Day.js |
| node-cron (frontend) | Wrong layer - scheduling belongs in backend | robfig/cron in Go backend |
| GraphQL | Overkill for this app - simple REST endpoints suffice | Chi REST API (already using) |
| Redis for job queues | Adds operational complexity for POC | In-memory cron scheduler is sufficient for single-instance POC |
| Kubernetes CronJobs | Infrastructure complexity - app handles scheduling | robfig/cron for in-app scheduling |

## Stack Patterns by Variant

**If deploying multi-instance (HA):**
- Add Redis for distributed locks on scheduled jobs
- Consider gocron v2 with distributed locking support
- Use pg advisory locks for singleton job execution

**If adding more cloud providers (Azure, etc.):**
- Maintain provider abstraction pattern already in place
- Add Azure SDK v2 for Azure SQL
- Keep metrics interface generic (GetMetrics returns map[string]any)

**If scaling to thousands of instances:**
- Add message queue (SQS/Pub/Sub) for async operations
- Consider worker pool pattern for concurrent stop/start
- Rate limit cloud API calls per provider

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| pgx v5.8.0 | PostgreSQL 12-16 | Tested with PG 15 in docker-compose |
| aws-sdk-go-v2 v1.41.1 | All AWS regions | Auto-detects regions from credentials |
| robfig/cron v3.0.1 | Go 1.11+ | Uses Go modules, timezone-aware |
| React 18.2 | Vite 5.x-7.x | Vite 7 requires minor config updates |
| TanStack Query 5.x | React 18+ | Suspense support requires React 18 |

## CloudWatch Metrics for RDS Activity Detection

Key metrics to fetch for sleep scheduling decisions:

| Metric | Namespace | Purpose |
|--------|-----------|---------|
| DatabaseConnections | AWS/RDS | Primary activity indicator - 0 = inactive |
| CPUUtilization | AWS/RDS | Secondary indicator - low % = minimal activity |
| ReadIOPS, WriteIOPS | AWS/RDS | I/O activity - 0 = no database operations |

```go
// Example: GetMetricData for connections
input := &cloudwatch.GetMetricDataInput{
    StartTime: aws.Time(time.Now().Add(-24 * time.Hour)),
    EndTime:   aws.Time(time.Now()),
    MetricDataQueries: []types.MetricDataQuery{
        {
            Id: aws.String("connections"),
            MetricStat: &types.MetricStat{
                Metric: &types.Metric{
                    Namespace:  aws.String("AWS/RDS"),
                    MetricName: aws.String("DatabaseConnections"),
                    Dimensions: []types.Dimension{
                        {Name: aws.String("DBInstanceIdentifier"), Value: aws.String(instanceID)},
                    },
                },
                Period: aws.Int32(3600), // 1 hour
                Stat:   aws.String("Average"),
            },
        },
    },
}
```

## GCP Cloud Monitoring for Cloud SQL Activity Detection

Key metrics to fetch:

| Metric | Resource Type | Purpose |
|--------|---------------|---------|
| cloudsql.googleapis.com/database/network/connections | cloudsql_database | Primary activity indicator |
| cloudsql.googleapis.com/database/cpu/utilization | cloudsql_database | CPU load indicator |
| cloudsql.googleapis.com/database/disk/read_ops_count | cloudsql_database | Read operations |

```go
// Example: Cloud Monitoring query
req := &monitoringpb.ListTimeSeriesRequest{
    Name:   "projects/" + projectID,
    Filter: `metric.type="cloudsql.googleapis.com/database/network/connections" AND resource.labels.database_id="` + instanceID + `"`,
    Interval: &monitoringpb.TimeInterval{
        StartTime: timestamppb.New(time.Now().Add(-24 * time.Hour)),
        EndTime:   timestamppb.Now(),
    },
}
```

## Cron Scheduling Pattern

Use robfig/cron v3 with timezone support for sleep/wake schedules:

```go
import "github.com/robfig/cron/v3"

// Initialize with timezone support and panic recovery
c := cron.New(
    cron.WithLocation(time.UTC),
    cron.WithChain(
        cron.Recover(logger),
        cron.SkipIfStillRunning(logger),
    ),
)

// Add sleep/wake jobs from schedule
c.AddFunc("CRON_TZ=America/New_York 0 19 * * 1-5", func() {
    // Stop instances at 7 PM ET on weekdays
    scheduler.ExecuteSleepOperation(scheduleID)
})

c.AddFunc("CRON_TZ=America/New_York 0 7 * * 1-5", func() {
    // Start instances at 7 AM ET on weekdays
    scheduler.ExecuteWakeOperation(scheduleID)
})

c.Start()
defer c.Stop()
```

## Sources

- pkg.go.dev/github.com/robfig/cron/v3 — Cron library documentation, v3.0.1 features (2026-02-20)
- pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/cloudwatch — CloudWatch SDK v1.54.0 (2026-02-20)
- pkg.go.dev/cloud.google.com/go/monitoring — GCP Monitoring v1.24.3 (2026-02-20)
- registry.npmjs.org — npm package versions verified (2026-02-20)
- GitHub API releases — Chi v5.2.5, pgx v5.8.0 verified (2026-02-20)
- Existing codebase analysis — go.mod, package.json reviewed

---
*Stack research for: Database sleep scheduling (SnoozeQL)*
*Researched: 2026-02-20*
