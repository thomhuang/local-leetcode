package trapping_rain_water

func trap(height []int) int {
	/*
				Let's think about this in the simplest scenario, and from that we can
				extrapolate from that ...

				Notice that if we have some example, e.g. :

				[1, 0, 0, 0, 1]

				intiutively, we know that the trapped rain water would be 3,
				for the gaps between our two boundaries of height 1...

				If we have:

				[3, 1, 0, 0, 2]

				Notice that in this one, the troughs are noticeable in this "image":

		        |
				|X X X|
				|.|X X|

				Let's work through this example.
				Our boundary on the left is 3, and our boundary on the right is 2 ...
				As we're bottlenecked by our boundary on the right, let's move inward ...
				so we have: res = {boundary_R = 2} - {height[right] = 0} = 2
				And the same applies for the next index ...
				so we have res = 4 by similar logic. So we've gone from r = 3 -> 2 -> 1

				So at this point, we have l = 0, r = 1
				Our right boundary is still the lesser of the two, so we can it still dictates
				what the max amount of water we can trap is ...
				so we have: res += {boundary_R = 2} - {height[right] = 1} = 5

	*/

	n := len(height)
	l, r := 0, n-1
	lMax, rMax := height[l], height[r]
	res := 0

	for l < r {
		// if our right boundary is less than our left boundary, then we can move forward
		// and see if there's any "water" we can trap ...
		// Note that we must keep a running max of our height as we move inwards
		// to keep the heights aligned
		if rMax < lMax {
			// move inward ...
			r--

			// update rMax if applies ...
			rMax = max(rMax, height[r])

			// if height[r] is a larger height, nothing to update ...
			// otherwise we have a cavity and there is trapped water
			res += rMax - height[r]
		} else { // lMax <= rMax
			l++

			lMax = max(lMax, height[l])
			res += lMax - height[l]
		}
	}

	return res
}
