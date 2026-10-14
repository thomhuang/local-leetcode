package subtree_of_another_tree

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
	For every node in the main tree, we have to check if the sub tree is contained in it ...
	We first check the current node we're on, if it's contained, we return true ...
	otherwise we will check for both the left/right subtree, and so on.
*/
func isSubtree(root *TreeNode, subRoot *TreeNode) bool {
	// if sub root is nil, then there's nothing to check
	if subRoot == nil {
		return true
	}

	// if sub root != nil and root is nil, then no way it's contained
	if root == nil {
		return false
	}

	if isSameTree(root, subRoot) {
		return true
	}

	return isSubtree(root.Left, subRoot) ||
		isSubtree(root.Right, subRoot)
}

// in `isSubTree`, if, while traversing root, we find some node \in root
// that has the same value as the subRoot, we will run this to see if they're
// identical
func isSameTree(node1, node2 *TreeNode) bool {
	if node1 == nil && node2 == nil {
		return true
	}

	if node1 == nil || node2 == nil {
		return false
	}

	if node1.Val != node2.Val {
		return false
	}

	return isSameTree(node1.Left, node2.Left) &&
		isSameTree(node1.Right, node2.Right)
}
