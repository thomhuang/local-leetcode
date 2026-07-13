package accounts_merge

func accountsMerge(accounts [][]string) [][]string {
	parents := make(map[string]string)
	ranks := make(map[string]int)
	emailToName := make(map[string]string)

	for _, userAccount := range accounts {
		name := userAccount[0]
		firstEmail := userAccount[1]

		for i := 1; i < len(userAccount); i++ {
			email := userAccount[i]
			emailToName[email] = name

			if _, exists := parents[email]; !exists {
				parents[email] = email
				ranks[email] = 1
			}
			union(firstEmail, email, parents, ranks)
		}
	}

	rootToEmails := make(map[string][]string)
	for email := range parents {
		root := find(email, parents)
		rootToEmails[root] = append(rootToEmails[root], email)
	}

	res := make([][]string, 0, len(rootToEmails))
	for _, emails := range rootToEmails {
		name := emailToName[emails[0]]

		merged := make([]string, 0, len(emails)+1)
		merged = append(merged, name)
		merged = append(merged, emails...)
		res = append(res, merged)
	}

	return res
}

func find(email string, parents map[string]string) string {
	if parents[email] != email {
		parents[email] = find(parents[email], parents)
	}

	return parents[email]
}

func union(email1, email2 string, parents map[string]string, ranks map[string]int) bool {
	p1, p2 := find(email1, parents), find(email2, parents)
	if p1 == p2 {
		return false
	}

	r1, r2 := ranks[p1], ranks[p2]
	if r1 == r2 {
		parents[p2] = p1
		ranks[p1]++
	} else if r1 > r2 {
		parents[p2] = p1
	} else {
		parents[p1] = p2
	}

	return true
}
