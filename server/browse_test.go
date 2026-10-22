package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestBrowseListViewFitsWindowWhenShrunk(t *testing.T) {
	sizes := [][2]int{{80, 30}, {60, 24}, {40, 20}, {30, 15}, {24, 12}}
	for _, size := range sizes {
		width, height := size[0], size[1]
		m := testBrowseModel()
		m.width, m.height = width, height

		view := m.View()
		if got := strings.Count(view, "\n") + 1; got > height {
			t.Fatalf("size %v: list view has %d lines, over the %d-line height:\n%s", size, got, height, view)
		}
		for i, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Fatalf("size %v: line %d width %d exceeds %d: %q", size, i, w, width, line)
			}
		}
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

func TestBrowseChatReusesSessionForSameProblem(t *testing.T) {
	t.Setenv(deepseekAPIKeyEnv, "test-key")

	m := testBrowseModel()
	m.app = &App{}
	m.mode = modeDetail
	m.detailIndex = 0 // slug "a1"
	m.chatSlug = "a1"
	m.chat = testChatModel()
	m.chat.input = "kept"

	m.openChat()

	if m.mode != modeChat {
		t.Fatalf("mode = %v, want modeChat", m.mode)
	}
	if m.chat.input != "kept" {
		t.Fatalf("chat input = %q, want the existing session reused", m.chat.input)
	}
	if m.chatSlug != "a1" {
		t.Fatalf("chatSlug = %q, want a1", m.chatSlug)
	}
}

func TestBrowseResetConfirmation(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}
	m.mode = modeDetail
	m.detailIndex = 0

	// 'r' only enters confirmation when a starter file exists on disk; the
	// fake slugs used by the test have no files.
	if updated, cmd := m.Update(runeKey('r')); updated.(browseModel).confirmReset || cmd != nil {
		t.Fatal("r with no starter file should not confirm or run a command")
	}

	// 'y' confirms: it clears the flag, marks busy, and returns the reset cmd.
	m.confirmReset = true
	updated, cmd := m.Update(runeKey('y'))
	m = updated.(browseModel)
	if m.confirmReset {
		t.Fatal("y should clear confirmReset")
	}
	if !m.busy {
		t.Fatal("y should mark the model busy")
	}
	if cmd == nil {
		t.Fatal("y should return the reset command")
	}

	// 'n' cancels without running anything.
	m.busy = false
	m.confirmReset = true
	updated, cmd = m.Update(runeKey('n'))
	m = updated.(browseModel)
	if m.confirmReset || m.busy || cmd != nil {
		t.Fatal("n should cancel the confirmation and run nothing")
	}

	// 'esc' also cancels.
	m.confirmReset = true
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(browseModel)
	if m.confirmReset {
		t.Fatal("esc should cancel the confirmation")
	}

	// The confirmation prompt is rendered in the detail view.
	m.confirmReset = true
	if view := m.View(); !strings.Contains(view, "Reset implementation to starter? (y/n)") {
		t.Fatalf("detail view missing confirmation prompt:\n%s", view)
	}
}

func TestExtractImageURLs(t *testing.T) {
	markdown := "Intro\n\n![](https://assets.leetcode.com/a.png)\n\n**![](https://assets.leetcode.com/b.png)**\n\n![alt](https://assets.leetcode.com/c.png)\n\n![](https://assets.leetcode.com/a.png)\n"
	got := extractImageURLs(markdown)
	want := []string{
		"https://assets.leetcode.com/a.png",
		"https://assets.leetcode.com/b.png",
		"https://assets.leetcode.com/c.png",
	}
	if len(got) != len(want) {
		t.Fatalf("extractImageURLs = %d urls, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("url %d = %q, want %q", i, got[i], want[i])
		}
	}

	if urls := extractImageURLs("no images here"); len(urls) != 0 {
		t.Fatalf("extractImageURLs(no images) = %v, want empty", urls)
	}
}
