package matrix01

/*
I think the idea is to perform a BFS on this guy, and on each "step" level, we increment
to determine the number of "steps" taken from a given "1" to a "0" if found

so we can initialize our queue with all the position of the '0's and initialize
the distance for any 0 to be the largest step distance it could be, e.g. from corner to corner
*/
func updateMatrix(mat [][]int) [][]int {
	if mat == nil || len(mat) == 0 || len(mat[0]) == 0 {
		return [][]int{}
	}

	m, n := len(mat), len(mat[0])
	queue := make([][]int, 0)
	MAX_VALUE := m * n

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if mat[i][j] == 0 { // NOTE: 0s are 0 distance away from itself!
				queue = append(queue, []int{i, j})
			} else {
				mat[i][j] = MAX_VALUE
			}
		}
	}

	directions := [][]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	// Queue is initialized with all positions of '0's
	for len(queue) > 0 {
		cell := queue[0]
		queue = queue[1:]
		for _, dir := range directions {
			r, c := cell[0]+dir[0], cell[1]+dir[1]
			// If our bounds are valid and the neighboring node's current distance state to a 0 (default m*n) is
			// is greater than a neighboring cell's distance to a 0, plus one, add it to the queue and continue
			if r >= 0 && r < m && c >= 0 && c < n && mat[r][c] > mat[cell[0]][cell[1]]+1 {
				queue = append(queue, []int{r, c})
				mat[r][c] = mat[cell[0]][cell[1]] + 1
			}
		}
	}

	return mat
}
