package same_tree

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
	Go through each node of p,q and their values and see if they're identical ...
	do the same for its left/right subtrees.
*/
func isSameTree(p *TreeNode, q *TreeNode) bool {
	// if they're both null, then valid
	if p == nil && q == nil {
		return true
	}

	// we already checked if both p/q are null,
	// if one of them is null and the other isn't, then invalid
	if p == nil || q == nil {
		return false
	}

	// both p/q should be valid here,
	// check their values respectively, and then recurse on their left/right nodes
	if p.Val != q.Val {
		return false
	}

	// check their subtrees
	return isSameTree(p.Left, q.Left) && isSameTree(p.Right, q.Right)
}
