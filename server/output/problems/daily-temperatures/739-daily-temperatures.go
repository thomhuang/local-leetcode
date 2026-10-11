package daily_temperatures

/*
I think the intention for us here is to have it so that whenever we process
a given temperature, we have a stack that indicates days that
hasn't found any future days that are warmer

the stack, if it has elements, is always in a decreasing order
as if we found a day with some greater value, it would've been popped off already,
so that guarantees that when we DO pop, it's the FIRST warmer day, not some random
warmer day ..
*/
func dailyTemperatures(temperatures []int) []int {
	// initialize result, default 0
	res := make([]int, len(temperatures))

	stack := make([]int, 0)
	for i, temp := range temperatures {
		for len(stack) > 0 && temperatures[stack[len(stack)-1]] < temp {
			res[stack[len(stack)-1]] = i - stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}

		stack = append(stack, i)
	}

	return res
}
