package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomhuang/local-leetcode/internal/assistant"
	"github.com/thomhuang/local-leetcode/util"
)

func testChatModel() chatModel {
	return newChatModel(nil, assistant.DefaultConfig(), "system prompt", "1. Two Sum", 80, 30)
}

func TestChatTypingAndSend(t *testing.T) {
	m := testChatModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hint")})
	m = updated
	if m.input != "hint" {
		t.Fatalf("input = %q, want hint", m.input)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated
	if cmd == nil {
		t.Fatalf("send should return a command")
	}
	if !m.streaming {
		t.Fatalf("streaming = false after send")
	}
	if m.input != "" {
		t.Fatalf("input not cleared: %q", m.input)
	}
	if len(m.turns) != 2 || m.turns[0].role != "you" || m.turns[1].role != "assistant" {
		t.Fatalf("turns = %+v", m.turns)
	}
	if len(m.messages) != 2 || m.messages[1].Role != "user" {
		t.Fatalf("messages = %+v", m.messages)
	}
}

func TestChatQIsTypedNotQuit(t *testing.T) {
	m := testChatModel()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated
	if m.input != "q" {
		t.Fatalf("input = %q, want q", m.input)
	}
	if cmd != nil {
		t.Fatalf("typing q returned a command")
	}
}

func TestChatEscapeExits(t *testing.T) {
	m := testChatModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated
	if !m.exited {
		t.Fatalf("exited = false after esc")
	}
}

func TestChatAppendChunkGrowsReply(t *testing.T) {
	m := testChatModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")})
	m = updated
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	m.appendChunk("Hello", "")
	m.appendChunk(" world", "")

	if got := m.turns[1].text; got != "Hello world" {
		t.Fatalf("reply = %q, want %q", got, "Hello world")
	}
	if got := m.messages[len(m.messages)-1].Content; got != "Hello world" {
		t.Fatalf("history reply = %q, want %q", got, "Hello world")
	}
}

func TestChatReasoningKeepsThinkingIndicator(t *testing.T) {
	m := testChatModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")})
	m = updated
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	m.appendChunk("", "reasoning only")
	if !m.thinking {
		t.Fatalf("thinking = false after reasoning delta")
	}
	m.appendChunk("answer", "")
	if m.thinking {
		t.Fatalf("thinking = true after content delta")
	}
}

func TestChatErrorDropsEmptyPlaceholder(t *testing.T) {
	m := testChatModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")})
	m = updated
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	updated, _ = m.Update(chatChunkMsg{err: errors.New("boom")})
	m = updated

	if m.streaming {
		t.Fatalf("streaming = true after error")
	}
	if len(m.turns) != 1 {
		t.Fatalf("turns = %+v, want only the user turn", m.turns)
	}
	if !strings.Contains(m.status, "Assistant error") {
		t.Fatalf("status = %q", m.status)
	}
}

func TestChatResetClearsConversation(t *testing.T) {
	m := testChatModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")})
	m = updated
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated
	m.appendChunk("hello", "")
	m.finishStream()

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = updated

	if len(m.messages) != 1 {
		t.Fatalf("messages after reset = %d, want just the system message", len(m.messages))
	}
	if len(m.turns) != 0 {
		t.Fatalf("turns after reset = %d, want 0", len(m.turns))
	}
	if m.streaming {
		t.Fatalf("streaming = true after reset")
	}
	if m.input != "" {
		t.Fatalf("input = %q after reset, want empty", m.input)
	}
}

func TestBrowseAssistantRequiresKey(t *testing.T) {
	t.Setenv(deepseekAPIKeyEnv, "")

	m := testBrowseModel()
	m.app = &App{}
	m.mode = modeDetail
	m.detailIndex = 0

	updated, _ := m.Update(runeKey('a'))
	m = updated.(browseModel)

	if m.mode != modeDetail {
		t.Fatalf("mode = %v, want modeDetail", m.mode)
	}
	if !strings.Contains(m.message, deepseekAPIKeyEnv) {
		t.Fatalf("message = %q, want a hint about %s", m.message, deepseekAPIKeyEnv)
	}
}

func TestBrowseChatKeyRouting(t *testing.T) {
	m := testBrowseModel()
	m.app = &App{}
	m.mode = modeChat
	m.chat = testChatModel()

	updated, _ := m.Update(runeKey('x'))
	m = updated.(browseModel)
	if m.chat.input != "x" {
		t.Fatalf("chat input = %q, want x", m.chat.input)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(browseModel)
	if m.mode != modeDetail {
		t.Fatalf("mode = %v, want modeDetail after esc", m.mode)
	}
}

func TestChatCycleEffort(t *testing.T) {
	m := testChatModel()
	if got := m.currentEffort(); got != "medium" {
		t.Fatalf("currentEffort() = %q, want medium", got)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = updated
	if got := m.currentEffort(); got != "high" {
		t.Fatalf("after ctrl+r effort = %q, want high", got)
	}
	if !m.cfg.Thinking {
		t.Fatalf("thinking should stay enabled for high")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = updated
	if got := m.currentEffort(); got != "max" {
		t.Fatalf("effort = %q, want max", got)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = updated
	if got := m.currentEffort(); got != "none" {
		t.Fatalf("effort = %q, want none", got)
	}
	if m.cfg.Thinking {
		t.Fatalf("thinking should be disabled for none")
	}
}

func TestWrapText(t *testing.T) {
	got := wrapText("abcdef\nhi", 4)
	want := []string{"abcd", "ef", "hi"}
	if len(got) != len(want) {
		t.Fatalf("wrapText = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("wrapText = %v, want %v", got, want)
		}
	}
}

func TestChatInputWrapsWhileTyping(t *testing.T) {
	m := testChatModel()

	if got := len(m.visibleInputLines()); got != 1 {
		t.Fatalf("empty input lines = %d, want 1", got)
	}

	long := strings.Repeat("word ", 40)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(long)})
	m = updated

	lines := m.visibleInputLines()
	if len(lines) < 2 {
		t.Fatalf("wrapped input lines = %d, want at least 2", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > m.width {
			t.Fatalf("input line %d width = %d, exceeds view width %d", i, w, m.width)
		}
	}

	// The transcript shrinks by one line for every extra input line so the
	// chat history still fits above the wrapped input.
	if got, want := m.transcriptHeight(), m.height-5-len(lines); got != want {
		t.Fatalf("transcriptHeight = %d, want %d", got, want)
	}
}

func TestChatInputLinesCapToScreen(t *testing.T) {
	m := testChatModel()

	long := strings.Repeat("word ", 1000)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(long)})
	m = updated

	if got, max := len(m.visibleInputLines()), m.height-8; got > max {
		t.Fatalf("visible input lines = %d, want at most %d", got, max)
	}
}

func TestChatLogsStreamLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chat.log")
	log := util.NewLog(path)
	t.Cleanup(log.Close)

	app := &App{Log: log}
	m := newChatModel(app, assistant.DefaultConfig(), "system prompt", "1. Two Sum", 80, 30)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")})
	m = updated
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated

	m.appendChunk("hello", "")
	m.appendChunk(" world", "")
	m.finishStream()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"chat.send", "chat.first_content", "chat.stream.done"} {
		if !strings.Contains(text, want) {
			t.Fatalf("chat log missing %q:\n%s", want, text)
		}
	}

	// Verbose bodies are only written once verbose recording is enabled.
	if strings.Contains(text, "chat.send.body") {
		t.Fatalf("verbose body logged while verbose was off:\n%s", text)
	}
}

