package container_with_most_water

func maxArea(height []int) int {
	// two window approach. ..
	l, r := 0, len(height)-1

	res := 0
	for l < r {
		lHeight, rHeight := height[l], height[r]

		// area is determined by lowest height between the two ends
		maxArea := min(lHeight, rHeight) * (r - l)
		res = max(res, maxArea)

		// we shrink the space of the lower height as
		// we will always get a bigger potential area by keeping the largest height!
		if lHeight > rHeight {
			r--
		} else {
			l++
		}
	}

	return res
}
