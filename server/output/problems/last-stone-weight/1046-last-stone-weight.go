package last_stone_weight

import "container/heap"

/*
We represent our list of stones as a max heap to always get the two heaviest stones ...
For as long as our heap is of length > 2, we will pop the two largest, apply
the game rules and push it back on unti lwe're left with one remaining stone, or 0.
*/
func lastStoneWeight(stones []int) int {
	maxHeap := MaxHeap(stones)
	h := &maxHeap
	heap.Init(h)

	for h.Len() >= 2 {
		first, second := heap.Pop(h).(int), heap.Pop(h).(int)
		if first == second {
			continue
		} else {
			heap.Push(h, abs(first-second))
		}
	}

	if h.Len() == 0 {
		return 0
	}

	return heap.Pop(h).(int)
}

func abs(num int) int {
	if num < 0 {
		return -num
	}

	return num
}

type MaxHeap []int

func (h MaxHeap) Len() int {
	return len(h)
}

func (h MaxHeap) Less(i, j int) bool {
	return h[i] > h[j]
}

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(int))
}

func (h *MaxHeap) Pop() any {
	n := len(*h)
	popped := (*h)[n-1]
	*h = (*h)[:n-1]
	return popped
}
