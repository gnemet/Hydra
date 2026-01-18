# 🐉 Hydra: NAS Password Recovery Suite

Hydra is a high-performance, parallelized password recovery tool written in Go. It is specifically optimized for recovering forgotten credentials from web-based interfaces, such as those found on **Thecus N4100/PRO** and other NAS devices.

## 🚀 Key Features

*   **Multi-Seed Mutation**: Generate thousands of password variants from a few basic "seed" memory fragments.
*   **Parallel Orchestration**: Utilizes multiple CPU cores to run brute-force attacks in parallel for maximum speed.
*   **Intelligent Resume**: Built-in state management. If the process is interrupted, it resumes exactly where it left off.
*   **NAS Simulator**: Includes a built-in Thecus N4100 simulator for testing your configuration before the real attack.
*   **Regex-Constrained Search**: Generates passwords that strictly follow your known habits (length, special characters).
*   **Radical Fallback**: If "Smart Mutations" fail, Phase 2 automatically triggers a broad pattern-based search.

---

## 🛠️ Configuration (.env)

All settings are managed in your `.env` file for portability.

| Variable | Description |
| :--- | :--- |
| `HYDRA_URL` | The login URL of your NAS (e.g., `http://192.168.1.100/adm/login.php`). |
| `HYDRA_USER_FIELD` | The HTML `name` attribute for the username field (Thecus: `u_name`). |
| `HYDRA_PASS_FIELD` | The HTML `name` attribute for the password field (Thecus: `u_pwd`). |
| `HYDRA_SUCCESS_TEXT` | String that appear only on successful login (e.g., "Control Panel"). |
| `HYDRA_ERROR_TEXT` | String that appears on failed attempts (e.g., "Retry"). |
| `HYDRA_PASS_REGEX` | Regex rule defining the password structure. |
| `HYDRA_THREAD_COUNT` | Number of parallel worker threads. |
| `HYDRA_GEN_COUNT` | Total number of unique passwords to generate per seed. |

---

## 📖 Usage Guide

### 1. Direct Recovery
If you have the NAS connected and `.env` configured:
```bash
./parallel.sh
```

### 2. Testing with Simulator
To verify your settings work against a simulated Thecus NAS:
1. Set `HYDRA_URL=http://localhost:8082/adm/login.php` in `.env`.
2. Run `./parallel.sh`. 
3. The script will automatically start the `testserver` and verify the logic.

### 3. Deploying to another machine
1. Run the packager:
   ```bash
   ./package.sh
   ```
2. Copy the resulting `hydra_dist/` folder to the target machine.

---

## ♻️ State & Resuming
Hydra automatically creates `.state` files in the `temp_lists/` directory. 
- When re-running `parallel.sh`, follow the prompt: `Resume last session? (y/n)`.
- Choosing **y** will skip regeneration and pick up the brute-force exactly at the last tested index.

---

## 📁 Project Structure

*   `cmd/hydra-gen/`: Password mutation and pattern engine.
*   `cmd/hydra-brute/`: HTTP form submission and response evaluator.
*   `cmd/testserver/`: Thecus N4100 web interface simulator.
*   `parallel.sh`: The master orchestrator script.
*   `internal/generator/`: Core mutation algorithms (Leetspeak, Case-toggling, Padding).

---

## 🛡️ Best Practices for NAS Recovery
1. **Thread Count**: Keep `HYDRA_THREAD_COUNT` low (4-6) for older NAS devices to avoid crashing their HTTP service.
2. **Timeout**: Use a generous `HYDRA_TIMEOUT` (10s+) as old NAS CPUs take time to hash passwords.
3. **Connectivity**: Always use a direct UTP cable and set a static IP on your host laptop (e.g., `192.168.1.101`).
