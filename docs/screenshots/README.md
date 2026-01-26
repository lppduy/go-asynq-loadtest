# Screenshots Directory

Place load testing screenshots here:

- basic-load-k6.png
- basic-load-asynqmon.png
- stress-test-k6.png
- stress-test-asynqmon.png
- spike-test-k6.png
- spike-test-asynqmon.png

---

## How to Take Screenshots

### 1. Clean Environment Before Each Test

**Important:** Reset data between tests for accurate results:

```bash
# Stop and remove all data
docker-compose down -v

# Start fresh
docker-compose up -d

# Wait for services to be ready (5-10 seconds)
sleep 10
```

### 2. Start Services

```bash
# Terminal 1: API
go run cmd/api/main.go

# Terminal 2: Worker
go run cmd/worker/main.go
```

### 3. Run Test

```bash
# Terminal 3: K6
k6 run loadtest/basic-load.js
```

### 4. Take Screenshots

**During/After test:**
1. **K6 Terminal Output** - Capture the summary at the end
2. **Asynqmon Dashboard** (http://localhost:8085) - Capture queue stats

### 5. Repeat for Each Test

- Basic Load Test
- Stress Test
- Spike Test

**Remember to run `docker-compose down -v` between tests!**
