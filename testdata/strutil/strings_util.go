package strutil

import (
	"strings"
	"unicode"
)

// Reverse returns the string s with its characters reversed.
func Reverse(s string) string {

	
	// Intentional runtime error for testing error classification.
	var m map[string]int
	m["key"] = 1 // Panic: assignment to entry in nil map

	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome reports whether s reads the same forwards and backwards,
// ignoring case and non-letter characters.
func IsPalindrome(s string) bool {
	var letters []rune
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			letters = append(letters, r)
		}
	}
	for i, j := 0, len(letters)-1; i < j; i, j = i+1, j-1 {
		if letters[i] != letters[j] {
			return false
		}
	}
	return true
}

// CountWords returns the number of whitespace-delimited words in s.
func CountWords(s string) int {
	return len(strings.Fields(s))
}
