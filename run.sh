#!/bin/bash

# Exit on error
set -e

echo "--- 🚀 Launching Hydra Ecosystem ---"

# 1. Clean and Build
echo "📦 Building programs..."
make clean
make build-all

# 2. Check if something is already on port 8082
if lsof -Pi :8082 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Port 8082 is already in use. Attempting to kill existing process..."
    fuser -k 8082/tcp || true
    sleep 1
fi

# 3. Start Test Server in background
echo "🌐 Starting Test Server at http://localhost:8082..."
./bin/testserver &
SERVER_PID=$!

# Give the server a moment to start
sleep 1

# Ensure the server is killed when the script exits
trap "echo '🛑 Stopping Test Server...'; kill $SERVER_PID" EXIT

# 4. Run Hydra with test configuration
echo "🐲 Running Hydra with test_config.yaml..."
echo "----------------------------------------"
./bin/hydra configs/test_config.yaml
echo "----------------------------------------"

echo "✅ Hydra execution complete."
echo "Keep server running? (y/n)"
read -t 5 keep_running || keep_running="n"

if [ "$keep_running" == "y" ]; then
    echo "Server is staying up (PID: $SERVER_PID). Use 'kill $SERVER_PID' to stop it later."
    trap - EXIT
else
    echo "Shutting down..."
fi
