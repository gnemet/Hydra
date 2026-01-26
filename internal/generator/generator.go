package generator

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type RegexSegment struct {
	Charset  string
	Literal  string
	Variants []RegexSegment // For alternation (a|b|c)
	Min      int
	Max      int
}

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	specChars   = "!@#_-$"
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
// PatternedSequentialIterator keeps track of the state for exhaustive brute force with specific charsets per position
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

// GenerateRandomPatterned generates a single random string matching the position-based charsets
func GenerateRandomPatterned(segments []RegexSegment) string {
	res := ""
	for _, s := range segments {
		res += s.GenerateRandom()
	}
	return res
}

// GenerateRandomPatternedWithSeeds injects a random seed into the generation
func GenerateRandomPatternedWithSeeds(segments []RegexSegment, seeds []string) string {
	if len(seeds) == 0 {
		return GenerateRandomPatterned(segments)
	}

	idx, _ := GetRandIdx(int64(len(seeds)))
	seed := seeds[idx]
	mutatedSeed := []rune(RandomizeCase(seed))

	res := ""
	for _, s := range segments {
		if s.IsWordLike() && len(mutatedSeed) > 0 {
			take := s.Min
			if s.Max > take || s.Max < 0 {
				// If this is a multi-char segment, let it consume the rest of the seed or at least a good chunk
				if s.Max > 2 || s.Max < 0 {
					take = len(mutatedSeed)
				} else if s.Max > take {
					take = s.Max
				}
			}
			if take > len(mutatedSeed) {
				take = len(mutatedSeed)
			}
			if take < s.Min {
				take = s.Min // respect Min if possible, will be filled by randoms below if seed is short
			}

			if take > 0 && len(mutatedSeed) >= take {
				res += string(mutatedSeed[:take])
				mutatedSeed = mutatedSeed[take:]
			} else if len(mutatedSeed) > 0 {
				res += string(mutatedSeed)
				mutatedSeed = nil
			}

			// If this segment has a count > what we put in, or if we want to add a separator
			// from its charset after the seed word.
			remMin := s.Min - take
			if remMin < 0 {
				remMin = 0
			}
			remMax := s.Max - take
			if remMax < 0 {
				remMax = 0
			}

			if remMax > 0 || s.Max < 0 {
				// Chance to add a separator if the charset allows it
				if strings.ContainsAny(s.Charset, "_#-$!@") {
					prob, _ := GetRandIdx(10)
					if prob < 4 { // 40% chance for a separator if segment allows
						sepIdx, _ := GetRandIdx(int64(len(specChars)))
						sep := specChars[sepIdx]
						if strings.Contains(s.Charset, string(sep)) {
							res += string(sep)
						}
					}
				}
				// If we still haven't met the Min length, fill with randoms
				if remMin > 0 {
					extra, _ := generateWithCharset(remMin, remMin, s.Charset)
					res += extra
				}
			}
		} else {
			res += s.GenerateRandom()
		}
	}

	// If seed wasn't fully consumed (rare with our regex), just prepend the rest
	if len(mutatedSeed) > 0 {
		res = string(mutatedSeed) + res
	}

	return res
}

func RandomizeCase(s string) string {
	n, _ := GetRandIdx(10)
	if n < 3 { // 30% All Lower
		return strings.ToLower(s)
	}
	if n < 6 { // 30% Title Case
		if len(s) > 0 {
			return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		}
		return s
	}
	if n < 7 { // 10% All Upper
		return strings.ToUpper(s)
	}

	// 30% Random case (original behavior)
	runes := []rune(s)
	for i, r := range runes {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			rn, _ := GetRandIdx(2)
			if rn == 0 {
				runes[i] = []rune(strings.ToUpper(string(r)))[0]
			} else {
				runes[i] = []rune(strings.ToLower(string(r)))[0]
			}
		}
	}
	return string(runes)
}

func (s RegexSegment) GenerateRandom() string {
	if len(s.Variants) > 0 {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.Variants))))
		return s.Variants[n.Int64()].GenerateRandom()
	}
	if s.Literal != "" {
		return s.Literal
	}
	if s.Charset != "" {
		res, _ := generateWithCharset(s.Min, s.Max, s.Charset)
		return res
	}
	return ""
}

// RunSegmentedBrute executes exhaustive search for a given list of segments
func RunSegmentedBrute(segments []RegexSegment, count int, prefix string, callback func(string)) {
	generatedCount := 0

	// We need to handle segment length variations (e.g. [a-z]{1,2})
	// For each segment, we find all possible outputs
	var allPosOutputs [][]string
	for _, s := range segments {
		allPosOutputs = append(allPosOutputs, s.Expand())
	}

	indices := make([]int, len(allPosOutputs))
	for {
		// Build string
		res := prefix
		for i, idx := range indices {
			res += allPosOutputs[i][idx]
		}
		callback(res)
		generatedCount++
		if count > 0 && generatedCount >= count {
			return
		}

		// Increment
		for i := len(indices) - 1; i >= 0; i-- {
			indices[i]++
			if indices[i] < len(allPosOutputs[i]) {
				goto next
			}
			indices[i] = 0
			if i == 0 {
				return
			}
		}
	next:
	}
}

