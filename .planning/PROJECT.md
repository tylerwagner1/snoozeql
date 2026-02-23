# SnoozeQL

## What This Is

A database lifecycle management application that helps users reduce cloud costs by identifying inactive database instances and either manually or intelligently scheduling sleep/wake cycles. The system monitors AWS RDS and GCP Cloud SQL instances, detects periods of low/no activity, and provides both manual control and automatic schedule recommendations.

## Core Value

Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## Requirements

### Validated

- ✓ Cloud provider abstraction layer supports AWS RDS and GCP Cloud SQL — v1.0
- ✓ Discovery service polls providers for database instances — v1.0
- ✓ Background scheduling system for recurring sleep/wake operations — v1.0
- ✓ PostgreSQL persistence for schedules, recommendations, and account configurations — v1.0
- ✓ Instance discovery with multi-account AWS/GCP support — v1.0
- ✓ Manual sleep/wake with confirmation dialogs — v1.0
- ✓ Schedule creation with regex-based instance assignment — v1.0
- ✓ SchedulesPage with instance count and audit logging — v1.0
- ✓ Activity analysis with idle period detection — v1.0
- ✓ Intelligent schedule recommendations based on activity patterns — v1.0
- ✓ Recommendation workflow (view → confirm → create schedule) — v1.0

### Active

- [ ] **UI-01**: Export schedule as IaC configuration (Terraform/CloudFormation)
- [ ] **UI-02**: Cost savings dashboard with actual vs projected costs
- [ ] **UI-03**: Bulk operation presets for common patterns

### Out of Scope

- Multi-user support with RBAC — Single-user POC only
- Billing reports or cost projections — Focus is on sleep scheduling, not analytics
- Email notifications or alerts — No notification infrastructure needed for POC
- Real-time wake-on-connect — Manual wake-up only for now
- Advanced scheduling patterns (timezone-aware, holidays, etc.) — Basic time-based scheduling only

## Context

**Current State (v1.0 shipped 2026-02-23):**

SnoozeQL v1.0 is fully functional with:
- Go 1.24.0 backend (Chi router, PostgreSQL with pgx)
- React 18.2 frontend (Vite, React Router DOM, Recharts)
- AWS SDK v2 and Google Cloud API support
- Docker/Docker Compose deployment
- 25/26 v1 requirements shipped (ACT-02 GCP deferred)

**Key Achievements:**
- Multi-cloud discovery (AWS RDS + GCP Cloud SQL)
- Manual sleep/wake with audit logging
- Time-based scheduling with regex-based instance assignment
- Activity analysis with idle period detection
- Intelligent schedule recommendations

**Known Issues:**
- AWS 7-day auto-restart: implement re-stop mechanism (deferred)
- Instance state race conditions: implement proper state machine (deferred)

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
| ROADMAP.md split (one line per completed milestone) | Constant context cost for large project | Scales to 100+ phases without overflow |
| Decimal phase numbering (2.1, 2.2) | Clear insertion semantics | Easy to insert phases without renumbering |

---
*Last updated: 2026-02-23 after v1.0 milestone*
