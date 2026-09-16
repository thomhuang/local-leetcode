package merge_intervals

import (
	"sort"
)

func merge(intervals [][]int) [][]int {
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i][0] == intervals[j][0] {
			return intervals[i][1] < intervals[j][1]
		}

		return intervals[i][0] < intervals[j][0]
	})

	merged := [][]int{intervals[0]}

	for i := 1; i < len(intervals); i++ {
		last := len(merged) - 1

		// e.g. merged[last] = [1,4], intervals[i] = []
		if merged[last][1] >= intervals[i][0] {
			merged[last][1] = max(merged[last][1], intervals[i][1])
		} else {
			merged = append(merged, intervals[i])
		}
	}

	return merged
}
