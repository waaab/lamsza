#!/bin/bash

# Configuration
BACKEND_DIR="backend"
BACKEND_PID_FILE="$BACKEND_DIR/server.pid"
FRONTEND_PID_FILE="server.pid"
BACKEND_LOG="backend/server_backend.log"
FRONTEND_LOG="server_frontend.log"

echo "🔄 Restarting all services..."

# 1. Stop Services
echo "🛑 Stopping services..."

# Stop Frontend
if [ -f "$FRONTEND_PID_FILE" ]; then
    PID=$(cat "$FRONTEND_PID_FILE")
    kill $PID 2>/dev/null || pkill -f vite
    rm "$FRONTEND_PID_FILE"
else
    pkill -f vite
fi

# Stop Backend
if [ -f "$BACKEND_PID_FILE" ]; then
    PID=$(cat "$BACKEND_PID_FILE")
    kill $PID 2>/dev/null || pkill -f "go run main.go"
    rm "$BACKEND_PID_FILE"
else
    pkill -f "go run main.go"
fi

# Stop Database
docker compose down

echo "✅ All services stopped."

# 2. Start Services
echo "🚀 Starting services in separate background processes..."

# Start Database
echo "📦 Starting Database..."
docker compose up -d

# Start Backend
echo "⚙️ Starting Backend (log: $BACKEND_LOG)..."
(cd "$BACKEND_DIR" && go run main.go > server_backend.log 2>&1 & echo $! > server.pid)

# Start Frontend
echo "🌐 Starting Frontend (log: $FRONTEND_LOG)..."
(npm run dev > "$FRONTEND_LOG" 2>&1 & echo $! > "$FRONTEND_PID_FILE")

echo "✨ All services restarted successfully!"
echo "--------------------------------------------------"
echo "To view logs:"
echo "  Backend: tail -f $BACKEND_LOG"
echo "  Frontend: tail -f $FRONTEND_LOG"
echo "--------------------------------------------------"
