package reverse_linked_list

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
Example: 1 -> 2 -> 3 -> 4 -> 5

We can do this in place, we just have to have "dummy" variable
that we leverage to help us rearrange our nodes ...

We want a "node" to indicate the previous node we've interacted with so that
we can set the "next" of the reversed list of the previous of the list

e.g. prev := nil, curr = 1
We want curr.Next = nil alongside a pointer to our head,
and our "prev" can point to the curr ...
*/
func reverseList(head *ListNode) *ListNode {
	var prev *ListNode

	for head != nil {
		curr := head
		head = head.Next
		curr.Next = prev
		prev = curr
	}

	return prev
}
