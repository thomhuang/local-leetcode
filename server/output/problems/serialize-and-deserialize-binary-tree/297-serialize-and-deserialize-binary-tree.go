package serialize_and_deserialize_binary_tree

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

type Codec struct {
}

func Constructor() Codec {
	panic("not implemented")

}

// Serializes a tree to a single string.
func (this *Codec) serialize(root *TreeNode) string {
	panic("not implemented")

}

// Deserializes your encoded data to tree.
func (this *Codec) deserialize(data string) *TreeNode {
	panic("not implemented")

}

/**
 * Your Codec object will be instantiated and called as such:
 * ser := Constructor();
 * deser := Constructor();
 * data := ser.serialize(root);
 * ans := deser.deserialize(data);
 */