func TestWrapTextKeepsWordsWhole(t *testing.T) {
	cases := []struct {
		text  string
		width int
		want  []string
	}{
		{"hello world", 5, []string{"hello", "world"}},
		{"keep these words", 10, []string{"keep these", "words"}},
		{"a b c d", 3, []string{"a b", "c d"}},
		{"one\ntwo", 5, []string{"one", "two"}},
	}
	for _, tc := range cases {
		got := wrapText(tc.text, tc.width)
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Fatalf("wrapText(%q, %d) = %v, want %v", tc.text, tc.width, got, tc.want)
		}
	}
}

func TestWrapTextHardSplitsLongWord(t *testing.T) {
	got := wrapText("supercalifragilistic", 5)
	want := []string{"super", "calif", "ragil", "istic"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("wrapText = %v, want %v", got, want)
	}
}

func TestWrapTextPreservesIndentAndNeverExceedsWidth(t *testing.T) {
	got := wrapText("    func main() {", 12)
	if len(got) == 0 || !strings.HasPrefix(got[0], "    func") {
		t.Fatalf("first wrapped line = %v, want the indented start", got)
	}
	for _, line := range got {
		if w := len([]rune(line)); w > 12 {
			t.Fatalf("wrapped line %q width = %d, exceeds 12", line, w)
		}
	}
}

func TestChatViewFitsWindowWhenShrunk(t *testing.T) {
	sizes := [][2]int{{80, 30}, {60, 24}, {40, 20}, {30, 15}, {24, 12}}
	for _, size := range sizes {
		width, height := size[0], size[1]
		m := newChatModel(nil, assistant.DefaultConfig(), "system prompt", "1. Two Sum", width, height)
		m.input = strings.Repeat("wrap me ", 20)
		m.turns = []chatTurn{
			{role: "you", text: strings.Repeat("some words ", 20)},
			{role: "assistant", text: "a reply"},
		}

		view := m.View()
		if got := strings.Count(view, "\n") + 1; got > height {
			t.Fatalf("size %v: view has %d lines, over the %d-line height", size, got, height)
		}
		for i, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Fatalf("size %v: line %d width %d exceeds %d: %q", size, i, w, width, line)
			}
		}
	}
}

func TestChatFooterWrapsInsteadOfOverflowing(t *testing.T) {
	m := newChatModel(nil, assistant.DefaultConfig(), "system prompt", "1. Two Sum", 30, 15)
	if got := len(m.helpLines()); got < 2 {
		t.Fatalf("narrow footer help lines = %d, want it to wrap", got)
	}
	for _, line := range m.helpLines() {
		if w := lipgloss.Width(line); w > m.textWidth() {
			t.Fatalf("footer line %q width = %d, exceeds %d", line, w, m.textWidth())
		}
	}
}
