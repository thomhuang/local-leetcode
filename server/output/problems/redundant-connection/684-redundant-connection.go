package redundant_connection

func findRedundantConnection(edges [][]int) []int {
	n := len(edges)
	parents, ranks := make([]int, n+1), make([]int, n+1)
	for i := 1; i < n+1; i++ {
		parents[i] = i
		ranks[i] = 1
	}

	for _, edge := range edges {
		// find what component either edge is connected to
		p1, p2 := find(edge[0], parents), find(edge[1], parents)
		if p1 == p2 {
			// if p1 == p2, they're already a part of the same component
			// adding them adds creates a cycle from their existing connection
			return edge
		}

		union(edge[0], edge[1], parents, ranks)
	}

	return []int{}
}

func find(node int, parents []int) int {
	if node == parents[node] {
		return node
	}

	return find(parents[node], parents)
}

func union(n1, n2 int, parents []int, ranks []int) bool {
	p1, p2 := find(n1, parents), find(n2, parents)

	if p1 == p2 { // already a part of the same component, return false
		return false
	}

	r1, r2 := ranks[p1], ranks[p2]
	if r1 == r2 {
		ranks[p1]++
		parents[p2] = p1
	} else if r1 > r2 {
		parents[p2] = p1
	} else {
		parents[p1] = p2
	}

	return true
}
