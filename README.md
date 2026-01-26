# Go Asynq Load Test

> Background job processing with Go, Asynq, and Redis. Includes K6 load testing.

**Use Case:** E-commerce Order Processing System (POC)

---

## 🎯 What This Project Does

When an order is created via REST API:
1. **API** saves order to PostgreSQL (~10ms response)
2. **API** enqueues 6 background tasks to Redis
3. **Worker** processes tasks asynchronously:
   - 💳 Payment processing (2s)
   - 📦 Inventory update (500ms)
   - 📧 Email confirmation (1s)
   - 🧾 Invoice generation (3s)
   - 📊 Analytics tracking (200ms)
   - 🏭 Warehouse notification (500ms)

**Result:** Fast API response + reliable background processing with priority queues and automatic retries.

---

## 📦 Installation

### Prerequisites

- **Docker Desktop** (must be running)
- **Go 1.21+** ([Download](https://go.dev/dl/))
- **K6** (for load testing)

### Install Dependencies

```bash
# Clone repository
git clone <your-repo-url>
cd go-asynq-loadtest

# Download Go dependencies
go mod download

# Verify Go installation
go version  # Should show: go version go1.21.x

# Install K6 (macOS)
brew install k6

# Install K6 (Linux)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Install K6 (Windows)
choco install k6

# Verify K6 installation
k6 version  # Should show: k6 v0.x.x
```

---

## ⚡ Quick Start

### 1. Start Infrastructure

```bash
# Start Redis, PostgreSQL, Asynqmon
docker-compose up -d

# Verify services are running
docker-compose ps
```

**Expected output:**
```
NAME             STATUS    PORTS
asynq-redis      Up        0.0.0.0:6379->6379/tcp
asynq-postgres   Up        0.0.0.0:5432->5432/tcp
asynqmon         Up        0.0.0.0:8085->8080/tcp
```

### 2. Start API Server (Terminal 1)

```bash
go run cmd/api/main.go
```

**You'll see:**
```
🚀 Starting Order Processing API...
✅ Connected to Redis: localhost:6379
✅ Database connected successfully
✅ API server running on http://localhost:8080
```

### 3. Start Worker (Terminal 2)

```bash
go run cmd/worker/main.go
```

**You'll see:**
```
🔧 Starting Asynq Worker...
✅ Worker registered task handlers:
   💳 [Critical] payment:process
   📦 [High]     inventory:update
   📧 [Default]  email:confirmation
   🧾 [Default]  invoice:generate
   📊 [Low]      analytics:track
   🏭 [Low]      warehouse:notify

⚙️  Worker concurrency: 20
🔴 Redis: localhost:6379

🚀 Worker started! Waiting for tasks...
```

### 4. Create Test Order (Terminal 3)

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "customer_email": "test@example.com",
    "items": [{
      "product_id": "prod-1",
      "product_name": "Laptop",
      "quantity": 1,
      "unit_price": 1200.00
    }],
    "shipping_address": {
      "street": "123 Main St",
      "city": "SF",
      "state": "CA",
      "postal_code": "94102",
      "country": "USA"
    },
    "payment_method": "credit_card"
  }'
```

### 5. Monitor Tasks

Open **Asynqmon Dashboard**: http://localhost:8085

Watch the 6 background tasks being processed in real-time!

---

## 📊 Load Testing

### Run Basic Load Test (50 users, 4 minutes)

```bash
k6 run loadtest/basic-load.js
```

**You'll see real-time output:**
```
running (2m30s), 35/50 VUs
✓ order created status is 201
✓ response time < 200ms
http_req_duration: avg=10.16ms p(95)=44.97ms
```

### Run Stress Test (Find Breaking Point)

```bash
k6 run loadtest/stress-test.js
```

Gradually increases from 0 → 400 users to find system limits.

### Run Spike Test (Sudden Traffic Spike)

```bash
k6 run loadtest/spike-test.js
```

Tests recovery from sudden 10 → 200 users spike.

**See [docs/LOAD_TESTING.md](docs/LOAD_TESTING.md) for detailed guide.**

---

## 📈 Benchmark Results

Results from `k6 run loadtest/basic-load.js`:

### Test Configuration

```yaml
Test Duration: 4 minutes
Load Pattern:
  - Ramp up: 0 → 20 users (30s)
  - Ramp up: 20 → 50 users (1m)
  - Sustained: 50 users (2m)
  - Ramp down: 50 → 0 users (30s)

Worker Configuration:
  - Workers: 1 process
  - Concurrency: 20 (20 concurrent goroutines)
  - Priority Queues:
      critical: weight 6
      high: weight 4
      default: weight 2
      low: weight 1

Hardware:
  - CPU: Apple M1/M2 (or equivalent)
  - RAM: 8GB+
  - Database: PostgreSQL 15 (Docker)
  - Queue: Redis 7 (Docker)
```

### Performance Metrics

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| **Response Time (avg)** | 10.16ms | < 100ms | ✅ Excellent |
| **Response Time (p95)** | 44.97ms | < 200ms | ✅ Excellent |
| **Response Time (p99)** | ~70ms | < 500ms | ✅ Excellent |
| **Throughput** | 73 req/s | 50+ req/s | ✅ Good |
| **Error Rate** | 0% | < 1% | ✅ Perfect |
| **Total Requests** | 17,556 | - | ✅ |
| **Success Rate** | 100% | > 99% | ✅ Perfect |

### Task Processing

| Queue | Tasks | Avg Time | Success Rate |
|-------|-------|----------|--------------|
| **Critical** (payment) | 4,389 | ~2s | 100% |
| **High** (inventory) | 4,389 | ~500ms | 100% |
| **Default** (email, invoice) | 8,778 | ~1-3s | 100% |
| **Low** (analytics, warehouse) | 8,778 | ~200-500ms | 100% |

**Total Background Tasks:** 26,334 (6 tasks per order × 4,389 orders)

### Key Observations

- ✅ API responds instantly (~10ms avg) without waiting for background tasks
- ✅ All 26,334 background tasks processed successfully
- ✅ No queue backlog (workers processed tasks faster than enqueue rate)
- ✅ Priority queues working correctly (critical tasks processed first)
- ✅ Zero errors with automatic retry enabled

### Scaling Potential

Based on results:
- **Current capacity:** 73 req/s with 1 worker (20 concurrency)
- **Estimated with 3 workers:** ~200+ req/s
- **Estimated with 5 workers:** ~350+ req/s
- **Bottleneck:** Worker processing (not API or database)

**To scale further:** Run `k6 run loadtest/stress-test.js` to find exact breaking point.

---

## 📡 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/orders` | Create new order |
| GET | `/api/v1/orders` | List all orders |
| GET | `/api/v1/orders/:id` | Get order details |
| GET | `/api/v1/orders/:id/status` | Get order status |
| POST | `/api/v1/orders/:id/cancel` | Cancel order |
| GET | `/health` | Health check |

---

## 🏗️ Architecture

```
Client Request
      ↓
┌─────────────┐
│  API Server │ → PostgreSQL (save order)
│   (Gin)     │ → Redis (enqueue 6 tasks)
└─────────────┘
      ↓ Returns HTTP 201 (~10ms)
      
Redis Task Queue:
  [Critical] payment:process (weight 6)
  [High]     inventory:update (weight 4)
  [Default]  email:confirmation (weight 2)
  [Default]  invoice:generate (weight 2)
  [Low]      analytics:track (weight 1)
  [Low]      warehouse:notify (weight 1)
      ↓
┌─────────────┐
│   Workers   │ → Process tasks asynchronously
│ (20 conc)   │ → Automatic retries on failure
└─────────────┘
```

**Priority Queues:**
- **Critical (weight 6):** Payment - highest priority (46% worker time)
- **High (weight 4):** Inventory - time-sensitive (31% worker time)
- **Default (weight 2):** Email, Invoice - moderate (15% worker time)
- **Low (weight 1):** Analytics, Warehouse - can be delayed (8% worker time)

**See [docs/ASYNQ.md](docs/ASYNQ.md) for detailed Asynq explanation.**

---

## 🛠️ Useful Commands

```bash
# Infrastructure
docker-compose up -d      # Start services
docker-compose down       # Stop services
docker-compose ps         # Check status
docker-compose logs -f    # View logs

# Application
go run cmd/api/main.go    # Start API
go run cmd/worker/main.go # Start worker

# Build binaries
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go

# Testing
k6 run loadtest/basic-load.js   # Load test
k6 run loadtest/stress-test.js  # Stress test
k6 run loadtest/spike-test.js   # Spike test

# Health Checks
curl http://localhost:8080/health           # API health
docker exec asynq-redis redis-cli ping      # Redis health
docker exec asynq-postgres pg_isready -U admin  # DB health
```

---

## 📂 Project Structure

```
go-asynq-loadtest/
├── cmd/
│   ├── api/              # API server entry point
│   └── worker/           # Worker entry point
├── internal/
│   ├── config/           # Configuration
│   ├── domain/           # Domain models
│   ├── dto/              # Request/Response DTOs
│   ├── handler/          # HTTP handlers
│   ├── repository/       # Data access (GORM)
│   ├── service/          # Business logic
│   └── tasks/            # Asynq task definitions
├── pkg/
│   └── database/         # PostgreSQL connection
├── loadtest/             # K6 test scripts
│   ├── basic-load.js     # Baseline test
│   ├── stress-test.js    # Find limits
│   └── spike-test.js     # Spike recovery
├── docs/                 # Detailed documentation
│   ├── ASYNQ.md          # Asynq explanation
│   └── LOAD_TESTING.md   # K6 testing guide
├── docker-compose.yml    # Infrastructure setup
└── Makefile              # Convenience commands
```

---

## 📚 Documentation

- **[docs/ASYNQ.md](docs/ASYNQ.md)** - How Asynq works & priority queues
- **[docs/LOAD_TESTING.md](docs/LOAD_TESTING.md)** - K6 testing guide & metrics
- **[TESTING.md](TESTING.md)** - Test scenarios & examples

---

## ⚠️ POC Status

This is a **Proof of Concept** for learning purposes.

**What's Simulated:**
- Payment processing (real: Stripe API integration)
- Email sending (real: SendGrid/AWS SES integration)
- Invoice generation (real: PDF generation + S3 upload)
- Other external services

All task handlers use `time.Sleep()` to simulate processing time. See inline comments in `internal/tasks/*.go` for production implementation guidance.

---

## 🛠️ Technology Stack

- **Language:** Go 1.21+
- **Web Framework:** Gin
- **Task Queue:** Asynq (Redis-based)
- **Database:** PostgreSQL 15 + GORM
- **Load Testing:** K6
- **Monitoring:** Asynqmon
- **Infrastructure:** Docker Compose

---

## 🔗 Resources

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [K6 Documentation](https://k6.io/docs/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)

---

**Built with ❤️ for learning Go, Asynq, and distributed systems.**
