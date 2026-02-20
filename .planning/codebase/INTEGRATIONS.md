# External Integrations

**Analysis Date:** 2026-02-20

## APIs & External Services

**AWS RDS:**
- Purpose: Database instance discovery, start/stop operations, metrics
- SDK: `github.com/aws/aws-sdk-go-v2` v1.41.1
- Implementation: `internal/provider/aws/rds.go`
- Auth: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` env vars or instance role
- Operations:
  - `DescribeDBInstances` - List all RDS instances
  - `StartDBInstance` - Start stopped instance
  - `StopDBInstance` - Stop running instance
- Metrics: CPU, connections, IOPS (stub implementations)

**GCP Cloud SQL:**
- Purpose: Database instance discovery, start/stop operations
- SDK: `google.golang.org/api/sqladmin/v1`
- Implementation: `internal/provider/gcp/cloudsql.go`
- Auth: GCP Application Default Credentials
- Operations:
  - `Instances.List` - List Cloud SQL instances
  - `Instances.Patch` - Start/stop via ActivationPolicy
  - `Instances.Get` - Get instance status
- Metrics: Not implemented (stubs return error)

**Slack (Planned):**
- Purpose: Pre-stop notifications
- Config: `SLACK_WEBHOOK_URL`, `SLACK_APP_TOKEN` env vars
- Implementation: `internal/notifier/` (directory exists but empty)
- Status: Not implemented

## Data Storage

**Databases:**
- PostgreSQL 15
  - Connection: `DATABASE_URL` env var
  - Client: `github.com/jackc/pgx/v5` via `database/sql`
  - Implementation: `internal/store/postgres.go`
  - Schema: `deployments/docker/migrations/001_base_schema.sql`

**Tables:**
| Table | Purpose |
|-------|---------|
| `cloud_accounts` | Configured AWS/GCP accounts with credentials (JSONB) |
| `instances` | Discovered database instances |
| `schedules` | Sleep/wake schedules with cron patterns |
| `settings` | Application configuration (JSONB) |
| `recommendations` | AI-suggested schedules |
| `overrides` | Manual keep-alive/skip-next overrides |
| `events` | Audit log of state changes |
| `savings` | Daily cost savings aggregations |
| `api_keys` | API authentication keys |

**File Storage:**
- None (no file uploads)

**Caching:**
- None (no Redis/Memcached)

## Authentication & Identity

**API Authentication:**
- Bearer token in `Authorization` header
- Implementation: `internal/api/middleware/chi.go`
- Development bypass: Token `dev-key` skips validation
- Production: Validates against `api_keys` table (not fully implemented)

**Cloud Provider Auth:**
- AWS: Static credentials from `cloud_accounts.credentials` JSONB field
  - Falls back to instance role/SSO if credentials empty
- GCP: Application Default Credentials (ADC)

## Monitoring & Observability

**Error Tracking:**
- None (no Sentry/Rollbar)

**Logs:**
- Go standard `log` package
- Format: Timestamp + microseconds + short file (`log.LstdFlags | log.Lmicroseconds | log.Lshortfile`)
- Output: stdout
- Debug logging via `log.Printf("DEBUG: ...")` pattern

**Metrics:**
- OpenTelemetry instrumentation available (indirect dependency)
- Not actively configured

## CI/CD & Deployment

**Hosting:**
- Docker containers (development and production)
- Kubernetes manifests: `deployments/kubernetes/` (directory exists)

**CI Pipeline:**
- None detected (no GitHub Actions, GitLab CI, etc.)

**Container Images:**
- `Dockerfile` - Production multi-stage build
- `Dockerfile.dev` - Development with live reload
- `Dockerfile.frontend` - Frontend nginx container

## Environment Configuration

**Required env vars for production:**
```bash
DATABASE_URL=postgresql://user:pass@host:5432/snoozeql
AWS_ACCESS_KEY_ID=...     # Or use instance role
AWS_SECRET_ACCESS_KEY=... # Or use instance role
```

**Optional env vars:**
```bash
GCP_PROJECT=...           # For GCP Cloud SQL support
SLACK_WEBHOOK_URL=...     # For notifications (not implemented)
DISCOVERY_INTERVAL_SECONDS=30
PRE_STOP_MINUTES=10
```

**Secrets location:**
- Cloud credentials stored in PostgreSQL `cloud_accounts.credentials` JSONB column
- Connection strings via environment variables
- No external secrets manager integration

## Webhooks & Callbacks

**Incoming:**
- None (no webhook endpoints)

**Outgoing:**
- Slack webhook (planned, not implemented)

## Provider Registry

The application uses a provider registry pattern for cloud integrations:

**Location:** `internal/provider/`

**Interface:**
```go
type Provider interface {
    ListDatabases(ctx context.Context) ([]models.Instance, error)
    StartDatabase(ctx context.Context, id string) error
    StopDatabase(ctx context.Context, id string) error
    GetDatabaseStatus(ctx context.Context, id string) (string, error)
    GetMetrics(ctx context.Context, providerName, id, period string) (map[string]any, error)
}
```

**Registered Providers:**
- `aws_{region}` - AWS RDS provider per region
- GCP Cloud SQL provider (not auto-registered from accounts)

**Registration Flow:**
1. Load `cloud_accounts` from database at startup
2. For each AWS account, create `RDSProvider` per region
3. Register with key `aws_{region}`
4. Discovery service polls all registered providers

---

*Integration audit: 2026-02-20*
