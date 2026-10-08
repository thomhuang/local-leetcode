package valid_parentheses

func isValid(s string) bool {
	/*
		I think the idea here, is that everytime we
		run into an open parentheses, we know that
		it must be closed with its associated parentheses
		prior to closing another type ...

		so what i'm thinking, is that we keep a stack
		of the open parentheses, so the last one in
		is the first one out if we run into its matching
		parentheses ...

		If the stack is empty, we have to check if it's a
		closing parentheses, cause that's also invalid as
		there is no matching
	*/

	complements := make(map[rune]rune)
	complements[')'] = '('
	complements[']'] = '['
	complements['}'] = '{'

	stack := make([]rune, 0)
	for _, bracket := range s {
		complement, exists := complements[bracket]

		if !exists {
			stack = append(stack, bracket)
		} else {
			if len(stack) == 0 {
				return false
			}

			curr := stack[len(stack)-1]
			if curr != complement {
				return false
			}

			stack = stack[:len(stack)-1]
		}
	}
	return len(stack) == 0
}
