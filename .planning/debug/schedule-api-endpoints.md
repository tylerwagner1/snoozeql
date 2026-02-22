---
status: resolved
trigger: "Schedule CRUD endpoints needed implementation"
created: "2026-02-22T10:00:00Z"
updated: "2026-02-22T10:25:00Z"
---

## Current Focus

hypothesis: Schedule CRUD endpoints are stubbed but not implemented with real store
test: Implemented ScheduleStore in postgres.go with full CRUD operations
expecting: Backend endpoints now work with real database operations
next_action: Verify Create Schedule UI works

## Summary

Implemented complete Schedule CRUD backend for the SnoozeQL application:

### Changes Made:

1. **internal/store/postgres.go** - Added ScheduleStore:
   - GetSchedule(id) - Retrieves single schedule by UUID
   - ListSchedules() - Returns all schedules sorted by created_at DESC
   - CreateSchedule(schedule) - Creates new schedule with JSONB selectors
   - UpdateSchedule(schedule) - Updates existing schedule
   - DeleteSchedule(id) - Deletes schedule by UUID
   - GetMatchingSchedules(instance) - Finds schedules matching an instance

2. **internal/api/handlers/schedules.go** - Updated to use real store:
   - Full CRUD with proper error handling
   - 404 for non-existent schedules
   - 400 for invalid request bodies
   - Enable/Disable endpoints for toggle operations

3. **cmd/server/main.go** - Added schedule routes:
   - GET /schedules - List all schedules
   - GET /schedules/{id} - Get single schedule
   - POST /schedules - Create new schedule
   - PUT /schedules/{id} - Update schedule
   - DELETE /schedules/{id} - Delete schedule
   - POST /schedules/{id}/enable - Enable schedule
   - POST /schedules/{id}/disable - Disable schedule

4. **Matcher functions** - Implemented using store helpers:
   - matchesMatcher - Pattern matching with exact, contains, prefix, suffix, regex
   - selectorMatchesInstance - Check if instance matches selector
   - matchesInstance - Check if instance matches any selector in list

### Verification:

```
✓ Store compiles: go build ./internal/store/...
✓ Handlers compile: go build ./internal/api/handlers/...
✓ Server compiles: go build ./cmd/server
```

### Database:

The `schedules` table already existed in the database schema:
```sql
CREATE TABLE schedules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    selectors JSONB NOT NULL,  -- Array of Selector objects
    timezone VARCHAR(100) NOT NULL,
    sleep_cron VARCHAR(100) NOT NULL,
    wake_cron VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

### Frontend:

Already implemented:
- ScheduleNewPage.tsx - Full form for creating schedules
- ScheduleEditPage.tsx - Full form for editing schedules
- Routes added to main.tsx: /schedules/new, /schedules/:id
- API calls updated to use /api/v1/schedules endpoints

### Test Data:

Create a test schedule via curl:
```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Schedule",
    "description": "A test schedule for verification",
    "timezone": "America/New_York",
    "sleep_cron": "0 22 * * *",
    "wake_cron": "0 8 * * 1-5",
    "enabled": true,
    "selectors": [{
      "name": {"pattern": "test", "type": "contains"}
    }]
  }'
```

### Next Steps:

1. Build and deploy the updated Docker container
2. Test Create Schedule UI in browser
3. Test Edit Schedule UI for existing schedules
4. Verify schedules appear in the list
5. Test enabling/disabling schedules