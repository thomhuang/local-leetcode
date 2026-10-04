package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// problemRowText renders the columns that follow a status symbol.
func problemRowText(problem question.ProblemSetProblem, category string, showCategory bool) string {
	const maxTitle = 46
	title := problem.Title
	if runes := []rune(title); len(runes) > maxTitle {
		title = string(runes[:maxTitle-3]) + "..."
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

// printProblemSet renders a plain checklist, used by the -list flag and the
// REPL so the output can be piped or grepped.
func (app *App) printProblemSet(out io.Writer, setName, category, status string) error {
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

	fmt.Fprintf(out, "%s\n\n", set.Name)

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
		fmt.Fprintf(out, "%s  %d/%d\n", cat.Name, catAccepted, len(cat.Problems))
		for _, line := range lines {
			fmt.Fprintln(out, "  "+line)
		}
		fmt.Fprintln(out)

		acceptedTotal += catAccepted
		total += len(cat.Problems)
	}

	if !anyShown {
		fmt.Fprintln(out, "No problems matched the filter.")
		return nil
	}
	fmt.Fprintf(out, "accepted %d/%d\n", acceptedTotal, total)
	return nil
}
