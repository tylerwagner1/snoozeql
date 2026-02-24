# Phase 9: Complete Savings Removal & Cache Validation

## Objective

Systematically remove all savings-related code from both frontend and backend, rebuild Docker containers with fresh artifacts, and validate that no savings functionality or cached content remains.

## Background

Quick Task #002 removed the visible savings page (route, navigation, components) but left behind:
- Frontend API methods and type definitions in `web/src/lib/api.ts`
- Backend savings endpoints in `cmd/server/main.go`
- Backend savings handler in `internal/api/handlers/savings.go`
- Backend savings store in `internal/store/savings_store.go`
- Backend savings calculator in `internal/savings/calculator.go`
- Backend event decorator in `internal/savings/event_decorator.go`
- Docker frontend container serving stale build artifacts

## Success Criteria

1. No savings-related code in frontend (`web/src/`)
2. No savings endpoints registered in backend
3. No savings handlers, stores, or calculators in backend
4. Docker containers rebuilt with fresh artifacts
5. Application builds and runs successfully
6. Navigation shows only: Dashboard, Accounts, Instances, Schedules, Recommendations, Audit Log

## Tasks

### Task 1: Remove Frontend Savings API Methods and Types

**Files to modify:**
- `web/src/lib/api.ts`

**Actions:**
1. Remove `SavingsSummary` interface (lines 122-132)
2. Remove `DailySavingsResponse` interface (lines 136-143)
3. Remove `InstanceSavingsItem` interface (lines 145-152)
4. Remove `InstanceSavingsDetail` interface (lines 154-164)
5. Remove `getSavingsSummary` method (lines 291-293)
6. Remove `getDailySavings` method (lines 295-296)
7. Remove `getOngoingCost` method (lines 298-299)
8. Remove `getSavingsByInstance` method (lines 301-302)
9. Remove `getInstanceSavings` method (lines 304-305)
10. Remove `savings_7d` from `Stats` interface if present

**Verification:**
```bash
grep -n "savings\|Savings" web/src/lib/api.ts
# Should return only lines related to recommendations (estimated_daily_savings)
```

### Task 2: Remove Backend Savings Routes

**Files to modify:**
- `cmd/server/main.go`

**Actions:**
1. Remove savings handler import usage (line 668-674)
2. Remove savings routes:
   - `r.Get("/savings", savingsHandler.GetSavingsSummary)`
   - `r.Get("/savings/daily", savingsHandler.GetDailySavings)`
   - `r.Get("/savings/by-instance", savingsHandler.GetSavingsByInstance)`
   - `r.Get("/savings/ongoing", savingsHandler.GetOngoingCost)`
3. Remove `savingsStore` initialization (line 192)
4. Remove `savingsCalculator` initialization (lines 194-195)
5. Remove `decoratedEventStore` initialization (lines 197-201)
6. Remove savings-related imports from import block (line 28)
7. Remove savings-related global variables (lines 41, 44)

**Note:** Keep `savings_7d` in stats endpoint if it's needed for dashboard display. Otherwise remove.

**Verification:**
```bash
grep -n "savings\|Savings" cmd/server/main.go
# Should return minimal or no results
```

### Task 3: Delete Backend Savings Handler

**Files to delete:**
- `internal/api/handlers/savings.go`

**Verification:**
```bash
ls internal/api/handlers/savings.go
# Should return "No such file or directory"
```

### Task 4: Delete Backend Savings Store

**Files to delete:**
- `internal/store/savings_store.go`

**Verification:**
```bash
ls internal/store/savings_store.go
# Should return "No such file or directory"
```

### Task 5: Delete Backend Savings Package

**Directory to delete:**
- `internal/savings/` (contains calculator.go, event_decorator.go)

**Verification:**
```bash
ls internal/savings/
# Should return "No such file or directory"
```

### Task 6: Update Event Store (Remove Decorator Wrapper)

**Files to modify:**
- `cmd/server/main.go`

**Actions:**
1. Change event store usage from `decoratedEventStore` back to plain `eventStore`
2. Update any handlers that use the decorated store to use plain event store

**Verification:**
```bash
go build -o /dev/null ./cmd/server
# Should compile without errors
```

### Task 7: Rebuild and Restart Docker Containers

**Commands:**
```bash
# Rebuild frontend container with fresh build
docker-compose build --no-cache frontend

# Rebuild app container (backend)
docker-compose build --no-cache app

# Restart containers
docker-compose down
docker-compose up -d

# Verify new build artifacts are being served
curl -s http://localhost:3001 | grep -oE "index-[^.]+\.js"
```

**Verification:**
- Check that the JS filename served matches the one in `web/dist/assets/`
- Check that `/savings` route returns 404
- Check that navigation has exactly 6 items

### Task 8: Final Validation

**Verification steps:**
1. Build frontend: `cd web && npm run build`
2. Build backend: `go build ./cmd/server`
3. Run backend tests: `go test ./...`
4. Check no savings references:
   ```bash
   grep -r "savings\|Savings" web/src/ --include="*.ts" --include="*.tsx"
   # Should return only recommendation-related (estimated_daily_savings)
   
   grep -r "savings\|Savings" internal/ cmd/
   # Should return minimal or no results
   ```
5. Manual verification in browser:
   - Navigation shows 6 items only
   - `/savings` route returns 404 or redirects
   - No console errors related to savings

## Files Summary

**To Delete:**
- `internal/api/handlers/savings.go`
- `internal/store/savings_store.go`
- `internal/savings/calculator.go`
- `internal/savings/event_decorator.go`
- `internal/savings/` directory

**To Modify:**
- `web/src/lib/api.ts` - Remove savings types and API methods
- `cmd/server/main.go` - Remove savings routes, handlers, stores, imports

## Risk Assessment

**Low Risk:**
- Removing unused code that's no longer referenced
- Frontend changes are straightforward type/method removal

**Medium Risk:**
- Backend changes involve removing the event decorator wrapper
- Need to ensure event store still works correctly without savings calculation

**Mitigation:**
- Build and test after each major change
- Keep savings database table (don't run migration to drop it) in case user wants to restore later

## Estimated Duration

- Task 1: ~5 min
- Task 2: ~10 min
- Task 3-5: ~5 min
- Task 6: ~10 min
- Task 7: ~5 min
- Task 8: ~10 min

**Total: ~45 minutes**

## Dependencies

- Phase 8 complete
- Quick Task #002 complete (frontend page/components already removed)
- Docker daemon running
