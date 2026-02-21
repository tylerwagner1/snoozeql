# Requirements: SnoozeQL

**Defined:** 2026-02-20
**Core Value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Instance Discovery

- [x] **DISC-01**: User can view all database instances from multiple AWS and GCP accounts
- [x] **DISC-02**: User can see instance status (running/stopped/pending) for each database

### Manual Control

- [ ] **SLEEP-01**: User can manually select one or many instances and trigger sleep operation
- [ ] **SLEEP-02**: User sees confirmation dialog showing selected instance count before sleep
- [ ] **WAKE-01**: User can manually select instances and trigger wake operation
- [ ] **WAKE-02**: User sees confirmation dialog showing selected instance count before wake

### Schedule Management

- [ ] **SCH-01**: User can create schedules with start time, end time, and days of week
- [ ] **SCH-02**: User can create schedule filters based on instance name (regex)
- [ ] **SCH-03**: User can create schedule filters based on instance tags (regex)
- [ ] **SCH-04**: User can create schedule filters based on cloud provider (AWS/GCP)
- [ ] **SCH-05**: Schedule filters support AND/OR operators with regex patterns
- [ ] **SCH-06**: User can view filtered instance list showing what would match schedule
- [ ] **SCH-07**: User can confirm schedule and apply it to matched instances
- [ ] **SCH-08**: User can view all created schedules in a dedicated tab

### Intelligent Recommendations

- [ ] **REC-01**: System analyzes activity metrics and recommends potential sleep schedules
- [ ] **REC-02**: Recommendations display as cards on the main dashboard
- [ ] **REC-03**: Recommendations available in a dedicated recommendations tab
- [ ] **REC-04**: Recommendation shows suggested sleep/wake times based on activity patterns
- [ ] **REC-05**: User can review recommendation and confirm to create schedule
- [ ] **REC-06**: Confirmed recommendations create new schedule entry

### Activity Analysis

- [ ] **ACT-01**: System ingests CloudWatch metrics (AWS) for RDS instances
- [ ] **ACT-02**: System ingests Cloud Monitoring metrics (GCP) for Cloud SQL instances
- [ ] **ACT-03**: Activity analysis identifies patterns of low/zero activity
- [ ] **ACT-04**: System detects nightly idle periods for recommendation generation

### Operation Tracking

- [ ] **AUDIT-01**: System logs all sleep/wake operations with timestamp
- [ ] **AUDIT-02**: System logs schedule creation and updates
- [ ] **AUDIT-03**: User can view operation history/audit log

### Multi-Cloud Support

- [x] **AWS-01**: System supports multiple AWS accounts/connections
- [x] **AWS-02**: System uses AWS SDK v2 for RDS operations
- [x] **GCP-01**: System supports multiple GCP projects/connections
- [x] **GCP-02**: System uses Google Cloud API for Cloud SQL operations

### Phase 1 Requirements Status

**Completed:** 2026-02-21

| Requirement | Status |
|-------------|--------|
| DISC-01, DISC-02 | ✅ Complete | Instance discovery with AWS/GCP support |
| AWS-01, AWS-02 | ✅ Complete | AWS multi-account support, SDK v2 |
| GCP-01, GCP-02 | ✅ Complete | GCP multi-project support, Cloud API |

**Summary:** All Phase 1 v1 requirements complete. Phase 2 (Manual Control & Audit) ready to start.

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Scheduling

- **SCH-09**: Schedule conflict detection when new schedules overlap
- **SCH-10**: Override support for temporary "keep awake" periods
- **SCH-11**: Tag-based filtering additional fields beyond regex
- **SCH-12**: Recurring schedule templates (weekly patterns)

### Enhanced Insights

- **ACT-05**: Activity visualization charts for historical usage
- **ACT-06**:Predictive start time estimation for wake scheduling
- **ACT-07**: Cost savings estimate with actual vs projected costs

### User Experience

- **UI-01**: Bulk operations for applying schedules to multiple instances
- **UI-02**: Email notification system (deferred per research)
- **UI-03**: Multi-user authentication (deferred per research)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Real-time wake-on-connect | Databases take 5-15 minutes to start; complex proxy infrastructure needed |
| Auto-scaling to zero | RDS/Cloud SQL don't support scale-to-zero |
| Email/SMS notifications | Infrastructure complexity for POC; in-app logging sufficient |
| Multi-user RBAC | Single-user POC only; no auth infrastructure needed |
| Holiday calendar integration | Timezone complexity; manual override adequate for POC |
| Complex recurrence patterns | Weekly patterns sufficient; simple time windows work for v1 |
| Billing API integration | Estimation from instance specs + uptime sufficient for POC |
| Terraform/IaC export | API-first design; IaC integration can be built later |
| Database query analysis | CloudWatch/Cloud Monitoring metrics sufficient for activity detection |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| DISC-01 | Phase 1: Multi-Cloud Discovery | Pending |
| DISC-02 | Phase 1: Multi-Cloud Discovery | Pending |
| AWS-01 | Phase 1: Multi-Cloud Discovery | Pending |
| AWS-02 | Phase 1: Multi-Cloud Discovery | Pending |
| GCP-01 | Phase 1: Multi-Cloud Discovery | Pending |
| GCP-02 | Phase 1: Multi-Cloud Discovery | Pending |
| SLEEP-01 | Phase 2: Manual Control & Audit | Pending |
| SLEEP-02 | Phase 2: Manual Control & Audit | Pending |
| WAKE-01 | Phase 2: Manual Control & Audit | Pending |
| WAKE-02 | Phase 2: Manual Control & Audit | Pending |
| AUDIT-01 | Phase 2: Manual Control & Audit | Pending |
| AUDIT-02 | Phase 2: Manual Control & Audit | Pending |
| AUDIT-03 | Phase 2: Manual Control & Audit | Pending |
| SCH-01 | Phase 3: Basic Scheduling | Pending |
| SCH-02 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-03 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-04 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-05 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-06 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-07 | Phase 4: Advanced Schedule Filtering | Pending |
| SCH-08 | Phase 4: Advanced Schedule Filtering | Pending |
| ACT-01 | Phase 5: Activity Analysis | Pending |
| ACT-02 | Phase 5: Activity Analysis | Pending |
| ACT-03 | Phase 5: Activity Analysis | Pending |
| ACT-04 | Phase 5: Activity Analysis | Pending |
| REC-01 | Phase 6: Intelligent Recommendations | Pending |
| REC-02 | Phase 6: Intelligent Recommendations | Pending |
| REC-03 | Phase 6: Intelligent Recommendations | Pending |
| REC-04 | Phase 6: Intelligent Recommendations | Pending |
| REC-05 | Phase 6: Intelligent Recommendations | Pending |
| REC-06 | Phase 6: Intelligent Recommendations | Pending |

**Coverage:**
- v1 requirements: 26 total
- Mapped to phases: 26
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-20*
*Last updated: 2026-02-20 after requirements scoping*
