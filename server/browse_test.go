package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomhuang/local-leetcode/internal/question"
)

func testSet() question.ProblemSet {
	return question.ProblemSet{
		Name: "Test",
		Slug: "test",
		Categories: []question.ProblemSetCategory{
			{Name: "A", Problems: []question.ProblemSetProblem{{ID: 1, Slug: "a1"}, {ID: 2, Slug: "a2"}}},
			{Name: "B", Problems: []question.ProblemSetProblem{{ID: 3, Slug: "b1"}}},
		},
	}
}

func testBrowseModel() browseModel {
	m := browseModel{set: testSet(), height: 30, width: 80}
	m.rows = []browseRow{
		{category: "A", problem: question.ProblemSetProblem{ID: 1, Slug: "a1"}, state: stateTodo},
		{category: "A", problem: question.ProblemSetProblem{ID: 2, Slug: "a2"}, state: stateWritten},
		{category: "B", problem: question.ProblemSetProblem{ID: 3, Slug: "b1"}, state: stateTodo},
	}
	m.applyFilter()
	return m
}

func TestBrowseApplyFilter(t *testing.T) {
	m := testBrowseModel()
	if len(m.visible) != 3 {
		t.Fatalf("unfiltered visible = %d, want 3", len(m.visible))
	}

	m.category = "A"
	m.applyFilter()
	if len(m.visible) != 2 {
		t.Fatalf("category filter visible = %d, want 2", len(m.visible))
	}

	m.category = ""
	m.status = "todo"
	m.applyFilter()
	if len(m.visible) != 2 {
		t.Fatalf("status filter visible = %d, want 2", len(m.visible))
	}

	m.category = "A"
	m.applyFilter()
	if len(m.visible) != 1 || m.rows[m.visible[0]].problem.ID != 1 {
		t.Fatalf("combined filter visible = %v, want just problem 1", m.visible)
	}
}

func TestBrowseCycleFilters(t *testing.T) {
	m := testBrowseModel()

	m.cycleCategory()
	if m.category != "A" {
		t.Fatalf("cycleCategory() = %q, want A", m.category)
	}
	m.category = "B"
	m.cycleCategory()
	if m.category != "" {
		t.Fatalf("cycleCategory() wrapped to %q, want all", m.category)
	}

	m.cycleStatus()
	if m.status != "todo" {
		t.Fatalf("cycleStatus() = %q, want todo", m.status)
	}
}

func TestBrowseNextTodo(t *testing.T) {
	m := testBrowseModel()
	m.cursor = 0
	m.nextTodo()
	if m.rows[m.visible[m.cursor]].problem.ID != 3 {
		t.Fatalf("nextTodo() landed on problem %d, want 3", m.rows[m.visible[m.cursor]].problem.ID)
	}
}

func TestMatchCategory(t *testing.T) {
	set := testSet()

	if got, err := matchCategory(set, ""); err != nil || got != "" {
		t.Fatalf("matchCategory(\"\") = %q, %v; want empty", got, err)
	}
	if got, err := matchCategory(set, "all"); err != nil || got != "" {
		t.Fatalf("matchCategory(all) = %q, %v; want empty", got, err)
	}
	if got, err := matchCategory(set, "a"); err != nil || got != "A" {
		t.Fatalf("matchCategory(a) = %q, %v; want A", got, err)
	}
	if _, err := matchCategory(set, "missing"); err == nil {
		t.Fatalf("matchCategory(missing) error = nil, want an error")
	}
}

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestBrowseListViewRenders(t *testing.T) {
	m := testBrowseModel()
	view := m.View()
	if !strings.Contains(view, "Test") {
		t.Fatalf("list view missing set name:\n%s", view)
	}
	if !strings.Contains(view, "1. ") {
		t.Fatalf("list view missing problem rows:\n%s", view)
	}
}

func TestBrowseUpdateKeepsModelConsistent(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}

	// s cycles the status filter.
	updated, _ := m.Update(runeKey('s'))
	m = updated.(browseModel)
	if m.status != "todo" {
		t.Fatalf("status after 's' = %q, want todo", m.status)
	}
	m.status = ""

	// enter opens the detail view, esc returns.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(browseModel)
	if m.mode != modeDetail {
		t.Fatalf("mode after enter = %v, want modeDetail", m.mode)
	}
	if view := m.View(); !strings.Contains(view, "status:") {
		t.Fatalf("detail view missing status line:\n%s", view)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(browseModel)
	if m.mode != modeList {
		t.Fatalf("mode after esc = %v, want modeList", m.mode)
	}
}
