# Load Testing Results

Performance results from K6 load tests with screenshots.

---

## Test Environment

### Configuration

```yaml
Worker:
  - Processes: 1
  - Concurrency: 20 goroutines
  - Queue Weights:
      critical: 6
      high: 4
      default: 2
      low: 1

Infrastructure:
  - PostgreSQL 15 (Docker)
  - Redis 7 (Docker)
  - Asynqmon (monitoring)

Hardware:
  - CPU: Apple M1/M2 (or equivalent)
  - RAM: 8GB+
  - OS: macOS / Linux
```

---

## 1. Basic Load Test

**Command:** `k6 run loadtest/basic-load.js`

**Load Pattern:**
- Ramp up: 0 → 20 users (30s)
- Ramp up: 20 → 50 users (1m)
- Sustained: 50 users (2m)
- Ramp down: 50 → 0 users (30s)
- **Total Duration:** 4 minutes

### K6 Output

<!-- Screenshot: Terminal output showing K6 results -->
![Basic Load Test Results](screenshots/basic-load-k6.png)

**Key Metrics:**
- Response Time (avg): 10.16ms
- Response Time (p95): 44.97ms
- Throughput: 73 req/s
- Error Rate: 0%
- Total Requests: 17,556

### Asynqmon Dashboard

<!-- Screenshot: Asynqmon showing queue stats -->
![Asynqmon Queue Stats](screenshots/basic-load-asynqmon.png)

**Queue Processing:**
- All 6 task types processed successfully
- No queue backlog
- 100% success rate

---

## 2. Stress Test

**Command:** `k6 run loadtest/stress-test.js`

**Load Pattern:**
- Gradual increase: 50 → 100 → 200 → 300 → 400 users
- **Total Duration:** 10 minutes

### K6 Output

<!-- Screenshot: Terminal output showing stress test results -->
![Stress Test Results](screenshots/stress-test-k6.png)

**Breaking Point:**
- (Add results after running test)
- System capacity: XXX req/s
- Performance degradation at: XXX users

### Asynqmon During Peak Load

<!-- Screenshot: Asynqmon during high load -->
![Asynqmon Under Stress](screenshots/stress-test-asynqmon.png)

---

## 3. Spike Test

**Command:** `k6 run loadtest/spike-test.js`

**Load Pattern:**
- Warm up: 10 users (30s)
- **SPIKE:** 10 → 200 users (10s)
- Sustained: 200 users (1m)
- Drop: 200 → 10 users (10s)
- **Total Duration:** 2.5 minutes

### K6 Output

<!-- Screenshot: Terminal output showing spike test results -->
![Spike Test Results](screenshots/spike-test-k6.png)

**Recovery:**
- System recovery time: XX seconds
- Error rate during spike: X%

### Asynqmon Recovery

<!-- Screenshot: Asynqmon showing queue recovery -->
![Asynqmon Queue Recovery](screenshots/spike-test-asynqmon.png)

---

## Summary

### Performance Highlights

| Test | Users | Duration | Throughput | Avg Response | Error Rate |
|------|-------|----------|------------|--------------|------------|
| **Basic Load** | 50 | 4m | 73 req/s | 10ms | 0% |
| **Stress** | 400 | 10m | TBD | TBD | TBD |
| **Spike** | 200 | 2.5m | TBD | TBD | TBD |

### Key Findings

- ✅ System handles 50 concurrent users with excellent performance
- ✅ API response time consistently under 50ms (p95)
- ✅ Zero errors with automatic retry mechanism
- ✅ Workers process background tasks faster than enqueue rate
- ✅ Priority queues working as expected

### Bottlenecks

- **Current:** Worker processing capacity
- **Recommendation:** Scale workers horizontally for higher throughput
- **Estimated capacity:** 200+ req/s with 3 workers

---

## How to Reproduce

```bash
# 1. Start system
docker-compose up -d
go run cmd/api/main.go       # Terminal 1
go run cmd/worker/main.go    # Terminal 2

# 2. Run tests
k6 run loadtest/basic-load.js   # Terminal 3
k6 run loadtest/stress-test.js
k6 run loadtest/spike-test.js

# 3. Monitor
# Open http://localhost:8085 (Asynqmon)
```

---

## Screenshots

To add screenshots after running tests:

1. Run each test script
2. Take screenshot of K6 terminal output
3. Take screenshot of Asynqmon dashboard
4. Save to `docs/screenshots/` directory
5. Update image links in this file

**Recommended screenshot tool:**
- macOS: `Cmd + Shift + 4`
- Linux: `gnome-screenshot` or `spectacle`
- Windows: `Win + Shift + S`

---

**Last Updated:** [Add date after running tests]
