# Technology Stack

**Analysis Date:** 2026-02-20

## Languages

**Primary:**
- Go 1.24.0 - Backend API server (`cmd/server/main.go`, `internal/`)
- TypeScript 5.3.3 - Frontend React application (`web/src/`)

**Secondary:**
- SQL - Database migrations and schema (`deployments/docker/migrations/`)

## Runtime

**Environment:**
- Go 1.24.0 with toolchain go1.24.4
- Node.js (version not pinned, no .nvmrc)

**Package Manager:**
- Go modules (`go.mod`, `go.sum`)
- npm for frontend (`web/package.json`)
- Lockfile: `go.sum` present, `package-lock.json` not committed

## Frameworks

**Core:**
- Chi v5.2.0 - HTTP router and middleware (`github.com/go-chi/chi/v5`)
- React 18.2.0 - Frontend UI (`web/src/`)
- React Router DOM 6.20.0 - Client-side routing

**Testing:**
- Go standard `testing` package
- No frontend testing framework configured

**Build/Dev:**
- Vite 5.0.8 - Frontend bundler and dev server
- Docker - Containerization (`Dockerfile`, `Dockerfile.dev`, `Dockerfile.frontend`)
- Docker Compose - Multi-container orchestration (`docker-compose.yml`)

## Key Dependencies

**Critical - Backend:**
- `github.com/aws/aws-sdk-go-v2` v1.41.1 - AWS API integration
- `google.golang.org/api` v0.267.0 - GCP API integration
- `github.com/jackc/pgx/v5` v5.8.0 - PostgreSQL driver

**Critical - Frontend:**
- `react` 18.2.0 - UI framework
- `recharts` 2.10.0 - Dashboard charting
- `lucide-react` 0.300.0 - Icon library

**Infrastructure:**
- `tailwindcss` 3.4.0 - CSS framework
- `clsx` + `tailwind-merge` - CSS class utilities
- `go.opentelemetry.io/otel` - Observability (indirect dependency)

## Configuration

**Environment:**
- Environment variables via `os.Getenv()` in `internal/config/config.go`
- No `.env` file in repository (uses docker-compose environment)
- Key config loaded at startup with defaults

**Required Environment Variables:**
| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | `postgresql://localhost:5432/snoozeql?sslmode=disable` | PostgreSQL connection |
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `AWS_REGION` | `us-east-1` | Default AWS region |
| `AWS_ACCESS_KEY_ID` | (empty) | AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | (empty) | AWS credentials |
| `GCP_PROJECT` | (empty) | GCP project ID |
| `SLACK_WEBHOOK_URL` | (empty) | Slack notifications |
| `SLACK_APP_TOKEN` | (empty) | Slack App integration |
| `PRE_STOP_MINUTES` | `10` | Pre-stop notification lead time |
| `DISCOVERY_ENABLED` | `true` | Enable instance discovery |
| `DISCOVERY_INTERVAL_SECONDS` | `30` | Discovery polling interval |
| `AGGREGATION_ENABLED` | `true` | Enable notification aggregation |
| `CURRENCY` | `USD` | Cost display currency |

**Build:**
- `Makefile` - Development commands
- `go.mod` - Go module definition
- `web/package.json` - Frontend dependencies and scripts

**Build Commands:**
```bash
make build              # Build Go binary to bin/snoozeql
make run                # Build and run
make test               # Run Go tests
make lint               # Run golangci-lint
make docker-up          # Start Docker containers
make docker-down        # Stop Docker containers
```

**Frontend Commands:**
```bash
npm run dev             # Start Vite dev server
npm run build           # Build for production
npm run lint            # Run ESLint
```

## Platform Requirements

**Development:**
- Go 1.24+
- Node.js (18+ recommended)
- Docker & Docker Compose
- PostgreSQL 15 (via Docker)
- golangci-lint (optional, for linting)

**Production:**
- Alpine Linux container
- PostgreSQL 15+
- AWS/GCP credentials for cloud provider access
- Port 8080 (API), Port 3001 (frontend)

## Build Artifacts

**Backend:**
- `bin/snoozeql` - Compiled Go binary (via make build)
- `server` - Pre-built binary in repo root

**Frontend:**
- `web/dist/` - Vite production build output

---

*Stack analysis: 2026-02-20*
