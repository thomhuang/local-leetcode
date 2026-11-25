//go:build math

package task_scheduler

import "sort"

func leastInterval(tasks []byte, n int) int {
	counts := make([]int, 26)
	for _, task := range tasks {
		counts[int(task-'A')]++
	}

	sort.Ints(counts)
	maxFrequency := counts[25]
	idleTime := (maxFrequency - 1) * n

	for i := 24; i >= 0; i-- {
		// in the case of counts[i] == maxFrequency, it goes to the last "interval cycle"
		idleTime -= min(counts[i], maxFrequency-1)
	}

	return max(0, idleTime) + len(tasks)
}
