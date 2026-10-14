package binary_tree_right_side_view

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
	BFS, but only "store" the last element processed in a given level
*/
func rightSideView(root *TreeNode) []int {
	res := make([]int, 0)
	if root == nil {
		return res
	}

	queue := [](*TreeNode){root}

	// standard BFS here ...
	for len(queue) > 0 {
		level := len(queue)
		for i := range level {
			curr := queue[0]
			queue = queue[1:]

			// if this is the last node for the given level,
			// this corresponds to right side view, so add it to our result
			if i == level-1 {
				res = append(res, curr.Val)
			}

			if curr.Left != nil {
				queue = append(queue, curr.Left)
			}

			if curr.Right != nil {
				queue = append(queue, curr.Right)
			}
		}
	}

	return res
}
