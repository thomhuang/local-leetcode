package reorder_list

type ListNode struct {
	Val  int
	Next *ListNode
}

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

/*
I think what makes sense here, is to do the following:
The first half of the linked list is in order
the second half of the linked list is in reverse order

We can use tortoise-hair algorithm to get a pointer to the "middle"
of our list
and that will be the "half" that we will reverse

Then, we can, at the pointer of the start, and the reverse half,
add them in alternating order until we have our fully reordered list
*/
func reorderList(head *ListNode) {
	if head == nil || head.Next == nil {
		return
	}

	// Find the end of the first half.
	slow, fast := head, head
	for fast.Next != nil && fast.Next.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}

	// "break off" the first/second half into
	// distinct partitions
	second := slow.Next
	slow.Next = nil

	var prev *ListNode
	for second != nil {
		next := second.Next
		second.Next = prev
		prev = second
		second = next
	}

	// add each in alternating order
	first, second := head, prev
	for second != nil {
		nextFirst, nextSecond := first.Next, second.Next

		first.Next = second
		second.Next = nextFirst

		first, second = nextFirst, nextSecond
	}
}
