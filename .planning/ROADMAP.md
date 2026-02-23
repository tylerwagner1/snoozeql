# Roadmap: SnoozeQL

## Overview

This roadmap delivers a database sleep scheduling system that minimizes cloud costs by intelligently managing AWS RDS and GCP Cloud SQL instance lifecycles. The journey progresses from multi-cloud discovery through manual control, scheduling automation, and culminates in activity-based intelligent recommendations—the key differentiator that sets SnoozeQL apart from basic schedulers.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Multi-Cloud Discovery** - Connect to AWS/GCP and display all database instances ✅ Complete (2026-02-21)
- [x] **Phase 2: Manual Control & Audit** - Sleep/wake instances with confirmation and operation logging ✅ Complete (2026-02-23)
- [ ] **Phase 3: Basic Scheduling** - Create time-based sleep/wake schedules
- [ ] **Phase 4: Advanced Schedule Filtering** - Regex-based instance assignment with preview
- [ ] **Phase 5: Activity Analysis** - Collect metrics and detect inactivity patterns
- [ ] **Phase 6: Intelligent Recommendations** - Generate and apply activity-based schedule suggestions

## Phase Details

### Phase 1: Multi-Cloud Discovery
**Goal**: Users can view all database instances across multiple AWS and GCP accounts
**Depends on**: Nothing (first phase)
**Requirements**: DISC-01, DISC-02, AWS-01, AWS-02, GCP-01, GCP-02
**Success Criteria** (what must be TRUE):
  1. User can add multiple AWS account connections and see their RDS instances
  2. User can add multiple GCP project connections and see their Cloud SQL instances
  3. User can see instance status (running/stopped/pending) for each database in the UI
  4. Instances from all connected accounts appear in a unified list
**Plans**: 6 plans

Plans:
- [x] 01-01-PLAN.md — Instance persistence and connection status tracking
- [x] 01-02-PLAN.md — Multi-account provider registration and GCP credentials
- [x] 01-03-PLAN.md — Sortable/filterable instances table with account column
- [x] 01-04-PLAN.md — Connection status chips, skeleton loading, toasts
- [x] 01-05-PLAN.md — Clickable dashboard stats and add account CTAs
- [x] 01-06-PLAN.md — End-to-end verification checkpoint

#### Phase 1 Completion Details

**Completed:** 2026-02-21  
**Status:** ✅ Complete  
**Verification:** see [VERIFICATION.md](./phases/01-multi-cloud-discovery/VERIFICATION.md)

**Key Accomplishments:**
- Instance persistence with database syncing
- Multi-account provider registration (AWS + GCP)
- Sortable/filterable instances table with account column
- Connection status tracking with chips and toasts
- Clickable dashboard stats with URL-based filtering
- CTAs for adding cloud accounts
- Complete end-to-end verification

---

### Phase 2: Manual Control & Audit
**Goal**: Users can manually sleep/wake instances with confirmation and see operation history
**Depends on**: Phase 1 ✅ (Phase 1 completed 2026-02-21)
**Requirements**: SLEEP-01, SLEEP-02, WAKE-01, WAKE-02, AUDIT-01, AUDIT-03
**Success Criteria** (what must be TRUE):
  1. User can select one or multiple instances and trigger sleep with confirmation dialog
  2. User can select one or multiple instances and trigger wake with confirmation dialog
  3. All sleep/wake operations are logged with timestamps
  4. User can view operation history/audit log showing all past operations
**Plans**: 5 plans

Plans:
- [x] 02-01-PLAN.md — EventStore and Events API endpoint (Wave 1)
- [x] 02-02-PLAN.md — ConfirmDialog component with Headless UI (Wave 1)
- [x] 02-03-PLAN.md — Bulk stop/start API endpoints with audit logging (Wave 2)
- [x] 02-04-PLAN.md — Multi-select and bulk actions in InstancesPage (Wave 3)
- [x] 02-05-PLAN.md — AuditLogPage, navigation, and verification (Wave 4)

#### Phase 2 Completion Details

**Completed:** 2026-02-23  
**Status:** ✅ Complete  
**Verification:** see [02-VERIFICATION.md](./phases/02-manual-control-audit/02-VERIFICATION.md)

**Key Accomplishments:**
- EventStore with CreateEvent, ListEvents methods for audit logging
- ConfirmDialog component with Headless UI for confirmation flows
- Bulk stop/start API endpoints with event logging
- Multi-select table with bulk action buttons in InstancesPage
- AuditLogPage with event filtering and navigation link
- Human-verified complete end-to-end flow

---

### Phase 3: Basic Scheduling
**Goal**: Users can create time-based sleep/wake schedules
**Depends on**: Phase 2
**Requirements**: SCH-01
**Success Criteria** (what must be TRUE):
  1. User can create a schedule specifying start time, end time, and days of week
  2. Created schedules appear in the schedules list
**Plans**: TBD

Plans:
- [ ] 03-01: TBD

### Phase 4: Advanced Schedule Filtering
**Goal**: Users can assign schedules to instances using flexible regex-based filters
**Depends on**: Phase 3
**Requirements**: SCH-02, SCH-03, SCH-04, SCH-05, SCH-06, SCH-07, SCH-08
**Success Criteria** (what must be TRUE):
  1. User can create schedule filters based on instance name using regex patterns
  2. User can create schedule filters based on instance tags using regex patterns
  3. User can create schedule filters based on cloud provider (AWS/GCP)
  4. User can combine filters with AND/OR operators
  5. User can preview which instances will match a filter before applying
  6. User can view all created schedules in a dedicated schedules tab
**Plans**: TBD

Plans:
- [ ] 04-01: TBD

### Phase 5: Activity Analysis
**Goal**: System collects and analyzes metrics to detect inactivity patterns
**Depends on**: Phase 4
**Requirements**: ACT-01, ACT-02, ACT-03, ACT-04
**Success Criteria** (what must be TRUE):
  1. System ingests CloudWatch metrics for AWS RDS instances
  2. System ingests Cloud Monitoring metrics for GCP Cloud SQL instances
  3. System identifies periods of low/zero activity from collected metrics
  4. System detects nightly idle periods suitable for sleep scheduling
**Plans**: TBD

Plans:
- [ ] 05-01: TBD

### Phase 6: Intelligent Recommendations
**Goal**: System generates schedule recommendations based on activity analysis that users can review and apply
**Depends on**: Phase 5
**Requirements**: REC-01, REC-02, REC-03, REC-04, REC-05, REC-06
**Success Criteria** (what must be TRUE):
  1. System analyzes activity metrics and generates potential sleep schedule recommendations
  2. Recommendations display as cards on the main dashboard
  3. Recommendations are available in a dedicated recommendations tab
  4. Each recommendation shows suggested sleep/wake times based on activity patterns
  5. User can review a recommendation and confirm to create a schedule
**Plans**: TBD

Plans:
- [ ] 06-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Multi-Cloud Discovery | 6/6 | Complete | 2026-02-21 ✅ |
| 2. Manual Control & Audit | 5/5 | Complete | 2026-02-23 ✅ |
| 3. Basic Scheduling | 0/? | Not started | - |
| 4. Advanced Schedule Filtering | 0/? | Not started | - |
| 5. Activity Analysis | 0/? | Not started | - |
| 6. Intelligent Recommendations | 0/? | Not started | - |

---
*Created: 2026-02-20*
*Depth: comprehensive*
*Coverage: 26/26 v1 requirements mapped*
