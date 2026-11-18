package letter_combinations_of_a_phone_number

func letterCombinations(digits string) []string {
	// create lookup between digits and the possible letters it can produce
	lookup := make(map[byte][]rune)
	lookup['2'] = []rune{'a', 'b', 'c'}
	lookup['3'] = []rune{'d', 'e', 'f'}
	lookup['4'] = []rune{'g', 'h', 'i'}
	lookup['5'] = []rune{'j', 'k', 'l'}
	lookup['6'] = []rune{'m', 'n', 'o'}
	lookup['7'] = []rune{'p', 'q', 'r', 's'}
	lookup['8'] = []rune{'t', 'u', 'v'}
	lookup['9'] = []rune{'w', 'x', 'y', 'z'}

	res := make([]string, 0)

	// we will bactrack on each digit, and at each new iteration
	// we will pass the index of the next character we're processing
	// we will add a character to the path, continue backtracking on it, then pop it off
	var backtrack func(pos int, path []rune)
	backtrack = func(pos int, path []rune) {
		if pos == len(digits) {
			res = append(res, string(path))
			return
		}

		for _, c := range lookup[digits[pos]] {
			backtrack(pos+1, append(path, c))
		}
	}

	backtrack(0, make([]rune, 0))
	return res
}
