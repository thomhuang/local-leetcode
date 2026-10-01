package contains_duplicate

func containsDuplicate(nums []int) bool {
	lookup := make(map[int]bool)

	for _, num := range nums {
		if _, exists := lookup[num]; exists {
			return true
		}

		lookup[num] = true
	}

	return false

}
