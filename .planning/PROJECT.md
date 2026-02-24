# SnoozeQL

## What This Is

A database lifecycle management application that helps users reduce cloud costs by identifying inactive database instances and either manually or intelligently scheduling sleep/wake cycles. The system monitors AWS RDS and GCP Cloud SQL instances, detects periods of low/no activity, and provides both manual control and automatic schedule recommendations.

## Core Value

Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## Current State

**v1.1 shipped 2026-02-24**

SnoozeQL is fully functional with:
- Go 1.24.0 backend (Chi router, PostgreSQL with pgx)
- React 18.2 frontend (Vite, React Router DOM, Recharts)
- AWS SDK v2 and Google Cloud API support
- Docker/Docker Compose deployment
- 9 phases, 33 plans shipped across v1.0 and v1.1

**v1.1 shipped:** Cost savings tracking system built, evaluated, then removed per product direction change. Instance metrics display added to Instance Details page.

## Current Milestone: v1.2 Metrics & Recommendations

**Goal:** Metrics working flawlessly with time-series visualization, powering intelligent schedule recommendations.

**Target features:**
- CPU, Memory, Connections metrics collected every 15 min (AWS RDS)
- 7-day retention for metrics data
- Time-series charts on Instance Details page
- Recommendation engine detects idle patterns via simple threshold
- Grouped recommendations with per-instance overrides
- User approval workflow creates and assigns schedules

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
- ✓ SAV-01: System calculates cost savings from stop/start events — v1.1 (built, then removed)
- ✓ SAV-02: Savings dashboard shows estimated vs projected costs — v1.1 (built, then removed)
- ✓ SAV-03: Historical activity charts visualize usage patterns — v1.1 (built, then removed)
- ✓ SAV-04: Per-instance savings attribution — v1.1 (built, then removed)
- ✓ SAV-05: Cost projection compares expected vs actual — v1.1 (built, then removed)
- ✓ AUD-01: All cost calculations logged — v1.1 (built, then removed)
- ✓ AUD-02: System stores hourly rate at stop event time — v1.1 (built, then removed)

### Active

(v1.2 requirements to be defined after research)

### Out of Scope

- Multi-user support with RBAC — Single-user POC only
- Email notifications or alerts — No notification infrastructure needed for POC
- Real-time wake-on-connect — Manual wake-up only for now
- Advanced scheduling patterns (timezone-aware, holidays, etc.) — Basic time-based scheduling only
- Billing API integration — Estimation from instance specs + uptime sufficient for POC

## Context

**Technical Environment:**

- **Backend:** Go 1.24.0, Chi router, PostgreSQL with pgx
- **Frontend:** React 18.2, Vite, React Router DOM, Recharts
- **Cloud SDKs:** AWS SDK v2, Google Cloud API
- **Deployment:** Docker/Docker Compose

**Known Issues (deferred to future):**
- AWS 7-day auto-restart: implement re-stop mechanism
- Instance state race conditions: implement proper state machine
- GCP Cloud Monitoring metrics: defer until AWS validation proves value
- Minor tech debt from savings removal (orphaned files, ~70 lines total)

## Constraints

- **Tech Stack:** Must leverage existing Go/React/PostgreSQL stack
- **Timeline:** POC hackathon — rapid development, minimum viable features
- **Scope:** Focus on RDS and Cloud SQL only (no Spanner, DocumentDB, etc.)
- **Users:** Single-user POC, no RBAC or multi-tenancy needed
- **Cloud APIs:** Must use native provider APIs for stop/start operations (not just scaling)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use existing discovery service pattern | Leverages working provider polling, only needs UI enhancements | ✓ Good |
| Single-user POC scope | Hackathon constraints, no need for complex auth | ✓ Good |
| Cloud provider stop/start APIs | Requirement to achieve $0 cost during sleep | ✓ Good |
| Regex-based schedule assignment | User requested flexible filtering with wildcards | ✓ Good |
| Manual confirmation before applying recommendations | Prevents accidental schedule changes | ✓ Good |
| Activity metrics from CloudWatch/Cloud Monitoring | Standard provider metrics, well-documented | ✓ Good |
| ROADMAP.md split (one line per completed milestone) | Constant context cost for large project | ✓ Good |
| Decimal phase numbering (2.1, 2.2) | Clear insertion semantics | ✓ Good |
| Integer cents for money calculations | Avoid float64 precision issues | ✓ Good |
| Remove savings feature after evaluation | Product direction change after build | ✓ Executed cleanly |

---
*Last updated: 2026-02-24 after v1.2 milestone started*
