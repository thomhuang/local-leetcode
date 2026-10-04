package main

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomhuang/local-leetcode/util"
)

func TestTUIFitHelpers(t *testing.T) {
	if got := fitLine("hello", 10); got != "hello" {
		t.Fatalf("fitLine short = %q, want hello", got)
	}
	if got := fitLine("hello world", 5); got != "hell…" {
		t.Fatalf("fitLine long = %q, want hell…", got)
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

// stubModel records the last message it received so tests can verify that the
// screen host forwards to its inner view.
type stubModel struct {
	last tea.Msg
}

func (s stubModel) Init() tea.Cmd { return nil }
func (s stubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	s.last = msg
	return s, nil
}
func (s stubModel) View() string { return "stub" }

func newTestScreenHost(t *testing.T) (screenHost, *util.Log) {
	t.Helper()
	log := util.NewLog(filepath.Join(t.TempDir(), "debug.log"))
	t.Cleanup(log.Close)
	return newScreenHost(&App{Log: log}, stubModel{}), log
}

func TestScreenHostDebugOverlayToggle(t *testing.T) {
	host, log := newTestScreenHost(t)
	log.Info("test.event", "hello world", "k", "v")

	updated, cmd := host.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	host = updated.(screenHost)
	if !host.debug.open {
		t.Fatal("debug overlay did not open on ctrl+o")
	}
	if cmd == nil {
		t.Fatal("opening the overlay returned no refresh command")
	}

	view := host.View()
	if !strings.Contains(view, "Debug log") || !strings.Contains(view, "test.event") {
		t.Fatalf("debug view missing content:\n%s", view)
	}

	updated, _ = host.Update(tea.KeyMsg{Type: tea.KeyEsc})
	host = updated.(screenHost)
	if host.debug.open {
		t.Fatal("debug overlay did not close on esc")
	}
}

func TestScreenHostForwardsMessagesToInner(t *testing.T) {
	host, _ := newTestScreenHost(t)

	updated, _ := host.Update(chatChunkMsg{content: "hello"})
	host = updated.(screenHost)

	inner, ok := host.inner.(stubModel)
	if !ok {
		t.Fatalf("inner type = %T, want stubModel", host.inner)
	}
	if _, ok := inner.last.(chatChunkMsg); !ok {
		t.Fatalf("inner did not receive the chat chunk, got %T", inner.last)
	}
}

func TestScreenHostBrowseExitQuits(t *testing.T) {
	host, _ := newTestScreenHost(t)

	_, cmd := host.Update(browseExitMsg{})
	if cmd == nil {
		t.Fatal("browseExitMsg returned no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("browseExitMsg command = %T, want tea.QuitMsg", cmd())
	}
}

func TestDebugPanelFiltersAndVerbose(t *testing.T) {
	log := util.NewLog(filepath.Join(t.TempDir(), "debug.log"))
	t.Cleanup(log.Close)
	log.Debug("dbg", "debug only")
	log.Error("err", "error only")

	panel := newDebugPanel(log)
	panel.setHeight(20)
	panel.level = util.LevelError

	view := panel.view(100, 20)
	if strings.Contains(view, "debug only") {
		t.Fatalf("error filter still showed debug entry:\n%s", view)
	}
	if !strings.Contains(view, "error only") {
		t.Fatalf("error filter dropped error entry:\n%s", view)
	}

	panel.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if !panel.verbose || !log.Verbose() {
		t.Fatal("v did not enable verbose recording")
	}
}
