# Requirements: SnoozeQL

**Defined:** 2026-02-24
**Core Value:** Minimize database costs by automatically sleeping instances during inactive periods while ensuring they wake up when needed.

## v1.2 Requirements

Requirements for Metrics & Recommendations milestone. Each maps to roadmap phases.

### Metrics Collection

- [x] **METR-01**: System reliably collects CPU, Memory (FreeableMemory), and Connections from AWS CloudWatch every 15 minutes
- [x] **METR-02**: Collected metrics are stored in metrics_hourly table and persisted correctly
- [ ] **METR-03**: Metrics retention is enforced at 7 days (cleanup removes older data)

*Note: Collection infrastructure exists but needs verification. Memory is a new metric to add.*

### Metrics Visualization

- [ ] **VIS-01**: User can view time-series charts for CPU, Memory, and Connections on Instance Details page
- [ ] **VIS-02**: User can select time range (1h, 6h, 24h, 7d) for metrics charts
- [ ] **VIS-03**: Multiple metrics are visible together on Instance Details page
- [ ] **VIS-04**: Metrics charts show appropriate loading and empty states

### Recommendations

- [ ] **REC-01**: Idle detection uses compound threshold (CPU < 5% AND connections = 0)
- [ ] **REC-02**: Recommendations are grouped by similar idle patterns
- [ ] **REC-03**: Each recommendation displays estimated daily savings

## Future Requirements

Deferred to future release. Tracked but not in current roadmap.

### Metrics Visualization

- **VIS-05**: Responsive charts resize with viewport
- **VIS-06**: Anomaly highlighting on charts
- **VIS-07**: Correlation view (CPU vs Connections)
- **VIS-08**: Export metrics as CSV

### Recommendations

- **REC-04**: Per-instance override within group
- **REC-05**: Confidence score badge (HIGH/MEDIUM/LOW)
- **REC-06**: Recommendation history/audit
- **REC-07**: Bulk confirm/dismiss

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| GCP Cloud Monitoring metrics | Focus on AWS first, defer GCP until AWS is flawless |
| Real-time streaming metrics | 15-min collection sufficient for idle detection |
| Enhanced Monitoring (OS-level) | Requires per-instance AWS setup, adds cost |
| ML-based pattern detection | Overkill for POC, simple thresholds are effective |
| Auto-apply recommendations | User explicitly wants confirmation workflow |
| Custom alerting/thresholds | Requires notification infrastructure |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| METR-01 | Phase 10 | Complete |
| METR-02 | Phase 10 | Complete |
| METR-03 | Phase 12 | Pending |
| VIS-01 | Phase 11 | Pending |
| VIS-02 | Phase 11 | Pending |
| VIS-03 | Phase 11 | Pending |
| VIS-04 | Phase 11 | Pending |
| REC-01 | Phase 13 | Pending |
| REC-02 | Phase 14 | Pending |
| REC-03 | Phase 14 | Pending |

**Coverage:**
- v1.2 requirements: 10 total
- Mapped to phases: 10 ✓
- Unmapped: 0

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-24 after roadmap creation*
