package longest_palindromic_substring

import "testing"

func TestLongestPalindrome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "babad", input: "babad", want: "bab"},
		{name: "empty", input: "", want: ""},
		{name: "single character", input: "a", want: "a"},
		{name: "even length", input: "cbbd", want: "bb"},
		{name: "whole string", input: "racecar", want: "racecar"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := longestPalindrome(test.input)
			if got != test.want && !(test.input == "babad" && got == "aba") {
				t.Fatalf("longestPalindrome(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
