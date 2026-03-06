---
description: start, stop and restart project services (db, backend, frontend)
---

This workflow provides standardized commands to manage the development environment. **Always run these in separate terminals when working manually.**

### 1. Start All Services
Starts the database, Go backend, and Vite frontend.

// turbo
```bash
# Start Database
docker compose up -d
# Start Backend
cd backend && go run main.go > server_backend.log 2>&1 & echo $! > server.pid
# Start Frontend
cd .. && npm run dev > server_frontend.log 2>&1 & echo $! > server.pid
```

**Manual Start (Use 3 Separate Terminals):**
1. **Terminal 1 (DB):** `docker compose up`
2. **Terminal 2 (Backend):** `cd backend && go run main.go`
3. **Terminal 3 (Frontend):** `npm run dev`

### 2. Stop All Services
Terminates all running processes and stops the database.

// turbo
```bash
# Stop Frontend
if [ -f server.pid ]; then kill $(cat server.pid) || pkill -f vite; rm server.pid; else pkill -f vite; fi
# Stop Backend
if [ -f backend/server.pid ]; then kill $(cat backend/server.pid) || pkill -f "go run main.go"; rm backend/server.pid; else pkill -f "go run main.go"; fi
# Stop Database
docker compose down
```

**Manual Stop:**
- `pkill -f vite`
- `pkill -f "go run main.go"`
- `docker compose down`

### 3. Restart All Services
Performs a full stop and start.

// turbo
```bash
# Stop All
if [ -f server.pid ]; then kill $(cat server.pid) || pkill -f vite; rm server.pid; else pkill -f vite; fi
if [ -f backend/server.pid ]; then kill $(cat backend/server.pid) || pkill -f "go run main.go"; rm backend/server.pid; else pkill -f "go run main.go"; fi
docker compose down

# Start All
docker compose up -d
cd backend && go run main.go > server_backend.log 2>&1 & echo $! > server.pid
cd ..
npm run dev > server_frontend.log 2>&1 & echo $! > server.pid
```
鼓