func (s RegexSegment) Expand() []string {
	if len(s.Variants) > 0 {
		var res []string
		for _, v := range s.Variants {
			res = append(res, v.Expand()...)
		}
		return res
	}
	if s.Literal != "" {
		return []string{s.Literal}
	}
	if s.Charset != "" {
		var res []string
		// Brute force expansion of charset combinations for small lengths
		// If length is too large, this might be slow, but for NAS it's usually small segments
		for l := s.Min; l <= s.Max; l++ {
			it := NewSequentialIterator(s.Charset, l, l)
			for {
				p, ok := it.Next()
				if !ok {
					break
				}
				res = append(res, p)
			}
		}
		return res
	}
	return []string{""}
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

func (s RegexSegment) IsWordLike() bool {
	if s.Literal != "" {
		return true
	}
	if s.Charset == "" {
		return false
	}
	letters := 0
	for _, r := range s.Charset {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			letters++
		}
	}
	// If more than 50% letters or specifically has alpha ranges
	return letters > (len(s.Charset) / 2)
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

// ParseLengthsFromRegex extracts total {min,max} or [min-max] from a regex
func ParseLengthsFromRegex(regex string) (int, int) {
	// First check if there is a global length constraint at the end: [6-10] or {6,10}
	globalRe := regexp.MustCompile(\`[\[\{](\d+)[-,\s](\d+)[\]\}]$\`)
	if match := globalRe.FindStringSubmatch(regex); match != nil {
		min, _ := strconv.Atoi(match[1])
		max, _ := strconv.Atoi(match[2])
		return min, max
	}

	segments := ParseSegmentedRegex(regex)
	if len(segments) == 0 {
		return 6, 12 // Fallback to original default
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
	// Strip global length constraint if present
	regex = CleanRegex(regex)

	// Strip outer wrapping parens if they exist for the whole thing
	if strings.HasPrefix(regex, "(") && strings.HasSuffix(regex, ")") {
		// Only strip if they don't have a quantifier after them (handled by CleanRegex already)
		regex = regex[1 : len(regex)-1]
	}

	var segments []RegexSegment
	i := 0
	for i < len(regex) {
		if regex[i] == '[' {
			j := i
			for j < len(regex) && regex[j] != ']' {
				j++
			}
			if j < len(regex) {
				segment := RegexSegment{Charset: ParseCharsetFromRegex(regex[i : j+1]), Min: 1, Max: 1}
				i = j + 1
				// Check for quantifier
				i = parseQuantifier(regex, i, &segment)
				segments = append(segments, segment)
			} else {
				i++
			}
		} else if regex[i] == '(' {
			j := findClosingParen(regex, i)
			if j != -1 {
				content := regex[i+1 : j]
				segment := parseGroup(content)
				i = j + 1
				i = parseQuantifier(regex, i, &segment)
				segments = append(segments, segment)
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return segments
}

func parseGroup(content string) RegexSegment {
	if strings.Contains(content, "|") {
		parts := strings.Split(content, "|")
		var variants []RegexSegment
		for _, p := range parts {
			// Each part can be a literal or a sub-segment
			if strings.HasPrefix(p, "[") {
				sub := ParseSegmentedRegex(p)
				if len(sub) > 0 {
					variants = append(variants, sub[0])
				}
			} else {
				variants = append(variants, RegexSegment{Literal: p})
			}
		}
		return RegexSegment{Variants: variants}
	}
	// Not an alternation, just a group
	sub := ParseSegmentedRegex(content)
	if len(sub) == 1 {
		return sub[0]
	}
	// Multiple segments in a group is tricky, let's just return a placeholder or handle it
	return RegexSegment{Literal: content}
}

func parseQuantifier(regex string, i int, s *RegexSegment) int {
	if i >= len(regex) {
		return i
	}
	if regex[i] == '+' {
		s.Min, s.Max = 1, 8
		return i + 1
	}
	if regex[i] == '*' {
		s.Min, s.Max = 0, 8
		return i + 1
	}
	if regex[i] == '{' {
		j := i
		for j < len(regex) && regex[j] != '}' {
			j++
		}
		if j < len(regex) {
			parts := strings.Split(regex[i+1:j], ",")
			if len(parts) == 2 {
				if parts[0] == "" {
					s.Min = 0
				} else {
					s.Min, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
				}
				if parts[1] != "" {
					s.Max, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				} else {
					s.Max = 12
				}
			} else {
				s.Min, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
				s.Max = s.Min
			}
			return j + 1
		}
	}
	return i
}

func findClosingParen(s string, start int) int {
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// CleanRegex removes our custom global length constraints from a regex so it can be used with the standard regexp package
func CleanRegex(regex string) string {
	globalRe := regexp.MustCompile(\`[\[\{]\d+[-,\s]\d+[\]\}]$\`)
	return globalRe.ReplaceAllString(regex, "")
}
