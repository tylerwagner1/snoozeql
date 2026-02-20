# Codebase Concerns

**Analysis Date:** 2026-02-20

## Tech Debt

**Monolithic main.go with inline handlers:**
- Issue: `cmd/server/main.go` (484 lines) contains all API route handlers inline instead of using the handler files in `internal/api/handlers/`
- Files: `cmd/server/main.go`
- Impact: Code duplication, harder to test, harder to maintain. Handler files exist but are unused stubs.
- Fix approach: Refactor inline handlers to use the existing handler structs in `internal/api/handlers/`, inject dependencies properly.

**Stub handler implementations:**
- Issue: Multiple handler files contain stub implementations that return empty arrays or placeholder responses
- Files: `internal/api/handlers/schedules.go`, `internal/api/handlers/recommendations.go`, `internal/api/handlers/settings.go`, `internal/api/handlers/savings.go`
- Impact: These handlers are not connected to the store layer; actual implementation is duplicated in `cmd/server/main.go`
- Fix approach: Wire up handlers with proper store dependencies and route them in main.go

**Stub metrics implementations returning nil:**
- Issue: AWS RDS provider metrics functions return `nil, nil` without actual CloudWatch integration
- Files: `internal/provider/aws/rds.go` (lines 264-278)
- Impact: Pattern analysis (`internal/analyzer/analyzer.go`) cannot work without real metrics data; recommendations will be unreliable
- Fix approach: Implement CloudWatch metrics fetching using AWS SDK CloudWatch client

**GCP Cloud Monitoring not implemented:**
- Issue: GCP provider GetMetrics returns error placeholders instead of actual metrics
- Files: `internal/provider/gcp/cloudsql.go` (lines 98-123)
- Impact: GCP instances cannot be analyzed for activity patterns
- Fix approach: Integrate with GCP Cloud Monitoring API

**Database migration stub:**
- Issue: `Migrate()` function is empty and does nothing
- Files: `internal/store/postgres.go` (lines 36-39)
- Impact: Schema changes require manual intervention; no migration history tracking
- Fix approach: Implement migration runner using golang-migrate or similar

**Discovery sync not implemented:**
- Issue: TODO comments indicate database sync is not implemented
- Files: `internal/discovery/discovery.go` (lines 84, 136)
- Impact: Discovered instances are not persisted; lost on restart
- Fix approach: Implement SyncInstance to write to instances table via store layer

**Duplicate ProviderRegistry implementations:**
- Issue: There are two provider registry implementations with duplicated logic
- Files: `internal/discovery/discovery.go` (lines 147-207), `internal/provider/registry.go`
- Impact: Confusion about which to use; code duplication
- Fix approach: Remove duplicate in discovery.go; use only `internal/provider/registry.go`

## Security Considerations

**Hardcoded dev-key bypass:**
- Risk: API authentication can be bypassed with "dev-key" in production
- Files: `internal/api/middleware/chi.go` (lines 54-58)
- Current mitigation: None - any request with "dev-key" is authorized
- Recommendations: Add environment check to only allow dev-key in development; implement proper API key validation against database

**API key validation not implemented:**
- Risk: Non-dev-key API keys are not validated at all; middleware passes all requests through after dev-key check
- Files: `internal/api/middleware/chi.go` (lines 60-63)
- Current mitigation: Config passed in but unused
- Recommendations: Implement key validation against `api_tokens` table; hash keys before comparison

**Cloud credentials stored in plaintext:**
- Risk: AWS access keys and secrets stored as plain JSON in database
- Files: `internal/store/postgres.go` (lines 258-266)
- Current mitigation: None
- Recommendations: Encrypt credentials at rest; consider using HashiCorp Vault or AWS Secrets Manager

**DEBUG logs expose partial credentials:**
- Risk: Log statements print first 10 characters of AWS access keys
- Files: `cmd/server/main.go` (lines 90, 98)
- Current mitigation: None
- Recommendations: Remove or redact credential logging; use structured logging with sensitive field masking

**CORS allows all origins:**
- Risk: `Access-Control-Allow-Origin: *` allows any website to make API requests
- Files: `internal/api/middleware/chi.go` (line 71)
- Current mitigation: API key required
- Recommendations: Configure allowed origins via environment variable; restrict to known frontend domains

