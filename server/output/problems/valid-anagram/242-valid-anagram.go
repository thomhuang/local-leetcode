package valid_anagram

func isAnagram(s string, t string) bool {
	if len(s) != len(t) {
		return false
	}

	var m1, m2 [26]int
	n := len(s)
	for i := range n {
		m1[s[i] - 'a']++
		m2[t[i] - 'a']++
	}

	return m1 == m2
}

