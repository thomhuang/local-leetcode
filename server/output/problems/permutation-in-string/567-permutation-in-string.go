package permutation_in_string

func checkInclusion(s1 string, s2 string) bool {
	/*
		Very clear sliding window problem ...

		We see from the constraints that s1, s2 consist of only lowercase
		english letters

		we know as well that we are checking s1 \in s2 ...

		so we can have a sliding window of length s1, as well
		as some store that represents the characters in s1.

		As a result, we can keep track of the characters
		within that sliding window within s2, and if it ever
		equals the store that we have for s1, then we can return true ...
	*/

	// as we are checking if s1 is contained in s2,
	// if s2 is shorter than s1, that's not possible ..
	if len(s2) < len(s1) {
		return false
	}

	n := len(s1)
	var s1Chars, window [26]int

	// first, we set our character counts of s1
	// and seed our window character counts ...
	for i := range n { // from 0 -> n - 1
		s1Chars[int(s1[i]-'a')]++
		window[int(s2[i]-'a')]++
	}

	// if their lengths are equal, check if their
	// character counts match
	if len(s1) == len(s2) {
		return s1Chars == window
	}

	// otherwise, we move across our window ...
	for i := n; i < len(s2); i++ {
		if window == s1Chars {
			return true
		}

		// decrement the beginning of the window
		window[int(s2[i-n]-'a')]--
		// increment the end of the window, where we're extending to
		window[int(s2[i]-'a')]++
	}

	if window == s1Chars {
		return true
	}

	return false
}
