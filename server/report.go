package main

import (
	"fmt"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// problemRowText renders the columns that follow a status symbol.
func problemRowText(problem question.ProblemSetProblem, category string, showCategory bool) string {
	const maxTitle = 46
	title := problem.Title
	if len(title) > maxTitle {
		title = title[:maxTitle-3] + "..."
	}

	row := fmt.Sprintf("%4d. %-*s %-6s", problem.ID, maxTitle, title, problem.Difficulty)
	if showCategory {
		row += " " + category
	}
	return row
}

func categoryNames(set question.ProblemSet) []string {
	names := make([]string, 0, len(set.Categories))
	for _, category := range set.Categories {
		names = append(names, category.Name)
	}
	return names
}

// matchCategory resolves a user-supplied category filter to its canonical name.
// An empty value or "all" means no category filter.
func matchCategory(set question.ProblemSet, want string) (string, error) {
	if want == "" || strings.EqualFold(want, "all") {
		return "", nil
	}
	for _, category := range set.Categories {
		if strings.EqualFold(category.Name, want) {
			return category.Name, nil
		}
	}
	return "", fmt.Errorf("unknown category %q, want one of: %s", want, strings.Join(categoryNames(set), ", "))
}

// printProblemSet renders a plain checklist, used by the -list flag so the
// output can be piped or grepped.
func (app *App) printProblemSet(setName, category, status string) error {
	set, err := app.LoadProblemSet(setName)
	if err != nil {
		return err
	}

	statusKey, err := parseStatusFilter(status)
	if err != nil {
		return err
	}

	category, err = matchCategory(set, category)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n\n", set.Name)

	acceptedTotal, total := 0, 0
	anyShown := false
	for _, cat := range set.Categories {
		if category != "" && cat.Name != category {
			continue
		}

		catAccepted := 0
		for _, problem := range cat.Problems {
			if app.problemState(problem) == stateAccepted {
				catAccepted++
			}
		}

		lines := make([]string, 0, len(cat.Problems))
		for _, problem := range cat.Problems {
			state := app.problemState(problem)
			if statusKey != "" && state.StatusKey() != statusKey {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s %s", state.Symbol(), problemRowText(problem, cat.Name, category == "")))
		}
		if len(lines) == 0 {
			continue
		}

		anyShown = true
		fmt.Printf("%s  %d/%d\n", cat.Name, catAccepted, len(cat.Problems))
		for _, line := range lines {
			fmt.Println("  " + line)
		}
		fmt.Println()

		acceptedTotal += catAccepted
		total += len(cat.Problems)
	}

	if !anyShown {
		fmt.Println("No problems matched the filter.")
		return nil
	}
	fmt.Printf("accepted %d/%d\n", acceptedTotal, total)
	return nil
}
