package longest_substring_without_repeating_characters

func lengthOfLongestSubstring(s string) int {
	/*
		Straight forward sliding window problem; we extend our window so long as
		we haven't run into any duplicate characters, once we have, we shrink the front of the window
		to the last position that can guarantee a unique set in the window. Keep track of the longest
		length we've encountered as we do this.
	*/
    longest := 0
	left := 0

	visited := make(map[rune]int)
	for right, c := range s {
		/*
			The naive solution here would be to always take the last position of the duplicate character ...
			but if you take a look at the example: 'abba'
			and we're at the last 'a', we have:
				visited = {'a': 0, 'b': 2}
				left = 2
			and if we take the last position we've seen 'a', then we're moving our window backwards ... Bad!
		*/
		if prev, seen := visited[c]; seen {
			left = max(prev + 1, left)
		}

		longest = max(longest, right - left + 1)
		visited[c] = right
	}

	return longest
}