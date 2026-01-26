# Asynq - Distributed Task Queue

Complete guide to understanding Asynq and how it's used in this project.

---

## 🎯 What is Asynq?

**Asynq** is a Go library for **distributed task queuing** backed by Redis.

**Think of it as:**
- A job queue system (like Sidekiq in Ruby, Celery in Python, BullMQ in Node.js)
- Producer-Consumer pattern
- Reliable background job processing

**Why use Asynq?**
- ✅ **Fast API responses** - Don't wait for slow operations
- ✅ **Reliable** - Automatic retries, persistence
- ✅ **Scalable** - Add more workers easily
- ✅ **Priority queues** - Critical tasks first
- ✅ **Monitoring** - Built-in Asynqmon dashboard
- ✅ **Simple** - Redis-based, easy to setup

---

## 🧾 Task Retention (Completed/Failed visibility in Asynqmon)

By default, Asynq removes completed tasks from Redis immediately after processing. This means the **Processed counter** increases, but the **Completed tab may show 0 tasks**.

In this project we enable retention on enqueue so completed tasks remain visible in Asynqmon for debugging.

### Configure

- **Env var**: `ASYNQ_RETENTION_MINUTES`
- **Default**: `30`
- **Disable**: set to `0` (reduces Redis memory usage during large load tests)

### Notes

- Retention is helpful in development and for capturing screenshots.
- For heavy load tests, you may want to reduce retention to keep Redis memory stable.

---

## 🏗️ Architecture

```
┌──────────┐         ┌──────────┐
│   API    │ Enqueue │  Redis   │
│  Server  ├────────►│  Queue   │
│(Producer)│  Tasks  │ (Broker) │
└──────────┘         └────┬─────┘
                          │ Poll
                          ↓
                    ┌──────────┐
                    │ Workers  │
                    │(Consumer)│
                    └──────────┘
```

### Components

**1. Producer (API Server)**
- Creates tasks
- Enqueues to Redis
- Returns immediately

**2. Broker (Redis)**
- Stores task queues
- Manages task state
- Handles scheduling

**3. Consumer (Workers)**
- Polls for tasks
- Executes task handlers
- Reports completion/failure

---

## 📦 Task Lifecycle

### 1. Task Creation (Producer)

```go
// In API handler
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // 1. Save order to database
    order, _ := h.orderService.CreateOrder(...)
    
    // 2. Create Asynq task
    task, _ := tasks.NewPaymentProcessTask(order.ID, amount)
    
    // 3. Enqueue task to Redis
    h.asynqClient.Enqueue(task)
    
    // 4. Return immediately (don't wait for task)
    c.JSON(201, order)  // Fast response!
}
```

**Task Options:**
```go
task, _ := asynq.NewTask(
    "payment:process",  // Task type
    payload,            // JSON data
    asynq.Queue("critical"),     // Queue name
    asynq.MaxRetry(3),           // Max retry attempts
    asynq.Timeout(30*time.Second), // Timeout
    asynq.ProcessIn(2*time.Second), // Delay before processing
)
```

### 2. Task Storage (Redis)

```
Redis Data Structure:
┌─────────────────────────────────┐
│ asynq:queues:critical           │ → List of task IDs
│ asynq:queues:high               │
│ asynq:queues:default            │
│ asynq:queues:low                │
├─────────────────────────────────┤
│ asynq:{namespace}:t:{task_id}   │ → Task payload (Hash)
│   - type: "payment:process"     │
│   - payload: {...}               │
│   - retry: 3                     │
│   - timeout: 30s                 │
└─────────────────────────────────┘
```

### 3. Task Processing (Worker)

```go
// Worker polls Redis every 100ms
func (w *Worker) Start() {
    for {
        // 1. Check for tasks (respecting priority)
        task := pollRedis()
        
        // 2. Execute handler
        err := handler(task)
        
        // 3. Handle result
        if err != nil {
            // Retry with exponential backoff
            scheduleRetry(task)
        } else {
            // Mark as completed
            markCompleted(task)
        }
    }
}
```

**Task Handler:**
```go
func HandlePaymentProcessTask(ctx context.Context, task *asynq.Task) error {
    // 1. Parse payload
    var payload PaymentPayload
    json.Unmarshal(task.Payload(), &payload)
    
    // 2. Process payment
    err := processPayment(payload)
    if err != nil {
        return err  // Will retry
    }
    
    // 3. Update database
    updateOrderStatus(payload.OrderID, "paid")
    
    return nil  // Success!
}
```

---

## 🎚️ Priority Queues

### How Priority Works

Asynq uses **weighted priority**, not strict priority.

```go
// Worker configuration
asynq.Config{
    Queues: map[string]int{
        "critical": 6,  // Weight 6
        "high":     4,  // Weight 4
        "default":  2,  // Weight 2
        "low":      1,  // Weight 1
    },
}
```

