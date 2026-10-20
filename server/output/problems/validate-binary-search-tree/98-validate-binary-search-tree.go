package validate_binary_search_tree

import "math"

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
func isValidBST(root *TreeNode) bool {
	return isValidPath(root, math.MaxInt32+1, math.MinInt32-1)
}

/*
	At each node, with a proper binary search tree ...
		- A valid Left node has to be less than the node above it
		- A valid Right node has to be greater than the node above it
	On top of every of node in the path/subtree so far ...
*/
func isValidPath(node *TreeNode, pathMax, pathMin int) bool {
	if node == nil {
		return true
	}

	// if not within our bounds, return false
	if !(node.Val < pathMax && node.Val > pathMin) {
		return false
	}

	// traverse both L/R subtrees
	return isValidPath(node.Left, node.Val, pathMin) && isValidPath(node.Right, pathMax, node.Val)
}
