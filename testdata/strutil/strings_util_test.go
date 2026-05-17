package strutil

import (
	"testing"
)

func TestReverse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"normal", "hello", "olleh"},
		{"empty", "", ""},
		{"single", "a", "a"},
		{"unicode", "测试", "试测"},
		{"palindrome", "aibohphobia", "aibohphobia"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Reverse(tc.in)
			if got != tc.want {
				t.Errorf("Reverse(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsPalindrome(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"palindrome", "A man a plan a canal Panama", true},
		{"not palindrome", "hello", false},
		{"empty", "", true},
		{"single char", "a", true},
		{"with punctuation", "Able was I, I saw Elba", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPalindrome(tc.in)
			if got != tc.want {
				t.Errorf("IsPalindrome(%q) = %v; want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"normal", "Hello world", 2},
		{"multiple spaces", "Hello    world", 2},
		{"empty", "", 0},
		{"single word", "Hello", 1},
		{"with punctuation", "Hello, world! How are you?", 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CountWords(tc.in)
			if got != tc.want {
				t.Errorf("CountWords(%q) = %d; want %d", tc.in, got, tc.want)
			}
		})
	}
}