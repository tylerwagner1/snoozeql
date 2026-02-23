# Requirements: SnoozeQL

**Defined:** 2026-02-23
**Core Value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## v1 Requirements

### Cost Tracking

- ✓ **SAV-01**: System calculates cost savings from stop/start events using instance hourly cost and stopped duration
- ✓ **SAV-02**: Savings dashboard shows estimated vs projected costs with summary cards and time-series chart
- [ ] **SAV-03**: Historical activity charts visualize usage patterns over configurable time ranges (7d, 30d, 90d, custom)
- [ ] **SAV-04**: Per-instance savings attribution shows which instances contributed most to overall savings
- [ ] **SAV-05**: Cost projection compares expected vs actual for billing forecasts with clear disclaimers

### Audit & Compliance

- ✓ **AUD-01**: All cost calculations are logged with instance ID, date, stopped minutes, and estimated savings
- ✓ **AUD-02**: System stores hourly rate at stop event time to handle instance resizing accurately

## v2 Requirements

### Advanced Analytics

- **SAV-10**: Actual billing API integration (AWS Cost Explorer, GCP Cloud Billing) for precise cost matching
- **SAV-11**: Per-schedule savings attribution to measure ROI of automated schedules
- **SAV-12**: Export savings data to CSV for external analysis

### User Experience

- **SAV-20**: Email notifications for savings milestones
- **SAV-21**: PDF report generation for monthly savings summaries
- **SAV-22**: ML-based forecasting of future savings based on historical patterns

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-user support with RBAC | Single-user POC only; add after v1 validation |
| Real-time wake-on-connect | Manual wake-up only for v1; automatic wake adds complexity |
| Advanced scheduling patterns (timezone-aware, holidays) | Basic time-based scheduling sufficient for POC |
| Billing API integration for定价 accuracy | Estimation from instance specs + uptime is sufficient for POC; defer to v2 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SAV-01 | Phase 7 | Complete |
| SAV-02 | Phase 7 | Complete |
| SAV-03 | Phase 8 | Pending |
| SAV-04 | Phase 8 | Pending |
| SAV-05 | Phase 8 | Pending |
| AUD-01 | Phase 7 | Complete |
| AUD-02 | Phase 7 | Complete |

**Coverage:**

- v1.1 requirements: 7 total
- Phase 7 (Core Savings): SAV-01, SAV-02, AUD-01, AUD-02 (4 requirements)
- Phase 8 (Visualization): SAV-03, SAV-04, SAV-05 (3 requirements)
- Mapped: 7/7 ✓
- Unmapped: 0 ✓

---

*Requirements defined: 2026-02-23*
*Last updated: 2026-02-23 (Phase 7 complete - SAV-01, SAV-02, AUD-01, AUD-02)*
