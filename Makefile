.PHONY: help build run-api run-worker test clean docker-up docker-down loadtest fmt lint install

# Variables
APP_NAME=asynq-loadtest
API_BINARY=bin/api
WORKER_BINARY=bin/worker
DOCKER_IMAGE=$(APP_NAME):latest

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## install: Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

## build: Build API and Worker binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o $(API_BINARY) ./cmd/api
	go build -o $(WORKER_BINARY) ./cmd/worker
	@echo "✅ Build complete: $(API_BINARY), $(WORKER_BINARY)"

## run-api: Run API server
run-api:
	@echo "Starting API server..."
	go run ./cmd/api/main.go

## run-worker: Run worker
run-worker:
	@echo "Starting worker..."
	go run ./cmd/worker/main.go

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting services..."
	docker-compose up -d
	@echo "✅ Services started"
	@echo "   - API: http://localhost:8080"
	@echo "   - Asynqmon: http://localhost:8085"
	@echo "   - Prometheus: http://localhost:9090"
	@echo "   - Grafana: http://localhost:3000"

## docker-down: Stop all services
docker-down:
	@echo "Stopping services..."
	docker-compose down
	@echo "✅ Services stopped"

## docker-down-volumes: Stop services and remove volumes
docker-down-volumes:
	@echo "Stopping services and removing volumes..."
	docker-compose down -v
	@echo "✅ Services stopped and volumes removed"

## docker-logs: View logs
docker-logs:
	docker-compose logs -f

## docker-build: Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker build -t $(DOCKER_IMAGE) .

## loadtest: Run K6 load test
loadtest:
	@echo "Running load test..."
	k6 run loadtest/basic-load.js

## loadtest-stress: Run stress test
loadtest-stress:
	@echo "Running stress test..."
	k6 run loadtest/stress-test.js

## loadtest-spike: Run spike test
loadtest-spike:
	@echo "Running spike test..."
	k6 run loadtest/spike-test.js

## dev: Run in development mode with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	air

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	go mod tidy

## migrate-up: Run database migrations
migrate-up:
	@echo "Running migrations..."
	# TODO: Add migration command

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	# TODO: Add migration rollback command
