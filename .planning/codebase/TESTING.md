# Testing Patterns

**Analysis Date:** 2026-02-20

## Test Status

**Critical Finding: No tests exist in this codebase.**

No `*_test.go` files found in Go code. No test configuration or test files found in the frontend.

## Go Test Framework

**Runner:**
- Go standard testing (`go test`)
- Config: None (uses Go defaults)

**Assertion Library:**
- None configured (would use standard `testing` package)

**Run Commands:**
```bash
make test            # Run all tests (currently no tests)
go test ./... -v     # Direct Go command
```

**Linting:**
```bash
make lint            # golangci-lint run ./...
```

## Frontend Test Framework

**Runner:**
- None configured
- No `jest.config.*` or `vitest.config.*` found

**Expected Setup (not implemented):**
- Vitest (pairs with Vite build tool)
- React Testing Library

**Run Commands:**
```bash
npm run test         # Not configured
```

## Recommended Test Structure

### Go Backend

**Location Pattern:**
- Co-locate tests with source: `internal/store/postgres_test.go`

**Naming:**
- `{filename}_test.go`
- Test functions: `func TestXxx(t *testing.T)`

**Recommended Structure:**
```
internal/
├── api/
│   └── handlers/
│       ├── instances.go
│       └── instances_test.go
├── store/
│   ├── postgres.go
│   └── postgres_test.go
├── provider/
│   ├── aws/
│   │   ├── rds.go
│   │   └── rds_test.go
│   └── gcp/
│       ├── cloudsql.go
│       └── cloudsql_test.go
└── models/
    ├── models.go
    └── models_test.go
```

### TypeScript Frontend

**Location Pattern:**
- Co-locate or in `__tests__` directory

**Recommended Structure:**
```
web/src/
├── __tests__/
│   ├── pages/
│   │   └── InstancesPage.test.tsx
│   └── lib/
│       └── api.test.ts
├── pages/
│   └── InstancesPage.tsx
└── lib/
    └── api.ts
```

## Recommended Test Patterns

### Go Unit Tests

**Table-Driven Tests:**
```go
func TestParsePeriod(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     time.Duration
        wantErr  bool
    }{
        {"1 hour", "1h", time.Hour, false},
        {"24 hours", "24h", 24 * time.Hour, false},
        {"7 days", "7d", 7 * 24 * time.Hour, false},
        {"invalid", "invalid", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parsePeriod(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parsePeriod() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("parsePeriod() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**HTTP Handler Tests:**
```go
func TestGetAllInstances(t *testing.T) {
    handler := NewInstanceHandler(mockStore)
    
    req := httptest.NewRequest("GET", "/api/v1/instances", nil)
    w := httptest.NewRecorder()
    
    handler.GetAllInstances(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

### TypeScript Unit Tests

**Component Tests (with React Testing Library):**
```typescript
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import InstancesPage from './InstancesPage'
import api from '../lib/api'

jest.mock('../lib/api')

test('renders instances list', async () => {
    (api.getInstances as jest.Mock).mockResolvedValue([
        { id: '1', name: 'test-db', status: 'running' }
    ])

    render(
        <BrowserRouter>
            <InstancesPage />
        </BrowserRouter>
    )

    await waitFor(() => {
        expect(screen.getByText('test-db')).toBeInTheDocument()
    })
})
```

**API Client Tests:**
```typescript
import api from './api'

describe('api', () => {
    beforeEach(() => {
        global.fetch = jest.fn()
    })

    test('getInstances fetches from correct endpoint', async () => {
        (fetch as jest.Mock).mockResolvedValue({
            ok: true,
            json: () => Promise.resolve([])
        })

        await api.getInstances()

        expect(fetch).toHaveBeenCalledWith(
            'http://localhost:8080/api/v1/instances',
            expect.objectContaining({
                headers: expect.objectContaining({
                    'Authorization': 'Bearer dev-key'
                })
            })
        )
    })
})
```

## Mocking

### Go Backend

**Interface-Based Mocking:**

Define interfaces for dependencies (already done in codebase):
```go
// internal/scheduler/scheduler.go
type Store interface {
    GetSchedule(id string) (*models.Schedule, error)
    ListSchedules() ([]models.Schedule, error)
    // ...
}
```

Create mock implementations:
```go
type mockStore struct {
    schedules []models.Schedule
    err       error
}

func (m *mockStore) ListSchedules() ([]models.Schedule, error) {
    return m.schedules, m.err
}
```

**HTTP Test Servers:**
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode([]models.Instance{})
}))
defer server.Close()
```

### TypeScript Frontend

**Jest Mocking (recommended setup):**
```typescript
// Mock entire module
jest.mock('../lib/api')

