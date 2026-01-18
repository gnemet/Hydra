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
	userFile := os.Getenv("HYDRA_USER_FILE")
	passFile := os.Getenv("HYDRA_PASS_FILE")

	if url == "" || selector == "" {
		log.Fatal("HYDRA_URL and HYDRA_TARGET_SELECTOR must be set in .env")
	}

	timeout, _ := strconv.Atoi(timeoutStr)
	if timeout == 0 {
		timeout = 5
	}

	// 2. Load lists
	users, err := readLines(userFile)
	if err != nil {
		log.Fatalf("Error reading users: %v", err)
	}
	passwords, err := readLines(passFile)
	if err != nil {
		log.Fatalf("Error reading passwords: %v", err)
	}

	// 3. Initialize Hydra components
	f := fetcher.NewFetcher(timeout)
	e := evaluator.NewEvaluator(selector)

	fmt.Printf("--- 🐲 Hydra Brute Force Mode ---\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Users: %d, Passwords: %d\n", len(users), len(passwords))
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
				fmt.Printf("[ERROR: %v]\n", err)
				continue
			}

			result, err := e.Elevate(content)
			if err != nil {
				// We expect some elements might not be found if authentication fails
				// depending on how the server responds.
				// In our test server, the ID #response is always returned.
				fmt.Printf("[EVAL ERROR: %v]\n", err)
				continue
			}

			cleanResult := strings.TrimSpace(result)
			if strings.Contains(cleanResult, successText) {
				fmt.Printf("✅ SUCCESS!\n")
				fmt.Printf(">> Response: %s\n", cleanResult)
				found = true
				// We can break or continue. Break to find first valid.
				break
			} else {
				fmt.Printf("❌ FAILED\n")
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
