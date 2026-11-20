package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thomhuang/local-leetcode/internal/question"
)

func TestSanitizeVariantName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"bf", "bf"},
		{"two-pointer", "two_pointer"},
		{"Two Pointer!", "two_pointer"},
		{"2", "v2"},
		{"BF", "bf"},
		{"!!!", ""},
		{"  ", ""},
		{"primary", "primary"},
	}
	for _, test := range tests {
		if got := sanitizeVariantName(test.name); got != test.want {
			t.Fatalf("sanitizeVariantName(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestSolutionPath(t *testing.T) {
	base := t.TempDir()
	if got, want := solutionPath(base, 1, "two-sum", ""), base+"/two-sum/1-two-sum.go"; got != want {
		t.Fatalf("primary solutionPath = %q, want %q", got, want)
	}
	if got, want := solutionPath(base, 1, "two-sum", "primary"), base+"/two-sum/1-two-sum.go"; got != want {
		t.Fatalf("primary solutionPath = %q, want %q", got, want)
	}
	if got, want := solutionPath(base, 1, "two-sum", "bf"), base+"/two-sum/1-two-sum_bf.go"; got != want {
		t.Fatalf("variant solutionPath = %q, want %q", got, want)
	}
}

func TestStripBuildTags(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "variant file",
			source: "//go:build bf\n\npackage two_sum\n\nfunc twoSum(nums []int, target int) []int {}\n",
			want:   "package two_sum\n\nfunc twoSum(nums []int, target int) []int {}\n",
		},
		{
			name:   "no build tag",
			source: "package two_sum\n\nfunc twoSum(nums []int, target int) []int {}\n",
			want:   "package two_sum\n\nfunc twoSum(nums []int, target int) []int {}\n",
		},
	}
	for _, test := range tests {
		if got := stripBuildTags(test.source); got != test.want {
			t.Fatalf("stripBuildTags(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestPrepareSubmissionWithBuildTag(t *testing.T) {
	source := "//go:build bf\n\npackage two_sum\n\nfunc twoSum(nums []int, target int) []int { return nil }\n"
	got := prepareSubmission(source, "two_sum")
	want := "func twoSum(nums []int, target int) []int { return nil }\n"
	if got != want {
		t.Fatalf("prepareSubmission() = %q, want %q", got, want)
	}
}

func TestListSolutionVariants(t *testing.T) {
	base := t.TempDir()
	slug := "two-sum"
	dir := filepath.Join(base, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeVariant := func(name, content string) {
		t.Helper()
		path := variantSolutionPath(base, 1, slug, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := os.WriteFile(primarySolutionPath(base, 1, slug), []byte("package two_sum\n\nfunc twoSum(nums []int, target int) []int { return nil }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeVariant("bf", "//go:build bf\n\npackage two_sum\n\nfunc twoSum(nums []int, target int) []int { return nil }\n")
	writeVariant("dp", "//go:build dp\n\npackage two_sum\n\nfunc twoSum(nums []int, target int) []int { return nil }\n")

	if err := setActiveVariant(base, 1, slug, "dp"); err != nil {
		t.Fatal(err)
	}

	variants := listSolutionVariants(base, 1, slug)
	if len(variants) != 3 {
		t.Fatalf("listSolutionVariants() len = %d, want 3", len(variants))
	}
	if !variants[0].Primary || variants[0].Name != "" {
		t.Fatalf("first variant = %+v, want primary", variants[0])
	}
	if variants[1].Name != "bf" || variants[1].Selected {
		t.Fatalf("second variant = %+v, want bf unselected", variants[1])
	}
	if variants[2].Name != "dp" || !variants[2].Selected {
		t.Fatalf("third variant = %+v, want dp selected", variants[2])
	}
}

func TestActiveVariantSelection(t *testing.T) {
	base := t.TempDir()
	slug := "two-sum"
	if err := os.MkdirAll(filepath.Join(base, slug), 0o755); err != nil {
		t.Fatal(err)
	}

	if got := activeVariantName(base, 1, slug); got != "" {
		t.Fatalf("activeVariantName() with no selection = %q, want empty", got)
	}

	// A selection whose file does not exist falls back to primary.
	if err := setActiveVariant(base, 1, slug, "bf"); err != nil {
		t.Fatal(err)
	}
	if got := activeVariantName(base, 1, slug); got != "" {
		t.Fatalf("activeVariantName() with missing file = %q, want empty", got)
	}

	// With the file present, the selection is returned.
	if err := os.WriteFile(variantSolutionPath(base, 1, slug, "bf"), []byte("//go:build bf\n\npackage two_sum\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := activeVariantName(base, 1, slug); got != "bf" {
		t.Fatalf("activeVariantName() = %q, want bf", got)
	}

	// "primary" resets to empty.
	if err := setActiveVariant(base, 1, slug, "primary"); err != nil {
		t.Fatal(err)
	}
	if got := activeVariantName(base, 1, slug); got != "" {
		t.Fatalf("activeVariantName() after primary = %q, want empty", got)
	}
}

func TestStarterFilePathSkipsVariants(t *testing.T) {
	base := t.TempDir()
	slug := "two-sum"
	dir := filepath.Join(base, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1-two-sum.go"), []byte("package two_sum\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1-two-sum_bf.go"), []byte("//go:build bf\n\npackage two_sum\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	problem := question.ProblemSetProblem{ID: 1, Slug: slug}
	got := starterFilePath(base, problem)
	want := base + "/" + slug + "/1-two-sum.go"
	if got != want {
		t.Fatalf("starterFilePath() = %q, want %q", got, want)
	}
}

func TestIsVariantFile(t *testing.T) {
	base := t.TempDir()
	variant := filepath.Join(base, "v.go")
	primary := filepath.Join(base, "p.go")
	if err := os.WriteFile(variant, []byte("//go:build bf\n\npackage x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(primary, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isVariantFile(variant) {
		t.Fatalf("isVariantFile(variant) = false, want true")
	}
	if isVariantFile(primary) {
		t.Fatalf("isVariantFile(primary) = true, want false")
	}
}
