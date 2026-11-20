package k_closest_points_to_origin

import (
	"container/heap"
)

/*
K Closest points, we can use a max heap and ensure
the size is always <= k ... whenever we pop an item,
that would be outside of the k closest points and in the end we can
just return the heap casted as [][]int
*/
func kClosest(points [][]int, k int) [][]int {
	h := &MaxHeap{}
	// instead of heapify the initial points and
	// popping until we have k elements, we can have it so that
	// we will push points in order, and pop whenever we're > k
	for _, point := range points {
		heap.Push(h, point)

		if h.Len() > k {
			heap.Pop(h)
		}
	}

	return [][]int(*h)
}

type MaxHeap [][]int

func (h MaxHeap) Len() int {
	return len(h)
}

func (h MaxHeap) Less(i, j int) bool {
	return squaredDistance(h[i]) > squaredDistance(h[j])
}

func squaredDistance(point []int) int {
	return point[0]*point[0] + point[1]*point[1]
}

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.([]int))
}

func (h *MaxHeap) Pop() any {
	n := len(*h)
	popped := (*h)[:n-1]
	*h = (*h)[:n-1]

	return popped
}
