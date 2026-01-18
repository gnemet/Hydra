package generator

import (
	"strings"
)

// TrigramSimilarity calculates the similarity between two strings using trigrams.
// This is similar to PostgreSQL's similarity() function from pg_trgm.
func TrigramSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 1.0
	}

	t1 := getTrigrams(s1)
	t2 := getTrigrams(s2)

	if len(t1) == 0 && len(t2) == 0 {
		return 1.0
	}
	if len(t1) == 0 || len(t2) == 0 {
		return 0.0
	}

	common := 0
	for trigram := range t1 {
		if _, ok := t2[trigram]; ok {
			common++
		}
	}

	// Jaccard similarity: |A ∩ B| / |A ∪ B|
	// |A ∪ B| = |A| + |B| - |A ∩ B|
	union := len(t1) + len(t2) - common
	return float64(common) / float64(union)
}

func getTrigrams(s string) map[string]struct{} {
	// Standard pg_trgm behavior: pad with two spaces at the beginning and one at the end
	padded := "  " + s + " "
	trigrams := make(map[string]struct{})
	for i := 0; i < len(padded)-2; i++ {
		trigrams[padded[i:i+3]] = struct{}{}
	}
	return trigrams
}
