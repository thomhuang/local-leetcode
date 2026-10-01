package two_sum

func twoSum(nums []int, target int) []int {
	lookup := make(map[int]int)

	for curr, num := range nums {
		want := target - num
		if index, exists := lookup[want]; exists {
			return []int {index, curr}
		}

		lookup[num] = curr
	}

	return []int{}
}
