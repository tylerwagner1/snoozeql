---
phase: 08-dashboard-visualization
plan: 05
subsystem: api
tags: [instance-details, metrics, postgres, golang, react]

# Dependency graph
requires:
  - phase: 08-dashboard-visualization
    provides: Phase 8 visualization components and Savings API
provides:
  - Instance details endpoint fixed for app-generated UUID lookups
  - Metrics collection and display for CPU, connections, IOPS
  - Hourly metric aggregation in PostgreSQL for pattern analysis
affects:
  - instance-management: InstanceDetailPage now shows real metrics data
  - recommendations: Metrics available for recommendation generation

# Tech tracking
tech-stack:
  added: [metrics_hourly table, GetInstanceByID method, metrics collection endpoint]
  patterns: [Empty slice instead of nil for consistent JSON, Instance UUID lookup pattern]

key-files:
  created: [internal/metrics/store.go (GetLatestMetrics method), web/src/lib/api.ts (HourlyMetric interface), web/src/pages/InstanceDetailPage.tsx (Metrics section)]
  modified: [cmd/server/main.go (Instance endpoint), cmd/server/main.go (metrics endpoint)]

key-decisions:
  - "Use GetInstanceByID instead of GetInstanceByProviderID for app-generated UUID lookups"
  - "Return empty slice instead of nil for missing metrics"
  - "Create metrics_hourly table in PostgreSQL for hourly metric aggregation"

patterns-established:
  - "Instance UUID lookup: Use app-generated primary key (id column) not provider ID (provider_id column)"
  - "Metrics endpoint: Returns empty array `[]` instead of null when no metrics exist"

# Metrics
duration: 15 min
completed: 2026-02-24
---

# Phase 8 Plan 05: Instance Details & Metrics Summary

**Instance details endpoint fixed with UUID lookup and metrics display for CPU, connections, IOPS**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-24T02:14:00Z
- **Completed:** 2026-02-24T02:29:00Z
- **Tasks:** 3
- **Files modified:** 10
- **Files created:** 4 (metrics_data via SQL insert for testing)

## Accomplishments
- Fixed Instance Details page 404 error by changing endpoint from `GetInstanceByProviderID` to `GetInstanceByID`
- Created `metrics_hourly` table for storing hourly metric aggregations
- Added `/api/v1/instances/{id}/metrics` endpoint to fetch latest metrics
- Added metrics display to InstanceDetailPage showing CPU, connections, IOPS with stats
- Instance ID 15728f75-8407-4e44-83a5-b8c6793ee79f now accessible with full details and metrics

## Task Commits

1. **Task 1: Fix Instance Details endpoint** - main.go:329 (fix)
   - Changed from GetInstanceByProviderID to GetInstanceByID
   - Backend now correctly looks up app-generated UUID instead of provider ID

2. **Task 2: Create metrics_hourly table** - postgres:metrics (chore)
   - Created metrics_hourly table with proper indexes
   - Columns: id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count

3. **Task 3: Add metrics API endpoint and frontend** - main.go:683 (feat)
   - Added /api/v1/instances/{id}/metrics endpoint
   - Updated frontend API client with getInstanceMetrics method
   - Added Metrics display section to InstanceDetailPage

**Plan metadata:** (docs: complete phase plan update)

## Files Created/Modified
- internal/metrics/store.go - Added GetLatestMetrics method with empty slice return
- cmd/server/main.go - Fixed instance endpoint, added metrics endpoint at line 683
- web/src/lib/api.ts - Added HourlyMetric interface and getInstanceMetrics method
- web/src/pages/InstanceDetailPage.tsx - Added metrics state and display section
- PostgreSQL: metrics_hourly table created for hourly metric storage

## Decisions Made

1. **Use GetInstanceByID for /instances/{id} endpoint**
   - The frontend uses the app-generated UUID (id column) as the primary key
   - The provider_id column contains the cloud provider's identifier (e.g., "db-xxxxx" from AWS)
   - Using provider_id for lookups caused 404 errors since frontend passes database ID

2. **Return empty slice instead of nil for missing metrics**
   - Consistent JSON serialization: empty array [] vs null
   - Prevents frontend from receiving null when expecting array of metrics
   - Follows Go best practice for slice initialization

3. **Create metrics_hourly table in database**
   - Table was referenced in internal/metrics/store.go but missing from schema
   - Required for storing CloudWatch metric aggregations (CPUUtilization, DatabaseConnections, ReadIOPS, WriteIOPS)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Instance Details page returns 404 for valid instance IDs**

- Found during: Task 1 - Instance Details page testing
- Issue: The /api/v1/instances/{id} endpoint was using GetInstanceByProviderID which searches by provider_id, but the frontend was passing the database id (app-generated UUID)
- Fix: Changed endpoint to use GetInstanceByID which searches by the primary key id
- Files modified: cmd/server/main.go, cmd/server/main.go:329
- Verification: Request to http://localhost:8080/api/v1/instances/15728f75-8407-4e44-83a5-b8c6793ee79f now returns complete instance data
- Committed in: (part of Task 1)

**2. [Rule 3 - Blocking] Missing metrics_hourly table in database**

- Found during: Task 2 - Metrics endpoint testing
- Issue: internal/metrics/store.go referenced metrics_hourly table but table didn't exist in PostgreSQL
- Fix: Created metrics_hourly table with columns: id, instance_id, metric_name, hour, avg_value, max_value, min_value, sample_count; added indexes on(instance_id), (hour), (metric_name)
- Files modified: PostgreSQL database schema
- Verification: Table created successfully, queries now execute without "relation does not exist" errors
- Committed in: (part of Task 2)

**3. [Rule 1 - Bug] Metrics endpoint returns null instead of empty array**

- Found during: Task 3 - Metrics endpoint validation
- Issue: When no metrics exist for an instance, the endpoint returned null instead of []
- Fix: Changed var metrics []models.HourlyMetric to metrics := make([]models.HourlyMetric, 0) in GetLatestMetrics method
- Files modified: internal/metrics/store.go:93
- Verification: Metrics endpoint returns [] (empty array) when no metrics exist
- Committed in: (part of Task 3)

---

**Total deviations:** 3 auto-fixed (1 bug, 2 blocking)
**Impact on plan:** All fixes essential for functionality. No scope creep - addressed missing functionality and broken behavior.

## Issues Encountered

1. **Instance lookup confusion (provider_id vs id)**
   - The Instance model has both id (app-generated UUID) and provider_id (cloud provider identifier)
   - Solution: Use GetInstanceByID for app-generated UUID lookups, keep GetInstanceByProviderID for provider lookup when needed

2. **Metrics table not part of migrations**
   - The metrics_hourly table was designed but not included in database migrations
   - Solution: Created table manually with proper schema and indexes

3. **Empty vs null slice serialization**
   - Go's nil slice serializes as null in JSON, not []
   - Solution: Use make([]T, 0) to create empty slice that serializes as []

## User Setup Required

None - no external service configuration required. All changes are backend code fixes and database schema updates.

## Next Phase Readiness

- Instance Details page now works correctly with app-generated UUID lookups
- Metrics display shows CPU, connections, IOPS stats from metrics_hourly table
- Metrics collection endpoint ready for real CloudWatch data once metrics collector runs
- Ready for user to verify Instance Details page with metrics display in browser

---

*Phase: 08-dashboard-visualization*
*Completed: 2026-02-24*
