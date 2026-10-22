package word_search

/*
	Backtracking here, but on startup, employ it only when we run into a (r,c) in the board
	that matches the first letter of our 'word'

	At it's simplest, if we've traversed the len(word) in our backtracking, we've found our word
	as we only continue on if the next traversed (r,c) matches the next letter in the word. Continue
	this until we find the entire word!

	One thing we have to be careful about is to not revisit the same path, so that's
	where the backtracking comes in. If we've visited a given (r,c) in the current
	path we're on, then we skip, otherwise we set it to visited, continue backtracking
	on that particular path, then unvisit it so future backtracking doesn't falsely
	skip
*/
func exist(board [][]byte, word string) bool {
	k := len(word)
	n, m := len(board), len(board[0])
	used := make(map[[2]int]bool)
	var backtrack func(pos, r, c int) bool
	backtrack = func(pos, r, c int) bool {
		if pos == k {
			return true
		}

		if r < 0 || c < 0 || r >= n || c >= m {
			return false
		}

		key := [2]int{r, c}
		if used[key] {
			return false
		}

		if board[r][c] != word[pos] {
			return false
		}

		used[key] = true
		res := backtrack(pos+1, r+1, c) || backtrack(pos+1, r-1, c) ||
			backtrack(pos+1, r, c+1) || backtrack(pos+1, r, c-1)
		used[key] = false
		return res
	}

	for r := range n {
		for c := range m {
			if board[r][c] == word[0] {
				found := backtrack(0, r, c)
				if found {
					return true
				}
			}
		}
	}

	return false
}
