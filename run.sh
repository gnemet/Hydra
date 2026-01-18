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

# 3. Mode Selection
echo "Select option:"
echo "1) Single Scrape (configs/test_config.yaml)"
echo "2) Brute Force Mode (.env + lists)"
echo "3) Password Generator (regex variations)"
read -p "Selection (1/2/3): " mode

if [ "$mode" == "3" ]; then
    echo "🔑 Generator Settings:"
    read -p "How many passwords? (default 10): " count
    count=${count:-10}
    echo "Generating $count varied variations (6-10 chars) into passwords.txt..."
    ./bin/hydra-gen -n $count >> passwords.txt
    echo "✅ Done. You can now run Brute Force mode."
    exit 0
fi

# 4. Start Test Server in background
echo "🌐 Starting Test Server at http://localhost:8082..."
./bin/testserver &
SERVER_PID=$!

# Give the server a moment to start
sleep 1

# Ensure the server is killed when the script exits
trap "echo '🛑 Stopping Test Server...'; kill $SERVER_PID" EXIT

if [ "$mode" == "2" ]; then
    echo "🐲 Running Hydra Brute Force..."
    echo "----------------------------------------"
    ./bin/hydra-brute
    echo "----------------------------------------"
else
    echo "🐲 Running Hydra Single Scrape..."
    echo "----------------------------------------"
    ./bin/hydra configs/test_config.yaml
    echo "----------------------------------------"
fi

echo "✅ Hydra execution complete."
echo "Keep server running? (y/n)"
read -t 5 keep_running || keep_running="n"

if [ "$keep_running" == "y" ]; then
    echo "Server is staying up (PID: $SERVER_PID). Use 'kill $SERVER_PID' to stop it later."
    trap - EXIT
else
    echo "Shutting down..."
fi
