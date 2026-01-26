# 📊 Load Testing với K6

Complete guide về load testing cho Asynq Order Processing system.

---

## 🎯 **What is Load Testing?**

**Load testing** = Testing the system with simulated traffic to:
- 📈 Find performance limits (how many concurrent users?)
- 🐛 Identify bottlenecks (DB? Redis? Worker?)
- 🔥 Verify system stability under load
- 📊 Measure response times & throughput

---

## 🛠️ **Install K6**

### **macOS:**
```bash
brew install k6
```

### **Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

### **Windows:**
```bash
choco install k6
```

### **Docker (no install needed):**
```bash
docker run --rm -i --network=host grafana/k6 run - <loadtest/basic-load.js
```

---

## 📁 **Test Scripts**

### **1. basic-load.js** - Baseline Performance
**Purpose:** Test performance with normal load

**Profile:**
```
0s ────► 30s: 0 → 20 users  (warm up)
30s ───► 1m30s: 20 → 50 users (ramp up)
1m30s ─► 3m30s: 50 users (sustained load)
3m30s ─► 4m: 50 → 0 users (cool down)
```

**Use case:** 
- Daily traffic simulation
- Baseline metrics
- Performance regression testing

---

### **2. stress-test.js** - Find Breaking Point
**Purpose:** Gradually increase load until system reaches stress point

**Profile:**
```
0 → 50 → 100 → 200 → 300 → 400 users (gradual increase)
```

**Use case:**
- Find capacity limits
- Identify bottlenecks
- Plan scaling strategy

**Expected behavior:**
- Response time increases gradually
- Error rate stays low until breaking point
- System recovers when load decreases

---

### **3. spike-test.js** - Sudden Traffic Spike
**Purpose:** Test recovery from sudden traffic spike

**Profile:**
```
10 users ─► 200 users (in 10 seconds!) ─► back to 10
```

**Use case:**
- Flash sales
- Marketing campaigns
- Viral events
- DDoS simulation

**Expected behavior:**
- Some errors acceptable during spike
- System should NOT crash
- Should recover after spike

---

## 🚀 **Run Tests**

### **Before running tests:**

1. **Start infrastructure:**
```bash
docker-compose up -d
```

2. **Start API & Worker:**
```bash
# Terminal 1
go run cmd/api/main.go

# Terminal 2
go run cmd/worker/main.go
```

3. **Verify system:**
```bash
curl http://localhost:8080/health
```

4. **Open Asynqmon:**
```
http://localhost:8085
```

---

### **Run Basic Load Test:**

```bash
cd /Users/lppduy/learn/go-asynq-loadtest

# Run test
k6 run loadtest/basic-load.js

# Run with detailed output
k6 run --out json=loadtest/results/basic-load.json loadtest/basic-load.js
```

**You'll see:**
```
running (4m00s), 00/50 VUs, 1234 complete and 0 interrupted iterations
default ✓ [======================================] 00/50 VUs  4m0s

✅ Basic Load Test Summary
==================================================

📊 Requests:
  Total: 4936
  Rate: 20.56 req/s

⏱️  Response Time:
  Avg: 45.23ms
  Min: 12.45ms
  Max: 234.56ms
  p(95): 89.12ms
  p(99): 145.34ms

❌ Errors:
  Failed Requests: 0
  Error Rate: 0.00%

👥 Virtual Users:
  Max: 50
```

---

### **Run Stress Test:**

```bash
k6 run loadtest/stress-test.js
```

**Watch for:**
- Response time increasing
- Error rate rising
- Queue depth in Asynqmon

---

### **Run Spike Test:**

```bash
k6 run loadtest/spike-test.js
```

**Watch for:**
- System behavior during spike
- Recovery time
- Error handling

---

## 📊 **Understanding K6 Output**

### **Key Metrics:**

#### **1. Virtual Users (VUs)**
```
vus......................: 50   min=0    max=50
vus_max..................: 50
```
- **Meaning:** Number of concurrent users being simulated
- **Good:** Matches predefined profile
- **Bad:** VUs drop unexpectedly (K6 crashed)

---

#### **2. HTTP Request Duration**
```
http_req_duration........: avg=45ms min=12ms max=234ms p(95)=89ms p(99)=145ms
```
- **avg:** Average response time
- **p(95):** 95% requests nhanh hơn giá trị này
- **p(99):** 99% requests nhanh hơn giá trị này

**Good values:**
- avg < 100ms (excellent)
- p(95) < 200ms (good)
- p(99) < 500ms (acceptable)

**Red flags:**
- avg > 500ms (slow)
- p(99) > 2000ms (very slow)
- High variance (max >> avg) → inconsistent

---

#### **3. HTTP Request Rate**
```
http_reqs................: 4936   20.56/s
```
- **Meaning:** Requests per second (throughput)
- **Good:** High & stable
- **Bad:** Drops during test (bottleneck)

---

#### **4. Failed Requests**
```
http_req_failed..........: 0.00%  ✓ 4936  ✗ 0
```
- **Good:** < 1% (production)
- **Acceptable:** < 5% (stress test)
- **Bad:** > 10% (system overloaded)

---

#### **5. Checks**
```
✓ order created status is 201
✓ response time < 200ms
✗ order has ID  (5 failed)
```
- **Green ✓:** Check passed
- **Red ✗:** Check failed
- **Use:** Verify business logic under load

---

## 🔍 **How K6 Works**

### **Architecture:**

