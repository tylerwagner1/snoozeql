# Architecture

**Analysis Date:** 2026-02-20

## Pattern Overview

**Overall:** Multi-tier Monolith with Provider Abstraction

**Key Characteristics:**
- Go backend with REST API serving a React SPA frontend
- Provider abstraction pattern for multi-cloud support (AWS, GCP)
- Registry pattern for managing multiple cloud provider instances
- Background discovery service for continuous instance polling
- PostgreSQL for persistence with JSONB for flexible data

## Layers

**API Layer:**
- Purpose: HTTP endpoints, request validation, response formatting
- Location: `cmd/server/main.go` (inline routes), `internal/api/handlers/`
- Contains: Route definitions, middleware, HTTP handlers
- Depends on: Discovery Service, Store, Config
- Used by: Frontend SPA, external API consumers

**Discovery Layer:**
- Purpose: Background service that polls cloud providers for database instances
- Location: `internal/discovery/discovery.go`
- Contains: DiscoveryService with continuous polling, instance sync logic
- Depends on: Provider Registry
- Used by: API Layer (for instance listing, start/stop operations)

**Provider Layer:**
- Purpose: Abstract cloud provider operations into a common interface
- Location: `internal/provider/`
- Contains: Provider interface, Registry, AWS RDS implementation, GCP Cloud SQL implementation
- Depends on: Models, Cloud SDKs (AWS SDK v2, Google Cloud API)
- Used by: Discovery Service

**Store Layer:**
- Purpose: Database access and persistence
- Location: `internal/store/postgres.go`
- Contains: PostgreSQL connection wrapper, domain-specific stores (CloudAccountStore, RecommendationStore)
- Depends on: Models, pgx driver
- Used by: API Layer, handlers

**Models Layer:**
- Purpose: Domain types shared across layers
- Location: `internal/models/models.go`
- Contains: CloudAccount, Instance, Schedule, Recommendation, Override, Event, Saving, Settings
- Depends on: Nothing (pure data types)
- Used by: All layers

**Config Layer:**
- Purpose: Environment configuration loading and validation
- Location: `internal/config/config.go`
- Contains: Config struct, Load(), Validate()
- Depends on: Environment variables
- Used by: Server entry point

**Frontend Layer:**
- Purpose: React SPA dashboard
- Location: `web/src/`
- Contains: React components, pages, API client
- Depends on: Backend REST API
- Used by: End users via browser

## Data Flow

**Instance Discovery:**

1. Server starts → loads CloudAccounts from database (`internal/store/postgres.go`)
2. For each CloudAccount, creates provider (AWS/GCP) and registers with Registry
3. DiscoveryService.RunContinuous() polls registry at configured interval
4. Each provider's ListDatabases() calls cloud API (RDS DescribeDBInstances, Cloud SQL List)
5. Instances returned to API via DiscoveryService.ListAllDatabases()

**Start/Stop Instance:**

1. Frontend calls POST `/api/v1/instances/{id}/start` or `/stop`
2. API handler finds instance in discovery cache
3. Resolves provider name from instance (e.g., "aws_us-east-1")
4. Calls DiscoveryService.StartDatabase() or StopDatabase()
5. DiscoveryService delegates to Registry.StartDatabase()
6. Registry looks up Provider by name, calls provider.StartDatabase()
7. Provider calls cloud SDK (RDS StartDBInstance / Cloud SQL Patch)

**State Management:**
- Server state: In-memory provider registry, discovery service cache
- Persistent state: PostgreSQL (cloud_accounts, instances, schedules, etc.)
- Frontend state: React useState/useEffect, no global state management

## Key Abstractions

**Provider Interface:**
- Purpose: Unified API for multi-cloud database operations
- Examples: `internal/provider/provider.go`, `internal/provider/aws/rds.go`, `internal/provider/gcp/cloudsql.go`
- Pattern: Strategy pattern - each cloud implements same interface

```go
type Provider interface {
    ListDatabases(ctx context.Context) ([]models.Instance, error)
    StartDatabase(ctx context.Context, id string) error
    StopDatabase(ctx context.Context, id string) error
    GetDatabaseStatus(ctx context.Context, id string) (string, error)
    GetMetrics(ctx context.Context, providerName string, id string, period string) (map[string]any, error)
}
```

**Provider Registry:**
- Purpose: Manage multiple provider instances with named lookup
- Examples: `internal/provider/registry.go`
- Pattern: Registry pattern with key format "aws_{region}" or "gcp"

**Store Pattern:**
- Purpose: Domain-specific database access with raw SQL
- Examples: `internal/store/postgres.go` (CloudAccountStore, RecommendationStore)
- Pattern: Repository-like pattern wrapping raw `database/sql`

## Entry Points

**Server (cmd/server/main.go):**
- Location: `cmd/server/main.go`
- Triggers: `go run cmd/server/main.go` or built binary
- Responsibilities:
  - Load config from environment
  - Connect to PostgreSQL
  - Load cloud accounts and create providers
  - Start discovery service goroutine
  - Set up Chi router with middleware
  - Listen on configured host:port

**Discover CLI (cmd/discover/main.go):**
- Location: `cmd/discover/main.go`
- Triggers: Manual CLI invocation
- Responsibilities: One-shot database discovery for testing

**Frontend (web/src/main.tsx):**
- Location: `web/src/main.tsx`
- Triggers: Browser navigation
- Responsibilities: React app bootstrap, routing setup

## Error Handling

**Strategy:** Log and return HTTP errors; no global error types

**Patterns:**
- Errors wrapped with `fmt.Errorf("context: %w", err)` for stack context
- HTTP handlers return JSON error responses: `{"error": "message"}`
- Background discovery logs errors but continues running
- No panic recovery except Chi middleware

## Cross-Cutting Concerns

**Logging:** 
- Standard `log` package with flags `LstdFlags | Lmicroseconds | Lshortfile`
- Output to stdout
- DEBUG-prefixed messages for troubleshooting

**Validation:**
- Config validation via `config.Validate()`
- Request body validation inline in handlers
- No centralized validation framework

**Authentication:**
- API key via Authorization header
- Development mode: "dev-key" bypasses validation
- Middleware: `internal/api/middleware/chi.go` APIKeyAuth()

**CORS:**
- Wide open: `Access-Control-Allow-Origin: *`
- Middleware: `internal/api/middleware/chi.go` CORS()

---

*Architecture analysis: 2026-02-20*
