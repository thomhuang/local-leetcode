package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomhuang/local-leetcode/internal/question"
)

func TestBuildProblemContextIn(t *testing.T) {
	base := t.TempDir()
	slug := "two-sum"
	if err := os.MkdirAll(filepath.Join(base, slug), 0o755); err != nil {
		t.Fatal(err)
	}
	statement := "# Easy: 1. Two Sum\n\nGiven an array of integers..."
	if err := os.WriteFile(filepath.Join(base, slug, slug+".md"), []byte(statement), 0o644); err != nil {
		t.Fatal(err)
	}
	code := "package twosum\n\nfunc twoSum(nums []int, target int) []int { return nil }\n"
	if err := os.WriteFile(filepath.Join(base, slug, "1-two-sum.go"), []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}

	problem := question.ProblemSetProblem{ID: 1, Title: "Two Sum", Slug: slug, Difficulty: "Easy"}
	got := buildProblemContextIn(base, problem, "Expected: [0 1]\nActual: []")

	for _, want := range []string{"Two Sum", "Easy", statement, "func twoSum", "Expected: [0 1]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("context missing %q:\n%s", want, got)
		}
	}
}

func TestBuildProblemContextWithoutFiles(t *testing.T) {
	problem := question.ProblemSetProblem{ID: 42, Title: "Unknown", Slug: "missing"}
	got := buildProblemContextIn(t.TempDir(), problem, "")
	if !strings.Contains(got, "42. Unknown") {
		t.Fatalf("context missing title:\n%s", got)
	}
}

func TestHasAssistantKey(t *testing.T) {
	t.Setenv(deepseekAPIKeyEnv, "")
	if hasAssistantKey() {
		t.Fatalf("hasAssistantKey() = true with empty env")
	}

	t.Setenv(deepseekAPIKeyEnv, "  sk-test  ")
	if !hasAssistantKey() {
		t.Fatalf("hasAssistantKey() = false with key set")
	}
}
