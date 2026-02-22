# SnoozeQL Development Makefile

.PHONY: all build run test clean deps docker-build docker-up docker-down migrate

all: build

# Build the Go binary
build:
	@echo "Building SnoozeQL..."
	go build -o bin/snoozeql ./cmd/server

# Run the application
run: build
	@echo "Starting SnoozeQL..."
	./bin/snoozeql

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean -cache
	go clean -testcache

# Run database migrations
migrate:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	go run cmd/migrate/main.go down

# Create new migration
migrate-create:
	@echo "Creating new migration: $(NAME)"
	migrate create -ext sql -dir deployments/docker/migrations -seq $(NAME)

# Run Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t snoozeql:latest .

# Run Docker containers
docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

# Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

# Install Go tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/psampaz/go-mod-outdated@latest
