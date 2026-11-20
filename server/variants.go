package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/question"
)

// activeVariantFile records which implementation of a problem is currently
// selected. An empty value means the primary implementation.
const activeVariantFile = ".active"

// solutionVariant is one submit-able implementation of a problem: the primary
// file or a named variant.
type solutionVariant struct {
	Name     string // empty for the primary implementation
	Path     string
	Primary  bool
	Selected bool
}

// label returns the human-facing name of a variant.
func (v solutionVariant) label() string {
	if v.Primary {
		return "primary"
	}
	return v.Name
}

// primarySolutionPath returns the path of the primary (build-tag-free) solution.
func primarySolutionPath(baseDir string, id int, slug string) string {
	return baseDir + "/" + slug + "/" + strconv.Itoa(id) + "-" + slug + ".go"
}

// variantSolutionPath returns the path of a named variant solution.
func variantSolutionPath(baseDir string, id int, slug, name string) string {
	return baseDir + "/" + slug + "/" + strconv.Itoa(id) + "-" + slug + "_" + sanitizeVariantName(name) + ".go"
}

// solutionPath returns the primary path for an empty or "primary" name, and the
// named variant's path otherwise.
func solutionPath(baseDir string, id int, slug, name string) string {
	if name == "" || name == "primary" {
		return primarySolutionPath(baseDir, id, slug)
	}
	return variantSolutionPath(baseDir, id, slug, name)
}

// sanitizeVariantName converts a free-form variant label into a filename and
// build-tag fragment: lowercase alphanumerics and underscores, never starting
// with a digit. It returns "" when nothing valid remains.
func sanitizeVariantName(name string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		default:
			sb.WriteByte('_')
		}
	}
	out := strings.Trim(sb.String(), "_")
	if out == "" {
		return ""
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "v" + out
	}
	return out
}

// variantBuildTag returns the //go:build constraint that keeps a variant file
// out of the default go vet/test build while still compiling under its tag.
func variantBuildTag(name string) string {
	return "//go:build " + sanitizeVariantName(name) + "\n\n"
}

// activeVariantName reads the selected variant for a problem, falling back to
// the primary (empty) when nothing is selected or the selection is gone.
func activeVariantName(baseDir string, id int, slug string) string {
	data, err := os.ReadFile(baseDir + "/" + slug + "/" + activeVariantFile)
	if err != nil {
		return ""
	}
	name := sanitizeVariantName(strings.TrimSpace(string(data)))
	if name == "" || name == "primary" {
		return ""
	}
	if _, err := os.Stat(variantSolutionPath(baseDir, id, slug, name)); err != nil {
		return ""
	}
	return name
}

// setActiveVariant records the selected implementation for a problem. "primary"
// or an empty name selects the primary file.
func setActiveVariant(baseDir string, id int, slug, name string) error {
	dir := baseDir + "/" + slug
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}
	value := sanitizeVariantName(strings.TrimSpace(name))
	if value == "primary" {
		value = ""
	}
	return os.WriteFile(dir+"/"+activeVariantFile, []byte(value), filePerm)
}

