package generator

import (
	"crypto/rand"
	"math/big"
)

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	specChars   = "_"
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
