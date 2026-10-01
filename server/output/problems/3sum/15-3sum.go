package sum3

import "sort"

func threeSum(nums []int) [][]int {
	/*
		The naiive solution is to do a nested loop between i, j = i + 1, k = j + 1 and search the combiations s.t. == 0

		But that's an O(n^3) solution, and we can probably do better than that.

		Notice that if we sort our input array and iterate through 'nums', based on the value of 'nums' we can use a two
		pointer appraoch for j, k starting at i + 1 and the end of the array, and update our search space depending
		on the current sums ...

		as we don't want any repeats, we have to be sure to move pointers forward if the previous element we processed
		is identical to the current ..
	*/

	sort.Ints(nums)

	res := make([][]int, 0)
	// as j, k start from after j, and end of 'nums', index i as a result ...
	for i := 0; i < len(nums)-2; i++ {
		// skip duplicate processed `i`s
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}

		j, k := i+1, len(nums)-1
		// as long as our search space is valid
		for j < k {
			// get the tuplet sum
			curr := nums[i] + nums[j] + nums[k]

			// if valid, add to result and shrink our window appropriately
			if curr == 0 {
				res = append(res, []int{nums[i], nums[j], nums[k]})

				j++
				k--

				for j < k && nums[j] == nums[j-1] {
					j++
				}

				for j < k && nums[k] == nums[k+1] {
					k--
				}
			} else { // otherwise, move j, k depending on the tuplet sum
				if curr < 0 {
					j++
				} else {
					k--
				}
			}
		}
	}

	return res
}
