package permutations

func permute(nums []int) [][]int {
	res := make([][]int, 0)

	used := make([]bool, len(nums))
	var backtrack func(pos int, path []int)
	backtrack = func(pos int, path []int) {
		if len(path) == len(nums) {
			validPath := make([]int, len(path))
			copy(validPath, path)
			res = append(res, validPath)
			return
		}

		for i := 0; i < len(nums); i++ {
			if used[i] {
				continue
			}

			used[i] = true
			backtrack(pos+1, append(path, nums[i]))
			used[i] = false
		}
	}

	backtrack(0, make([]int, 0))

	return res
}
