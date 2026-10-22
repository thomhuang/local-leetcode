package permutations

/*
	This is backtracking, but we have to fill out len(nums) for each valid path...

	Notice that from the constraints that all the input integers within `nums` are unique
	So, what we can do, is to iterate 0 -> n - 1, but in our given path
	we need to keep track, at the given recurision level, what elements
	we've used so far to not repeat in nums...

	If, in our path, we have len(nums) elements, we've generated a permutation ...

*/
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
