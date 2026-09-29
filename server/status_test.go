package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thomhuang/local-leetcode/internal/progress"
	"github.com/thomhuang/local-leetcode/internal/question"
)

func writeProblemFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestProblemStateIn(t *testing.T) {
	base := t.TempDir()
	twoSum := question.ProblemSetProblem{ID: 1, Slug: "two-sum", Title: "Two Sum"}

	if got := problemStateIn(base, twoSum, nil); got != stateMissing {
		t.Fatalf("no directory: got %v, want stateMissing", got)
	}

	writeProblemFile(t, filepath.Join(base, "two-sum", "1-two-sum.go"),
		"package two_sum\n\nfunc twoSum(nums []int, target int) []int {\n\tpanic(\"not implemented\")\n}\n")
	if got := problemStateIn(base, twoSum, nil); got != stateTodo {
		t.Fatalf("stub file: got %v, want stateTodo", got)
	}

	writeProblemFile(t, filepath.Join(base, "two-sum", "1-two-sum.go"),
		"package two_sum\n\nfunc twoSum(nums []int, target int) []int {\n\treturn []int{0, 1}\n}\n")
	if got := problemStateIn(base, twoSum, nil); got != stateWritten {
		t.Fatalf("solved file: got %v, want stateWritten", got)
	}

	store, err := progress.Load(filepath.Join(base, "progress.json"))
	if err != nil {
		t.Fatalf("progress.Load: %v", err)
	}
	store.MarkAccepted("two-sum", time.Now())
	if got := problemStateIn(base, twoSum, store); got != stateAccepted {
		t.Fatalf("accepted record: got %v, want stateAccepted (should override the file)", got)
	}
}

func TestProblemStateInPremium(t *testing.T) {
	base := t.TempDir()
	premium := question.ProblemSetProblem{ID: 286, Slug: "walls-and-gates", Premium: true}
	writeProblemFile(t, filepath.Join(base, "walls-and-gates", "walls-and-gates_test.go"),
		"package walls_and_gates\n\nimport \"testing\"\n\nfunc TestNotImported(t *testing.T) { t.Skip(\"premium\") }\n")

	if got := problemStateIn(base, premium, nil); got != statePremium {
		t.Fatalf("premium placeholder: got %v, want statePremium", got)
	}

	// A non-premium directory with only a test file is treated as missing.
	odd := question.ProblemSetProblem{ID: 1, Slug: "only-test"}
	writeProblemFile(t, filepath.Join(base, "only-test", "only-test_test.go"), "package only_test\n")
	if got := problemStateIn(base, odd, nil); got != stateMissing {
		t.Fatalf("test-only directory: got %v, want stateMissing", got)
	}
}

func TestParseStatusFilter(t *testing.T) {
	for _, value := range []string{"", "all", "todo", "written", "accepted", "premium", "missing"} {
		if _, err := parseStatusFilter(value); err != nil {
			t.Fatalf("parseStatusFilter(%q) error = %v", value, err)
		}
	}
	if _, err := parseStatusFilter("nope"); err == nil {
		t.Fatalf("parseStatusFilter(\"nope\") error = nil, want an error")
	}
}
