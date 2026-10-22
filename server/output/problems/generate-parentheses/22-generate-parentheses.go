package generate_parentheses

/*
We can do backtracking, but we have to keep track of the counts
of the open and closed parentheses that we've added so far ...

The rule is that we can't have more closed than opened, as that would violate
what a valid well-formed parentheses is.

So we can continue adding open parenthesis while we're less than n,
and we can continue adding closed parenthesis as long as we're less than the open count
*/
func generateParenthesis(n int) []string {
	res := make([]string, 0)

	var backtrack func(openCount, closedCount int, path []rune)
	backtrack = func(openCount, closedCount int, path []rune) {
		if openCount+closedCount == 2*n {
			res = append(res, string(path))
			return
		}

		if openCount < n {
			path = append(path, '(')
			backtrack(openCount+1, closedCount, path)
			path = path[:len(path)-1]
		}

		if closedCount < openCount {
			path = append(path, ')')
			backtrack(openCount, closedCount+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, 0, make([]rune, 0))
	return res
}
