package subsets

/*
For a given input, the intuition is prety straight forward ...
at each element in nums, moving strictly forward, we have the option of
either keeping, or skipping the current element in nums

at each step we have the option of including/excluding the element
so that gives us a runtime of O(2^n)
*/
func subsets(nums []int) [][]int {
	res := make([][]int, 0)

	var backtrack func(curr int, path []int)
	backtrack = func(curr int, path []int) {
		validPath := make([]int, len(path))
		copy(validPath, path)
		res = append(res, validPath)

		for i := curr; i < len(nums); i++ {
			path = append(path, nums[i])
			backtrack(i+1, path)
			path = path[:len(path)-1]
		}
	}

	backtrack(0, make([]int, 0))
	return res
}
