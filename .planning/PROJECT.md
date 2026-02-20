# SnoozeQL

## What This Is

A database lifecycle management application that helps users reduce cloud costs by identifying inactive database instances and either manually or intelligently scheduling sleep/wake cycles. The system monitors AWS RDS and GCP Cloud SQL instances, detects periods of low/no activity, and provides both manual control and automatic schedule recommendations.

## Core Value

Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## Requirements

### Validated

- ✓ Cloud provider abstraction layer supports AWS RDS and GCP Cloud SQL
- ✓ Discovery service polls providers for database instances
- ✓ Background scheduling system for recurring sleep/wake operations
- ✓ PostgreSQL persistence for schedules, recommendations, and account configurations

### Active

- [ ] **UI-01**: User can view all database instances from multiple AWS and GCP accounts
- [ ] **UI-02**: User can manually select one or many instances and trigger immediate sleep
- [ ] **SCH-01**: User can create schedules with regex-based instance assignment (name, tags, provider)
- [ ] **SCH-02**: Schedule assignment logic supports AND/OR operators with regex patterns
- [ ] **SCH-03**: User can create schedules manually by specifying start/sleep times
- [ ] **REC-01**: System analyzes activity metrics and recommends sleep schedules
- [ ] **REC-02**: RecommendationsDisplay shows recommendations as cards on dashboard and in separate tab
- [ ] **REC-03**: User can confirm and apply schedule recommendations to create new schedules
- [ ] **ACT-01**: System ingests CloudWatch metrics (AWS) and Cloud Monitoring metrics (GCP)
- [ ] **ACT-02**: Activity analysis identifies patterns of low/zero activity (e.g., nightly off-hours)
- [ ] **CON-01**: Support multiple AWS "connections" (different accounts/instances)
- [ ] **CON-02**: Support multiple GCP "connections" (different projects/instances)
- [ ] **CONF-01**: Sleep/wake operations use cloud provider APIs to physically stop/start instances

### Out of Scope

- Multi-user support with RBAC — Single-user POC only
- Billing reports or cost projections — Focus is on sleep scheduling, not analytics
- Email notifications or alerts — No notification infrastructure needed for POC
- Real-time wake-on-connect — Manual wake-up only for now
- Advanced scheduling patterns (timezone-aware, holidays, etc.) — Basic time-based scheduling only

## Context

**Current State (from codebase analysis):**

The existing codebase has a working foundation:
- Go backend with REST API serving a React SPA frontend
- Provider abstraction pattern for AWS RDS and GCP Cloud SQL
- Discovery service that continuously polls cloud providers
- PostgreSQL persistence with JSONB support
- Models for CloudAccount, Instance, Schedule, Recommendation, Override, Event, Saving, Settings

**What Needs Work:**

The existing codebase needs enhancement to support:
1. Enhanced UI for multi-account instance viewing and selection
2. Manual sleep operations for selected instances
3. Schedule creation with regex-based assignment
4. Activity-based recommendation engine
5. Recommendation workflow (view → confirm → create schedule)

**Technical Environment:**

- **Backend:** Go 1.24.0, Chi router, PostgreSQL with pgx
- **Frontend:** React 18.2, Vite, React Router DOM, Recharts
- **Cloud SDKs:** AWS SDK v2, Google Cloud API
- **Deployment:** Docker/Docker Compose

## Constraints

- **Tech Stack:** Must leverage existing Go/React/PostgreSQL stack
- **Timeline:** POC hackathon — rapid development, minimum viable features
- **Scope:** Focus on RDS and Cloud SQL only (no Spanner, DocumentDB, etc.)
- **Users:** Single-user POC, no RBAC or multi-tenancy needed
- **Cloud APIs:** Must use native provider APIs for stop/start operations (not just scaling)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use existing discovery service pattern | Leverages working provider polling, only needs UI enhancements | Faster development, less risk |
| Single-user POC scope | Hackathon constraints, no need for complex auth | Simplified implementation, faster delivery |
| Cloud provider stop/start APIs | Requirement to achieve $0 cost during sleep | Direct cost savings, matches user intent |
| Regex-based schedule assignment | User requested flexible filtering with wildcards | Powerful and familiar pattern |
| Manual confirmation before applying recommendations | Prevents accidental schedule changes | Safety for user, clear workflow |
| Activity metrics from CloudWatch/Cloud Monitoring | Standard provider metrics, well-documented | Reliable data source for analysis |

---
*Last updated: 2026-02-20 after initialization*
