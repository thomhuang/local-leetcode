package count_good_nodes_in_binary_tree

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
/*
	DFS for each root -> node path, and keep track of the max node
	encountered in that path ... if a "good" code, then increment our counter
	do the same for the left/right subtrees
*/
func goodNodes(root *TreeNode) int {
	totalGoodNodes := 0

	var countGoodNodes func(node *TreeNode, maxSoFar int)
	countGoodNodes = func(node *TreeNode, maxSoFar int) {
		if node == nil {
			return
		}

		if node.Val >= maxSoFar {
			totalGoodNodes++
		}

		pathMax := max(maxSoFar, node.Val)
		countGoodNodes(node.Left, pathMax)
		countGoodNodes(node.Right, pathMax)
	}

	countGoodNodes(root, -math.MaxInt32)
	return totalGoodNodes
}
