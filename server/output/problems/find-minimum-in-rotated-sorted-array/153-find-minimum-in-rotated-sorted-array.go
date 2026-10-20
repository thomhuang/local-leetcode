package find_minimum_in_rotated_sorted_array

/*
	In any search window we have, the minimum will always be < the end of the search window ...
	so we continue shrinking the search space until l == r essentially.
*/
func findMin(nums []int) int {
	l, r := 0, len(nums)-1

	for l < r {
		m := (l + r) / 2
		// if our midpoint is < than the end, then there are "smaller"
		// elements between [m,r], so move r back
		if nums[m] < nums[r] {
			r = m
		} else {
			// otherwise, move l up ...
			l = m + 1
		}
	}

	return nums[r]
}
