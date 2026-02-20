# Codebase Structure

**Analysis Date:** 2026-02-20

## Directory Layout

```
snoozeql/
├── cmd/                    # Application entry points
│   ├── server/             # Main API server
│   │   └── main.go
│   └── discover/           # Discovery CLI tool
│       └── main.go
├── internal/               # Private Go packages
│   ├── analyzer/           # Activity analysis (stub)
│   ├── api/                # HTTP layer
│   │   ├── handlers/       # Route handlers
│   │   └── middleware/     # Chi middleware
│   ├── config/             # Configuration loading
│   ├── cost/               # Cost calculation (stub)
│   ├── discovery/          # Background discovery service
│   ├── models/             # Domain data types
│   ├── notifier/           # Notifications (stub)
│   ├── provider/           # Cloud provider abstraction
│   │   ├── aws/            # AWS RDS implementation
│   │   └── gcp/            # GCP Cloud SQL implementation
│   ├── scheduler/          # Schedule execution (stub)
│   └── store/              # Database access
├── web/                    # React SPA frontend
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── lib/            # Utilities (API client)
│   │   └── pages/          # Route pages
│   └── dist/               # Built frontend assets
├── deployments/            # Deployment configurations
│   ├── docker/
│   │   └── migrations/     # SQL migrations
│   └── kubernetes/         # K8s manifests (empty)
├── vendor/                 # Go vendored dependencies
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
├── Dockerfile              # Production container
├── Dockerfile.dev          # Development container
├── Dockerfile.frontend     # Frontend build container
├── docker-compose.yml      # Local dev environment
└── Makefile                # Build commands
```

## Directory Purposes

**cmd/:**
- Purpose: Executable entry points (main packages)
- Contains: `main.go` files only, minimal logic
- Key files: `cmd/server/main.go` (primary server), `cmd/discover/main.go` (CLI tool)

**internal/:**
- Purpose: Private application packages (cannot be imported externally)
- Contains: All business logic, organized by domain concern
- Key files: All core Go packages

**internal/provider/:**
- Purpose: Cloud provider abstractions and implementations
- Contains: Provider interface, Registry, cloud-specific implementations
- Key files:
  - `provider.go`: Provider interface definition
  - `registry.go`: Multi-provider management
  - `aws/rds.go`: AWS RDS implementation
  - `gcp/cloudsql.go`: GCP Cloud SQL implementation

**internal/api/:**
- Purpose: HTTP API layer
- Contains: Handlers and middleware
- Key files:
  - `handlers/instances.go`: Instance CRUD handlers (stub)
  - `handlers/schedules.go`: Schedule handlers (stub)
  - `handlers/recommendations.go`: Recommendation handlers (stub)
  - `middleware/chi.go`: Request ID, CORS, API key auth

**internal/store/:**
- Purpose: Database access layer
- Contains: PostgreSQL connection, domain stores
- Key files: `postgres.go` (Postgres wrapper, CloudAccountStore, RecommendationStore)

**internal/models/:**
- Purpose: Shared domain data types
- Contains: Struct definitions with JSON/DB tags
- Key files: `models.go` (all models in one file)

**web/src/:**
- Purpose: React frontend source code
- Contains: React components, pages, utilities
- Key files:
  - `main.tsx`: App bootstrap and routing
  - `App.tsx`: Root layout component
  - `lib/api.ts`: API client wrapper

**deployments/docker/migrations/:**
- Purpose: Database schema migrations
- Contains: SQL migration files
- Key files: `001_base_schema.sql` (initial schema)

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Main server entry point
- `cmd/discover/main.go`: Discovery CLI entry point
- `web/src/main.tsx`: Frontend entry point

**Configuration:**
- `internal/config/config.go`: Environment config loading
- `docker-compose.yml`: Local development config
- `web/vite.config.ts`: Frontend build config

**Core Logic:**
- `internal/provider/provider.go`: Provider interface
- `internal/provider/registry.go`: Provider registry
- `internal/provider/aws/rds.go`: AWS RDS operations
- `internal/discovery/discovery.go`: Discovery service
- `internal/store/postgres.go`: Database access

**Testing:**
- No test files currently exist in the codebase

**Database:**
- `deployments/docker/migrations/001_base_schema.sql`: Schema definition

## Naming Conventions

**Files:**
- Go files: lowercase with underscores if needed (`cloud_account.go`)
- TSX files: PascalCase for components (`Dashboard.tsx`), lowercase for utilities (`api.ts`)
- SQL files: numbered prefix with snake_case (`001_base_schema.sql`)

**Directories:**
- Go packages: lowercase, single word preferred (`provider`, `store`, `models`)
- Frontend: lowercase, descriptive (`components`, `pages`, `hooks`, `lib`)

**Go Code:**
- Packages: lowercase, match directory name
- Exported types: PascalCase (`CloudAccount`, `Provider`, `Registry`)
- Unexported: camelCase (`parsePeriod`, `getInstanceCost`)
- Interface implementations: type name ends with descriptor (`RDSProvider`, `CloudSQLProvider`)

**Frontend Code:**
- Components: PascalCase matching filename (`Dashboard`, `Navigation`)
- Hooks: camelCase with `use` prefix (no custom hooks yet)
- API functions: camelCase descriptive (`getInstances`, `startInstance`)

## Where to Add New Code

**New Cloud Provider:**
- Implementation: `internal/provider/{provider_name}/{service}.go`
- Must implement `Provider` interface from `internal/provider/provider.go`
- Register in `cmd/server/main.go` during startup

**New API Endpoint:**
- Handler: `internal/api/handlers/{resource}.go`
- Route: Add to `cmd/server/main.go` in `r.Route("/api", ...)` block
- Frontend API: Add method to `web/src/lib/api.ts`

**New Domain Model:**
- Model: Add struct to `internal/models/models.go`
- Database: Add migration to `deployments/docker/migrations/`
- Store: Add store methods to `internal/store/postgres.go` or new store file

**New Frontend Page:**
- Page component: `web/src/pages/{PageName}.tsx`
- Route: Add to `web/src/main.tsx` in Routes block

**New Frontend Component:**
- Reusable: `web/src/components/{ComponentName}.tsx`
- UI primitives: `web/src/components/ui/{component}.tsx`

**New Middleware:**
- Location: `internal/api/middleware/{name}.go`
- Register: Add `r.Use(middleware.Name)` in `cmd/server/main.go`

**Utilities:**
- Go utilities: Create new package in `internal/{name}/`
- Frontend utilities: `web/src/lib/{name}.ts`

## Special Directories

**vendor/:**
- Purpose: Vendored Go dependencies
- Generated: Yes (`go mod vendor`)
- Committed: No (currently empty, may have issues per go.mod warning)

**web/dist/:**
- Purpose: Built frontend assets
- Generated: Yes (`npm run build`)
- Committed: Appears committed (contains index-KWk7TmbP.js)

**web/node_modules/:**
- Purpose: Node.js dependencies
- Generated: Yes (`npm install`)
- Committed: No

**.git/:**
- Purpose: Git repository data
- Generated: Yes
- Committed: No (is the repo itself)

---

*Structure analysis: 2026-02-20*
