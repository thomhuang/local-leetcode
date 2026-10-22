package combination_sum

/*
Backtrack, and since we allow repeat elements, we'll
go ahead and go through all elements of `nums` and add that
to our path

if our path sum is > target, or if we start at target and decrement, then < 0,
then we can break off that "branch"

if we ever reach pathSum == target, add the path to our result
*/
func combinationSum(candidates []int, target int) [][]int {
	res := make([][]int, 0)

	var backtrack func(path []int, curr, remaining int)
	backtrack = func(path []int, curr, remaining int) {
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
			path = append(path, candidates[i])
			backtrack(path, i, remaining-candidates[i])
			path = path[:len(path)-1]
		}
	}

	backtrack(make([]int, 0), 0, target)
	return res
}
