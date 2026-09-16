package n_queens

/*
We place one queen per row with backtracking. For each row we try every column.
isValid checks the column above and both upper diagonals for an existing queen.
When a placement is valid, we place the queen, recurse to the next row, then remove it.
When row == n, every row has a queen, so we copy the board into the result.
*/
func solveNQueens(n int) [][]string {
	res := make([][]string, 0)
	if n == 0 {
		return res
	}
	board := make([][]byte, n)
	for i := range n {
		row := make([]byte, n)
		for j := range n {
			row[j] = '.'
		}

		board[i] = row
	}

	var backtrack func(row int)
	backtrack = func(row int) {
		if row == n {
			validBoard := make([]string, n)
			for r := range board {
				validBoard[r] = string(board[r])
			}

			res = append(res, validBoard)
			return
		}

		for c := 0; c < len(board[0]); c++ {
			if !isValid(row, c, board) {
				continue
			}

			board[row][c] = 'Q'
			backtrack(row + 1)
			board[row][c] = '.'
		}
	}

	backtrack(0)

	return res
}

func isValid(row, col int, board [][]byte) bool {
	// first we need to check every other row with the given column to see if
	// there's a conflicting queen

	for r := range row {
		if board[r][col] == 'Q' {
			return false
		}
	}

	// now we need to check each upper diagonal, left/right from the position

	// this first checks the upper right diagonal
	for i, j := row, col; i >= 0 && j < len(board); i, j = i-1, j+1 {
		if board[i][j] == 'Q' {
			return false
		}
	}

	// now this checks the upper left diagonal
	for i, j := row, col; i >= 0 && j >= 0; i, j = i-1, j-1 {
		if board[i][j] == 'Q' {
			return false
		}
	}

	return true
}
