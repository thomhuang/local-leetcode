package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIMenuViewRenders(t *testing.T) {
	m := tuiModel{}
	view := m.menuView()
	for _, label := range mainMenu {
		if !strings.Contains(view, label) {
			t.Fatalf("menu view missing %q:\n%s", label, view)
		}
	}
}

func TestTUIMenuNavigation(t *testing.T) {
	m := tuiModel{}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tuiModel)
	if m.menuIndex != 1 {
		t.Fatalf("menuIndex after down = %d, want 1", m.menuIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(tuiModel)
	if m.menuIndex != 0 {
		t.Fatalf("menuIndex after up = %d, want 0", m.menuIndex)
	}

	// Wrap around to the last entry.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(tuiModel)
	if m.menuIndex != len(mainMenu)-1 {
		t.Fatalf("menuIndex after wrap = %d, want %d", m.menuIndex, len(mainMenu)-1)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(tuiModel)
	if m.menuIndex != len(mainMenu)-2 {
		t.Fatalf("menuIndex after second up = %d, want %d", m.menuIndex, len(mainMenu)-2)
	}
	if cmd != nil {
		t.Fatalf("plain navigation returned a command, want nil")
	}
}

func TestTUIInputValidationAndSubmit(t *testing.T) {
	var submitted string
	m := tuiModel{}
	m.startInput(
		"How many?",
		"",
		func(input string) string {
			if input == "" {
				return "required"
			}
			return ""
		},
		func(input string) tea.Cmd {
			submitted = input
			return nil
		},
	)

	// Empty submit keeps the screen and shows the error.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tuiModel)
	if m.screen != screenInput || m.inputError != "required" {
		t.Fatalf("empty submit: screen=%v error=%q", m.screen, m.inputError)
	}
	if !strings.Contains(m.View(), "required") {
		t.Fatalf("input view missing validation error:\n%s", m.View())
	}

	// Type a value and submit.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("42")})
	m = updated.(tuiModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tuiModel)
	if submitted != "42" {
		t.Fatalf("submitted = %q, want 42", submitted)
	}
	if m.screen != screenBusy {
		t.Fatalf("screen after submit = %v, want screenBusy", m.screen)
	}
}

func TestTUIInputBackspaceAndEscape(t *testing.T) {
	m := tuiModel{}
	m.startInput("prompt", "", nil, nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("12")})
	m = updated.(tuiModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(tuiModel)
	if m.inputValue != "1" {
		t.Fatalf("inputValue after backspace = %q, want 1", m.inputValue)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tuiModel)
	if m.screen != screenMenu {
		t.Fatalf("screen after esc = %v, want screenMenu", m.screen)
	}
}

func TestTUISelectPick(t *testing.T) {
	var picked string
	m := tuiModel{}
	m.startSelect("Pick one", []string{"a", "b", "c"}, "working", func(choice string) tea.Cmd {
		picked = choice
		return nil
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tuiModel)
	if m.selectIndex != 1 {
		t.Fatalf("selectIndex after down = %d, want 1", m.selectIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(tuiModel)
	if picked != "b" {
		t.Fatalf("picked = %q, want b", picked)
	}
	if m.screen != screenBusy || m.busyText != "working" {
		t.Fatalf("screen=%v busyText=%q, want busy/working", m.screen, m.busyText)
	}
}

func TestTUISelectEscape(t *testing.T) {
	m := tuiModel{}
	m.startSelect("Pick one", []string{"a"}, "", nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tuiModel)
	if m.screen != screenMenu {
		t.Fatalf("screen after esc = %v, want screenMenu", m.screen)
	}
}

func TestTUIBrowseExit(t *testing.T) {
	m := tuiModel{browsing: true, screen: screenBrowse}
	updated, _ := m.Update(browseExitMsg{})
	m = updated.(tuiModel)
	if m.browsing || m.screen != screenMenu {
		t.Fatalf("browsing=%v screen=%v, want false/menu", m.browsing, m.screen)
	}
}

func TestTUIResultScrollAndBack(t *testing.T) {
	m := tuiModel{height: 6}
	m.showResult("Result", "line1\nline2\nline3\nline4\nline5\nline6\nline7")
	if m.screen != screenResult {
		t.Fatalf("screen = %v, want screenResult", m.screen)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(tuiModel)
	if m.resultOffset != 1 {
		t.Fatalf("resultOffset after down = %d, want 1", m.resultOffset)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(tuiModel)
	if m.screen != screenMenu {
		t.Fatalf("screen after esc = %v, want screenMenu", m.screen)
	}
}

func TestTUIFitHelpers(t *testing.T) {
	if got := fitLine("hello", 10); got != "hello" {
		t.Fatalf("fitLine short = %q, want hello", got)
	}
	if got := fitLine("hello world", 5); got != "hell…" {
		t.Fatalf("fitLine long = %q, want hell…", got)
	}
	if got := fitTail("hello world", 5); got != "…orld" {
		t.Fatalf("fitTail long = %q, want …orld", got)
	}
	if got := fitTail("hi", 5); got != "hi" {
		t.Fatalf("fitTail short = %q, want hi", got)
	}
}

func TestTUIProgressBar(t *testing.T) {
	if got := progressBar(0, 0, 8); got != "░░░░░░░░" {
		t.Fatalf("progressBar empty = %q", got)
	}
	if got := progressBar(4, 4, 8); got != "████████" {
		t.Fatalf("progressBar full = %q", got)
	}
	if got := progressBar(1, 2, 8); got != "████░░░░" {
		t.Fatalf("progressBar half = %q", got)
	}
}

func TestTUIConfirmHandler(t *testing.T) {
	newConfirm := func() tuiModel {
		m := tuiModel{screen: screenConfirm, confirmPrompt: "Sure?"}
		m.confirmAnswer = func(model tuiModel, yes bool) (tea.Model, tea.Cmd) {
			if yes {
				model.screen = screenInput
			} else {
				model.screen = screenMenu
			}
			return model, nil
		}
		return m
	}

	m := newConfirm()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = updated.(tuiModel)
	if m.screen != screenInput {
		t.Fatalf("screen after yes = %v, want screenInput", m.screen)
	}

	m = newConfirm()
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = updated.(tuiModel)
	if m.screen != screenMenu {
		t.Fatalf("screen after no = %v, want screenMenu", m.screen)
	}
}
