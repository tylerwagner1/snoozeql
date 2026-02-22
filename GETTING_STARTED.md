# SnoozeQL - Getting Started Guide

## Quick Start (Recommended)

```bash
# Clone the repository (if not already done)
cd snoozeql

# Set up your cloud credentials (optional for demo mode)
export AWS_ACCESS_KEY_ID=your_aws_key
export AWS_SECRET_ACCESS_KEY=your_aws_secret
export GCP_SERVICE_ACCOUNT_JSON=$(cat your_gcp_key.json | base64)

# Start everything with Docker Compose
docker-compose up -d

# Visit the application
# Web UI: http://localhost:3000
# API: http://localhost:8080/api/v1
```

## What Gets Installed

Docker Compose will start:

1. **PostgreSQL** (port 5432) - Database for storing schedules, instances, recommendations
2. **Go API Server** (port 8080) - Backend with all business logic
3. **React Frontend** (port 3000) - Web UI with dashboard

## Development Mode (Alternative)

If you prefer to run components separately for development:

```bash
# Terminal 1: Start PostgreSQL
docker-compose up -d postgres

# Terminal 2: Run Go API server
cd cmd/server
go run main.go

# Terminal 3: Run React dev server
cd web
npm install
npm run dev
```

## Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for development)
- Node.js 20+ (for React development)
- AWS account (for RDS instances)
- GCP account (for Cloud SQL instances)

## Configuration

Create `.env` file in project root:

```bash
# PostgreSQL
POSTGRES_USER=snoozeql
POSTGRES_PASSWORD=snoozeql
POSTGRES_DB=snoozeql
POSTGRES_PORT=5432

# Server
SERVER_PORT=8080
DATABASE_URL=postgresql://snoozeql:snoozeql@postgres:5432/snoozeql?sslmode=disable
AWS_REGION=us-east-1
PRE_STOP_MINUTES=10

# Optional: Cloud Credentials (for actual database management)
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
GCP_SERVICE_ACCOUNT_JSON=your_base64_encoded_json

# Optional: Notifications
SLACK_WEBHOOK_URL=your_slack_webhook
```

## Next Steps

1. **Add Cloud Accounts**: Use the API or UI to register your AWS/GCP accounts
2. **Discover Databases**: Run discovery to find existing databases
3. **Create Schedules**: Set up sleep/wake schedules based on your needs
4. **Review Recommendations**: Check AI-generated schedule suggestions
5. **Monitor Savings**: Track cost savings over time

## Commands

```bash
# View logs
docker-compose logs -f

# Stop the application
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Run migrations manually
docker-compose exec postgres psql -U snoozeql -d snoozeql -f /docker-entrypoint-initdb.d/001_base_schema.sql

# Access database shell
docker-compose exec postgres psql -U snoozeql -d snoozeql

# Build for production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Troubleshooting

**App can't connect to database:**
```bash
docker-compose up -d postgres
docker-compose logs app
```

**React frontend shows errors:**
```bash
cd web
npm run dev  # Check for TypeScript errors
```

**AWS/GCP doesn't show instances:**
- Verify credentials are set correctly
- Check AWS_REGION matches where your RDS instances are
- Verify GCP project ID is correct

## API Documentation

Once the server is running:

-Swagger UI: http://localhost:8080/swagger (if Swagger enabled)
-API Base: http://localhost:8080/api/v1

### Endpoints

- `GET /instances` - List all database instances
- `POST /instances/{id}/start` - Start a database
- `POST /instances/{id}/stop` - Stop a database
- `GET /schedules` - List schedules
- `POST /schedules` - Create schedule
- `GET /recommendations` - Get recommendations
- `POST /instances/discover` - Run discovery

## Support

For issues or questions, check the logs first:
```bash
docker-compose logs
```
