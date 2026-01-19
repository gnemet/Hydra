package generator

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type RegexSegment struct {
	Charset string
	Min     int
	Max     int
}

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
		strategy, _ := rand.Int(rand.Reader, big.NewInt(11))
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
		case 8: // Double Suffix (Digit + Special Char)
			digit := digitChars[getRandIdxString(digitChars)]
			special := specChars[getRandIdxString(specChars)]
			mut = append(mut, rune(digit), rune(special))
		case 9: // Prepend/Append years for legacy devices (1990-2010)
			years := []string{"1998", "1999", "2000", "2001", "2002", "2003", "2004", "2005", "2006", "2007", "2008", "2009", "2010"}
			yIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(years))))
			mut = append(mut, []rune(years[yIdx.Int64()])...)
		case 10: // Add padding characters
			pads := []string{"---", "___", "...", "!!!"}
			pIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pads))))
			mut = append([]rune(pads[pIdx.Int64()]), mut...)
			mut = append(mut, []rune(pads[pIdx.Int64()])...)
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

// PatternedSequentialIterator supports different charsets for different positions
type PatternedSequentialIterator struct {
	charsets [][]rune
	indices  []int
	done     bool
}

func NewPatternedSequentialIterator(posCharsets []string) *PatternedSequentialIterator {
	charsets := make([][]rune, len(posCharsets))
	for i, s := range posCharsets {
		charsets[i] = []rune(s)
	}
	return &PatternedSequentialIterator{
		charsets: charsets,
		indices:  make([]int, len(posCharsets)),
	}
}

func (it *PatternedSequentialIterator) Next() (string, bool) {
	if it.done {
		return "", false
	}

	// Build current string
	res := make([]rune, len(it.indices))
	for i, idx := range it.indices {
		res[i] = it.charsets[i][idx]
	}

	// Increment indices
	for i := len(it.indices) - 1; i >= 0; i-- {
		it.indices[i]++
		if it.indices[i] < len(it.charsets[i]) {
			return string(res), true
		}
		it.indices[i] = 0
		if i == 0 {
			it.done = true
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

// ParseCharsetFromRegex extracts a flat charset from a simple [a-zA-Z] style regex
func ParseCharsetFromRegex(regex string) string {
	if regex == "" {
		return commonChars
	}

	// Simple heuristic: find content between brackets
	start := 0
	for i, r := range regex {
		if r == '[' {
			start = i + 1
			break
		}
	}
	end := len(regex)
	for i := len(regex) - 1; i >= 0; i-- {
		if regex[i] == ']' {
			end = i
			break
		}
	}

	if start >= end {
		return commonChars
	}

	content := regex[start:end]
	res := ""
	for i := 0; i < len(content); i++ {
		if i+2 < len(content) && content[i+1] == '-' {
			// Range detected: a-z, A-Z, 0-9
			for c := content[i]; c <= content[i+2]; c++ {
				res += string(c)
			}
			i += 2
		} else {
			res += string(content[i])
		}
	}

	// Deduplicate
	seenChar := make(map[rune]bool)
	final := ""
	for _, r := range res {
		if !seenChar[r] {
			final += string(r)
			seenChar[r] = true
		}
	}

	return final
}

// ParseLengthsFromRegex extracts total {min,max} or {n} from a regex across all segments
func ParseLengthsFromRegex(regex string) (int, int) {
	segments := ParseSegmentedRegex(regex)
	if len(segments) == 0 {
		return 6, 10
	}

	totalMin, totalMax := 0, 0
	for _, s := range segments {
		totalMin += s.Min
		totalMax += s.Max
	}
	return totalMin, totalMax
}

// ParseSegmentedRegex splits a complex regex into individual segments
func ParseSegmentedRegex(regex string) []RegexSegment {
	// Matches like [a-z]{1,2} or [A-Z]
	re := regexp.MustCompile(`\[([^\]]+)\](?:\{([^\}]*)\})?`)
	matches := re.FindAllStringSubmatch(regex, -1)

	var segments []RegexSegment
	for _, m := range matches {
		charsetPart := "[" + m[1] + "]"
		charset := ParseCharsetFromRegex(charsetPart)

		min, max := 1, 1 // Default
		if m[2] != "" {
			parts := strings.Split(m[2], ",")
			if len(parts) == 2 {
				if parts[0] == "" {
					min = 0
				} else {
					min, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
				}

				if len(parts) > 1 && parts[1] != "" {
					max, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				} else {
					max = min
				}
			} else {
				min, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
				max = min
			}
		}
		segments = append(segments, RegexSegment{Charset: charset, Min: min, Max: max})
	}
	// If no segments found, return a default
	if len(segments) == 0 {
		// Try to fallback to the old simple search if it doesn't match [..] pattern
		// but for now we expect [..]
	}
	return segments
}
