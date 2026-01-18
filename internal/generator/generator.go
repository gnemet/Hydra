package generator

import (
	"crypto/rand"
	"math/big"
)

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	specChars   = "!@#_-"
	commonChars = lowerChars + upperChars + digitChars + specChars
)

// GenerateRandomFromSet generates a random password of length [minLen, maxLen] using the provided charset
func GenerateRandomFromSet(minLen, maxLen int) (string, error) {
	return generateWithCharset(minLen, maxLen, commonChars)
}

// GenerateVaried picks a random 'style' for the password to create more realistic variations
func GenerateVaried(minLen, maxLen int) (string, error) {
	styles := []string{
		lowerChars,
		digitChars,
		lowerChars + digitChars,
		upperChars + digitChars,
		lowerChars + upperChars + digitChars,
		commonChars,
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(styles))))
	selectedCharset := styles[n.Int64()]

	return generateWithCharset(minLen, maxLen, selectedCharset)
}

func generateWithCharset(minLen, maxLen int, charset string) (string, error) {
	length := minLen
	if maxLen > minLen {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(maxLen-minLen+1)))
		if err != nil {
			return "", err
		}
		length = minLen + int(n.Int64())
	}

	res := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		res[i] = charset[idx.Int64()]
	}
	return string(res), nil
}

// GenerateByBlockPattern generates a password based on the ([a-z][A-Z][0-9][_]){min,max} pattern
func GenerateByBlockPattern(minBlocks, maxBlocks int) (string, error) {
	numBlocks := minBlocks
	if maxBlocks > minBlocks {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(maxBlocks-minBlocks+1)))
		if err != nil {
			return "", err
		}
		numBlocks = minBlocks + int(n.Int64())
	}

	res := ""
	for i := 0; i < numBlocks; i++ {
		res += getRandomChar(lowerChars)
		res += getRandomChar(upperChars)
		res += getRandomChar(digitChars)
		res += getRandomChar(specChars)
	}
	return res, nil
}

func getRandomChar(charset string) string {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	return string(charset[idx.Int64()])
}

// Mutate takes a seed string and applies variations requested by the user
func Mutate(seed string, minLen, maxLen int) (string, error) {
	mut := []rune(seed)
	if len(mut) == 0 {
		return GenerateRandomFromSet(minLen, maxLen)
	}

	// Apply 1 to 3 random mutations to increase the state space
	numMuts, _ := rand.Int(rand.Reader, big.NewInt(3))
	for i := 0; i <= int(numMuts.Int64()); i++ {
		strategy, _ := rand.Int(rand.Reader, big.NewInt(8)) // Increased to 8 strategies
		switch strategy.Int64() {
		case 0: // Variations in case: Randomly toggle case of letters
			for j, r := range mut {
				c, _ := rand.Int(rand.Reader, big.NewInt(10))
				if c.Int64() < 2 {
					if r >= 'a' && r <= 'z' {
						mut[j] = r - 'a' + 'A'
					} else if r >= 'A' && r <= 'Z' {
						mut[j] = r - 'A' + 'a'
					}
				}
			}
		case 1: // Start with a random number
			n := rune(digitChars[getRandIdxString(digitChars)])
			mut = append([]rune{n}, mut...)
		case 2: // End with a common sequence (expanded)
			vals := []string{
				"123", "2010", "2011", "2012", "2013", "2014", "2015", "2016", "2017", "2018", "2019",
				"2020", "2021", "2022", "2023", "2024", "2025", "2026", "3125", "5213", "!", "!!", "@", "1",
			}
			suffixIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(vals))))
			mut = append(mut, []rune(vals[suffixIdx.Int64()])...)
		case 3: // Include underscore
			pos, _ := rand.Int(rand.Reader, big.NewInt(int64(len(mut)+1)))
			mut = append(mut[:pos.Int64()], append([]rune{'_'}, mut[pos.Int64():]...)...)
		case 4: // Leetspeak
			subs := map[rune]rune{'a': '4', 'e': '3', 'i': '1', 'o': '0', 's': '5', 't': '7'}
			for j, r := range mut {
				if s, ok := subs[r]; ok {
					c, _ := rand.Int(rand.Reader, big.NewInt(10))
					if c.Int64() < 5 {
						mut[j] = s
					}
				}
			}
		case 5: // Capitalize First Letter only
			if len(mut) > 0 {
				if mut[0] >= 'a' && mut[0] <= 'z' {
					mut[0] = mut[0] - 'a' + 'A'
				}
			}
		case 6: // Append a single digit
			n := rune(digitChars[getRandIdxString(digitChars)])
			mut = append(mut, n)
		case 7: // Randomly capitalize an internal letter
			if len(mut) > 1 {
				pos, _ := rand.Int(rand.Reader, big.NewInt(int64(len(mut)-1)))
				idx := pos.Int64() + 1
				if mut[idx] >= 'a' && mut[idx] <= 'z' {
					mut[idx] = mut[idx] - 'a' + 'A'
				}
			}
		}
	}

	// Ensure the result stays within length constraints
	res := string(mut)

	// Clean up: if we mutated too much, we might have added chars we shouldn't (though unlikely)
	// But mostly we need to handle length.
	if len(res) < minLen {
		extra, _ := generateWithCharset(minLen-len(res), minLen-len(res), commonChars)
		res += extra
	}
	if len(res) > maxLen {
		// Truncate from random side to keep it interesting
		side, _ := rand.Int(rand.Reader, big.NewInt(2))
		if side.Int64() == 0 {
			res = res[:maxLen]
		} else {
			res = res[len(res)-maxLen:]
		}
	}

	if res == seed && len(seed) > 0 {
		return Mutate(seed, minLen, maxLen)
	}

	return res, nil
}

