.PHONY: build run test load-test clean docker-up docker-down db-shell lint

# Variables
APP_NAME=notes-api
DOCKER_COMPOSE=docker-compose

# Build
build:
	@echo "Building Go application..."
	go build -o bin/$(APP_NAME) ./cmd/server

# Run
run: build
	@echo "Starting application..."
	./bin/$(APP_NAME)

# Test
test:
	@echo "Running tests..."
	go test -v ./...

# Load test
load-test:
	@echo "Running load tests..."
	go run scripts/load_test.go

# Database
docker-up:
	@echo "Starting PostgreSQL..."
	$(DOCKER_COMPOSE) up -d postgres
	@echo "Waiting for database to be ready..."
	sleep 5

docker-down:
	@echo "Stopping PostgreSQL..."
	$(DOCKER_COMPOSE) down

db-shell:
	@echo "Connecting to database..."
	$(DOCKER_COMPOSE) exec postgres psql -U user -d notes

# Diagnostics
db-diag:
	@echo "Running diagnostics..."
	$(DOCKER_COMPOSE) exec postgres psql -U user -d notes -f /docker-entrypoint-initdb.d/diagnostic.sql

# Clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Lint
lint:
	@echo "Linting..."
	gofmt -d .
	go vet ./...