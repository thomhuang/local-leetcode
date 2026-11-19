package add_two_numbers

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
func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	res := &ListNode{}

	ptr := res

	carry := 0
	for l1 != nil || l2 != nil || carry > 0 {
		curr := 0

		if l1 != nil {
			curr += l1.Val
			l1 = l1.Next
		}

		if l2 != nil {
			curr += l2.Val
			l2 = l2.Next
		}

		if carry > 0 {
			curr += carry
		}

		carry = curr / 10
		curr %= 10

		ptr.Next = &ListNode{Val: curr}
		ptr = ptr.Next
	}

	return res.Next
}
