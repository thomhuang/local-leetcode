package cheapest_flights_within_k_stops

import (
	"container/heap"
	"math"
)

type item struct {
	city, cost, stops int
}

func findCheapestPrice(n int, flights [][]int, src int, dst int, k int) int {
	if src == dst {
		return 0
	}

	// 1. Build adjacency list
	adj := make([][][]int, n)
	for _, edge := range flights {
		s, d, cost := edge[0], edge[1], edge[2]
		adj[s] = append(adj[s], []int{d, cost})
	}

	// 2. Track the minimum STOPS taken to reach each city.
	// Initialize with MaxInt32 so any valid path will be less.
	minStops := make([]int, n)
	for i := range n {
		minStops[i] = math.MaxInt32
	}

	minHeap := &MinHeap{}
	heap.Init(minHeap)

	// Start at src with 0 cost and -1 stops (since src doesn't count as a stop)
	heap.Push(minHeap, item{src, 0, -1})

	for minHeap.Len() > 0 {
		node := heap.Pop(minHeap).(item)
		city, cost, stops := node.city, node.cost, node.stops

		// If we reached our destination, Dijkstra guarantees this is the cheapest valid path.
		if city == dst {
			return cost
		}

		// If we've already used k stops, we can't move to any neighbors.
		if stops >= k {
			continue
		}

		// Optimization: If we found a path to this city that used more or equal stops
		// than a previously recorded path to the same city, discard it.
		if stops >= minStops[city] {
			continue
		}
		minStops[city] = stops

		for _, neighbor := range adj[city] {
			nextCity, flightPrice := neighbor[0], neighbor[1]

			// Push all neighbor possibilities to the heap.
			// Heap sorting by cost ensures we always process cheaper overall paths first.
			heap.Push(minHeap, item{nextCity, cost + flightPrice, stops + 1})
		}
	}

	return -1
}

// --- Heap Implementation (Kept identical to yours) ---

type MinHeap []item

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].cost < h[j].cost }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Pop() any {
	prev := *h
	n := len(prev)
	popped := prev[n-1]
	*h = prev[:n-1]
	return popped
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(item))
}
