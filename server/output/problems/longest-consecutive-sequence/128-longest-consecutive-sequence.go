package longest_consecutive_sequence

func longestConsecutive(nums []int) int {
	lookup := make(map[int]bool)

	var res int
	// create a lookup for numbers that exist in nums
	for _, num := range nums {
		lookup[num] = true
	}

	// go through each num
	for num, _ := range lookup {
		// this is the important check here, this makes it so that
		// we never repeat lookup sequences, and we only start sequences at
		// the start of one. If we didn't have this check, our runtime would be O(n^2) vs O(n)
		if _, exists := lookup[num-1]; exists {
			continue
		}

		// Go through the sequence, and update our result if applies
		seq := 1
		curr := num + 1
		for lookup[curr] == true {
			seq++
			curr++
		}

		res = max(res, seq)
	}

	return res
}
