package places

import (
	"strings"
	"unicode"
)

// SanitizeString to unicode letters, spaces and minus
func SanitizeString(s string) string {

	// only unicode letters, spaces and minus
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return -1
	}, s)

	// remove spaces and minus from head and tail and lower case
	return strings.ToLower(strings.Trim(s, " -"))
}

// Min returns the minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
