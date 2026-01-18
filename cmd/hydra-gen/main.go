package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
)

const (
	lowerChars = "abcdefghijklmnopqrstuvwxyz"
	upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars = "0123456789"
	specChars  = "_"
)

func main() {
	count := flag.Int("n", 10, "Number of passwords to generate")
	minBlocks := flag.Int("min", 6, "Minimum number of [a-zA-Z0-9_] blocks")
	maxBlocks := flag.Int("max", 6, "Maximum number of [a-zA-Z0-9_] blocks")
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
		numBlocks := *minBlocks
		if *maxBlocks > *minBlocks {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(*maxBlocks-*minBlocks+1)))
			numBlocks = *minBlocks + int(n.Int64())
		}

		password := generatePassword(numBlocks)
		fmt.Fprintln(out, password)
	}

	if *outputFile != "" {
		fmt.Fprintf(os.Stderr, "Successfully generated %d passwords to %s\n", *count, *outputFile)
	}
}

func generatePassword(blocks int) string {
	res := ""
	for i := 0; i < blocks; i++ {
		res += getRandomChar(lowerChars)
		res += getRandomChar(upperChars)
		res += getRandomChar(digitChars)
		res += getRandomChar(specChars)
	}
	return res
}

func getRandomChar(charset string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	return string(charset[n.Int64()])
}
