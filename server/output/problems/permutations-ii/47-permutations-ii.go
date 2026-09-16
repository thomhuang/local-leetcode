package permutations

import "sort"

func permuteUnique(nums []int) [][]int {
	sort.Ints(nums)

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

			// for equal values, ensure we pick them in order
			// so we don't create identical permutations!
			// if we don't have the gate,
			// e.g. [1, 1, 2]
			// Pick first 1, then second 1  -> [1, 1, 2]
			// Pick second 1, then first 1  -> [1, 1, 2]  // same answer
			if i > 0 && nums[i] == nums[i-1] && !used[i-1] {
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
