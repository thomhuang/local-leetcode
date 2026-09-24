package number_of_provinces

/*
The problem notes that we want to find the number of connected components
after we've made all the connections within our n x n set of connections

What instantly pops out to me is to perform a union find on each
possible connection, and if we make successful union (i.e. we union two components)
(we start with n components), and decrement from the total number
*/
func findCircleNum(isConnected [][]int) int {
	n := len(isConnected)
	totalComponents := len(isConnected)

	parents, ranks := make([]int, n), make([]int, n)
	for i := range n {
		parents[i] = i
		ranks[i] = 1
	}

	for node := range isConnected { // O(N^2)
		for neighbor := node + 1; neighbor < n; neighbor++ {
			if isConnected[node][neighbor] != 1 {
				continue
			}

			if union(node, neighbor, parents, ranks) {
				totalComponents--
			}
		}
	}
	return totalComponents
}

func find(node int, parents []int) int {
	if parents[node] == node {
		return node
	}

	parents[node] = find(parents[node], parents)
	return parents[node]
}

func union(n1, n2 int, parents []int, ranks []int) bool {
	p1, p2 := find(n1, parents), find(n2, parents)

	if p1 == p2 {
		return false
	}

	r1, r2 := ranks[p1], ranks[p2]
	if r1 >= r2 {
		parents[p2] = p1
		ranks[p1]++
	} else {
		parents[p1] = p2
		ranks[p2]++
	}

	return true
}
