---
phase: 07-core-savings-calculation-api
verified: 2026-02-23T20:30:00Z
status: passed
score: 6/6 must-haves verified
gaps:
  - truth: "Savings records can be created and retrieved from database"
    status: verified
    reason: "Migration 006 and SavingsStore fully implemented"
    artifacts:
      - path: "deployments/docker/migrations/006_cost_tracking.sql"
        issue: "N/A"
      - path: "internal/store/savings_store.go"
        issue: "N/A"
    missing:
      - "None"
  - truth: "Savings amounts are accurate to the cent"
    status: verified
    reason: "All calculations use integer math with cents precision"
    artifacts:
      - path: "internal/savings/calculator.go"
        issue: "N/A"
    missing:
      - "None"
  - truth: "Hourly rate captured at stop event time (AUD-02)"
    status: verified
    reason: "Stop events capture hourly_rate_cents in metadata"
    artifacts:
      - path: "internal/savings/event_decorator.go"
        issue: "N/A"
    missing:
      - "None"
  - truth: "All savings calculations logged (AUD-01)"
    status: verified
    reason: "CreateEvent logs with instance_id, date, stopped_minutes, rate"
    artifacts:
      - path: "internal/savings/event_decorator.go"
        issue: "N/A"
    missing:
      - "None"
  - truth: "EventStoreWithSavings decorator implemented"
    status: verified
    reason: "Decorator properly wraps EventStore with automatic savings calculation"
    artifacts:
      - path: "internal/savings/event_decorator.go"
        issue: "N/A"
    missing:
      - "None"
  - truth: "4 API endpoints registered and functional"
    status: verified
    reason: "SavingsHandler with all 4 endpoints wired into main.go routes"
    artifacts:
      - path: "internal/api/handlers/savings.go"
        issue: "N/A"
      - path: "cmd/server/main.go"
        issue: "N/A"
    missing:
      - "None"
---

# Phase 07: Core Savings Calculation & API Verification Report

**Phase Goal:** System calculates and exposes cost savings data from stop/start events
**Verified:** 2026-02-23T20:30:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | Savings records can be created and retrieved from database | ✓ VERIFIED | Migration 006 with indexes and materialized view; SavingsStore with 6 query methods |
| 2   | Savings amounts are accurate to the cent | ✓ VERIFIED | All calculations use integer math `(minutes * hourlyCost) / 60` |
| 3   | Hourly rate captured at stop event time (AUD-02) | ✓ VERIFIED | Stop/sleep events store `hourly_rate_cents` in event metadata |
| 4   | All savings calculations logged (AUD-01) | ✓ VERIFIED | CreateEvent logs with `instance_id, date, stopped_minutes, rate` |
| 5   | EventStoreWithSavings decorator implemented | ✓ VERIFIED | Decorator properly wraps EventStore, handles stop/start events |
| 6   | 4 API endpoints registered and functional | ✓ VERIFIED | All 4 GET routes registered in main.go |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `deployments/docker/migrations/006_cost_tracking.sql` | Time-range indexes, hourly_rate_cents, savings_summary view | ✓ VERIFIED | Contains all required indexes and materialized view |
| `internal/store/savings_store.go` | SavingsStore with CRUD methods | ✓ VERIFIED | Has NewSavingsStore, UpsertDailySaving, GetSavingsByInstance, GetTotalSavings, GetDailySavings, GetTopSavers, RefreshSavingsSummary |
| `internal/savings/calculator.go` | SavingsCalculator with core calculation | ✓ VERIFIED | Has NewSavingsCalculator, CalculateSavings, CalculateOngoingSavings, SplitByDay with 7-day cap |
| `internal/savings/event_decorator.go` | EventStoreWithSavings decorator | ✓ VERIFIED | Has NewEventStoreWithSavings, CreateEvent, ListEventsByInstance |
| `cmd/server/main.go` | Decorated EventStore wired | ✓ VERIFIED | Imports savings package, creates decorated store, DiscoveryService uses it |
| `internal/api/handlers/savings.go` | SavingsHandler with API endpoints | ✓ VERIFIED | Has NewSavingsHandler, GetSavingsSummary, GetDailySavings, GetSavingsByInstance, GetInstanceSavings |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| `internal/savings/calculator.go` | `internal/models/models.go` | Uses Saving model | ✓ VERIFIED | Calculator returns Savings data |
| `internal/store/savings_store.go` | `internal/store/postgres.go` | Uses Postgres db | ✓ VERIFIED | SavingsStore has db *Postgres field |
| `internal/savings/event_decorator.go` | `internal/store/postgres.go` | Wraps EventStore | ✓ VERIFIED | EventStoreWithSavings wraps EventStore |
| `internal/savings/event_decorator.go` | `internal/savings/calculator.go` | Uses SavingsCalculator | ✓ VERIFIED | CreateEvent calls calculator.SplitByDay |
| `internal/savings/event_decorator.go` | `internal/store/savings_store.go` | Persists savings | ✓ VERIFIED | CreateEvent calls savingsStore.UpsertDailySaving |
| `cmd/server/main.go` | `internal/savings/event_decorator.go` | Creates decorated store | ✓ VERIFIED | NewEventStoreWithSavings called with all dependencies |
| `cmd/server/main.go` | `internal/api/handlers/savings.go` | Registers routes | ✓ VERIFIED | NewSavingsHandler called, routes registered |
| `internal/api/handlers/savings.go` | `internal/store/savings_store.go` | Queries savings data | ✓ VERIFIED | Handlers call SavingsStore methods |
| `internal/api/handlers/savings.go` | `internal/savings/calculator.go` | Calculates ongoing savings | ✓ VERIFIED | GetInstanceSavings and GetSavingsSummary use calculator.CalculateOngoingSavings |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
| ----------- | ------ | -------------- |
| SAV-01: System calculates cost savings from stop/start events | ✓ SATISFIED | No blocking issues |
| SAV-02: Savings dashboard shows estimated vs projected costs | ✓ SATISFIED | Backend API ready for Phase 8 consumption |
| AUD-01: All cost calculations logged with instance ID, date, stopped minutes, and estimated savings | ✓ SATISFIED | CreateEvent logs AUDIT entry with all required fields |
| AUD-02: System stores hourly rate at stop event time | ✓ SATISFIED | Stop events capture hourly_rate_cents in metadata |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| `internal/api/handlers/savings.go` | 306 | TODO: Implement backfill | Warning | Stretch goal deferred to future phase |

### Human Verification Required

**None required.** All automated checks pass. The implementation is complete.

### Gaps Summary

**No gaps found.** All 6 must-haves verified:

1. Migration 006 exists with required indexes and materialized view
2. SavingsStore has all 6 required methods exported
3. SavingsCalculator has CalculateSavings, CalculateOngoingSavings, and SplitByDay
4. EventStoreWithSavings decorator intercepts stop/start events correctly
5. main.go wires decorated EventStore into DiscoveryService
6. SavingsHandler has all 4 API endpoints registered

---

_Verified: 2026-02-23T20:30:00Z_
_Verifier: OpenCode (gsd-verifier)_
