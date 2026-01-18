package main

import (
	"fmt"
	"hydra/internal/generator"
)

func main() {
	s1 := "baseball"
	s2 := "Baseball12"
	fmt.Printf("Similarity '%s' vs '%s': %.4f\n", s1, s2, generator.TrigramSimilarity(s1, s2))

	s3 := "bAse12ball"
	fmt.Printf("Similarity '%s' vs '%s': %.4f\n", s1, s3, generator.TrigramSimilarity(s1, s3))
}
