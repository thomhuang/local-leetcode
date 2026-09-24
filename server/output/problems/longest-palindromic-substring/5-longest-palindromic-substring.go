package longest_palindromic_substring

func longestPalindrome(s string) string {
	if len(s) < 2 {
		return s
	}

	longest := s[:1]

	for i := range s {
		odd := getPalindrome(s, i, i)
		if len(odd) > len(longest) {
			longest = odd
		}

		even := getPalindrome(s, i, i+1)
		if len(even) > len(longest) {
			longest = even
		}
	}

	return longest
}

func getPalindrome(s string, left, right int) string {
	for left >= 0 && right < len(s) && s[left] == s[right] {
		left--
		right++
	}

	return s[left+1 : right]
}
