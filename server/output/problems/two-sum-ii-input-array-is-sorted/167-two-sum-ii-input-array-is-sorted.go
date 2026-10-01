package two_sum_ii_input_array_is_sorted

func twoSum(numbers []int, target int) []int {
	/*
		Pretty straight forward, as we know that `numbers` is sorted,
		if numbers[l] + numbers[r] < target, then move `l` forward, and vice versa
		this gives us an O(n) runtime worst case scenario, and we don't need any
		extra memory other than our pointers
	*/

	l, r := 0, len(numbers)-1

	// while our window is valid ...
	for l < r {
		// get current sum
		curr := numbers[l] + numbers[r]
		if curr == target { // if equal, we found our combination
			return []int{l + 1, r + 1}
		} else if curr < target { // if less than target, continue forward
			l++
		} else {
			r--
		}
	}

	return []int{}
}
