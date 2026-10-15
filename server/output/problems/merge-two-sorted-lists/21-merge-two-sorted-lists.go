package merge_two_sorted_lists

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
/*
	While list1/list2 aren't nil, we can just take the greater between the two, and keep it pushin
	but after list1/list2 are nil, we have to take the rest of the nodes in the other, if applies ...
*/
func mergeTwoLists(list1 *ListNode, list2 *ListNode) *ListNode {
	res := &ListNode{}

	curr := res
	for list1 != nil && list2 != nil {
		if list1.Val < list2.Val {
			curr.Next = list1
			list1 = list1.Next
		} else {
			curr.Next = list2
			list2 = list2.Next
		}

		curr = curr.Next
	}

	if list1 != nil {
		curr.Next = list1
		list1 = list1.Next
		curr = curr.Next
	}

	if list2 != nil {
		curr.Next = list2
		list2 = list2.Next
		curr = curr.Next
	}

	return res.Next
}
