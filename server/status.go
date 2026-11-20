package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/progress"
	"github.com/thomhuang/local-leetcode/internal/question"
)

type problemState int

const (
	stateMissing problemState = iota
	statePremium
	stateTodo
	stateWritten
	stateAccepted
)

// Symbol is the checklist glyph shown next to a problem.
func (s problemState) Symbol() string {
	switch s {
	case stateAccepted:
		return "[x]"
	case stateWritten:
		return "[-]"
	case statePremium:
		return "[p]"
	case stateMissing:
		return "[?]"
	default:
		return "[ ]"
	}
}

func (s problemState) Label() string {
	switch s {
	case stateAccepted:
		return "accepted"
	case stateWritten:
		return "written"
	case statePremium:
		return "premium"
	case stateMissing:
		return "not imported"
	default:
		return "todo"
	}
}

// StatusKey is the stable string used by filters and flags.
func (s problemState) StatusKey() string {
	switch s {
	case stateAccepted:
		return "accepted"
	case stateWritten:
		return "written"
	case statePremium:
		return "premium"
	case stateMissing:
		return "missing"
	default:
		return "todo"
	}
}

func parseStatusFilter(value string) (string, error) {
	switch value {
	case "", "all":
		return "", nil
	case "todo", "written", "accepted", "premium", "missing":
		return value, nil
	default:
		return "", fmt.Errorf("unknown status %q, want one of: all, todo, written, accepted, premium, missing", value)
	}
}

const notImplementedMarker = `panic("not implemented")`

// problemState classifies a problem for the checklist. Accepted is read from
// the progress store; written vs todo is inferred from the starter file.
func (app *App) problemState(problem question.ProblemSetProblem) problemState {
	return problemStateIn(problemsDir, problem, app.Progress)
}

func problemStateIn(baseDir string, problem question.ProblemSetProblem, store *progress.Store) problemState {
	if store != nil && store.Accepted(problem.Slug) {
		return stateAccepted
	}

	info, err := os.Stat(baseDir + "/" + problem.Slug)
	if err != nil || !info.IsDir() {
		return stateMissing
	}

	starter := starterFilePath(baseDir, problem)
	if starter == "" {
		if problem.Premium {
			return statePremium
		}
		return stateMissing
	}

	data, err := os.ReadFile(starter)
	if err != nil {
		return stateMissing
	}
	if strings.Contains(string(data), notImplementedMarker) {
		return stateTodo
	}
	return stateWritten
}

// starterFilePath returns the primary (non-variant) Go file in a problem
// directory. Variant files carry a //go:build constraint and are skipped so
// checklist state, reset, and open always operate on the canonical file.
func starterFilePath(baseDir string, problem question.ProblemSetProblem) string {
	dir := baseDir + "/" + problem.Slug
	primary := dir + "/" + strconv.Itoa(problem.ID) + "-" + problem.Slug + ".go"
	if fileExists(primary) {
		return primary
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if isVariantFile(dir + "/" + name) {
			continue
		}
		return dir + "/" + name
	}
	return ""
}

// isVariantFile reports whether a Go file begins with a //go:build constraint,
// which marks it as a non-primary variant implementation.
func isVariantFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	return strings.HasPrefix(strings.TrimSpace(line), "//go:build ")
}
