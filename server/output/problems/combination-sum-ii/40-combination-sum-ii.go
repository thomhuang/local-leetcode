package combination_sum_ii

import "sort"

func combinationSum2(candidates []int, target int) [][]int {
	res := make([][]int, 0)

	sort.Ints(candidates)
	var backtrack func(pos, difference int, path []int)
	backtrack = func(pos, difference int, path []int) {
		if difference < 0 {
			return
		}

		if difference == 0 {
			validPath := make([]int, len(path))
			copy(validPath, path)
			res = append(res, validPath)
			return
		}

		for i := pos; i < len(candidates); i++ {
			/*
				Example: [1, 1, 2]

				In backtrack(0), both i = 0 and i = 1 would add 1
				to the empty path:

				i = 0 -> path [1]
				i = 1 -> path [1] again

				Those choices lead to the same combinations, so we skip
				the second 1 at this recursion level.

				`i > pos` means this is not the first candidate considered
				at this level. If it matches the previous candidate, skip it.

				This still allows [1, 1], because after choosing index 0,
				the next call is backtrack(1), where i == pos == 1.
			*/
			if i > pos && candidates[i] == candidates[i-1] {
				continue
			}

			path = append(path, candidates[i])
			backtrack(i+1, difference-candidates[i], path)
			path = path[:len(path)-1]
		}

	}

	backtrack(0, target, make([]int, 0))
	return res
}
