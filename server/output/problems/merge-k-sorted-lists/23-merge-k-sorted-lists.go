package merge_k_sorted_lists

import "container/heap"

type ListNode struct {
	Val  int
	Next *ListNode
}

type MinHeap [](*ListNode)

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].Val < h[j].Val
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(*ListNode))
}

func (h *MinHeap) Pop() any {
	popped := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]

	return popped
}

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func mergeKLists(lists []*ListNode) *ListNode {
	if len(lists) == 0 {
		return nil
	}

	// As all lists are initially sorted,
	// we're guaranteed to have the smallest item in the initial heap
	// we will add the nodes, then heapify for linear time ...
	h := &MinHeap{}
	for _, node := range lists {
		if node != nil {
			*h = append(*h, node)
		}
	}

	heap.Init(h)

	res := &ListNode{}
	curr := res
	// While there are items in our heap, i.e. there are nodes to process
	// we will pop the smallest item across the lists and add that to our resulting list
	// If the processed node has a valid Next node, we will add it to the list
	for h.Len() > 0 {
		popped := heap.Pop(h).(*ListNode)

		if popped.Next != nil {
			heap.Push(h, popped.Next)
		}

		curr.Next = popped
		curr = curr.Next
	}

	return res.Next
}