**Weight Calculation:**
```
Total weight = 6 + 4 + 2 + 1 = 13

critical: 6/13 = 46% of worker time
high:     4/13 = 31% of worker time
default:  2/13 = 15% of worker time
low:      1/13 = 8% of worker time
```

### Why Weighted (Not Strict)?

**Strict Priority (Bad):**
```
✗ Low priority tasks may NEVER run
✗ Can cause starvation
✗ Queue backlog grows

Example:
- 1000 critical tasks/sec
- Low priority tasks waiting forever
```

**Weighted Priority (Good):**
```
✓ All tasks eventually run
✓ No starvation
✓ Proportional processing

Example:
- Critical tasks: 460ms
- High tasks: 310ms
- Default tasks: 150ms
- Low tasks: 80ms per second
```

### Task-to-Queue Assignment

In this project:

| Task | Queue | Weight | Rationale |
|------|-------|--------|-----------|
| **Payment Processing** | critical | 6 | Money involved, must process ASAP |
| **Inventory Update** | high | 4 | Stock management, time-sensitive |
| **Email Confirmation** | default | 2 | User notification, moderate priority |
| **Invoice Generation** | default | 2 | Important but not urgent |
| **Analytics Tracking** | low | 1 | Can be delayed without impact |
| **Warehouse Notification** | low | 1 | Background operation |

### Setting Task Priority

**Method 1: At Task Creation**
```go
// High priority task
task, _ := tasks.NewPaymentProcessTask(...)
// Already configured with asynq.Queue("critical")

// Low priority task
task, _ := tasks.NewAnalyticsTrackTask(...)
// Already configured with asynq.Queue("low")
```

**Method 2: At Enqueue Time**
```go
// Override queue when enqueueing
task, _ := tasks.NewPaymentProcessTask(...)
client.Enqueue(
    task,
    asynq.Queue("critical"),  // Override here
)
```

---

## 🔄 Retry Mechanism

### Automatic Retries

```go
// Task with retry configuration
task, _ := asynq.NewTask(
    "payment:process",
    payload,
    asynq.MaxRetry(3),  // Retry up to 3 times
)
```

**Retry Schedule (Exponential Backoff):**
```
Attempt 1: Immediate
Attempt 2: 1 minute later  (1 * 1²)
Attempt 3: 4 minutes later (1 * 2²)
Attempt 4: 9 minutes later (1 * 3²)
```

**Visualization:**
```
Initial → Failed
  ↓
Wait 1min → Retry 1 → Failed
  ↓
Wait 4min → Retry 2 → Failed
  ↓
Wait 9min → Retry 3 → Failed
  ↓
Move to Dead Letter Queue (DLQ)
```

### Custom Retry Logic

```go
// Custom retry delay
asynq.Config{
    RetryDelayFunc: func(n int, err error, task *Task) time.Duration {
        // Custom backoff
        return time.Duration(n*n) * time.Minute
    },
}
```

### Dead Letter Queue

After max retries exhausted:
```
Failed Task → archived:default → Manual review required
```

View in Asynqmon: http://localhost:8085 → Archived tab

---

## ⚙️ Worker Configuration

### Basic Worker Setup

```go
// cmd/worker/main.go
func main() {
    // 1. Create Redis connection
    redisOpt := asynq.RedisClientOpt{
        Addr: "localhost:6379",
    }
    
    // 2. Configure worker
    srv := asynq.NewServer(
        redisOpt,
        asynq.Config{
            Concurrency: 20,  // 20 concurrent workers
            Queues: map[string]int{
                "critical": 6,
                "high":     4,
                "default":  2,
                "low":      1,
            },
        },
    )
    
    // 3. Register handlers
    mux := asynq.NewServeMux()
    mux.HandleFunc("payment:process", HandlePaymentProcessTask)
    mux.HandleFunc("email:confirmation", HandleEmailConfirmationTask)
    
    // 4. Start processing
    srv.Run(mux)
}
```

### Concurrency

**What is Concurrency?**
```
Concurrency: 20 means:
- 20 goroutines processing tasks simultaneously
- Can process 20 tasks at the same time
```

**How to choose concurrency value?**
```
Low (5-10):   Light workload, simple tasks
Medium (20):  Moderate workload (default)
High (50+):   Heavy workload, I/O-bound tasks

Formula: 
Concurrency = (Desired Throughput) / (Task Duration)

Example:
- Want: 100 tasks/second
- Task takes: 2 seconds
- Concurrency = 100 * 2 = 200 workers
```

### Scaling Workers

**Horizontal Scaling:**
```bash
# Start multiple worker processes
# Terminal 1
go run cmd/worker/main.go

# Terminal 2
go run cmd/worker/main.go

# Terminal 3
go run cmd/worker/main.go

# All share same Redis queue
# Total: 20 * 3 = 60 concurrent workers
```

**Benefits:**
- ✅ Better CPU utilization
- ✅ Fault tolerance (one crashes, others continue)
- ✅ Easy to scale up/down
- ✅ Can run on different servers

