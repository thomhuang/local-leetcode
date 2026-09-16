package insert_interval

func insert(intervals [][]int, newInterval []int) [][]int {
	res := make([][]int, 0)

	i := 0

	// for all intervals that are placed before new interval,
	// they don't coincide in the sense that the start time
	// of our interval is strictly less than the end time of the interval
	// we propose to add
	for ; i < len(intervals) && intervals[i][1] < newInterval[0]; i++ {
		res = append(res, intervals[i])
	}

	// Now, newInterval may potentially overlap with intervals[i]
	// and as long as that's the case, merge the start/end times
	for ; i < len(intervals) && intervals[i][0] <= newInterval[1]; i++ {
		newInterval[0] = min(newInterval[0], intervals[i][0])
		newInterval[1] = max(newInterval[1], intervals[i][1])
	}

	res = append(res, newInterval)

	// now they don't overlap at all anymore, just add until we
	// finish processing all the intervals
	for ; i < len(intervals); i++ {
		res = append(res, intervals[i])
	}

	return res
}
