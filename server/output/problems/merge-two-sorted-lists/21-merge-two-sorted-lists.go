package merge_two_sorted_lists

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
Traverse list1/list2 while both are valid and add to our result node ...

if any of list1/list2 remain, add the rest!
*/
func mergeTwoLists(list1 *ListNode, list2 *ListNode) *ListNode {
	head := &ListNode{}
	curr := head

	for list1 != nil && list2 != nil {
		next := &ListNode{}
		if list1.Val < list2.Val {
			next.Val = list1.Val
			list1 = list1.Next
		} else {
			next.Val = list2.Val
			list2 = list2.Next
		}

		curr.Next = next
		curr = curr.Next
	}

	for list1 != nil {
		curr.Next = &ListNode{Val: list1.Val}
		curr = curr.Next
		list1 = list1.Next
	}

	for list2 != nil {
		curr.Next = &ListNode{Val: list2.Val}
		curr = curr.Next
		list2 = list2.Next
	}

	return head.Next
}
