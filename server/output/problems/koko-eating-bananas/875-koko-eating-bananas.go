package koko_eating_bananas

import "math"

/*
Binary search here ... our search space in this case is going to be size of
our pile ...
left boundary will be the minimum of the piles (e.g. [1, 2] h = 3 would be one pile size 1 an hour)
the right boundary will be the maximum of the piles (e.g. [10, 10, 10, 20], h = 5, one pile size 10 an hour)

We know that, within some range (min_pile, max_pile) there is some "threshold" pile where it won't
be able to be finished in h hours ...

so everytime we find a valid pile, that is a potential result but we continue shrinking until we reach
the end boundary of the invalid piles, or the beginning boundary of valid piles ...
*/

func isValidRate(piles []int, hours, pace int) bool {
	hoursTaken := 0
	for _, p := range piles {
		// essentially take the floor of pile size, and the pace of taking piles
		// to get hours it will take to eat a given pile
		hoursTaken += (p + pace - 1) / pace
	}

	// if hours taken to eat every pile is less than total alloted hours, return true
	return hoursTaken <= hours
}

func minEatingSpeed(piles []int, h int) int {
	l := 1
	r := math.MinInt32

	// set our boundaries ...
	for _, pile := range piles {
		r = max(r, pile)
	}

	for l < r {
		mid := (l + r) / 2
		if isValidRate(piles, h, mid) {
			r = mid
		} else {
			l = mid + 1
		}
	}

	return r
}
