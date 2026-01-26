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

## ⚡ Quick Start

### Prerequisites
- **Docker Desktop** (must be running)
- **Go 1.21+**
- **K6** (for load testing): `brew install k6`

### 1. Start Infrastructure

```bash
# Start Redis, PostgreSQL, Asynqmon
docker-compose up -d

# Verify running
docker-compose ps
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
   ...
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

**Expected Results:**
```
✅ Response Time: avg 10ms, p95 45ms
✅ Throughput: 73 requests/second
✅ Error Rate: 0%
✅ 100% tasks processed successfully
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
  [Critical] payment:process
  [High]     inventory:update
  [Default]  email:confirmation
  [Default]  invoice:generate
  [Low]      analytics:track
  [Low]      warehouse:notify
      ↓
┌─────────────┐
│   Workers   │ → Process tasks asynchronously
│ (20 conc)   │ → Automatic retries on failure
└─────────────┘
```

**Priority Queues:**
- **Critical (weight 6):** Payment - highest priority
- **High (weight 4):** Inventory - time-sensitive
- **Default (weight 2):** Email, Invoice - moderate priority
- **Low (weight 1):** Analytics, Warehouse - can be delayed

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

- **[QUICKSTART.md](QUICKSTART.md)** - Step-by-step setup guide
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

## 📈 Performance

From load test with 50 concurrent users:

```
✅ Response Time: avg 10ms, p95 45ms, p99 70ms
✅ Throughput: 73 requests/second
✅ Error Rate: 0%
✅ Task Processing: 100% success (6 tasks per order)
✅ Queue Depth: Stable (no backlog)
```

System can likely handle 300-500+ req/s (run stress test to find exact limit).

---

## 🔗 Resources

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [K6 Documentation](https://k6.io/docs/)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)

---

**Built with ❤️ for learning Go, Asynq, and distributed systems.**
