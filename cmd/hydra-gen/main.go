package main

import (
	"flag"
	"fmt"
	"os"

	"hydra/internal/generator"
)

func main() {
	count := flag.Int("n", 10, "Number of passwords to generate")
	minBlocks := flag.Int("min", 6, "Minimum number of blocks")
	maxBlocks := flag.Int("max", 10, "Maximum number of blocks")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
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
		password, _ := generator.GenerateByBlockPattern(*minBlocks, *maxBlocks)
		fmt.Fprintln(out, password)
	}
}
