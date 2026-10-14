package diameter_of_binary_tree

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
I think the naive way of thinking about this problem is to assume
that the longest "diameter" is the longest root to leaf node, but
there could be instances where the length is longer than when we
do a "^" shape path, from one leaf, to a shared node, to another leaf,
so let's keep that in mind.

I think what we can do, is to have a method that traverses the longest root to leaf
distance for a given node. We can perform that between the two left/right subtrees at each node
We can either take the sum of the two, or the larger between the two and continue on ...
*/
func diameterOfBinaryTree(root *TreeNode) int {
	if root == nil {
		return 0
	}

	maxDiameter := 0

	// take in node, get its left/right subtree path lengths
	// update maxDiamater as a result, return the larger of the two subtrees
	var traversePath func(node *TreeNode) int
	traversePath = func(node *TreeNode) int {
		if node == nil {
			return 0
		}

		leftPath, rightPath := traversePath(node.Left), traversePath(node.Right)
		maxDiameter = max(maxDiameter, leftPath+rightPath) // # of edges, don't count current node

		return 1 + max(leftPath, rightPath)
	}

	traversePath(root)

	return maxDiameter
}
