package reverse_linked_list

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func reverseList(head *ListNode) *ListNode {
	var prev *ListNode

	for head != nil {
		// set our pointer to the head
		curr := head
		// move head forwards
		head = head.Next

		// "swap" pointers ...
		curr.Next = prev

		// update previous to the node we just processed ...
		prev = curr
	}

	return prev
}
