package evaluate_division

/*
	I think the idea here, is for each given equation,
	e.g. a/b = 2 and b/c = 3
	we can rearrange the variables into terms of the other variable. For example,
	a = 2b, b = 1/2 * a and b = 3c, c = 1/3 * b

	In that case, we can create an "adjacency list" of sorts, that maps a variable
	to another variable with their factor.

	For each query that we receive, we can check if our adjacency list contains both
	of the variables, and if we do not, it's invalid and can return -1 for that particular query

	Otherwise, we can perform DFS into our adjacency list and find another relation that will
	give us what value we're looking for.
*/

type factor struct {
	variable string
	factor   float64
}

func calcEquation(equations [][]string, values []float64, queries [][]string) []float64 {
	adj := make(map[string][]factor)

	n := len(equations)

	// This will give is an adjacency list ... for the example a/b = 2 and b/c = 3,
	// a : {b, 2}
	// b : {a, 1/2}, {c, 3}
	// c : {b, 1/3}
	for i := range n {
		equation := equations[i]
		value := values[i]

		v1, v2 := equation[0], equation[1]

		if _, exists := adj[v1]; !exists {
			adj[v1] = make([]factor, 0)
		}

		if _, exists := adj[v2]; !exists {
			adj[v2] = make([]factor, 0)
		}

		adj[v1] = append(adj[v1], factor{variable: v2, factor: value})
		adj[v2] = append(adj[v2], factor{variable: v1, factor: 1 / value})
	}

	// Now we can go through each equation and perform our traversals
	res := make([]float64, len(queries))
	for i, query := range queries {
		if !isQueryValid(adj, query) {
			res[i] = float64(-1)
			continue
		}

		res[i] = traverse(query[0], query[1], adj, make(map[string]bool))
	}

	return res
}

// If either of the variables are not found in our adj, then it's an invalid query
func isQueryValid(adj map[string][]factor, query []string) bool {
	v1, v2 := query[0], query[1]

	if _, exists := adj[v1]; !exists {
		return false
	}

	if _, exists := adj[v2]; !exists {
		return false
	}

	return true
}

// backtrack on our set of mappings
func traverse(v1, v2 string, adj map[string][]factor, visited map[string]bool) float64 {
	if v1 == v2 { // if we encounter itself, then return 1
		return float64(1)
	}

	// set visited so we don't revisit node unnecessarily
	visited[v1] = true
	for _, neighbor := range adj[v1] {
		if visited[neighbor.variable] { // if visited, continue
			continue
		}

		// if we find our second variable, then we return the factor associated
		if neighbor.variable == v2 {
			return neighbor.factor
		}

		// go further ...
		curr := traverse(neighbor.variable, v2, adj, visited)
		if curr > 0 {
			/*
				a : {b, 2}
				b : {a, 1/2}, {c, 3}
				c : {b, 1/3}

				query: [a,c]

				=> traverse(a, c)
					=> traverse(b, c)
						returns 3 (as c is a neighbor of b with b = 3c)
				returns 3 * 2 (as b is a neighbor of a with a = 2b)


			*/
			return curr * neighbor.factor
		}
	}

	return float64(-1)
}
