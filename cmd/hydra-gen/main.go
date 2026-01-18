package main

import (
	"flag"
	"fmt"
	"os"

	"hydra/internal/generator"
)

func main() {
	count := flag.Int("n", 10, "Number of passwords to generate")
	minLen := flag.Int("min", 6, "Minimum length")
	maxLen := flag.Int("max", 10, "Maximum length")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	usePattern := flag.Bool("pattern", false, "Use complex block pattern ([a-z][A-Z][0-9]_){n}")
	flag.Parse()

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

	for i := 0; i < *count; i++ {
		var password string
		if *usePattern {
			password, _ = generator.GenerateByBlockPattern(*minLen, *maxLen)
		} else {
			password, _ = generator.GenerateVaried(*minLen, *maxLen)
		}
		fmt.Fprintln(out, password)
	}
}
