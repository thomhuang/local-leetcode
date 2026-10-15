package search_in_rotated_sorted_array

/*
This seems to be a binary search, but we need to be careful about how/when we move
our left/right pointers, as there are branches of decisions depending on
the rotation ...

If it's non-roatated, we can just do a normal binary search ...
*/
func search(nums []int, target int) int {
	l, r := 0, len(nums)-1

	for l <= r {
		m := (l + r) / 2
		if nums[m] == target {
			return m
		}

		// if our [l, m] is ascending, we have two cases ...
		if nums[l] <= nums[m] {
			// if target is within [l, m] then move r back
			if nums[l] <= target && target < nums[m] {
				r = m - 1
				// otherwise, move l up ...
			} else {
				// otherwise it's rotated
				l = m + 1
			}
			// otherwise, [l, m] contains some rotated segment, so [m, r] is ascending
		} else {
			// if target \in [m, r] then move l up
			if nums[m] < target && target <= nums[r] {
				l = m + 1
				// and vice versa
			} else {
				r = m - 1
			}
		}
	}

	return -1
}
