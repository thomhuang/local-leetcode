package find_minimum_in_rotated_sorted_array

/*
We want to find minimum element in a rotated array ...
notice that if we have some rotated array,

	e.g. [4, 1, 2, 3]
	the minimum array will always be s.t. it's < than the element
	that is at the end of its window that includes it ...

	We can continue shrinking this search space until l == r, then
	we can just return our l which will be the index
*/
func findMin(nums []int) int {
	l, r := 0, len(nums)-1

	// standard binary search procedure
	for l <= r {
		m := (l + r) / 2
		// if the middle of the search space is less than the
		// the end of the search space, then we move our end
		// boundary "back", but inclusive of it, cause
		// that may be the actual result!
		if nums[m] < nums[r] {
			r = m
			// otherwise, we know that the middle is > than the
			// the boundary, e.g. [6,5,4]
			// then the left search space is descending so move it up
		} else {
			l = m + 1
		}
	}

	return nums[r]
}
