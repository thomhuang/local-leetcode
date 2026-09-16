package subsets

func subsets(nums []int) [][]int {
	res := make([][]int, 0)

	var backtrack func(pos int, path []int)
	backtrack = func(pos int, path []int) {
		validPath := make([]int, len(path))
		copy(validPath, path)
		res = append(res, validPath)

		for i := pos; i < len(nums); i++ {
			path = append(path, nums[i])
			backtrack(i+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]int, 0))

	return res
}
