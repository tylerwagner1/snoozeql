---
phase: quick-001
plan: 01
subsystem: metrics
tags: ["metrics", "ui", "button", "api"]
---

# Phase quick-001 Plan 01: Add Test Metrics Button Summary

## One-Liner

Added "Test Metrics" button to Instance Details page for manual metrics collection trigger that calls backend endpoint on-demand.

## Dependency Graph

- **Requires:** Phase 10 (metrics collection infrastructure) - depends on existing metrics endpoints and collector
- **Provides:** Manual metrics collection endpoint, UI button for one-off collection
- **Affects:** Future phases may extend this pattern to other instances or add automation controls

### Tech Stack Changes

| Component | Action | Details |
|-----------|--------|---------|
| `internal/metrics/collector.go` | Added | Public `CollectInstance()` method for on-demand collection |
| `cmd/server/main.go` | Added | POST `/instances/:id/collect-metrics` endpoint |
| `web/src/lib/api.ts` | Added | `collectInstanceMetrics()` API method |
| `web/src/pages/InstanceDetailPage.tsx` | Added | Test Metrics button with loading state |

### Added Libraries/Patterns

| Tech | Purpose |
|------|---------|
| None | Pure implementation - used existing infrastructure |

### New Patterns Established

- Single-instance metrics trigger pattern
- UI button with loading state and disabled state for non-AWS instances
</file>
