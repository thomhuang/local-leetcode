package question

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ProblemSet is a named, ordered collection of LeetCode problems grouped into
// categories. Problem sets are stored as JSON files under
// server/output/ProblemSets and drive the bulk import action.
type ProblemSet struct {
	Name       string               `json:"name"`
	Slug       string               `json:"slug"`
	Source     string               `json:"source,omitempty"`
	Categories []ProblemSetCategory `json:"categories"`
}

type ProblemSetCategory struct {
	Name     string              `json:"name"`
	Problems []ProblemSetProblem `json:"problems"`
}

type ProblemSetProblem struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Difficulty string `json:"difficulty,omitempty"`
	Premium    bool   `json:"premium,omitempty"`
}

// ParseProblemSet decodes and validates a problem set document.
func ParseProblemSet(data []byte) (ProblemSet, error) {
	var set ProblemSet
	if err := json.Unmarshal(data, &set); err != nil {
		return ProblemSet{}, fmt.Errorf("parse problem set: %w", err)
	}
	if err := set.Validate(); err != nil {
		return ProblemSet{}, err
	}
	return set, nil
}

// Validate checks that a set has a slug, at least one category with at least
// one problem, and no duplicate problem slugs.
func (s ProblemSet) Validate() error {
	if strings.TrimSpace(s.Slug) == "" {
		return fmt.Errorf("problem set is missing a slug")
	}
	if len(s.Categories) == 0 {
		return fmt.Errorf("problem set %q has no categories", s.Slug)
	}

	seen := make(map[string]bool)
	for _, category := range s.Categories {
		if strings.TrimSpace(category.Name) == "" {
			return fmt.Errorf("problem set %q has a category with no name", s.Slug)
		}
		if len(category.Problems) == 0 {
			return fmt.Errorf("problem set %q category %q has no problems", s.Slug, category.Name)
		}
		for _, problem := range category.Problems {
			slug := strings.TrimSpace(problem.Slug)
			if slug == "" {
				return fmt.Errorf("problem set %q category %q has a problem with no slug", s.Slug, category.Name)
			}
			if seen[slug] {
				return fmt.Errorf("problem set %q lists %q more than once", s.Slug, slug)
			}
			seen[slug] = true
		}
	}
	return nil
}

// Problems returns every problem in the set in category order.
func (s ProblemSet) Problems() []ProblemSetProblem {
	var problems []ProblemSetProblem
	for _, category := range s.Categories {
		problems = append(problems, category.Problems...)
	}
	return problems
}
