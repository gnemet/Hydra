package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"hydra/internal/generator"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Default values from environment or fallback
	defCount := getEnvInt("HYDRA_GEN_COUNT", 10)
	defMin := getEnvInt("HYDRA_MIN_LEN", 6)
	defMax := getEnvInt("HYDRA_MAX_LEN", 10)

	// Attempt to extract range from HYDRA_PASS_REGEX (e.g., [a-z]{6,10})
	passRegex := os.Getenv("HYDRA_PASS_REGEX")
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

	defPattern := getEnvBool("HYDRA_USE_PATTERN", false)
	defSimFile := os.Getenv("HYDRA_SIMILARITY_FILE")
	defThreshold := getEnvFloat("HYDRA_SIMILARITY_THRESHOLD", 0.0)
	defPrefix := os.Getenv("HYDRA_PREFIX")

	defMaxRetriesFactor := getEnvInt("HYDRA_MAX_RETRIES_FACTOR", 20)

	count := flag.Int("n", defCount, "Number of passwords to generate")
	minLen := flag.Int("min", defMin, "Minimum length")
	maxLen := flag.Int("max", defMax, "Maximum length")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	usePattern := flag.Bool("pattern", defPattern, "Use complex block pattern ([a-z][A-Z][0-9]_){n}")
	simFile := flag.String("simfile", defSimFile, "File with base passwords for similarity check")
	threshold := flag.Float64("threshold", defThreshold, "Similarity threshold (0.0 to 1.0)")
	prefix := flag.String("prefix", defPrefix, "Constant prefix for all generated passwords")
	simPass := flag.String("simpass", "", "Single base password for similarity check")
	useMutation := flag.Bool("mutate", false, "Use mutation strategies instead of random generation")
	maxRetriesFactor := flag.Int("retries-factor", defMaxRetriesFactor, "Retries factor (max_retries = n * factor)")
	flag.Parse()

	var basePasswords []string
	if *simPass != "" {
		basePasswords = append(basePasswords, *simPass)
	}

	if *simFile != "" {
		data, err := os.ReadFile(*simFile)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					basePasswords = append(basePasswords, trimmed)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Could not read similarity file %s: %v\n", *simFile, err)
		}
	}

	var out *os.File = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	var compiledRegex *regexp.Regexp
	if passRegex != "" {
		compiledRegex, _ = regexp.Compile("^" + passRegex + "$")
	}

	generatedCount := 0
	seen := make(map[string]bool)
	maxRetries := *count * (*maxRetriesFactor)
	retries := 0

	// Strategy: If using mutation with a simpass, always try the seed(s) first
	if *useMutation {
		for _, seed := range basePasswords {
			if generatedCount < *count {
				if compiledRegex == nil || compiledRegex.MatchString(seed) {
					fmt.Fprintln(out, seed)
					seen[seed] = true
					generatedCount++
				}
			}
		}
	}

	for generatedCount < *count && retries < maxRetries {
		retries++
		var password string
		// Adjust length for prefix
		adjMin := *minLen - len(*prefix)
		adjMax := *maxLen - len(*prefix)
		if adjMin < 0 {
			adjMin = 0
		}
		if adjMax < adjMin {
			adjMax = adjMin
		}

		if *useMutation && len(basePasswords) > 0 {
			// Pick a random base password as seed
			idx, _ := generator.GetRandIdx(int64(len(basePasswords)))
			seed := basePasswords[idx]
			password, _ = generator.Mutate(seed, adjMin, adjMax)
		} else if *usePattern {
			password, _ = generator.GenerateByBlockPattern(adjMin, adjMax)
		} else {
			password, _ = generator.GenerateVaried(adjMin, adjMax)
		}

		password = *prefix + password

		if seen[password] {
			continue
		}

		// Regex Validation
		if compiledRegex != nil && !compiledRegex.MatchString(password) {
			continue
		}

		// Double check similarity to any base if threshold is set (safety check)
		if len(basePasswords) > 0 && *threshold > 0 {
			isSimilar := false
			for _, base := range basePasswords {
				if generator.TrigramSimilarity(password, base) >= *threshold {
					isSimilar = true
					break
				}
			}
			if !isSimilar {
				continue // Try again
			}
		}

		fmt.Fprintln(out, password)
		seen[password] = true
		generatedCount++
	}
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}
