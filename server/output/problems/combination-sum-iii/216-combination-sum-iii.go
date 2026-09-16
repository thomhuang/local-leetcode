package combination_sum_iii

func combinationSum3(k int, n int) [][]int {
	res := make([][]int, 0)

	var backtrack func(num, difference int, path []int)
	backtrack = func(num, difference int, path []int) {
		if len(path) == k {
			if difference == 0 {
				validPath := make([]int, len(path))
				copy(validPath, path)
				res = append(res, validPath)
			}

			return
		}

		for i := num; i <= n; i++ {
			path = append(path, i)
			if difference-i < 0 {
				continue
			}

			backtrack(i+1, difference-i, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(1, n, make([]int, 0))

	return res
}
