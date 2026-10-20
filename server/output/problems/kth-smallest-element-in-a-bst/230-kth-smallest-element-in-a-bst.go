package kth_smallest_element_in_a_bst

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
/*
	Brute force approach here would be to
*/
func kthSmallest(root *TreeNode, k int) int {
	// keep track of nodes we've processed ...
	var stack []*TreeNode
	curr := root

	for len(stack) > 0 || curr != nil {
		// go as deep left as possible
		for curr != nil {
			stack = append(stack, curr)
			curr = curr.Left
		}

		// process the current node
		curr = stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		k--
		if k == 0 {
			return curr.Val
		}

		// Move to the right subtree to get the next "smallest"
		curr = curr.Right
	}

	return -1 // Default fallback if k is out of bounds
}
