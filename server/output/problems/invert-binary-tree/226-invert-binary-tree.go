package invert_binary_tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
/*
	We can invert recursively ... pretty straight forward.
*/
func invertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}

	// swap the left, right nodes positionally
	root.Left, root.Right = root.Right, root.Left

	// do the same for the left/right subtrees
	invertTree(root.Left)
	invertTree(root.Right)

	return root
}
