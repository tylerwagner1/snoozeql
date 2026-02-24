# Project Milestones: SnoozeQL

## v1.1 Enhanced Insights & Savings (Shipped: 2026-02-24)

**Delivered:** Complete savings tracking system built and validated, then removed per product direction change — a full feature lifecycle (build → evaluate → remove).

**Phases completed:** 7-9 (9 plans total)

**Key accomplishments:**

- Built complete savings tracking with SavingsStore, SavingsCalculator, EventStoreWithSavings decorator
- Created savings dashboard UI with SavingsSummaryCards, SavingsChart, InstanceSavingsTable
- Implemented API endpoints for savings summary, daily breakdown, and per-instance attribution
- Added instance metrics display (CPU, connections, IOPS) to Instance Details page
- Pivoted direction and cleanly removed all savings code per user decision
- Rebuilt Docker containers with fresh artifacts for clean deployment

**Stats:**

- 63 files changed
- ~40,000 lines added / 632 deleted
- 3 phases, 9 plans
- 2 days from milestone start to ship

**Git range:** `d17a687c` → `b2c4d03e`

**What's next:** Define next milestone with `/gsd-new-milestone`

---

## v1.0 MVP (Shipped: 2026-02-23)

**Delivered:** Complete database lifecycle management with multi-cloud discovery, manual control, scheduling automation, activity analysis, and intelligent recommendations.

**Phases completed:** 1-6 (24 plans total)

**Key accomplishments:**

- Multi-cloud instance discovery (AWS RDS + GCP Cloud SQL) with provider abstraction layer
- Manual sleep/wake with confirmation dialogs and audit logging (EventStore)
- Time-based scheduling with visual weekly grid and CRON mode
- Regex-based schedule filtering with AND/OR operators and instance preview
- Activity analysis with CloudWatch metrics ingestion and idle period detection
- Intelligent schedule recommendations based on activity patterns

**Stats:**

- 24 files created/modified
- ~5,000 lines of Go and TypeScript
- 6 phases, 24 plans, 100+ tasks
- ~6 days from milestone start to ship

**Git range:** `feat(01-01)` → `docs(phase-06)`

**What's next:** Start v1.1 with enhanced insights and user experience features

---

