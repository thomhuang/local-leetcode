package kth_largest_element_in_an_array

import "container/heap"

/*
	Same idea as Kth largest/smallest in a set,
	we essentially make a heap in contrary to the problem set,
	e.g. Kth Largest -> MinHeap
	e.g. Kth Smallest -> MinHeap

	And we pop the heap while it has more than k elements!
*/
func findKthLargest(nums []int, k int) int {
	h := &MinHeap{}

	for _, num := range nums {
		heap.Push(h, num)

		if h.Len() > k {
			heap.Pop(h)
		}
	}

	return (*h)[0]
}

type MinHeap []int

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i] < h[j]
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(int))
}

func (h *MinHeap) Pop() any {
	n := len(*h)
	popped := (*h)[n-1]
	*h = (*h)[:n-1]
	return popped
}
