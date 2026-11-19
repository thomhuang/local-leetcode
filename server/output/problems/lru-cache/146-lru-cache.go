package lru_cache

type Node struct {
	Key, Value int
	Next, Prev *Node
}

type LRUCache struct {
	MaxCapacity, Capacity int
	Lookup                map[int]*Node
	// Head.Next will be true head
	// Tail.Prev will be true tail
	Head, Tail *Node
}

func Constructor(capacity int) LRUCache {
	head, tail := &Node{}, &Node{}
	head.Next = tail
	tail.Prev = head

	return LRUCache{
		MaxCapacity: capacity,
		Capacity:    0,
		Lookup:      make(map[int]*Node),
		Head:        head,
		Tail:        tail,
	}
}

func (this *LRUCache) Get(key int) int {
	// get node, and existence bool
	node, exists := this.Lookup[key]

	// if doesn't exist in lookup, return -1
	if !exists {
		return -1
	}

	// otherwise this node is most recently used
	// so we have to bump it to the top of our LL
	this.remove(node)
	this.accessed(node)

	return node.Value
}

func (this *LRUCache) Put(key int, value int) {
	node, exists := this.Lookup[key]

	var curr *Node
	// if exists, update existing node and remove it from its
	// place in list
	if exists {
		node.Value = value
		curr = node
		this.remove(curr)
	} else { // otherwise add new node
		curr = &Node{Key: key, Value: value}
		this.Lookup[key] = curr
		this.Capacity++
	}

	if this.Capacity > this.MaxCapacity {
		evicted := this.Tail.Prev
		this.remove(evicted)
		delete(this.Lookup, evicted.Key)
		this.Capacity--
	}

	this.accessed(curr)
}

// remove detaches a node from the linked list.
func (this *LRUCache) remove(node *Node) {
	// remove node from  linked list
	prev := node.Prev
	next := node.Next
	prev.Next = next
	next.Prev = prev

	// clean up the node after removing it
	node.Next = nil
	node.Prev = nil
}

func (this *LRUCache) accessed(node *Node) {
	// keep pointer to old head
	old := this.Head.Next

	// set head to accessed node
	this.Head.Next = node
	// insert node in front of head, and point its prev to the head ptr
	node.Prev = this.Head
	node.Next = old
	// previous head's prev node is now accessed node
	old.Prev = node

}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
