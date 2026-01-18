package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
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
	useSequential := flag.Bool("sequential", false, "Use exhaustive sequential brute force")
	useCombinatorial := flag.Bool("combine", false, "Combine seeds with each other")
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

	// 1. COMBINATORIAL: Combine seeds (Highest likelihood)
	if *useCombinatorial && len(basePasswords) > 1 {
		for _, s1 := range basePasswords {
			for idx2, s2 := range basePasswords {
				if generatedCount >= *count {
					break
				}
				// Skip same string unless combined with others
				combos := []string{
					s1 + s2,
					strings.Title(s1) + s2,
					s1 + strings.Title(s2),
					strings.Title(s1) + strings.Title(s2),
				}
				if idx2 < len(basePasswords)-1 {
					combos = append(combos, s1+basePasswords[idx2+1])
				}

				for _, combo := range combos {
					if len(combo) >= *minLen && len(combo) <= *maxLen {
						if compiledRegex == nil || compiledRegex.MatchString(combo) {
							if !seen[combo] {
								seen[combo] = true
								generatedCount++
							}
						}
					}
				}
			}
		}
	}

	// 2. SEQUENTIAL: Systematic systematic search
	if *useSequential {
		charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#_-"
		it := generator.NewSequentialIterator(charset, *minLen, *maxLen)
		for generatedCount < *count {
			p, ok := it.Next()
			if !ok {
				break
			}
			if compiledRegex != nil && !compiledRegex.MatchString(p) {
				continue
			}
			if !seen[p] {
				seen[p] = true
				generatedCount++
			}
		}
	}

	// 3. MUTATION / RANDOM: Final fill
	maxRetries := *count * (*maxRetriesFactor)
	retries := 0
	for generatedCount < *count && retries < maxRetries {
		retries++
		var password string
		adjMin := *minLen - len(*prefix)
		adjMax := *maxLen - len(*prefix)

		if *useMutation && len(basePasswords) > 0 {
			idx, _ := generator.GetRandIdx(int64(len(basePasswords)))
			seed := basePasswords[idx]
			password, _ = generator.Mutate(seed, adjMin, adjMax)
		} else if *usePattern {
			password, _ = generator.GenerateByBlockPattern(adjMin, adjMax)
		} else {
			password, _ = generator.GenerateVaried(adjMin, adjMax)
		}

		password = *prefix + password

		// Uniqueness
		if seen[password] {
			continue
		}

		// Regex
		if compiledRegex != nil && !compiledRegex.MatchString(password) {
			continue
		}

		// Similarity Check (restored)
		if len(basePasswords) > 0 && *threshold > 0 {
			isSimilar := false
			for _, base := range basePasswords {
				if generator.TrigramSimilarity(password, base) >= *threshold {
					isSimilar = true
					break
				}
			}
			if !isSimilar {
				continue
			}
		}

		seen[password] = true
		generatedCount++
	}

	// 4. Final Complexity Sort
	var finalPasswords []string
	for p := range seen {
		finalPasswords = append(finalPasswords, p)
	}

	sort.Slice(finalPasswords, func(i, j int) bool {
		return generator.CalculateComplexity(finalPasswords[i]) < generator.CalculateComplexity(finalPasswords[j])
	})

	for _, p := range finalPasswords {
		fmt.Fprintln(out, p)
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