// Mock specific function
(api.getInstances as jest.Mock).mockResolvedValue([...])

// Mock fetch
global.fetch = jest.fn()
```

**What to Mock:**
- API calls (`web/src/lib/api.ts`)
- External services
- Browser APIs (localStorage, window.location)

**What NOT to Mock:**
- React components being tested
- Utility functions (test them directly)
- Simple data transformations

## Fixtures and Factories

### Go Backend

**Test Data Pattern:**
```go
func testInstance() models.Instance {
    return models.Instance{
        ID:           "test-id",
        Name:         "test-db",
        Provider:     "aws",
        Region:       "us-east-1",
        Status:       "running",
        Engine:       "postgres",
        InstanceType: "db.t3.micro",
    }
}

func testSchedule() models.Schedule {
    return models.Schedule{
        ID:        "schedule-1",
        Name:      "Business Hours",
        Timezone:  "America/New_York",
        SleepCron: "0 19 * * 1-5",
        WakeCron:  "0 7 * * 1-5",
        Enabled:   true,
    }
}
```

### TypeScript Frontend

**Factory Pattern:**
```typescript
const createMockInstance = (overrides?: Partial<Instance>): Instance => ({
    id: 'test-id',
    name: 'test-db',
    provider: 'aws',
    region: 'us-east-1',
    status: 'running',
    engine: 'postgres',
    instance_type: 'db.t3.micro',
    managed: true,
    hourly_cost_cents: 50,
    cloud_account_id: 'account-1',
    provider_id: 'arn:aws:...',
    tags: {},
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
})
```

## Coverage

**Requirements:** None enforced (no tests exist)

**Recommended Targets:**
- Go: 70%+ for business logic (`internal/`)
- TypeScript: 60%+ for components and API client

**View Coverage (when tests exist):**
```bash
# Go
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# TypeScript (with vitest)
npm run test -- --coverage
```

## Test Types

### Unit Tests

**Go Scope:**
- Individual functions in isolation
- Store methods with mock database
- Provider methods with mock AWS/GCP clients

**TypeScript Scope:**
- Component rendering
- Event handlers
- API client methods

### Integration Tests

**Go Scope:**
- API handlers with real store (test database)
- Full request/response cycles

**TypeScript Scope:**
- Page components with mocked API
- User interaction flows

### E2E Tests

**Framework:** Not configured

**Recommended:**
- Playwright or Cypress for frontend
- Integration tests against test environment

## Priority Test Targets

Based on codebase analysis, these areas need tests first:

1. **`internal/store/postgres.go`** - Database operations
2. **`internal/provider/aws/rds.go`** - AWS SDK interactions
3. **`internal/scheduler/scheduler.go`** - Schedule matching logic
4. **`internal/analyzer/analyzer.go`** - Pattern detection
5. **`web/src/lib/api.ts`** - API client error handling
6. **`web/src/pages/InstancesPage.tsx`** - Core user interface

## Test Database Setup

**Recommended for Go integration tests:**

```go
func setupTestDB(t *testing.T) (*store.Postgres, func()) {
    testDBURL := os.Getenv("TEST_DATABASE_URL")
    if testDBURL == "" {
        testDBURL = "postgresql://localhost:5432/snoozeql_test?sslmode=disable"
    }
    
    db, err := store.NewPostgres(testDBURL)
    if err != nil {
        t.Fatalf("failed to connect to test db: %v", err)
    }
    
    cleanup := func() {
        db.Close()
    }
    
    return db, cleanup
}
```

---

*Testing analysis: 2026-02-20*
