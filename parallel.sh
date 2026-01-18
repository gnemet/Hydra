#!/bin/bash

# Load configuration
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

THREADS=${HYDRA_THREAD_COUNT:-4}
GEN_COUNT=${HYDRA_GEN_COUNT:-1000}
MASTER_PASS="master_passwords.txt"
TEMP_DIR="temp_lists"

echo "--- 🚀 Hydra Parallel Orchestrator ---"
echo "Threads: $THREADS"
echo "Target: $HYDRA_URL"

# 1. Build
make build-all

# 2. Prepare Directories
rm -rf $TEMP_DIR
mkdir -p $TEMP_DIR

# 3. Generate Master List
echo "📦 Generating $GEN_COUNT passwords..."
./bin/hydra-gen -n $GEN_COUNT > $MASTER_PASS

# 4. Split List
echo "✂️  Splitting into $THREADS parts..."
total_lines=$(wc -l < $MASTER_PASS)
lines_per_file=$(( (total_lines + THREADS - 1) / THREADS ))

split -l $lines_per_file -d --additional-suffix=.txt $MASTER_PASS $TEMP_DIR/part_

# 5. Start Test Server (if not running on 8082)
if ! lsof -i:8082 > /dev/null; then
    echo "🌐 Starting Test Server..."
    ./bin/testserver > /dev/null 2>&1 &
    SERVER_PID=$!
    sleep 1
    trap "kill $SERVER_PID" EXIT
fi

# 6. Launch Parallel Brute Force
echo "🔥 Launching $THREADS threads..."
echo "----------------------------------------"

for f in $TEMP_DIR/part_*.txt; do
    # Run in background, redirect output to a log file per thread
    ./bin/hydra-brute "$f" > "${f%.txt}.log" 2>&1 &
    pids+=($!)
done

# 7. Monitor (Simple wait)
echo "Waiting for threads to complete..."
for pid in "${pids[@]}"; do
    wait $pid
done

echo "----------------------------------------"
echo "✅ Parallel Scan Complete."

# 8. Check for successes in logs
if grep -r "SUCCESS" $TEMP_DIR/*.log; then
    echo "🎯 FOUND SUCCESSES:"
    grep -h -A 1 "SUCCESS" $TEMP_DIR/*.log | grep "Response:"
else
    echo "❌ No valid credentials found in this run."
fi

# Clean up
# rm -rf $TEMP_DIR $MASTER_PASS