---

## 📊 Monitoring with Asynqmon

### Dashboard Overview

Access: http://localhost:8085

**Features:**

**1. Queue Stats**
```
Queue: critical
├─ Pending: 45      (waiting to be processed)
├─ Active: 12       (currently processing)
├─ Completed: 5678  (successfully finished)
├─ Failed: 3        (moved to retry/archive)
└─ Latency: 0.2s    (time in queue)
```

**2. Task Details**
- Click any task to see:
  - Payload data
  - Retry count
  - Error messages
  - Processing time
  - Queue assignment

**3. Manual Actions**
- Retry failed tasks
- Delete tasks
- Archive tasks
- Pause/resume queues

**4. Real-time Updates**
- Watch tasks flow through system
- See processing in real-time
- Monitor queue depths

### Key Metrics to Watch

**1. Pending Count**
```
✅ Good:  Stable or decreasing
⚠️  Warning: Slowly increasing
❌ Bad:  Rapidly increasing (backlog!)
```

**2. Processing Rate**
```
✅ Good:  Processing rate > Enqueue rate
❌ Bad:  Processing rate < Enqueue rate
```

**3. Error Rate**
```
✅ Good:  < 1% errors
⚠️  Warning: 1-5% errors
❌ Bad:  > 5% errors
```

**4. Latency**
```
✅ Good:  < 1 second in queue
⚠️  Warning: 1-5 seconds
❌ Bad:  > 5 seconds
```

---

## 🎯 Best Practices

### 1. Task Design

**✅ DO:**
```go
// Idempotent tasks (safe to retry)
func HandlePaymentTask(ctx context.Context, task *asynq.Task) error {
    // Check if already processed
    if alreadyPaid(orderID) {
        return nil  // Skip, already done
    }
    
    // Process payment
    return processPayment(orderID)
}
```

**❌ DON'T:**
```go
// Non-idempotent (dangerous!)
func HandleIncrementTask(ctx context.Context, task *asynq.Task) error {
    // If retried, will increment multiple times!
    counter++
    return nil
}
```

### 2. Payload Size

**✅ DO:**
```go
// Small payload (just IDs)
payload := PaymentPayload{
    OrderID: "ORD-123",
    Amount:  1200.00,
}
```

**❌ DON'T:**
```go
// Large payload (entire order object)
payload := FullOrderWithAllItemsAndHistory{...}  // Too big!
```

### 3. Error Handling

**✅ DO:**
```go
func HandleTask(ctx context.Context, task *asynq.Task) error {
    // Transient error (retry)
    if err == ErrNetworkTimeout {
        return err  // Will retry
    }
    
    // Permanent error (don't retry)
    if err == ErrInvalidInput {
        log.Error("Invalid input, skipping")
        return nil  // Mark as done
    }
    
    return nil
}
```

### 4. Timeout Configuration

```go
// Set appropriate timeouts
asynq.NewTask(
    "payment:process",
    payload,
    asynq.Timeout(30*time.Second),  // Payment: 30s max
)

asynq.NewTask(
    "invoice:generate",
    payload,
    asynq.Timeout(2*time.Minute),  // PDF generation: 2min
)
```

---

## 🆚 Asynq vs Alternatives

| Feature | Asynq | Sidekiq | Celery | BullMQ |
|---------|-------|---------|--------|--------|
| **Language** | Go | Ruby | Python | Node.js |
| **Backend** | Redis | Redis | RabbitMQ/Redis | Redis |
| **Priority Queues** | ✅ Weighted | ✅ | ✅ | ✅ |
| **Retry** | ✅ Exponential | ✅ | ✅ | ✅ |
| **Monitoring** | ✅ Asynqmon | ✅ Web UI | ❌ | ✅ Bull Board |
| **Scheduled Tasks** | ✅ | ✅ | ✅ | ✅ |
| **Unique Tasks** | ✅ | ❌ | ❌ | ✅ |
| **Performance** | Very Fast | Fast | Medium | Fast |
| **Learning Curve** | Easy | Easy | Medium | Easy |

---

## 📚 Additional Resources

- **Asynq GitHub:** https://github.com/hibiken/asynq
- **Asynq Wiki:** https://github.com/hibiken/asynq/wiki
- **Asynqmon:** https://github.com/hibiken/asynqmon
- **Redis Commands:** https://redis.io/commands

---

## 🎓 Key Takeaways

1. **Asynq = Background job processing** - Don't block API responses
2. **Priority queues** - Process critical tasks first (weighted, not strict)
3. **Automatic retries** - Exponential backoff for resilience
4. **Scalable** - Add more workers easily
5. **Monitoring** - Asynqmon for real-time visibility
6. **Simple** - Redis-based, easy to setup and use

---

**Next:** Learn about load testing in [LOAD_TESTING.md](LOAD_TESTING.md)
