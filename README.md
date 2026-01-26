# Go Asynq Load Test

Background job processing with Go, Asynq, and Redis. Includes K6 load testing.

## 🎯 Features

- ✅ **REST API** - Gin framework with clean architecture
- ✅ **Background Jobs** - Asynq task queue with Redis
- ✅ **Priority Queues** - Critical, default, and low priority
- ✅ **Task Scheduling** - Delayed and periodic tasks
- ✅ **Load Testing** - K6 scripts for performance testing
- ✅ **Monitoring** - Asynqmon dashboard, Prometheus metrics
- ✅ **Docker Support** - Multi-container setup with Docker Compose

## 🏗️ Architecture

```
┌─────────────┐     ┌──────────┐     ┌─────────────┐
│   HTTP API  │────▶│  Redis   │◀────│   Workers   │
│  (Producer) │     │ (Asynq)  │     │ (Consumers) │
└─────────────┘     └──────────┘     └─────────────┘
```

## 📂 Project Structure

```
go-asynq-loadtest/
├── cmd/
│   ├── api/              # HTTP API server (producer)
│   └── worker/           # Background workers (consumer)
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Business entities
│   ├── dto/              # Request/Response DTOs
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   └── tasks/            # Asynq task definitions
├── pkg/
│   ├── logger/           # Logging utilities
│   └── monitoring/       # Metrics & monitoring
├── loadtest/             # K6 load test scripts
├── migrations/           # Database migrations
├── docker-compose.yml    # Multi-container setup
├── Makefile              # Build automation
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience)

### 1. Clone Repository

```bash
git clone https://github.com/lppduy/go-asynq-loadtest.git
cd go-asynq-loadtest
```

### 2. Copy Environment Variables

```bash
cp .env.example .env
```

### 3. Start Infrastructure

**Option A: Using Makefile** (Linux/macOS)
```bash
make docker-up
```

**Option B: Direct Command** (All platforms)
```bash
docker-compose up -d
```

### 4. Run API Server

**Option A: Using Makefile**
```bash
make run-api
```

**Option B: Direct Command**
```bash
go run cmd/api/main.go
```

### 5. Run Worker (in another terminal)

**Option A: Using Makefile**
```bash
make run-worker
```

**Option B: Direct Command**
```bash
go run cmd/worker/main.go
```

### 6. Access Services

- **API**: http://localhost:8080
- **Asynqmon**: http://localhost:8085 (Monitor tasks & queues)
- **Prometheus**: http://localhost:9090 (Metrics)
- **Grafana**: http://localhost:3000 (Dashboards - admin/admin)

## 📊 Load Testing

### Run K6 Tests

**Option A: Using Makefile**
```bash
# Basic load test
make loadtest

# Stress test
make loadtest-stress

# Spike test
make loadtest-spike
```

**Option B: Direct Commands**
```bash
# Basic load test
k6 run loadtest/basic-load.js

# Stress test
k6 run loadtest/stress-test.js

# Spike test
k6 run loadtest/spike-test.js

# Soak test (long-running)
k6 run loadtest/soak-test.js
```

### Test Scenarios

1. **Basic Load** - Ramp up to 100 RPS, sustained load
2. **Stress Test** - Gradually increase load to find breaking point
3. **Spike Test** - Sudden traffic surge simulation
4. **Soak Test** - Long-running stability test (30+ minutes)

## 🛠️ Development

> **Note:** Makefile commands are provided for convenience on Linux/macOS. 
> Windows users can use the direct commands or install [Make for Windows](https://gnuwin32.sourceforge.net/packages/make.htm) / use WSL.

### Install Dependencies

**Option A: Using Makefile**
```bash
make install
```

**Option B: Direct Command**
```bash
go mod download
go mod tidy
```

### Run Tests

**Option A: Using Makefile**
```bash
make test                # Run tests
make test-coverage       # Run tests with coverage report
```

**Option B: Direct Command**
```bash
go test -v ./...                              # Run tests
go test -coverprofile=coverage.out ./...      # With coverage
go tool cover -html=coverage.out              # View coverage in browser
```

### Build

**Option A: Using Makefile**
```bash
make build               # Build binaries to bin/
```

**Option B: Direct Command**
```bash
mkdir -p bin
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker
```

### Format Code

**Option A: Using Makefile**
```bash
make fmt
```

**Option B: Direct Command**
```bash
go fmt ./...
gofmt -s -w .
```

### Lint

**Option A: Using Makefile**
```bash
make lint
```

**Option B: Direct Command**
```bash
golangci-lint run
```

### Stop All Services

**Option A: Using Makefile**
```bash
make docker-down              # Stop services
make docker-down-volumes      # Stop and remove volumes
```

**Option B: Direct Command**
```bash
docker-compose down           # Stop services
docker-compose down -v        # Stop and remove volumes
```

## 📈 Monitoring

### Asynqmon Dashboard

Web UI for monitoring tasks, queues, and workers:
- View active, scheduled, and failed tasks
- Retry or delete tasks manually
- Monitor queue depth and worker status

### Prometheus Metrics

```
asynq_tasks_enqueued_total
asynq_tasks_processed_total
asynq_task_duration_seconds
asynq_queue_size
asynq_active_workers
```

### Grafana Dashboards

Import pre-built dashboards for:
- Task throughput
- Queue depth
- Latency percentiles
- Error rates

## 🔧 Configuration

Copy `.env.example` to `.env` and configure:

```env
# Server
SERVER_PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taskqueue
DB_USER=admin
DB_PASSWORD=secret

# Redis
REDIS_ADDR=localhost:6379

# Worker
WORKER_CONCURRENCY=20
```

## 📚 Documentation

- [Architecture](docs/architecture.md)
- [API Reference](docs/api.md)
- [Load Testing Guide](docs/loadtest.md)
- [Deployment](docs/deployment.md)

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 👤 Author

**Duy Le**
- GitHub: [@lppduy](https://github.com/lppduy)

## 🌟 Acknowledgments

- [Asynq](https://github.com/hibiken/asynq) - Simple, reliable task queue
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [K6](https://k6.io/) - Load testing tool
