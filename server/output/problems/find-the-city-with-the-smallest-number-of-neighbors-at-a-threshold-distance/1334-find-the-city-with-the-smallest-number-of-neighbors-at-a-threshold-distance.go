package find_the_city_with_the_smallest_number_of_neighbors_at_a_threshold_distance

import (
	"container/heap"
	"math"
)

type node struct {
	city, distance int
}

type MinHeap []node

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].distance < h[j].distance
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Pop() any {
	prev := *h
	n := len(prev)
	popped := prev[n-1]
	*h = prev[:n-1]
	return popped
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(node))
}

func findTheCity(n int, edges [][]int, distanceThreshold int) int {
	cityResult, minimumNeighbors := -1, math.MaxInt32
	adj := make([][]node, n)
	for _, edge := range edges {
		src, dst, cost := edge[0], edge[1], edge[2]
		adj[src] = append(adj[src], node{city: dst, distance: cost})
		adj[dst] = append(adj[dst], node{city: src, distance: cost})
	}

	for city := range n {
		cityNeighbors := djikstra(adj, city, n, distanceThreshold)
		if cityNeighbors <= minimumNeighbors {
			minimumNeighbors = cityNeighbors
			cityResult = city
		}
	}

	return cityResult
}

func djikstra(adj [][]node, currCity, numCities, distanceThreshold int) int {
	distances := make([]int, numCities)
	for city := range numCities {
		distances[city] = math.MaxInt32
	}

	distances[currCity] = 0

	h := &MinHeap{}
	heap.Init(h)
	heap.Push(h, node{city: currCity, distance: 0})

	for h.Len() > 0 {
		c := heap.Pop(h).(node)

		for _, neighbor := range adj[c.city] {
			pathLength := neighbor.distance + c.distance
			if pathLength < distances[neighbor.city] {
				distances[neighbor.city] = pathLength
				heap.Push(h, node{city: neighbor.city, distance: pathLength})
			}
		}
	}

	var res int
	for _, distance := range distances {
		if distance <= distanceThreshold {
			res++
		}
	}

	return res
}