func getRandIdxString(s string) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s))))
	return int(n.Int64())
}

// GetRandIdx returns a random index up to max
func GetRandIdx(max int64) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// SequentialIterator keeps track of the state for exhaustive brute force
type SequentialIterator struct {
	charset []rune
	indices []int
	minLen  int
	maxLen  int
	done    bool
}

func NewSequentialIterator(charset string, minLen, maxLen int) *SequentialIterator {
	return &SequentialIterator{
		charset: []rune(charset),
		indices: make([]int, minLen),
		minLen:  minLen,
		maxLen:  maxLen,
	}
}

func (it *SequentialIterator) Next() (string, bool) {
	if it.done {
		return "", false
	}

	// Build current string
	res := make([]rune, len(it.indices))
	for i, idx := range it.indices {
		res[i] = it.charset[idx]
	}

	// Increment indices
	for i := len(it.indices) - 1; i >= 0; i-- {
		it.indices[i]++
		if it.indices[i] < len(it.charset) {
			return string(res), true
		}
		it.indices[i] = 0
		if i == 0 {
			// Length increase
			if len(it.indices) < it.maxLen {
				it.indices = make([]int, len(it.indices)+1)
			} else {
				it.done = true
			}
		}
	}

	return string(res), true
}

// CalculateComplexity returns a score (lower is weaker/simpler).
// Weakest: all lowercase, short.
// Stronger: mixed case, numbers at end, special characters, longer.
func CalculateComplexity(p string) int {
	score := len(p)
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpec := false

	for i, r := range p {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
			if i == 0 {
				score += 1 // Common habit: starts with capital (slight bump)
			}
		} else if r >= 'a' && r <= 'z' {
			hasLower = true
		} else if r >= '0' && r <= '9' {
			hasDigit = true
			if i == len(p)-1 {
				score += 1 // Common habit: ends with number
			}
		} else {
			hasSpec = true
		}
	}

	if hasUpper && hasLower {
		score += 5
	}
	if hasDigit {
		score += 5
	}
	if hasSpec {
		score += 10
	}

	return score
}
