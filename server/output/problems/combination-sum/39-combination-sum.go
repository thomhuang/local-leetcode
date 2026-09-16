package combination_sum

func combinationSum(candidates []int, target int) [][]int {
	res := make([][]int, 0)

	var backtrack func(pos, difference int, path []int)
	backtrack = func(pos, difference int, path []int) {
		if difference < 0 {
			return
		}

		if difference == 0 {
			validPath := make([]int, len(path))
			copy(validPath, path)
			res = append(res, validPath)
		}

		for i := pos; i < len(candidates); i++ {
			path = append(path, candidates[i])
			backtrack(i, difference-candidates[i], path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, target, make([]int, 0))
	return res
}
