package product_of_array_except_self

func productExceptSelf(nums []int) []int {
	// we are going to keep track of the products from left/right ...
	leftProduct, rightProduct := 1, 1
	res := make([]int, len(nums))

	n := len(nums)
	// we can first start from the left ...
	// we first set the result to the running left product which excludes itself
	// and then update the prefix-product afterwards
	for i := range n {
		res[i] = leftProduct
		leftProduct *= nums[i]
	}

	// res[i] already is filled with the left-subproducts
	// so we update it with the running right-subproducts
	for i := n - 1; i >= 0; i-- {
		res[i] *= rightProduct
		rightProduct *= nums[i]
	}

	return res
}
