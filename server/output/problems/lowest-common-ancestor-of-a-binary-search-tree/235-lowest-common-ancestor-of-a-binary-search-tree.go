package lowest_common_ancestor_of_a_binary_search_tree

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val   int
 *     Left  *TreeNode
 *     Right *TreeNode
 * }
 */
func lowestCommonAncestor(root, p, q *TreeNode) *TreeNode {
	// if root is nil, no descendents ...
	if root == nil {
		return nil
	}

	// if the node we're processing while we go through L/R subtrees "sequentially",
	// then it's the LCA ...
	if root == p || root == q {
		return root
	}

	l, r := lowestCommonAncestor(root.Left, p, q), lowestCommonAncestor(root.Right, p, q)

	// if p or q is found between the l/r subtrees, then the root is a common ancestor
	if l != nil && r != nil {
		return root
	}

	// otherwise one of them is nil, return the found node that isn't ...
	if l != nil {
		return l
	}

	if r != nil {
		return r
	}

	return nil
}
