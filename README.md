# Go Asynq Load Test

Background job processing with Go, Asynq, and Redis. Includes K6 load testing.

**Use Case:** E-commerce Order Processing System

## 🎯 Features

- ✅ **REST API** - Order processing with Gin framework
- ✅ **Background Jobs** - Async payment, email, inventory, invoice generation (Asynq)
- ✅ **Priority Queues** - Critical (payment), high (inventory), default (email), low (analytics)
- ✅ **Task Retries** - Automatic retry with exponential backoff
- ✅ **Worker Pool** - Configurable concurrent workers (default: 20)
- ✅ **Load Testing** - K6 scripts for performance testing
- ✅ **Monitoring** - Asynqmon dashboard, Prometheus metrics
- ✅ **Docker Support** - Multi-container setup with Docker Compose
- ✅ **PostgreSQL + GORM** - Persistent data storage with ORM

## ⚠️ POC vs Production Implementation

### Current Status: **Proof of Concept / Demo**

This project demonstrates Asynq task queue architecture and patterns. Task handlers currently **simulate** external service calls for demonstration purposes with `time.Sleep()` and mock responses.

See inline comments in `internal/tasks/*.go` for specific integration recommendations.

## 🏗️ Architecture

```
                  HTTP Request
                       ↓
┌──────────────────────────────────────────────────┐
│                 API Server (Gin)                  │
│  POST /orders → Create order (50ms response)     │
└─────────────────────┬────────────────────────────┘
                      │
                      ↓ Enqueue background tasks
┌──────────────────────────────────────────────────┐
│              Redis (Task Queue)                   │
│  [Critical] payment:process                       │
│  [High]     inventory:update                      │
│  [Default]  email:confirmation, invoice:generate  │
│  [Low]      analytics:track, warehouse:notify     │
└─────────────────────┬────────────────────────────┘
                      │
                      ↓ Process async
┌──────────────────────────────────────────────────┐
│             Workers (Background)                  │
│  • Process payment (2s)                           │
│  • Update inventory (500ms)                       │
│  • Send confirmation email (1s)                   │
│  • Generate invoice PDF (3s)                      │
│  • Track analytics (200ms)                        │
│  • Notify warehouse (500ms)                       │
└──────────────────────────────────────────────────┘
```

## 📂 Project Structure

```
go-asynq-loadtest/
├── cmd/
│   ├── api/              # HTTP API server (order processing)
│   └── worker/           # Background workers (payment, email, etc)
├── internal/
│   ├── domain/           # Order, OrderItem, Address models
│   ├── dto/              # Request/Response DTOs
│   ├── handler/          # HTTP handlers (order_handler.go)
│   ├── repository/       # In-memory data storage
│   ├── service/          # Business logic (order_service.go)
│   ├── middleware/       # HTTP middleware (auth, logging, CORS)
│   ├── tasks/            # Asynq task definitions & handlers
│   └── config/           # Configuration management
├── pkg/
│   ├── logger/           # Structured logging
│   └── monitoring/       # Prometheus metrics
├── loadtest/             # K6 load test scripts
├── migrations/           # Database migrations (future)
├── docker-compose.yml    # Multi-container setup
├── Makefile              # Build automation
└── README.md
```

## 📡 API Endpoints

### Orders
- `POST /api/v1/orders` - Create new order
- `GET /api/v1/orders` - List all orders (query: `?customer_id=xxx`)
- `GET /api/v1/orders/:id` - Get order details
- `GET /api/v1/orders/:id/status` - Get order status
- `POST /api/v1/orders/:id/cancel` - Cancel order

### Health
- `GET /health` - Health check endpoint

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

### 7. Open Asynqmon Dashboard

Visit http://localhost:8085 to monitor background tasks in real-time.

### 8. Test API with cURL

```bash
# Health check
curl http://localhost:8080/health

# Create order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "customer_email": "customer@example.com",
    "items": [
      {
        "product_id": "prod-1",
        "product_name": "Laptop",
        "quantity": 1,
        "unit_price": 1200.00
      },
      {
        "product_id": "prod-2",
        "product_name": "Mouse",
        "quantity": 2,
        "unit_price": 25.00
      }
    ],
    "shipping_address": {
      "street": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94102",
      "country": "USA"
    },
    "payment_method": "credit_card",
    "notes": "Please deliver before 5 PM"
  }'

# List orders
curl http://localhost:8080/api/v1/orders

# Get order by ID (replace ORD-xxx with actual order ID)
curl http://localhost:8080/api/v1/orders/ORD-12345678

# Get order status
curl http://localhost:8080/api/v1/orders/ORD-12345678/status

# Cancel order
curl -X POST http://localhost:8080/api/v1/orders/ORD-12345678/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "Customer changed their mind"}'
```

## 🧪 Complete Testing Guide

For detailed end-to-end testing instructions, see [TESTING.md](TESTING.md).

Quick test:
```bash
# Terminal 1: Start API
go run cmd/api/main.go

# Terminal 2: Start Worker
go run cmd/worker/main.go

# Terminal 3: Create order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"test","customer_email":"test@example.com",...}'

# Watch logs in Terminal 1 & 2 to see tasks being processed!
```

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

