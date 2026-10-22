package combination_sum_ii

import "sort"

/*
Same idea as trhe first combination sum, however we need
to be careful as to have **unique** combinations of values ...

in the case of repeated elements like the example they provide:

	 candidates = [10,1,2,7,6,1,5], target = 8

	i = 1, 5 are duplicate elements, and if we perform the same implementation
	as the initial combination sum, we will 100% have duplicate elements.

So, in order to avoid this and ENSURE that duplicate elements are close-by, i.e.
next to each other, we can sort our input array.

In the case of some input like [1,1,1,2], target = 3, we want to make sure we only
have one instance of [1, 2] within our result
So what we can do, is that in our current iteration, we allow the processing of the
value itself, but for any subsequent duplicates, we want to ignore ...

so in a working example:

	let's say we're at curr = 1, remaining = 2, and path = [1]
		- At i = curr = 1, this is a valid addition, so we go ahead and add it to our path = [1,1]
			and continue
		- at i = 2, this is NOT a valid addition as this would make an identical permutation as [1,1] and we want to avoid
			duplicates ... SKIP!
		and continue!
*/
func combinationSum2(candidates []int, target int) [][]int {
	res := make([][]int, 0)

	sort.Ints(candidates)

	var backtrack func(curr, remaining int, path []int)
	backtrack = func(curr, remaining int, path []int) {
		if remaining == 0 {
			validPath := make([]int, len(path))
			copy(validPath, path)
			res = append(res, validPath)
			return
		}

		if remaining < 0 {
			return
		}

		for i := curr; i < len(candidates); i++ {
			if i > curr && candidates[i] == candidates[i-1] {
				continue
			}

			path = append(path, candidates[i])
			backtrack(i+1, remaining-candidates[i], path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, target, make([]int, 0))
	return res
}
