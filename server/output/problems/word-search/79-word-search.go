package word_search

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
