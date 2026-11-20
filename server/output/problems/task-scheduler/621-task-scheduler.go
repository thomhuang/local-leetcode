package task_scheduler

import "container/heap"

/*
	So I think the intention here, at least in the context of implementing this using
	heaps, is that we will have our heap contain elements in pairs of:
		{TaskName, TaskCount}
	Based on the most frequent task, we can estimate the maximum interval, and based
	on subsequent tasks we can determine the minimum interval.

	In the example of: tasks = ["A","A","A","B","B","B"], n = 2

	We have {A, 3}, {B, 3}, max task count of 3 with gap of 2 between another identical task

	In the worst case scenario we would have something like this:
		A _ _ A _ _ A _ _
	But if there are no other tasks with the same frequency as "A", then it would just be:
		A _ _ A _ _ A

	So, given the task with the highest frequency, let's say f_max, a "cycle" is the order of tasks prior to
	the next occurrence of f_max ...
*/

func leastInterval(tasks []byte, n int) int {
	// get frequency of each task
	frequencies := make(map[byte]int)
	for _, task := range tasks {
		frequencies[task]++
	}

	// hydrate our max heap with our tasks
	h := &MaxHeap{}
	for task, freq := range frequencies {
		heap.Push(h, Pair{Task: task, Frequency: freq})
	}

	intervals := 0
	// as long as there are tasks to process ...
	for h.Len() > 0 {
		// Within a cycle, there are a max of n + 1 tasks in it ...
		cycleSlots := n + 1
		var used []Pair
		// we have to check both cycleSLots and h.Len()
		// in the case of the final cycle, where there may more tasks
		// than can fit in a cycle slot ...
		for cycleSlots > 0 && h.Len() > 0 {
			curr := heap.Pop(h).(Pair)
			// valid task, increment interval
			intervals++
			// one less task in the cycle
			cycleSlots--

			// mark task as used, if
			// there are any remaining then add it back to reprocess
			curr.Frequency--
			if curr.Frequency > 0 {
				used = append(used, curr)
			}
		}

		// push back to heap
		for _, pair := range used {
			heap.Push(h, pair)
		}

		// if there are remaining items in the heap, then we have
		// to add any idle time
		if h.Len() > 0 {
			intervals += cycleSlots
		}
	}

	return intervals
}

type Pair struct {
	Task      byte
	Frequency int
}

type MaxHeap []Pair

func (h MaxHeap) Len() int {
	return len(h)
}

func (h MaxHeap) Less(i, j int) bool {
	return h[i].Frequency > h[j].Frequency
}

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(Pair))
}

func (h *MaxHeap) Pop() any {
	n := len(*h)

	popped := (*h)[n-1]
	*h = (*h)[:n-1]
	return popped
}
