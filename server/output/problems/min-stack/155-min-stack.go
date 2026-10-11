package min_stack

/*
I think the idea, is to have have our stack hold a pair ...
(x1, x2) where x1 is the value, and x2 is the minimum value at that point in time.
*/
type MinStack struct {
	values [][2]int
}

func Constructor() MinStack {
	return MinStack{}

}

func (this *MinStack) Push(value int) {
	var currMin int
	if len(this.values) > 0 {
		last := this.values[len(this.values)-1]
		// if there is already some values in the stack,
		// check if our new value isn't less than our current minimum
		currMin = min(last[1], value)
	} else {
		currMin = value
	}
	this.values = append(this.values, [2]int{value, currMin})
}

// operations will always be called on non-empty stacks
func (this *MinStack) Pop() {
	this.values = this.values[:len(this.values)-1]
}

// operations will always be called on non-empty stacks
func (this *MinStack) Top() int {
	return this.values[len(this.values)-1][0]
}

// operations will always be called on non-empty stacks
func (this *MinStack) GetMin() int {
	return this.values[len(this.values)-1][1]
}

/**
 * Your MinStack object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Push(value);
 * obj.Pop();
 * param_3 := obj.Top();
 * param_4 := obj.GetMin();
 */
