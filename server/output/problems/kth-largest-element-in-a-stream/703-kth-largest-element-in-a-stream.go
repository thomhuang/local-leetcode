package kth_largest_element_in_a_stream

import "container/heap"

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
	old := *h
	n := len(old)
	popped := old[n-1]
	*h = old[:n-1]

	return popped
}

// this style of question is pointing towards a heap implementation,
// and if we keep a minHeap and ensure that we only have K elements in the heap
// the front of the heap will always be the Kth largest ...
type KthLargest struct {
	capacity int
	minHeap  *MinHeap
}

func Constructor(k int, nums []int) KthLargest {
	h := &MinHeap{}
	heap.Init(h)
	for _, num := range nums {
		heap.Push(h, num)
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	return KthLargest{
		capacity: k,
		minHeap:  h,
	}
}

func (this *KthLargest) Add(val int) int {
	heap.Push(this.minHeap, val)

	if this.minHeap.Len() > this.capacity {
		heap.Pop(this.minHeap)
	}

	return (*this.minHeap)[0]
}

/**
 * Your KthLargest object will be instantiated and called as such:
 * obj := Constructor(k, nums);
 * param_1 := obj.Add(val);
 */
