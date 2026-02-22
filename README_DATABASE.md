# Database: SnoozeQL

PostgreSQL database schema for SnoozeQL application.

## Prerequisites

- PostgreSQL 14+ installed and running
- Access to create tables, indexes, and triggers

## Installation

### Using the migration script

```bash
# Run the migration
psql -U your_user -d postgres -f migrations/001_base_schema.sql
```

### Or using Go

```bash
# Build and run with migrations
go run cmd/server/main.go
```

## Database Configuration

The application expects a PostgreSQL connection string in this format:

```
postgresql://user:password@host:port/database?sslmode=[disable|require]
```

Or for Unix socket:

```
postgresql:///database?host=/path/to/socket
```

## Schema Overview

```
cloud_accounts ──> instances
                    │
                    ├─> schedules (via selectors)
                    ├─> recommendations
                    ├─> overrides
                    ├─> events
                    └─> savings

settings (global)
api_keys
```

## Tables

| Table | Description |
|-------|-------------|
| `cloud_accounts` | Configured cloud provider accounts (AWS/GCP) |
| `instances` | Discovered database instances |
| `schedules` | Sleep/wake schedules with dynamic selectors |
| `settings` | Application configuration and thresholds |
| `recommendations` | AI-generated schedule suggestions |
| `overrides` | Temporary manual overrides |
| `events` | Audit log of state changes |
| `savings` | Daily aggregated cost savings |
| `api_keys` | API authentication keys |

## Migrations

See [migrations/README.md](migrations/README.md) for migration instructions.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://localhost:5432/snoozeql?sslmode=disable` |
| `SERVER_HOST` | Server listen address | `0.0.0.0` |
| `SERVER_PORT` | Server listen port | `8080` |
| `AWS_REGION` | Default AWS region | `us-east-1` |
| `GCP_PROJECT` | GCP project ID | (empty) |
| `DISCOVERY_ENABLED` | Enable auto-discovery | `true` |
| `DISCOVERY_INTERVAL_HOURS` | Discovery scan interval | `6` |
| `SLACK_WEBHOOK_URL` | Slack webhook for notifications | (empty) |
| `SLACK_APP_TOKEN` | Slack app token for interactive buttons | (empty) |
| `PRE_STOP_MINUTES` | Minutes before scheduled stop to warn | `10` |

## License

MIT License - see LICENSE file for details.
