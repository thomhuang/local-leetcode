package path_sum_ii

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
func pathSum(root *TreeNode, targetSum int) [][]int {
	res := make([][]int, 0)

	var backtrack func(node *TreeNode, currSum int, path []int)

	backtrack = func(node *TreeNode, currSum int, path []int) {
		if node == nil || currSum < 0 {
			return
		}

		path = append(path, node.Val)
		currSum -= node.Val

		if node.Left == nil && node.Right == nil && currSum == 0 {
			validPath := append([]int(nil), path...)
			res = append(res, validPath)
			return
		}

		backtrack(node.Left, currSum, path)
		backtrack(node.Right, currSum, path)
	}

	backtrack(root, targetSum, make([]int, 0))
	return res
}