**Frontend hardcoded API URL and key:**
- Risk: API base URL and auth key hardcoded in frontend
- Files: `web/src/lib/api.ts` (lines 1, 69, 83, 98, 112)
- Current mitigation: None
- Recommendations: Use environment variables for API URL; implement proper auth flow

## Performance Bottlenecks

**N+1 query pattern on instance listing:**
- Problem: Each instance start/stop iterates through all instances to find match
- Files: `cmd/server/main.go` (lines 200-214, 247-262)
- Cause: No caching; full list fetched from providers on every operation
- Improvement path: Cache discovered instances; implement instance lookup by ID

**Provider registration on every request:**
- Problem: `store.NewCloudAccountStore(db)` creates new store instance per request
- Files: `cmd/server/main.go` (lines 388, 437, 453)
- Cause: Store not properly injected as singleton
- Improvement path: Create store once at startup; inject into handlers

**Discovery runs in goroutine without instance caching:**
- Problem: Discovery fetches fresh from AWS/GCP on interval but doesn't cache results
- Files: `internal/discovery/discovery.go` (lines 96-114)
- Cause: Run() method just logs count; no persistence
- Improvement path: Persist discovered instances to database; serve from cache

**Scheduler always returns "stop" action:**
- Problem: `determineAction()` is hardcoded to always return "stop"
- Files: `internal/scheduler/scheduler.go` (lines 180-184)
- Cause: Cron parsing not implemented
- Improvement path: Implement proper cron expression parsing; determine action based on current time

## Fragile Areas

**Request ID generation not unique:**
- Files: `internal/api/middleware/chi.go` (lines 86-88)
- Why fragile: Uses timestamp with second precision; concurrent requests get same ID
- Safe modification: Use UUID or add random suffix
- Test coverage: None

**Pattern detection relies on exact metric structure:**
- Files: `internal/analyzer/analyzer.go` (lines 86-96, 100-126)
- Why fragile: Type assertions on map[string]any without nil checks; panics if metrics structure changes
- Safe modification: Add defensive nil/type checks; use typed metrics struct
- Test coverage: None

**PostgreSQL array parsing manual string manipulation:**
- Files: `internal/store/postgres.go` (lines 238-246)
- Why fragile: Manual parsing of PostgreSQL text[] format; breaks on edge cases (quoted values, escaped chars)
- Safe modification: Use pq.StringArray or proper parsing library
- Test coverage: None

**Backup file in source tree:**
- Files: `cmd/server/main.go.bak`
- Why fragile: Contains duplicate code; may cause confusion
- Safe modification: Delete backup file
- Test coverage: N/A

## Test Coverage Gaps

**No Go tests exist:**
- What's not tested: Entire backend - all internal packages, cmd entrypoints
- Files: No `*_test.go` files found in codebase
- Risk: Regressions on any change; no confidence in refactoring
- Priority: High

**No frontend tests:**
- What's not tested: React components, API client, pages
- Files: `web/src/` has no test files
- Risk: UI regressions; API integration failures go undetected
- Priority: Medium

**Critical untested paths:**
- Discovery service flow
- Schedule matching logic (`internal/scheduler/scheduler.go`)
- Pattern detection (`internal/analyzer/analyzer.go`)
- Provider start/stop operations
- Store CRUD operations
- Priority: High

## Missing Critical Features

**Schedule CRUD not wired to database:**
- Problem: Schedule endpoints return empty arrays or placeholders
- Blocks: Users cannot create, view, or manage sleep schedules

**Instance detail endpoint returns 404:**
- Problem: GET `/instances/{id}` always returns not found
- Files: `cmd/server/main.go` (lines 190-194)
- Blocks: Instance detail page functionality

**Stats endpoint returns hardcoded zeros:**
- Problem: Dashboard stats are hardcoded placeholder values
- Files: `cmd/server/main.go` (lines 163-173)
- Blocks: Accurate dashboard statistics

**Recommendations not implemented:**
- Problem: Recommendations endpoints return empty arrays; analyzer not invoked
- Blocks: Automated schedule suggestions based on activity

**No pagination on list endpoints:**
- Problem: All list endpoints return unbounded results
- Files: All `/instances`, `/schedules`, `/recommendations` endpoints
- Blocks: Performance at scale

**Providers not reloaded after adding cloud account:**
- Problem: New cloud accounts require server restart to take effect
- Files: `cmd/server/main.go` - provider registration happens only at startup
- Blocks: Seamless cloud account management

---

*Concerns audit: 2026-02-20*
