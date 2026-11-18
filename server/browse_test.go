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

	m.setCategoryFlag("A", true)
	m.applyFilter()
	if len(m.visible) != 2 {
		t.Fatalf("category filter visible = %d, want 2", len(m.visible))
	}

	// Multi-select categories: A + B includes all 3 rows.
	m.setCategoryFlag("B", true)
	m.applyFilter()
	if len(m.visible) != 3 {
		t.Fatalf("multi-category filter visible = %d, want 3", len(m.visible))
	}

	m.categories = nil
	m.setStatusFlag(stateTodo, true)
	m.applyFilter()
	if len(m.visible) != 2 {
		t.Fatalf("status filter visible = %d, want 2", len(m.visible))
	}

	// Multi-select statuses: todo + written includes all 3 rows.
	m.setStatusFlag(stateWritten, true)
	m.applyFilter()
	if len(m.visible) != 3 {
		t.Fatalf("multi-status filter visible = %d, want 3", len(m.visible))
	}

	m.statuses = nil
	m.setStatusFlag(stateTodo, true)
	m.setCategoryFlag("A", true)
	m.applyFilter()
	if len(m.visible) != 1 || m.rows[m.visible[0]].problem.ID != 1 {
		t.Fatalf("combined filter visible = %v, want just problem 1", m.visible)
	}
}

func TestBrowseFilterMenusToggleAndClear(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}

	// 's' opens the status filter menu.
	updated, _ := m.Update(runeKey('s'))
	m = updated.(browseModel)
	if m.mode != modeFilterStatus {
		t.Fatalf("mode after 's' = %v, want modeFilterStatus", m.mode)
	}
	if view := m.View(); !strings.Contains(view, "Filter by status") {
		t.Fatalf("status filter view missing heading:\n%s", view)
	}

	// Space toggles index 0 (stateTodo); moving down and pressing space toggles stateWritten.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(browseModel)
	if !m.statuses[stateTodo] || len(m.visible) != 2 {
		t.Fatalf("after toggling todo: statuses=%v, visible=%d", m.statuses, len(m.visible))
	}

	updated, _ = m.Update(runeKey('j'))
	m = updated.(browseModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(browseModel)
	if !m.statuses[stateWritten] || len(m.visible) != 3 {
		t.Fatalf("after toggling written: statuses=%v, visible=%d", m.statuses, len(m.visible))
	}
	if got := m.statusFilterSummary(); got != "todo, written" {
		t.Fatalf("statusFilterSummary() = %q, want \"todo, written\"", got)
	}

	// 'c' switches directly from status filter to category filter.
	updated, _ = m.Update(runeKey('c'))
	m = updated.(browseModel)
	if m.mode != modeFilterCategory {
		t.Fatalf("mode after 'c' in status filter = %v, want modeFilterCategory", m.mode)
	}
	if view := m.View(); !strings.Contains(view, "Filter by category") {
		t.Fatalf("category filter view missing heading:\n%s", view)
	}

	// Toggle category "A" with space, leaving only A's todo+written problems (2).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(browseModel)
	if !m.categories["A"] || len(m.visible) != 2 {
		t.Fatalf("after toggling category A: categories=%v, visible=%d", m.categories, len(m.visible))
	}
	if got := m.categoryFilterSummary(); got != "A" {
		t.Fatalf("categoryFilterSummary() = %q, want A", got)
	}

	// 'a' clears active flags when any are checked; pressing 'a' again checks all.
	updated, _ = m.Update(runeKey('a'))
	m = updated.(browseModel)
	if m.hasCategoryFilter() {
		t.Fatalf("expected 'a' to clear category flags, got %v", m.categories)
	}
	updated, _ = m.Update(runeKey('a'))
	m = updated.(browseModel)
	if len(m.activeCategories()) != 2 || m.categoryFilterSummary() != "all" {
		t.Fatalf("expected 'a' on empty to check all categories, got %v", m.categories)
	}

	// 'enter' returns to modeList.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(browseModel)
	if m.mode != modeList {
		t.Fatalf("mode after enter in filter menu = %v, want modeList", m.mode)
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

func TestBrowseFilterViewsFitWindowWhenShrunk(t *testing.T) {
	var categories []question.ProblemSetCategory
	for i := 0; i < 20; i++ {
		categories = append(categories, question.ProblemSetCategory{
			Name:     "Category With A Reasonably Long Name",
			Problems: []question.ProblemSetProblem{{ID: i + 1, Slug: "p"}},
		})
	}
	sizes := [][2]int{{80, 30}, {60, 24}, {40, 20}, {30, 15}, {24, 12}}
	modes := []browseMode{modeFilterStatus, modeFilterCategory}

	for _, mode := range modes {
		for _, size := range sizes {
			width, height := size[0], size[1]
			m := browseModel{
				set:    question.ProblemSet{Name: "Many", Slug: "many", Categories: categories},
				width:  width,
				height: height,
				mode:   mode,
			}
			m.moveFilter(15)

			view := m.View()
			if got := strings.Count(view, "\n") + 1; got > height {
				t.Fatalf("mode %v size %v: filter view has %d lines, over %d:\n%s", mode, size, got, height, view)
			}
			for i, line := range strings.Split(view, "\n") {
				if w := lipgloss.Width(line); w > width {
					t.Fatalf("mode %v size %v: line %d width %d exceeds %d: %q", mode, size, i, w, width, line)
				}
			}
		}
	}
}

func TestBrowseUpdateKeepsModelConsistent(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}

	// s opens the status filter menu, space toggles todo, esc returns to list.
	updated, _ := m.Update(runeKey('s'))
	m = updated.(browseModel)
	if m.mode != modeFilterStatus {
		t.Fatalf("mode after 's' = %v, want modeFilterStatus", m.mode)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(browseModel)
	if !m.statuses[stateTodo] {
		t.Fatalf("statuses after space = %v, want stateTodo enabled", m.statuses)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(browseModel)
	if m.mode != modeList {
		t.Fatalf("mode after esc from filter = %v, want modeList", m.mode)
	}
	m.statuses = nil
	m.applyFilter()

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
