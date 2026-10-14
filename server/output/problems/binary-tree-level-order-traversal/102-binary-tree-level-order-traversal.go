package binary_tree_level_order_traversal

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
	BFS of course, at each level traversal, we will keep an array of the
	nodes we run into, L -> R
*/
func levelOrder(root *TreeNode) [][]int {
	res := make([][]int, 0)
	if root == nil {
		return res
	}

	queue := [](*TreeNode){root}
	for len(queue) > 0 {
		// only process the nodes in that level
		// as we'll be pushing more nodes as we process
		level := len(queue)
		// keep track of the nodes in a given level ...
		levelNodes := make([]int, 0)
		for i := 0; i < level; i++ {
			// pop curr node and add it to the list of nodes in the levels
			curr := queue[0]
			queue = queue[1:]
			levelNodes = append(levelNodes, curr.Val)

			// if l/r nodes are valid, add them in orders
			if curr.Left != nil {
				queue = append(queue, curr.Left)
			}

			if curr.Right != nil {
				queue = append(queue, curr.Right)
			}
		}

		// add the level nodes to the res
		res = append(res, levelNodes)
	}

	return res
}
