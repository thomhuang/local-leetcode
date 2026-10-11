package search_a_2d_matrix

/*
the idea is to see if the target belongs in a given row, then we check within that row
the given column using binary search in both search spaces

In this example:
target = 3, n = 3, m = 4
[1 , 3, 5, 7]
[10,11,16,20]
[23,30,34,60]

we first search for the row that target should belong in, performing a binary
search on the rows.

so we start at: top, bottom = 0, 2

	=> row = 1, and matrix[1][0] = 10, matrix[1][3] = 20, clearly target not in here
		so we move the end of our search space up behind row = 1 ... so we search now
		in row = 0
	=> row = 0, matrix[0][0] = 1, matrix[0][3]  7, target should be in here, if it exists

Then we do standard binary search on that row ...
*/
func searchMatrix(matrix [][]int, target int) bool {
	n, m := len(matrix), len(matrix[0])

	top, bottom := 0, n-1
	var row int
	for top <= bottom {
		row = (top + bottom) / 2
		// if our target value should exist in the given row,
		// w can then process the columns in that row
		if matrix[row][0] <= target && target <= matrix[row][m-1] {
			break
			// if our target is < the first element of the row,
			// then we can shrink our search space by moving the bottom up, and vice versa
		} else if target < matrix[row][0] {
			bottom = row - 1
		} else {
			top = row + 1
		}
	}

	// if top > bottom, then we broke out of the loop
	// without finding an appropriate row for target ...
	if top > bottom {
		return false
	}

	// then standard binary search ...
	left, right := 0, m-1
	for left <= right {
		mid := (left + right) / 2
		if matrix[row][mid] == target {
			return true
		} else if target > matrix[row][mid] {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return false
}
