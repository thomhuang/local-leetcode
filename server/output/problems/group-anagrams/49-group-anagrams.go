package group_anagrams

import (
	"slices"
)

func groupAnagrams(strs []string) [][]string {
	res := make([][]string, 0)

	// lookup between sorted word and its index in res if applies
	lookup := make(map[string]int)
	for _, s := range strs {
		// sort the word ...
		word := []rune(s)
		slices.Sort(word)

		// if the sorted word exists in lookup, append the original word to the corresponding index in res
		if idx, exists := lookup[string(word)]; exists {
			res[idx] = append(res[idx], s)
		} else { // otherwise, add a new entry in res and update the lookup
			res = append(res, []string{s})
			lookup[string(word)] = len(res) - 1
		}
	}

	return res
}
