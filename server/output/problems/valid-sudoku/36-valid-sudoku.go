package valid_sudoku

func isValidSudoku(board [][]byte) bool {
	rowMap, colMap, gridMap := make(map[int]map[int]bool), make(map[int]map[int]bool), make(map[int]map[int]bool)

	for i := range 9 {
		rowMap[i] = make(map[int]bool)
		colMap[i] = make(map[int]bool)
		gridMap[i] = make(map[int]bool)
	}

	for r := range 9 {
		for c := range 9 {
			if board[r][c] == '.' {
				continue
			}

			curr := int(board[r][c] - '0')
			if rowMap[r][curr] == true {
				return false
			} else {
				rowMap[r][curr] = true
			}

			if colMap[c][curr] == true {
				return false
			} else {
				colMap[c][curr] = true
			}

			grid := (r/3)*3 + c/3
			if gridMap[grid][curr] == true {
				return false
			} else {
				gridMap[grid][curr] = true
			}
		}
	}

	return true
}
