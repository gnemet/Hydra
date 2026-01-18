#!/bin/bash

# Load configuration
if [ -f .env ]; then
    set -a
    . .env
    set +a
fi

GEN_COUNT=${HYDRA_GEN_COUNT:-1000}
TEMP_DIR=${HYDRA_TEMP_DIR:-"temp_lists"}
BASE_PASS_FILE=${HYDRA_PASS_FILE:-"passwords.txt"}

# 1. Build
make build-all

# 2. Prepare Directories
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

echo "--- 🚀 Hydra Parallel Orchestrator (Multi-Seed Generation) ---"
echo "Target: $HYDRA_URL"

# 3. Read Base Passwords
if [ ! -f "$BASE_PASS_FILE" ]; then
    echo "❌ Error: Base password file $BASE_PASS_FILE not found."
    exit 1
fi

mapfile -t base_seeds < "$BASE_PASS_FILE"
num_seeds=${#base_seeds[@]}

if [ "$num_seeds" -eq 0 ]; then
    echo "❌ Error: No seeds found in $BASE_PASS_FILE."
    exit 1
fi

echo "Found $num_seeds base seeds in $BASE_PASS_FILE."
per_seed=$(( (GEN_COUNT + num_seeds - 1) / num_seeds ))

# 4. Parallel Generation (One thread per seed)
echo "📦 Phase 1/2: Generating passwords..."
pids_gen=()
for i in "${!base_seeds[@]}"; do
    seed="${base_seeds[$i]}"
    ./bin/hydra-gen -n "$per_seed" -simpass "$seed" -simfile "" -mutate > "$TEMP_DIR/part_$(printf "%02d" $i).txt" 2> /dev/null &
    pids_gen+=($!)
done

# Monitor Generation Progress
while true; do
    still_running=0
    for pid in "${pids_gen[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then ((still_running++)); fi
    done
    
    current_count=$(cat "$TEMP_DIR"/part_*.txt 2>/dev/null | wc -l)
    percent=$(( current_count * 100 / GEN_COUNT ))
    if [ "$percent" -gt 100 ]; then percent=100; fi
    
    printf "\r   ⮕  Progress: [%-50s] %d%% (%d/%d)" "$(printf "%$((percent/2))s" | tr ' ' '#')" "$percent" "$current_count" "$GEN_COUNT"
    
    if [ "$still_running" -eq 0 ]; then break; fi
    sleep 0.5
done
echo -e "\n✅ Generation Complete."

# 5. Start Test Server (if not running on 8082)
if ! lsof -i:8082 > /dev/null; then
    echo "🌐 Starting Test Server..."
    ./bin/testserver > /dev/null 2>&1 &
    SERVER_PID=$!
    sleep 1
    trap "kill $SERVER_PID" EXIT
fi

# 6. Launch Parallel Brute Force
TOTAL_FOR_BRUTE=$(cat "$TEMP_DIR"/part_*.txt 2>/dev/null | wc -l)
echo "🔥 Phase 2/2: Launching brute force ($TOTAL_FOR_BRUTE passwords)..."

pids_brute=()
for f in "$TEMP_DIR"/part_*.txt; do
    if [ -f "$f" ]; then
        ./bin/hydra-brute "$f" > "${f%.txt}.log" 2>&1 &
        pids_brute+=($!)
    fi
done

# 7. Monitor Brute Force Progress
success_file=""
while true; do
    still_running=0
    for pid in "${pids_brute[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then ((still_running++)); fi
    done
    
    # Check for success
    found_log=$(grep -l "SUCCESS" "$TEMP_DIR"/*.log 2>/dev/null | head -n 1)
    if [ -n "$found_log" ]; then
        success_file="$found_log"
        # Kill all remaining brute force processes
        for pid in "${pids_brute[@]}"; do
            kill "$pid" 2>/dev/null
        done
        break
    fi

    tested_count=$(grep -c "Testing:" "$TEMP_DIR"/*.log 2>/dev/null | awk -F: '{sum+=$2} END {print sum+0}')
    if [ "$TOTAL_FOR_BRUTE" -gt 0 ]; then
        percent=$(( tested_count * 100 / TOTAL_FOR_BRUTE ))
    else
        percent=0
    fi
    if [ "$percent" -gt 100 ]; then percent=100; fi
    
    printf "\r   ⮕  Progress: [%-50s] %d%% (%d/%d)" "$(printf "%$((percent/2))s" | tr ' ' '#')" "$percent" "$tested_count" "$TOTAL_FOR_BRUTE"
    
    if [ "$still_running" -eq 0 ]; then break; fi
    sleep 1
done

echo -e "\n----------------------------------------"
echo "✅ Parallel Scan Complete."

# 8. Check for successes in logs
if [ -n "$success_file" ]; then
    echo "🎯 FOUND SUCCESS!"
    # Extract the successful line, e.g., "Testing: admin:Secret123 ... ✅ SUCCESS!"
    success_line=$(grep "SUCCESS" "$success_file" | head -n 1)
    creds=$(echo "$success_line" | awk -F'Testing: ' '{print $2}' | awk -F' ' '{print $1}')
    echo "👤 User/Pass: $creds"
    grep -h "Response:" "$success_file"
else
    # Final check just in case it finished and then we checked
    if grep -r "SUCCESS" "$TEMP_DIR"/*.log > /dev/null; then
         echo "🎯 FOUND SUCCESS!"
         grep -h -B 1 "SUCCESS" "$TEMP_DIR"/*.log
    else
         echo "❌ No valid credentials found in this run."
    fi
fi
