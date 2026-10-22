package palindrome_partitioning

/*
	the backtracking here, is that in a given iteration, we will go from every
	index from the next character from the past iteration and loop to the end of the
	of 's' and if any of them are a palindrome, we will add it to our current path result,
	and then pop it out after we're done for the backtracking!
*/
func partition(s string) [][]string {
	res := make([][]string, 0)

	var backtrack func(start int, path []string)
	backtrack = func(start int, path []string) {
		if start == len(s) {
			validPath := make([]string, len(path))
			copy(validPath, path)
			res = append(res, validPath)
			return
		}

		for end := start; end < len(s); end++ {
			segment := s[start : end+1]
			if !isPalindrome(segment) {
				continue
			}

			path = append(path, segment)
			backtrack(end+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]string, 0))
	return res
}

func isPalindrome(s string) bool {
	l, r := 0, len(s)-1

	for l < r {
		if s[l] != s[r] {
			return false
		}
		l++
		r--
	}

	return true
}
