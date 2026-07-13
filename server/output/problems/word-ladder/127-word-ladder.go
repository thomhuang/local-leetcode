package word_ladder

/*
This is a standard BFS, where we can check the possibilities of 1-letter diff between
a given word in some wordList[i] and those would be the neighbors ...

Keep track of a visited so we don't ping pong between words that are 1-diff'd
*/
func ladderLength(beginWord string, endWord string, wordList []string) int {
	res := 1

	wordNeighbors := make(map[string][]string)
	wordNeighbors[beginWord] = make([]string, 0)
	endExists := false
	for _, word := range wordList {
		wordNeighbors[word] = make([]string, 0)
		if word == endWord {
			endExists = true
		}
	}

	if !endExists {
		return 0
	}

	setNeighbors(wordNeighbors)

	visited := make(map[string]bool)
	queue := []string{beginWord}
	visited[beginWord] = true
	for len(queue) > 0 {
		level := len(queue)
		for i := 0; i < level; i++ {
			curr := queue[0]
			queue = queue[1:]

			if curr == endWord {
				return res
			}

			for _, neighbor := range wordNeighbors[curr] {
				if visited[neighbor] {
					continue
				}

				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}

		res++
	}

	return 0
}

func setNeighbors(wordNeighbors map[string][]string) {
	for word := range wordNeighbors {
		for i := 0; i < len(word); i++ {
			for c := 'a'; c <= 'z'; c++ {
				if rune(word[i]) == c {
					continue
				}

				curr := word[:i] + string(c) + word[i+1:]
				if _, exists := wordNeighbors[curr]; exists {
					wordNeighbors[word] = append(wordNeighbors[word], curr)
				}
			}
		}
	}
}
