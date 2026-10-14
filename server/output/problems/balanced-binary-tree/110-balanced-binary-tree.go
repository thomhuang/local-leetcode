package balanced_binary_tree

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
We're considered height-balanced if the height of the left/right subtrees are <= 1
so we can go through each level of the subtree, calculate the height of the subtrees and see
if the heights are <= 1, and the same for its descendent subtrees as well
*/
func isBalanced(root *TreeNode) bool {
	if root == nil {
		return true
	}

	leftHeight, rightHeight := subTreeHeight(root.Left), subTreeHeight(root.Right)

	return abs(leftHeight-rightHeight) <= 1 && isBalanced(root.Left) && isBalanced(root.Right)
}

func subTreeHeight(node *TreeNode) int {
	if node == nil {
		return 0
	}

	return 1 + max(subTreeHeight(node.Left), subTreeHeight(node.Right))
}

func abs(num int) int {
	if num < 0 {
		return -num
	}

	return num
}