// listSolutionVariants returns the primary and any variant implementations for a
// problem, with the primary first and the active one flagged.
func listSolutionVariants(baseDir string, id int, slug string) []solutionVariant {
	active := activeVariantName(baseDir, id, slug)
	var out []solutionVariant

	if primary := primarySolutionPath(baseDir, id, slug); fileExists(primary) {
		out = append(out, solutionVariant{Name: "", Path: primary, Primary: true, Selected: active == ""})
	}

	dir := baseDir + "/" + slug
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	prefix := strconv.Itoa(id) + "-" + slug + "_"
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		names = append(names, strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".go"))
	}
	sort.Strings(names)
	for _, name := range names {
		out = append(out, solutionVariant{
			Name:     name,
			Path:     dir + "/" + prefix + name + ".go",
			Selected: active == name,
		})
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// listVariants resolves a problem and returns its implementations.
func (app *App) listVariants(id int) ([]solutionVariant, error) {
	meta, err := app.resolveQuestion(id)
	if err != nil {
		return nil, err
	}
	return listSolutionVariants(problemsDir, id, meta.QuestionTitleSlug), nil
}

// createVariant adds a new implementation starting from a fresh LeetCode
// starter stub and selects it. It is safe to call when the primary file already
// has a completed solution; the new variant starts empty.
func (app *App) createVariant(id int, name string) (string, error) {
	cleaned := sanitizeVariantName(name)
	if cleaned == "" {
		return "", fmt.Errorf("variant name %q is invalid; use letters, digits, - or _", name)
	}
	if cleaned == "primary" {
		return "", fmt.Errorf("%q is reserved for the primary implementation", name)
	}

	meta, err := app.resolveQuestion(id)
	if err != nil {
		return "", err
	}
	slug := meta.QuestionTitleSlug

	path := variantSolutionPath(problemsDir, id, slug, cleaned)
	if fileExists(path) {
		return "", fmt.Errorf("variant %q already exists for problem %d", cleaned, id)
	}

	ques, err := app.fetchQuestion(slug)
	if err != nil {
		app.log().Error("variant.create", "fetch failed", "problem", id, "slug", slug, "error", err.Error())
		return "", fmt.Errorf("fetch question: %w", err)
	}
	if !hasProblemContent(ques) {
		return "", fmt.Errorf("LeetCode returned no starter code for %s", slug)
	}

	if err := os.MkdirAll(problemsDir+"/"+slug, dirPerm); err != nil {
		return "", err
	}
	content := variantBuildTag(cleaned) + question.MakeStarterCompilable(ques.CodeSnippet)
	if err := os.WriteFile(path, []byte(content), filePerm); err != nil {
		app.log().Error("variant.create", "write failed", "path", path, "error", err.Error())
		return "", fmt.Errorf("write variant file: %w", err)
	}

	if err := setActiveVariant(problemsDir, id, slug, cleaned); err != nil {
		app.log().Error("variant.create", "select failed", "slug", slug, "variant", cleaned, "error", err.Error())
		return "", err
	}

	app.log().Info("variant.create", "variant created and selected", "problem", id, "slug", slug, "variant", cleaned, "path", path)
	return fmt.Sprintf("Created variant %q for %d. %s and set it active", cleaned, id, ques.Title), nil
}

// selectVariant sets the active implementation for a problem. "primary" selects
// the primary file.
func (app *App) selectVariant(id int, name string) (string, error) {
	meta, err := app.resolveQuestion(id)
	if err != nil {
		return "", err
	}
	slug := meta.QuestionTitleSlug

	cleaned := sanitizeVariantName(strings.TrimSpace(name))
	if strings.EqualFold(strings.TrimSpace(name), "primary") {
		cleaned = "primary"
	}
	if cleaned == "" {
		return "", fmt.Errorf("variant name %q is invalid", name)
	}
	if cleaned != "primary" && !fileExists(variantSolutionPath(problemsDir, id, slug, cleaned)) {
		return "", fmt.Errorf("variant %q does not exist for problem %d", cleaned, id)
	}

	if err := setActiveVariant(problemsDir, id, slug, cleaned); err != nil {
		return "", err
	}

	label := cleaned
	if cleaned == "primary" {
		label = "primary"
	}
	app.log().Info("variant.select", "active variant set", "problem", id, "slug", slug, "variant", label)
	return fmt.Sprintf("Active implementation for problem %d is now %q", id, label), nil
}

// deleteVariant removes a named variant. The primary implementation cannot be
// deleted; when the deleted variant was active, selection falls back to primary.
func (app *App) deleteVariant(id int, name string) (string, error) {
	meta, err := app.resolveQuestion(id)
	if err != nil {
		return "", err
	}
	slug := meta.QuestionTitleSlug

	cleaned := sanitizeVariantName(strings.TrimSpace(name))
	if cleaned == "" || cleaned == "primary" {
		return "", fmt.Errorf("the primary implementation cannot be deleted")
	}
	path := variantSolutionPath(problemsDir, id, slug, cleaned)
	if !fileExists(path) {
		return "", fmt.Errorf("variant %q does not exist for problem %d", cleaned, id)
	}

	wasActive := activeVariantName(problemsDir, id, slug) == cleaned
	if err := os.Remove(path); err != nil {
		app.log().Error("variant.delete", "remove failed", "path", path, "error", err.Error())
		return "", fmt.Errorf("delete variant file: %w", err)
	}
	if wasActive {
		if err := setActiveVariant(problemsDir, id, slug, "primary"); err != nil {
			return "", err
		}
	}

	app.log().Info("variant.delete", "variant deleted", "problem", id, "slug", slug, "variant", cleaned)
	return fmt.Sprintf("Deleted variant %q for problem %d", cleaned, id), nil
}
