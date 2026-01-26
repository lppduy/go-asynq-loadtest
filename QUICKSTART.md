# ⚡ Quick Start Guide

Get the Asynq POC running in 5 minutes!

---

## 📋 **Prerequisites**

- Go 1.21+
- Docker Desktop (running)
- 3 terminal windows

---

## 🚀 **Step-by-Step**

### **1. Start Infrastructure** (Terminal 1)

```bash
cd /Users/lppduy/learn/go-asynq-loadtest

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

---

### **2. Start API Server** (Terminal 1)

```bash
# Install dependencies (first time only)
go mod download && go mod tidy

# Start API
go run cmd/api/main.go
```

**You'll see:**
```
🚀 Starting Order Processing API...
✅ Connected to Redis: localhost:6379
✅ API server running on http://localhost:8080
```

---

### **3. Start Worker** (Terminal 2)

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

### **4. Create Test Order** (Terminal 3)

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "customer_email": "test@example.com",
    "items": [
      {
        "product_id": "prod-laptop",
        "product_name": "MacBook Pro",
        "quantity": 1,
        "unit_price": 2499.00
      },
      {
        "product_id": "prod-mouse",
        "product_name": "Magic Mouse",
        "quantity": 2,
        "unit_price": 79.00
      }
    ],
    "shipping_address": {
      "street": "123 Main Street",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94102",
      "country": "USA"
    },
    "payment_method": "credit_card",
    "notes": "Please deliver before 5 PM"
  }'
```

---

### **5. Watch the Magic!** ✨

**Terminal 1 (API):**
```
✅ Order created: ORD-a1b2c3d4 | Total: $2657.00 | Items: 2
📤 [Enqueued] Payment task for order: ORD-a1b2c3d4
📤 [Enqueued] Inventory task for order: ORD-a1b2c3d4
📤 [Enqueued] Email task for order: ORD-a1b2c3d4
📤 [Enqueued] Invoice task for order: ORD-a1b2c3d4
📤 [Enqueued] Analytics task for order: ORD-a1b2c3d4
📤 [Enqueued] Warehouse task for order: ORD-a1b2c3d4
✅ All background tasks enqueued
```

**Terminal 2 (Worker):**
```
💳 [Payment] Processing payment for order: ORD-a1b2c3d4
💳 [Payment] Amount: $2657.00 | Method: credit_card
✅ [Payment] Payment processed successfully

📦 [Inventory] Updating inventory for order: ORD-a1b2c3d4
📦 [Inventory] Items to update: 2
✅ [Inventory] All items updated

📧 [Email] Sending confirmation to: test@example.com
✅ [Email] Confirmation sent successfully

🧾 [Invoice] Generating invoice for order: ORD-a1b2c3d4
✅ [Invoice] Invoice generated

📊 [Analytics] Tracking order: ORD-a1b2c3d4
✅ [Analytics] Event tracked

📦 [Warehouse] Notifying warehouse about order: ORD-a1b2c3d4
✅ [Warehouse] Notification sent
```

---

### **6. Open Asynqmon Dashboard**

Open browser: **http://localhost:8085**

**You'll see:**
- 📊 Active tasks (currently processing)
- ⏳ Pending tasks (waiting in queue)
- ✅ Completed tasks (successful)
- ❌ Failed tasks (errors)
- 📈 Queue statistics
- ⚙️ Worker status

**Click around to explore:**
- See tasks by queue (critical, high, default, low)
- View task details (payload, retry count, timestamps)
- Monitor processing times

---

## 🧪 **More Test Commands**

### **Health Check**
```bash
curl http://localhost:8080/health
```

### **List All Orders**
```bash
curl http://localhost:8080/api/v1/orders
```

### **Get Specific Order**
```bash
curl http://localhost:8080/api/v1/orders/ORD-12345678
```

### **Check Order Status**
```bash
curl http://localhost:8080/api/v1/orders/ORD-12345678/status
```

### **Cancel Order**
```bash
curl -X POST http://localhost:8080/api/v1/orders/ORD-12345678/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "Customer changed their mind"}'
```

---

## 🔥 **Load Test (Create 10 Orders)**

```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -d '{
      "customer_id": "cust-'$i'",
      "customer_email": "customer'$i'@example.com",
      "items": [{
        "product_id": "prod-'$i'",
        "product_name": "Product '$i'",
        "quantity": 1,
        "unit_price": 100.00
      }],
      "shipping_address": {
        "street": "123 Main St",
        "city": "NYC",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA"
      },
      "payment_method": "credit_card"
    }' &
done
wait

echo "✅ Created 10 orders!"
```

**Then check Asynqmon:** You'll see 60 tasks (6 per order) being processed!

---

## 🛑 **Stop Everything**

```bash
# Stop API & Worker (Ctrl+C in terminals)

# Stop Docker services
docker-compose down

# Stop and remove data (clean slate)
docker-compose down -v
```

---

## 🎯 **Key Observations**

1. **API responds instantly** (~50ms) - doesn't wait for tasks
2. **Tasks processed in background** - non-blocking
3. **Priority queues work** - payment tasks first
4. **Real-time monitoring** - Asynqmon shows everything
5. **Automatic retries** - failed tasks retry automatically

---

## 📚 **Next Steps**

- Read `TESTING.md` for detailed scenarios
- Check `README.md` for architecture deep dive
- Explore priority queue tuning
- Try canceling orders mid-processing
- Monitor queue depths under load

---

**🎉 You're running an Asynq-powered order processing system!**
