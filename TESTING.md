# 🧪 Testing Guide

Complete guide to test the Order Processing POC with Asynq background jobs.

---

## 📋 Prerequisites

Before testing, ensure you have:

- ✅ Go 1.21+ installed
- ✅ Docker & Docker Compose installed
- ✅ Redis running (via Docker Compose)
- ✅ PostgreSQL running (via Docker Compose)

---

## 🚀 Step-by-Step Testing

### **Step 1: Start Infrastructure**

Start Redis, PostgreSQL, and Asynqmon:

```bash
cd /Users/lppduy/learn/go-asynq-loadtest
docker-compose up -d
```

Verify services are running:

```bash
docker-compose ps
```

You should see:
- ✅ redis (port 6379)
- ✅ postgres (port 5432)
- ✅ asynqmon (port 8085)

---

### **Step 2: Install Go Dependencies**

```bash
go mod download
go mod tidy
```

This will download:
- `github.com/hibiken/asynq` - Task queue library
- `github.com/gin-gonic/gin` - Web framework
- `gorm.io/gorm` - ORM
- `github.com/google/uuid` - UUID generation

---

### **Step 3: Start API Server**

In Terminal 1:

```bash
go run cmd/api/main.go
```

Expected output:

```
🚀 Starting Order Processing API...
✅ Connected to Redis: localhost:6379
✅ API server running on http://localhost:8080
📚 Endpoints:
   - POST   /api/v1/orders          (Create order)
   - GET    /api/v1/orders          (List orders)
   - GET    /api/v1/orders/:id      (Get order)
   - GET    /api/v1/orders/:id/status (Get status)
   - POST   /api/v1/orders/:id/cancel (Cancel order)

💡 Try: curl http://localhost:8080/health

📋 Background tasks will be processed by worker
   Start worker: go run cmd/worker/main.go
   Monitor tasks: http://localhost:8085 (Asynqmon)
```

---

### **Step 4: Start Worker**

In Terminal 2 (new terminal):

```bash
cd /Users/lppduy/learn/go-asynq-loadtest
go run cmd/worker/main.go
```

Expected output:

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

---

### **Step 5: Open Asynqmon Dashboard**

In your browser, open:

```
http://localhost:8085
```

You should see:
- 📊 Dashboard with queue stats
- 📋 Active, pending, scheduled tasks
- 🔴 Redis connection status
- ⚙️ Worker status

---

### **Step 6: Create Test Order**

In Terminal 3 (new terminal):

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

### **Step 7: Observe the Magic** ✨

#### **In API Server Terminal (Terminal 1):**

```
✅ Order created: ORD-a1b2c3d4 | Total: $2657.00 | Items: 2
📋 Background tasks enqueued asynchronously
📤 [Enqueued] Payment task for order: ORD-a1b2c3d4
📤 [Enqueued] Inventory task for order: ORD-a1b2c3d4
📤 [Enqueued] Email task for order: ORD-a1b2c3d4
📤 [Enqueued] Invoice task for order: ORD-a1b2c3d4
📤 [Enqueued] Analytics task for order: ORD-a1b2c3d4
📤 [Enqueued] Warehouse task for order: ORD-a1b2c3d4
✅ All background tasks enqueued for order: ORD-a1b2c3d4
```

#### **In Worker Terminal (Terminal 2):**

Watch tasks being processed in real-time:

```
💳 [Payment] Processing payment for order: ORD-a1b2c3d4
💳 [Payment] Amount: $2657.00 | Method: credit_card
✅ [Payment] Payment processed successfully for order: ORD-a1b2c3d4

📦 [Inventory] Updating inventory for order: ORD-a1b2c3d4
📦 [Inventory] Items to update: 2
📦 [Inventory] Updated: prod-laptop (qty: 1)
📦 [Inventory] Updated: prod-mouse (qty: 2)
✅ [Inventory] All items updated for order: ORD-a1b2c3d4

📧 [Email] Sending confirmation to: test@example.com
📧 [Email] Order: ORD-a1b2c3d4 | Amount: $2657.00
✅ [Email] Confirmation sent successfully to: test@example.com

🧾 [Invoice] Generating invoice for order: ORD-a1b2c3d4
🧾 [Invoice] Customer: cust-123 | Amount: $2657.00
✅ [Invoice] Invoice generated: https://storage.example.com/invoices/ORD-a1b2c3d4.pdf

📊 [Analytics] Tracking order: ORD-a1b2c3d4
📊 [Analytics] Customer: cust-123 | Amount: $2657.00 | Items: 2
✅ [Analytics] Event tracked for order: ORD-a1b2c3d4

📦 [Warehouse] Notifying warehouse about order: ORD-a1b2c3d4
📦 [Warehouse] Customer: cust-123 | Items: 2 | Priority: standard
✅ [Warehouse] Notification sent for order: ORD-a1b2c3d4
```

