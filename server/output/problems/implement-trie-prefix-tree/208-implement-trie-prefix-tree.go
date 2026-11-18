package implement_trie_prefix_tree

type Node struct {
	child map[rune]*Node
	end   bool
}

func NewNode() *Node {
	return &Node{
		child: make(map[rune]*Node),
		end:   false,
	}
}

type Trie struct {
	root *Node
}

func Constructor() Trie {
	return Trie{
		root: NewNode(),
	}
}

func (this *Trie) Insert(word string) {
	curr := this.root

	for _, ch := range word {
		if node, exists := curr.child[ch]; exists {
			curr = node
		} else {
			curr.child[ch] = NewNode()
			curr = curr.child[ch]
		}
	}
	curr.end = true
}

func (this *Trie) Search(word string) bool {
	curr := this.root

	for _, ch := range word {
		if node, exists := curr.child[ch]; exists {
			curr = node
		} else {
			return false
		}
	}

	return curr.end == true
}

func (this *Trie) StartsWith(prefix string) bool {
	curr := this.root

	for _, ch := range prefix {
		if node, exists := curr.child[ch]; exists {
			curr = node
		} else {
			return false
		}
	}

	return true
}

/**
 * Your Trie object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Insert(word);
 * param_2 := obj.Search(word);
 * param_3 := obj.StartsWith(prefix);
 */