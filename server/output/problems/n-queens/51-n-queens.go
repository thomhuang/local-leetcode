package n_queens

/*
We can do standard backtracking here, passing along our backtrack dfs the number of queens
placed so far, as well as the board state for that given path.

We'll have to make a validate method to see if a placed queen is in threat of another existing queen, etc.
*/
func solveNQueens(n int) [][]string {
	res := make([][]string, 0)

	// we initialize our board with no queens placed at all, i.e. a '.'
	board := make([][]rune, n)
	for r := range n { //O(n^2)
		board[r] = make([]rune, n)
		for c := range n {
			board[r][c] = '.'
		}
	}

	// Note that for a given queen in a row, we cannot have a queen within the same row
	// so what we can do is move down each row, and place a queen and check the PREVIOUS
	// rows to determine if it's a valid position to place the queen.
	var backtrack func(row int, path [][]rune) // O(n)
	backtrack = func(row int, path [][]rune) {
		if row == n { // if we've gone past the rows in the board, then we know we've placed n queens
			validBoard := make([]string, n)
			for i, row := range path { // O(n)
				validBoard[i] = string(row)
			}

			res = append(res, validBoard)
			return
		}

		// We know which row we're on from the passed parameter,
		// so we'll have to check if each column is valid
		for c := range n { // O(n)
			if isValid(row, c, path) { // O(n - row)
				// add the queen if valid
				path[row][c] = 'Q'
				// continue to next row
				backtrack(row+1, path)
				// revert path change for next column attempt
				path[row][c] = '.'
			}
		}
	}

	backtrack(0, board)

	return res
}

// we check any column, and diagonals above the (r,c) position ...
func isValid(row, col int, path [][]rune) bool { // O(n) worse case
	for r := 0; r < row; r++ {
		if path[r][col] == 'Q' {
			return false
		}
	}

	// upper left diagonal
	for r, c := row, col; r >= 0 && c >= 0; r, c = r-1, c-1 {
		if path[r][c] == 'Q' {
			return false
		}
	}

	// upper right diagonal
	for r, c := row, col; r >= 0 && c >= 0 && c < len(path); r, c = r-1, c+1 {
		if path[r][c] == 'Q' {
			return false
		}
	}

	return true
}