```
┌─────────────────────────────────────────┐
│         K6 Load Generator               │
│                                         │
│  ┌──────┐  ┌──────┐  ┌──────┐         │
│  │ VU 1 │  │ VU 2 │  │ VU N │  ...    │
│  └──┬───┘  └──┬───┘  └──┬───┘         │
│     │         │         │              │
└─────┼─────────┼─────────┼──────────────┘
      │         │         │
      HTTP      HTTP      HTTP
      ↓         ↓         ↓
┌─────────────────────────────────────────┐
│         Your API Server                 │
│      (http://localhost:8080)            │
└─────────────────────────────────────────┘
```

### **Virtual Users (VUs):**

Each VU = 1 independent user running your test script:
```javascript
export default function () {
  // This code runs in a loop for each VU
  http.post('http://localhost:8080/api/v1/orders', payload);
  sleep(1); // Wait 1s before next iteration
}
```

**Example:** 50 VUs với sleep(1):
- Each VU: 1 request/second
- Total: ~50 requests/second

---

### **Stages:**

```javascript
stages: [
  { duration: '30s', target: 20 },  // Stage 1
  { duration: '1m', target: 50 },   // Stage 2
  { duration: '2m', target: 50 },   // Stage 3
]
```

**Visual:**
```
VUs
50 ┤         ┌────────────────────┐
   │        ╱                    │
20 ┤───────╱                     │
   │      ╱                      │
0  ┤─────┘                        └────►
   └──30s──1m───────2m────────30s
```

---

### **Thresholds:**

```javascript
thresholds: {
  'http_req_duration': ['p(95)<500'],  // Test FAILS if p95 > 500ms
  'http_req_failed': ['rate<0.05'],    // Test FAILS if error rate > 5%
}
```

**Use case:** Automated pass/fail criteria for CI/CD pipelines

---

## 🎯 **What to Watch During Test**

### **1. Asynqmon (http://localhost:8085)**
```
Critical queue: 234 pending, 12 active, 5678 processed
High queue:     123 pending, 8 active, 3456 processed
Default queue:  456 pending, 15 active, 8901 processed
Low queue:      789 pending, 10 active, 12345 processed
```

**Good signs:**
- ✅ Pending count stable or decreasing
- ✅ Processing rate > enqueue rate
- ✅ Low latency (< 1s in queue)

**Bad signs:**
- ❌ Pending count growing (backlog!)
- ❌ Workers stuck (no progress)
- ❌ High failed count

---

### **2. API Logs (Terminal)**
```
✅ Order created: ORD-abc123
📤 [Enqueued] 6 tasks
✅ Order created: ORD-def456
📤 [Enqueued] 6 tasks
```

**Watch for:**
- Errors
- Slow queries
- Connection timeouts

---

### **3. Worker Logs (Terminal)**
```
💳 [Payment] Processing...
✅ [Payment] Success
📦 [Inventory] Processing...
✅ [Inventory] Success
```

**Watch for:**
- Tasks processing
- Errors/retries
- Processing speed

---

### **4. Docker Stats**
```bash
docker stats
```

**Watch:**
- CPU usage (should be < 80%)
- Memory usage (stable, not growing)
- Network I/O

---

## 📈 **Performance Tuning Tips**

### **If response times are slow:**

1. **Increase worker concurrency:**
```go
// cmd/worker/main.go
Concurrency: 20  // ← Try 50, 100
```

2. **Scale workers:**
```bash
# Start multiple workers
go run cmd/worker/main.go  # Terminal 2
go run cmd/worker/main.go  # Terminal 3
go run cmd/worker/main.go  # Terminal 4
```

3. **Database connection pool:**
```go
// pkg/database/postgres.go
db.DB().SetMaxOpenConns(100)
db.DB().SetMaxIdleConns(10)
```

4. **Redis optimization:**
```bash
# Check Redis performance
docker exec asynq-redis redis-cli INFO stats
```

---

### **If error rate is high:**

1. **Check timeout settings:**
```javascript
// K6 timeout
timeout: '10s'  // Increase if needed
```

2. **Check API logs:**
```
Database timeout?
Redis connection refused?
Worker backlog?
```

3. **Check thresholds:**
```javascript
// Maybe too strict?
'http_req_duration': ['p(95)<500']  // Try 1000ms
```

---

## 🎓 **Load Test Best Practices**

### **1. Start Small**
- Begin with `basic-load.js` to establish baseline
- Understand normal behavior first
- Then move to stress testing

### **2. Monitor Everything**
- **Asynqmon:** Queue depth and task processing
- **API logs:** Error messages and slow queries
- **Worker logs:** Task execution and retries
- **Docker stats:** Resource utilization (CPU, memory)

### **3. Test Realistic Scenarios**
- Use realistic payload sizes
- Mix different operations (GET, POST)
- Include think time (user delay simulation)

### **4. Incremental Tuning**
- Change ONE variable at a time
- Measure the impact
- Document all changes and results

### **5. Define Success Criteria**
```
Example targets:
✅ Support 100 req/s sustained throughput
✅ p95 response time < 200ms
✅ Error rate < 1%
✅ Queue depth stays < 1000 tasks
```

---

## 🚀 **Next Steps**

1. **Run basic-load.js** - Establish baseline
2. **Check Asynqmon** - Watch task processing
3. **Run stress-test.js** - Find limits
4. **Tune settings** - Worker concurrency, DB pool
5. **Run spike-test.js** - Test recovery
6. **Document results** - Save metrics

---

## 📝 **Results Location**

```
loadtest/results/
├── basic-load-summary.json
├── stress-test-summary.json
└── spike-test-summary.json
```

---

## 📚 **Additional Resources**

- [K6 Documentation](https://k6.io/docs/)
- [K6 Examples](https://k6.io/docs/examples/)
- [Asynq Monitoring Guide](https://github.com/hibiken/asynq/wiki/Monitoring)

---

**🎉 Happy Load Testing!**
