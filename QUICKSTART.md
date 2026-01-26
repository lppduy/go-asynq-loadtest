# ⚡ Quick Start Guide

Get the Asynq POC running in 3 minutes!

---

## 📋 **Prerequisites**

- Docker Desktop (running)
- Go 1.21+

---

## 🚀 **Steps**

### **1. Start Infrastructure** (Terminal 1)

```bash
cd /Users/lppduy/learn/go-asynq-loadtest

# Start Redis, PostgreSQL, Asynqmon
docker-compose up -d

# Verify
docker-compose ps
```

**Expected output:**
```
NAME             STATUS    PORTS
asynq-redis      Up        0.0.0.0:6379->6379/tcp
asynq-postgres   Up        0.0.0.0:5432->5432/tcp
asynqmon         Up        0.0.0.0:8085->8080/tcp
```

---

### **2. Start API Server** (Terminal 2)

```bash
cd /Users/lppduy/learn/go-asynq-loadtest

# First time: Install dependencies
go mod download

# Start API
go run cmd/api/main.go
```

**You'll see:**
```
🚀 Starting Order Processing API...
✅ Connected to Redis: localhost:6379
✅ Database connected successfully
✅ API server running on http://localhost:8080
```

---

### **3. Start Worker** (Terminal 3)

```bash
cd /Users/lppduy/learn/go-asynq-loadtest

# Start Worker
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

🚀 Worker started! Waiting for tasks...
```

---

## 🧪 **Test It!** (Terminal 4)

### **Create an order:**

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

### **Check results:**

1. **Terminal 2 (API):** See order created log
2. **Terminal 3 (Worker):** See 6 tasks being processed
3. **Browser:** Open http://localhost:8085 to see Asynqmon UI

---

## 📊 **What Just Happened?**

```
1. API received order → Saved to PostgreSQL
2. API enqueued 6 background tasks to Redis:
   - payment:process (critical queue)
   - inventory:update (high queue)  
   - email:confirmation (default queue)
   - invoice:generate (default queue)
   - analytics:track (low queue)
   - warehouse:notify (low queue)

3. Worker picked up tasks and processed them asynchronously!
```

---

## 🛑 **Stop Everything**

```bash
# Stop API (Terminal 2): Ctrl+C
# Stop Worker (Terminal 3): Ctrl+C

# Stop infrastructure
docker-compose down

# Clean all data
docker-compose down -v
```

---

## 📝 **Makefile Shortcuts**

```bash
# Infrastructure
make infra-up      # Start infrastructure
make infra-down    # Stop infrastructure
make infra-logs    # View logs
make infra-clean   # Remove all data

# Application
make api           # Run API
make worker        # Run Worker

# Helpers
make install       # Install dependencies
make build         # Build binaries
```

---

## ✅ **Next Steps**

- Load test with K6
- Check README.md for architecture details
- Explore Asynqmon UI at http://localhost:8085
