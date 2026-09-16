package combinations

/*
This here is a very obvious backtracking solution,
but the condition that we need to be wary about amongst our branches is:

Given n = 2 for example,

[1,2] and [2,1] are considered the same combination

I think what makes sense is to see if backtrack, iterating our position, but if candidates[i] > path[len(path)-1], then we can continue
so we can enforce ordering and don't reproduce the same combinations
*/
func combine(n int, k int) [][]int {
	res := make([][]int, 0)

	var backtrack func(num int, path []int)
	backtrack = func(num int, path []int) {
		if len(path) == k {
			validPath := make([]int, k)
			copy(validPath, path)
			res = append(res, validPath)
			return
		}

		for i := num + 1; i <= n; i++ {
			if len(path) > 0 && i <= path[len(path)-1] {
				continue
			}

			path = append(path, i)
			backtrack(i, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]int, 0))
	return res
}
