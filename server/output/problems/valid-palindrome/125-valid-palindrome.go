package valid_palindrome

import "unicode"

func isPalindrome(s string) bool {
	l, r := 0, len(s)-1

	// start from ends of string, and move inward
	for l < r {
		// so long as l < r and s[l] **isn't** alphanumeric, move forward
		// same idea for r
		for l < r && !isAlphanumeric(s[l]) {
			l++
		}

		for l < r && !isAlphanumeric(s[r]) {
			r--
		}

		// then check lowercase equality ...
		if toLower(s[l]) != toLower(s[r]) {
			return false
		}

		// shrink our search space if valid
		l++
		r--
	}

	return true
}

func isAlphanumeric(b byte) bool {
	r := rune(b)

	return unicode.IsDigit(r) || unicode.IsLetter(r)
}

func toLower(b byte) rune {
	return unicode.ToLower(rune(b))
}