## 🎚️ Priority Queues - Deep Dive

### How Priority Works in Asynq

Asynq uses a **weighted priority system**, not strict priority. This ensures low-priority tasks eventually get processed.

### Priority Configuration

In `cmd/worker/main.go`:

```go
srv := asynq.NewServer(
    redisOpt,
    asynq.Config{
        Concurrency: 20,  // Total number of concurrent workers
        
        Queues: map[string]int{
            "critical": 6,  // Weight: 6
            "high":     4,  // Weight: 4
            "default":  2,  // Weight: 2
            "low":      1,  // Weight: 1
        },
    },
)
```

### How Weights Work

**Weights = Time Allocation Ratio** (NOT strict priority)

```
Total Weight = 6 + 4 + 2 + 1 = 13

critical queue: 6/13 = ~46% of worker time
high queue:     4/13 = ~31% of worker time
default queue:  2/13 = ~15% of worker time
low queue:      1/13 = ~8% of worker time
```

### Example Scenario

With `Concurrency: 20` and weights `{critical:6, high:4, default:2, low:1}`:

```
┌─────────────────────────────────────────────────┐
│  20 Workers Polling Redis                       │
├─────────────────────────────────────────────────┤
│  ⚡ 9-10 workers  → Poll critical queue (46%)   │
│  🔥 6-7 workers   → Poll high queue (31%)       │
│  📧 3 workers     → Poll default queue (15%)    │
│  📊 1-2 workers   → Poll low queue (8%)         │
└─────────────────────────────────────────────────┘
```

**Key Points:**
- ✅ **Not blocking**: Low-priority tasks still get processed
- ✅ **Fair**: Prevents starvation
- ✅ **Flexible**: Adjust weights based on business needs
- ⚠️ **Not strict**: Critical tasks don't always go first

### Setting Task Priority

#### Method 1: Set Queue When Creating Task

```go
// In internal/tasks/payment.go
func NewPaymentProcessTask(...) (*asynq.Task, error) {
    return asynq.NewTask(
        TypePaymentProcess,
        payload,
        asynq.Queue("critical"),  // ← Specify queue here
        asynq.MaxRetry(3),
        asynq.Timeout(30*time.Second),
    ), nil
}
```

#### Method 2: Set Queue When Enqueueing

```go
// In handler
task, _ := tasks.NewPaymentProcessTask(...)

// Override queue at enqueue time
info, err := client.Enqueue(
    task,
    asynq.Queue("critical"),  // ← Can override here
)
```

#### Method 3: Default Queue

```go
// If no queue specified, uses "default" queue
task := asynq.NewTask(TypeSomeTask, payload)
client.Enqueue(task)  // Goes to "default" queue
```

### Priority Recommendations

| Queue | Use Cases | Timeout | Retry |
|-------|-----------|---------|-------|
| **critical** | Payment processing, refunds, auth | 30s | 3-5 |
| **high** | Inventory updates, order status | 15s | 3 |
| **default** | Emails, notifications, invoices | 20s | 5 |
| **low** | Analytics, logs, cleanup, reports | 10s | 2 |

### Advanced: Dynamic Priority

You can dynamically assign priority based on business logic:

```go
func (h *OrderHandler) enqueuePaymentTask(order *domain.Order) {
    queue := "default"
    
    // High-value orders get critical priority
    if order.TotalAmount > 10000 {
        queue = "critical"
    } else if order.TotalAmount > 1000 {
        queue = "high"
    }
    
    task, _ := tasks.NewPaymentProcessTask(...)
    h.asynqClient.Enqueue(task, asynq.Queue(queue))
}
```

### Tuning Priority Weights

**Example 1: Payment-Critical Business**
```go
Queues: map[string]int{
    "critical": 10,  // 77% - Focus on payments
    "high":     2,   // 15%
    "default":  1,   // 8%
    "low":      0,   // 0% - Disabled (won't process)
}
```

**Example 2: Balanced Processing**
```go
Queues: map[string]int{
    "critical": 4,  // 40%
    "high":     3,  // 30%
    "default":  2,  // 20%
    "low":      1,  // 10%
}
```

**Example 3: Equal Priority**
```go
Queues: map[string]int{
    "critical": 1,  // 25%
    "high":     1,  // 25%
    "default":  1,  // 25%
    "low":      1,  // 25%
}
```

### Monitoring Queue Performance

Check Asynqmon dashboard (http://localhost:8085):
- **Queue Depth**: Number of pending tasks
- **Processing Time**: Average task duration
- **Success Rate**: Completed vs failed tasks

If you see:
- ❌ **Critical queue growing**: Increase critical weight or concurrency
- ❌ **Low queue never processed**: Increase low weight
- ❌ **High latency**: Increase overall concurrency

### Best Practices

1. **Start Conservative**: Use default weights `{6,4,2,1}` first
2. **Monitor & Adjust**: Watch queue depths and adjust weights
3. **Business-Driven**: Priority should match business impact
4. **Avoid Weight 0**: Setting weight to 0 disables the queue entirely
5. **Consider SLAs**: Payment (seconds), Email (minutes), Analytics (hours)

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
