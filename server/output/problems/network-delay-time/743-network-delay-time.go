package network_delay_time

import (
	"container/heap"
	"math"
)

type MinHeap [][]int

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i][1] < h[j][1]
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.([]int))
}

func (h *MinHeap) Pop() any {
	n := len(*h)
	old := *h
	popped := old[n-1]
	*h = old[:n-1]

	return popped
}

func networkDelayTime(times [][]int, n int, k int) int {
	adj := make([][][]int, n+1)
	// for each node, keep track of any node connected, alongside the delay time
	for _, time := range times {
		adj[time[0]] = append(adj[time[0]], []int{time[1], time[2]})
	}

	// keep track of the minimum distance from source to a given network
	distances := make([]int, n+1)
	for i := 1; i <= n; i++ {
		distances[i] = math.MaxInt32
	}

	// initialize our heap, and distance to source is of course 0 ...
	h := &MinHeap{[]int{k, 0}}
	distances[k] = 0
	for h.Len() > 0 {
		// process the node we've traveled to
		node := heap.Pop(h).([]int)

		curr, distance := node[0], node[1]
		// for each neighbor from 'curr'
		// we will calculate the delay from our path so far alongside the delay
		// to the next node ... if it's less than what we've encountered so far,
		// then we can add it to our heap and continue processing
		for _, neighbor := range adj[curr] {
			destination, pathLength := neighbor[0], neighbor[1]
			if distance+pathLength >= distances[destination] {
				continue
			}

			distances[destination] = distance + pathLength
			heap.Push(h, []int{destination, distance + pathLength})
		}
	}

	maxDistance := 0
	for node := 1; node <= n; node++ {
		if distances[node] == math.MaxInt32 {
			return -1
		}
		maxDistance = max(maxDistance, distances[node])
	}
	return maxDistance
}
