# Coding Conventions

**Analysis Date:** 2026-02-20

## Languages & Stacks

**Backend (Go):**
- Go 1.24.0 (toolchain go1.24.4)
- Located in: `cmd/`, `internal/`

**Frontend (TypeScript/React):**
- TypeScript 5.3.3 with strict mode
- React 18.2.0 with functional components
- Located in: `web/src/`

## Naming Patterns

### Go Backend

**Files:**
- Lowercase single words: `config.go`, `models.go`, `postgres.go`
- Package-specific naming: `internal/provider/aws/rds.go`
- No underscores or hyphens in file names

**Packages:**
- Short, lowercase single words: `config`, `models`, `store`, `provider`, `discovery`
- Sub-packages mirror provider/category: `provider/aws`, `provider/gcp`

**Types:**
- PascalCase for exported types: `CloudAccount`, `Instance`, `Schedule`
- Struct fields: PascalCase with JSON/DB tags
- Pattern: `type CloudAccount struct { ... }`

**Functions:**
- PascalCase for exported: `NewPostgres()`, `ListDatabases()`
- camelCase for unexported: `getEnv()`, `parsePeriod()`, `isManaged()`
- Constructor pattern: `NewXxx()` returns `*Xxx` or `(*Xxx, error)`

**Variables:**
- camelCase: `ctx`, `cfg`, `err`, `db`
- Short names for common types: `r` for request, `w` for writer, `p` for provider

**Constants:**
- PascalCase for exported: `MatchExact`, `MatchContains`
- Grouped in `const` blocks with type

### TypeScript Frontend

**Files:**
- PascalCase for components: `Dashboard.tsx`, `InstancesPage.tsx`, `Navigation.tsx`
- camelCase for utilities: `api.ts`

**Components:**
- PascalCase: `Dashboard`, `InstancesPage`, `CloudAccountsPage`
- Arrow function pattern: `const Dashboard = () => { ... }`
- Default export at end: `export default Dashboard`

**Functions:**
- camelCase: `handleSubmit`, `handleDelete`, `loadAccounts`, `fetchData`
- Event handlers: `handle{Action}` pattern

**Types/Interfaces:**
- PascalCase: `Instance`, `Schedule`, `CloudAccount`, `Recommendation`
- Exported from `web/src/lib/api.ts`

**State Variables:**
- camelCase: `instances`, `loading`, `error`, `isModalOpen`
- Destructured from hooks: `const [instances, setInstances] = useState<Instance[]>([])`

## Code Style

### Go Formatting

**Tools:**
- `gofmt` (standard Go formatting)
- `golangci-lint` (configured in Makefile)

**Import Organization:**
```go
import (
    "context"        // stdlib
    "encoding/json"
    "fmt"
    
    "github.com/go-chi/chi/v5"  // external packages
    
    "snoozeql/internal/models"  // internal packages
)
```

**Struct Tags:**
- JSON and DB tags together: `json:"id" db:"id"`
- Sensitive fields omit JSON: `json:"-" db:"credentials"`

### TypeScript Formatting

**No explicit formatter config detected.** Follow these observed patterns:

**Indentation:** 2 spaces

**Quotes:** Single quotes for imports and strings

**Semicolons:** None (no semicolons style)

**Imports:**
```typescript
import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Activity, Database } from 'lucide-react'
import api from '../lib/api'
import type { Instance } from '../lib/api'
```

**Import Order:**
1. React hooks and core
2. Third-party libraries
3. Local modules (relative paths)
4. Type imports with `type` keyword

## Error Handling

### Go Backend

**Pattern:** Return errors with context wrapping using `fmt.Errorf`:
```go
if err := conn.Ping(); err != nil {
    return nil, fmt.Errorf("failed to ping database: %w", err)
}
```

**HTTP Error Responses:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusInternalServerError)
w.Write([]byte(`{"error":"Failed to list instances"}`))
```

**Logging Errors:**
```go
log.Printf("ERROR listing instances: %v", err)
```

**Warning for non-fatal:**
```go
log.Printf("Warning: Failed to load cloud accounts: %v", err)
```

### TypeScript Frontend

**Pattern:** Try/catch with state updates:
```typescript
try {
    const data = await api.getInstances()
    setInstances(data)
} catch (err) {
    setError('Failed to load instances')
    console.error(err)
} finally {
    setLoading(false)
}
```

**API Error Handling:**
```typescript
if (!response.ok) {
    throw new Error(`API error: ${response.statusText}`)
}
```

## Logging

### Go Backend

**Framework:** Standard library `log` package

**Configuration:**
```go
log.SetOutput(os.Stdout)
log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
```

**Patterns:**
- Success: `log.Printf("SUCCESS: %s", message)`
- Debug: `log.Printf("DEBUG: %s", detail)`
- Error: `log.Printf("ERROR %s: %v", operation, err)`
- Info with emoji: `log.Printf("✓ Connected to database")`

### TypeScript Frontend

**Framework:** Browser `console`

**Patterns:**
- `console.error('Failed to load instances:', err)`
- `console.error(err)` for caught exceptions

## Comments

### Go Backend

**Package Comments:**
```go
// Package provider provides interfaces for cloud database providers
package provider
```

**Type/Function Comments:**
```go
// CloudAccount represents a configured cloud provider account
type CloudAccount struct { ... }

