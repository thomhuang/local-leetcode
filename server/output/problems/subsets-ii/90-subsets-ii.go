package subsets_ii

import "sort"

func subsetsWithDup(nums []int) [][]int {
	res := make([][]int, 0)

	sort.Ints(nums)

	var backtrack func(pos int, path []int)
	backtrack = func(pos int, path []int) {
		validPath := make([]int, len(path))
		copy(validPath, path)
		res = append(res, validPath)

		for i := pos; i < len(nums); i++ {
			// Ensure we process every valid backtracking state
			// but also skip duplicate permutations in the same tree level
			// e.g. [1,2,2]
			// pos = 1, path = [1]
			// => i = 1, backtrack(2, [1,2])
			// => i = 2, skip as duplicate in level
			// pos = 2, path = [1,2]
			// => i = 2, !(i > pos), so backtgrack(3, [1,2,2])
			if i > pos && nums[i-1] == nums[i] {
				continue
			}

			path = append(path, nums[i])
			backtrack(i+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]int, 0))
	return res
}