#### **In Asynqmon Dashboard:**

Refresh the page and observe:
- 📈 Tasks moving through queues
- ✅ Completed tasks count increasing
- ⏱️ Processing time for each task
- 📊 Queue depth changes

---

### **Step 8: Verify Order in Database**

Check order was saved to PostgreSQL:

```bash
# List all orders
curl http://localhost:8080/api/v1/orders

# Get specific order (replace with actual order ID)
curl http://localhost:8080/api/v1/orders/ORD-a1b2c3d4

# Check order status
curl http://localhost:8080/api/v1/orders/ORD-a1b2c3d4/status
```

---

## 🧪 Additional Test Scenarios

### **Test 1: Multiple Orders (Load Test)**

Create 10 orders quickly:

```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -d '{
      "customer_id": "cust-'$i'",
      "customer_email": "customer'$i'@example.com",
      "items": [
        {
          "product_id": "prod-'$i'",
          "product_name": "Product '$i'",
          "quantity": 1,
          "unit_price": 100.00
        }
      ],
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
```

Watch Asynqmon to see:
- 🔥 60 tasks enqueued (6 tasks per order)
- ⚡ Tasks processed based on priority
- 📊 Queue depth visualization

---

### **Test 2: Cancel Order**

```bash
# Cancel an order
curl -X POST http://localhost:8080/api/v1/orders/ORD-a1b2c3d4/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "Customer changed their mind"}'
```

---

### **Test 3: Query by Customer**

```bash
# Get all orders for a customer
curl "http://localhost:8080/api/v1/orders?customer_id=cust-123"
```

---

## 🎯 Expected Behavior

### **✅ Success Criteria:**

1. **Fast API Response:** Order creation returns in <100ms
2. **Background Processing:** Tasks processed asynchronously
3. **Priority Queues:** Critical tasks (payment) processed first
4. **No Blocking:** API doesn't wait for background tasks
5. **Reliable:** Tasks retry on failure
6. **Observable:** Real-time monitoring in Asynqmon

### **📊 Performance Metrics:**

- API response time: ~50ms
- Payment processing: ~2 seconds
- Inventory update: ~500ms
- Email sending: ~1 second
- Invoice generation: ~3 seconds
- Analytics tracking: ~200ms
- Warehouse notification: ~500ms

**Total sequential time:** ~7.7 seconds  
**With Asynq (parallel):** API responds in ~50ms, tasks complete within ~5 seconds

---

## 🐛 Troubleshooting

### **Problem: Worker not receiving tasks**

**Solution:**
```bash
# Check Redis connection
docker-compose logs redis

# Check Asynq client connection in API logs
# Should see: "✅ Connected to Redis: localhost:6379"
```

---

### **Problem: Tasks failing**

**Solution:**
```bash
# Check worker logs for errors
# Check Asynqmon "Failed" tab
# Tasks auto-retry (max 3-5 times based on task)
```

---

### **Problem: Database connection error**

**Solution:**
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check connection string in logs
# Default: host=localhost port=5432 user=admin dbname=taskqueue
```

---

## 🧹 Cleanup

Stop all services:

```bash
# Stop API and Worker (Ctrl+C in terminals)

# Stop Docker services
docker-compose down

# Remove volumes (clean database)
docker-compose down -v
```

---

## 📊 Next Steps

1. ✅ Run load tests with K6 (see `loadtest/` folder)
2. ✅ Add monitoring with Prometheus + Grafana
3. ✅ Deploy to Kubernetes
4. ✅ Add more task types (refunds, notifications, etc.)

---

## 🎉 Success!

If you see:
- ✅ API responding quickly
- ✅ Tasks appearing in Asynqmon
- ✅ Worker processing tasks
- ✅ Logs showing task completion
- ✅ Orders saved in database

**🎊 Congratulations! Your Asynq POC is working!**
