package maximum_depth_of_binary_tree

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */
func maxDepth(root *TreeNode) int {
	// if node is nil, no depth ...
	if root == nil {
		return 0
	}

	// keep traversing left/right, take the max of either subtree ...
	return 1 + max(maxDepth(root.Left), maxDepth(root.Right))
}
