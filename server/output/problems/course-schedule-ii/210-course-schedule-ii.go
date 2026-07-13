package course_schedule_ii

func findOrder(numCourses int, prerequisites [][]int) []int {
	numPrerequisites := make([]int, numCourses)
	adj := make([][]int, numCourses)
	for _, prereq := range prerequisites {
		numPrerequisites[prereq[0]]++
		adj[prereq[1]] = append(adj[prereq[1]], prereq[0])
	}

	queue := make([]int, 0)
	for course, total := range numPrerequisites {
		if total == 0 {
			queue = append(queue, course)
		}
	}

	res := make([]int, 0)
	for len(queue) > 0 {
		course := queue[0]
		queue = queue[1:]

		res = append(res, course)

		for _, neighbor := range adj[course] {
			numPrerequisites[neighbor]--
			if numPrerequisites[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(res) != numCourses {
		return []int{}
	}

	return res
}
