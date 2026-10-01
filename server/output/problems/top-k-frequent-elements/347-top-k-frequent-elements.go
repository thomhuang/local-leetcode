package top_k_frequent_elements

import "sort"

func topKFrequent(nums []int, k int) []int {
	// keep map of frequencies of each num in nums
	freq := make(map[int]int)

	// get the actual frequency of each
	for _, num := range nums {
		freq[num]++
	}

	// keep track of (num, frequency)
	pairs := make([][]int, len(freq))

	i := 0
	for num, count := range freq {
		pairs[i] = []int{num, count}
		i++
	}

	// sort pairs by frequency, if tied then take smaller
	sort.Slice(pairs, func(i, j int) bool {
		n1, n2 := pairs[i][1], pairs[j][1]
		if n1 == n2 {
			return pairs[i][0] < pairs[j][0]
		}

		return n1 > n2
	})

	res := make([]int, k)
	for i := range k {
		res[i] = pairs[i][0]
	}

	return res
}