// NewPostgres creates a new PostgreSQL connection
func NewPostgres(url string) (*Postgres, error) { ... }
```

**Inline TODOs:**
```go
// TODO: Add store dependency
// TODO: Implement database sync in Phase 3
```

### TypeScript Frontend

**Inline Comments:** Minimal, mostly self-documenting code

## Function Design

### Go Backend

**Size:** Functions typically 10-50 lines

**Parameters:**
- Context first: `func (s *Store) GetRecommendation(ctx context.Context, id string)`
- Receiver methods on structs: `func (h *InstanceHandler) GetAllInstances(w http.ResponseWriter, r *http.Request)`

**Return Values:**
- Single value or `(value, error)` tuple
- Nil for errors when no error: `return &Postgres{...}, nil`

### TypeScript Frontend

**Size:** Components 50-200 lines

**Component Props:** Destructured in function signature when needed

**Hooks:** At top of component, in order: useState, useEffect, custom

## Module Design

### Go Backend

**Package Exports:**
- All public types, functions, constants are PascalCase
- Private helpers are camelCase

**Interface Pattern:** Define interfaces in the package that uses them:
```go
// In scheduler/scheduler.go
type Store interface {
    GetSchedule(id string) (*models.Schedule, error)
    ListSchedules() ([]models.Schedule, error)
    // ...
}
```

**Constructor Pattern:**
```go
func NewScheduler(store Store, registry *provider.Registry) *Scheduler {
    return &Scheduler{
        store:    store,
        registry: registry,
    }
}
```

### TypeScript Frontend

**Exports:**
- Default export for components: `export default Dashboard`
- Named exports for types: `export interface Instance { ... }`
- API object as default export: `export default api`

**Barrel Files:** Not used - direct imports

## API Response Patterns

### Go Backend

**JSON Responses:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(data)
```

**Success with message:**
```go
w.Write([]byte(fmt.Sprintf(`{"success":true,"instance_id":"%s","status":"starting"}`, instanceID)))
```

**Error response:**
```go
w.Write([]byte(`{"error":"Not found"}`))
```

### TypeScript Frontend

**API Client Pattern:**
```typescript
const api = {
    async get<T>(path: string): Promise<T> { ... },
    async post<T>(path: string, body?: unknown): Promise<T> { ... },
    
    // Domain methods use generic helpers
    getInstances: () => api.get<Instance[]>('/instances'),
    startInstance: (id: string) => api.post<void>(`/instances/${id}/start`),
}
```

## Configuration Conventions

### Go Backend

**Environment Variables:**
- SCREAMING_SNAKE_CASE: `DATABASE_URL`, `AWS_ACCESS_KEY_ID`, `SERVER_PORT`
- Loaded via helper functions: `getEnv()`, `getEnvBool()`, `getEnvInt()`
- Defaults provided in code

**Config Struct Fields:**
- Snake_case with underscores: `Server_host`, `Database_url`, `AWS_access_key`
- Not idiomatic Go but consistent in codebase

### TypeScript Frontend

**Constants:**
- SCREAMING_SNAKE_CASE: `const API_BASE_URL = 'http://localhost:8080/api/v1'`
- Hardcoded at top of files

## React Component Patterns

**Functional Components Only:**
```typescript
const InstancesPage = () => {
    const [instances, setInstances] = useState<Instance[]>([])
    // ...
}
```

**State Management:**
- Local state with `useState`
- No Redux or context providers

**Data Fetching:**
```typescript
useEffect(() => {
    const fetchData = async () => {
        try {
            const data = await api.getInstances()
            setInstances(data)
        } catch (err) {
            setError('...')
        } finally {
            setLoading(false)
        }
    }
    fetchData()
}, [])
```

**Styling:**
- Tailwind CSS classes inline
- No CSS modules or styled-components
- `clsx` and `tailwind-merge` for conditional classes

---

*Convention analysis: 2026-02-20*
