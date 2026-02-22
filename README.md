# SnoozeQL

SnoozeQL is a database management application that puts databases to sleep in dev/test environments to save cloud costs.

## Features

- **Multi-Cloud Support**: AWS RDS and GCP Cloud SQL
- **Manual Control**: Start/stop databases on demand
- **Automated Scheduling**: Define sleep/wake schedules with cron expressions
- **AI Recommendations**: Intelligent schedule suggestions based on activity patterns
- **Cost Tracking**: Monitor savings over time
- **Web UI**: Modern React-based interface for easy management

## Project Structure

```
snoozeql/
├── cmd/server/              # Go API server
├── internal/                # Core application logic
│   ├── api/                 # HTTP handlers and middleware
│   ├── analyzer/            # Activity pattern detection
│   ├── config/              # Configuration management
│   ├── discovery/           # Database discovery service
│   ├── models/              # Data models
│   ├── provider/            # Cloud provider implementations
│   │   ├── aws/            # AWS RDS
│   │   └── gcp/            # GCP Cloud SQL
│   ├── scheduler/          # Schedule matching and execution
│   └── store/              # Database access layer
├── deployments/            # Deployment files
│   ├── docker/            # Docker Compose setup
│   └── kubernetes/        # Kubernetes manifests
├── web/                    # React frontend
├── Dockerfile              # Production Docker image
├── Dockerfile.dev          # Development Docker image
├── Dockerfile.frontend     # Frontend Docker image
├── docker-compose.yml      # Docker Compose configuration
├── go.mod / go.sum         # Go dependencies
└── Makefile                # Build and run commands
```

## Quick Start

```bash
# Start all services with Docker Compose
docker-compose up -d

# Access the web UI
open http://localhost:3000
```

## Development

```bash
# Start PostgreSQL only
docker-compose up -d postgres

# Run Go API server
cd cmd/server && go run main.go

# Run React frontend (in another terminal)
cd web && npm install && npm run dev
```

## Configuration

Create a `.env` file with:

```bash
# PostgreSQL
POSTGRES_USER=snoozeql
POSTGRES_PASSWORD=snoozeql
POSTGRES_DB=snoozeql

# Server
SERVER_PORT=8080
AWS_REGION=us-east-1

# Optional: Cloud credentials
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
GCP_SERVICE_ACCOUNT_JSON=your_json
```

## Building

```bash
# Build Docker images
docker build -t snoozeql:latest -f Dockerfile .
docker build -t snoozeql-frontend:latest -f Dockerfile.frontend .

# Build Go binary
make build

# Build frontend
cd web && npm run build
```

## API Documentation

Once running, visit:
- Swagger UI: http://localhost:8080/swagger (if enabled)
- API Base: http://localhost:8080/api/v1

### Endpoints

- `GET /api/v1/instances` - List databases
- `POST /api/v1/instances/{id}/start` - Start database
- `POST /api/v1/instances/{id}/stop` - Stop database
- `GET /api/v1/schedules` - List schedules
- `POST /api/v1/schedules` - Create schedule
- `GET /api/v1/recommendations` - Get AI recommendations

## Database Schema

See `deployments/docker/migrations/001_base_schema.sql` for the complete schema including:
- `instances` - Database instance metadata
- `schedules` - Sleep/wake schedules
- `recommendations` - AI-generated suggestions
- `overrides` - Temporary manual overrides
- `events` - Audit log
- `savings` - Cost savings tracking

## Supported Cloud Providers

### AWS RDS
- Start/stop instances using `StartDBInstance`/`StopDBInstance`
- Activity detection via CloudWatch metrics
- Cost calculation based on instance type

### GCP Cloud SQL  
- Start/stop using `activationPolicy` field (ALWAYS/NEVER)
- Activity detection (Cloud Monitoring integration pending)
- Cost calculation based on instance configuration

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Support

For issues and questions, please open an issue on GitHub.
