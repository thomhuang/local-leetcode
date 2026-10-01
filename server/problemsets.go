package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/thomhuang/local-leetcode/internal/question"
)

const (
	// importDelay spaces out bulk requests so LeetCode does not rate limit us.
	importDelay      = 300 * time.Millisecond
	importMaxRetries = 3
)

// ProblemSetImportSummary reports the outcome of importing one problem set.
type ProblemSetImportSummary struct {
	Added   int
	Skipped int
	Premium int
	Failed  []question.ProblemSetProblem
}

func problemSetPath(name string) string {
	return problemSetsDir + "/" + name + ".json"
}

// availableProblemSets lists the set names found under server/output/ProblemSets.
func availableProblemSets() []string {
	entries, err := os.ReadDir(problemSetsDir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(names)
	return names
}

// LoadProblemSet reads and validates the named set from disk.
func (app *App) LoadProblemSet(name string) (question.ProblemSet, error) {
	data, err := os.ReadFile(problemSetPath(name))
	if err != nil {
		return question.ProblemSet{}, fmt.Errorf("read problem set %q: %w", name, err)
	}

	set, err := question.ParseProblemSet(data)
	if err != nil {
		return question.ProblemSet{}, err
	}
	return set, nil
}

// ImportProblemSet fetches every problem that is not already on disk. Problems
// that LeetCode will not return (premium) get a placeholder test file instead.
func (app *App) ImportProblemSet(set question.ProblemSet) ProblemSetImportSummary {
	var summary ProblemSetImportSummary

	for _, problem := range set.Problems() {
		imported, premium, err := app.ImportProblem(problem)
		if err != nil {
			app.Log.Append(fmt.Sprintf("Failed to import %s (%s): %s", problem.Title, problem.Slug, err.Error()))
			summary.Failed = append(summary.Failed, problem)
			continue
		}
		if !imported {
			summary.Skipped++
			continue
		}

		if premium {
			summary.Premium++
		} else {
			summary.Added++
		}
		time.Sleep(importDelay)
	}

	return summary
}

// ImportProblem fetches and saves a single problem unless it already exists. It
// reports whether the problem was imported and whether it is a premium
// placeholder (no statement or starter code).
func (app *App) ImportProblem(problem question.ProblemSetProblem) (imported, premium bool, err error) {
	if problemExists(problem.Slug) {
		return false, false, nil
	}

	ques, err := app.fetchQuestionWithRetry(problem.Slug)
	if err != nil {
		return false, false, err
	}

	if err := SaveMarkdownContent(ques); err != nil {
		return false, false, err
	}

	return true, !hasProblemContent(ques), nil
}

// fetchQuestionWithRetry retries transient failures a few times with a growing
// delay before giving up on a problem.
func (app *App) fetchQuestionWithRetry(slug string) (question.Question, error) {
	var lastErr error
	for attempt := 0; attempt < importMaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(importDelay * time.Duration(attempt+1))
		}

		ques, err := app.fetchQuestion(slug)
		if err == nil {
			return ques, nil
		}
		lastErr = err
	}
	return question.Question{}, lastErr
}

func problemExists(slug string) bool {
	info, err := os.Stat(problemsDir + "/" + slug)
	return err == nil && info.IsDir()
}

// importSetAndReport loads a set by name, imports it, and prints a summary.
func (app *App) importSetAndReport(name string) {
	set, err := app.LoadProblemSet(name)
	if err != nil {
		app.fail("Failed to load problem set", err)
		return
	}

	fmt.Printf("Importing %q (%d problems) from %s\n", set.Name, len(set.Problems()), problemSetPath(name))
	summary := app.ImportProblemSet(set)
	fmt.Printf("Done: %d added, %d skipped (already present), %d premium placeholders, %d failed\n",
		summary.Added, summary.Skipped, summary.Premium, len(summary.Failed))
	for _, problem := range summary.Failed {
		fmt.Printf("  failed: %s (%s)\n", problem.Title, problem.Slug)
	}
}
