package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomhuang/local-leetcode/util"
)

// debugOpenKey toggles the session debug overlay from any full-screen view. It
// is a control key so it never collides with text typed into the assistant chat.
const debugOpenKey = "ctrl+o"

// debugTickMsg refreshes the open debug overlay so entries logged from
// background goroutines appear without a keypress.
type debugTickMsg struct{}

func tickDebugCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return debugTickMsg{} })
}

// debugPanel is the scrollable session-log overlay. It is embedded by every
// full-screen host so ctrl+o works in the browse checklist and the chat.
type debugPanel struct {
	log     *util.Log
	open    bool
	offset  int
	follow  bool
	level   util.Level
	verbose bool
	height  int
}

func newDebugPanel(log *util.Log) debugPanel {
	return debugPanel{log: log, level: util.LevelDebug}
}

// setHeight records the terminal height so paging and clamping match the view.
func (p *debugPanel) setHeight(height int) {
	if height > 0 {
		p.height = height
	}
}

// toggle opens or closes the overlay and starts the refresh ticker when opening.
func (p *debugPanel) toggle() tea.Cmd {
	p.open = !p.open
	if !p.open {
		return nil
	}
	p.level = util.LevelDebug
	p.follow = true
	p.offset = 0
	p.verbose = p.log.Verbose()
	return tickDebugCmd()
}

func (p *debugPanel) handleKey(key tea.KeyMsg) {
	switch key.String() {
	case "esc", "q":
		p.open = false
		return
	case "up", "k":
		p.follow = false
		p.offset--
	case "down", "j":
		p.offset++
	case "pgup":
		p.follow = false
		p.offset -= p.viewHeight()
	case "pgdown":
		p.offset += p.viewHeight()
	case "home":
		p.follow = false
		p.offset = 0
	case "end":
		p.follow = true
	case "l":
		p.level = nextDebugLevel(p.level)
	case "v":
		p.verbose = !p.verbose
		p.log.SetVerbose(p.verbose)
	case "c":
		p.log.Clear()
		p.offset = 0
		p.follow = true
	}
	p.clamp()
}

func (p *debugPanel) clamp() {
	maxOffset := len(p.lines()) - p.viewHeight()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.follow {
		p.offset = maxOffset
	}
	if p.offset < 0 {
		p.offset = 0
	}
	if p.offset > maxOffset {
		p.offset = maxOffset
	}
}

func (p debugPanel) viewWidth(width int) int {
	if width <= 0 {
		return 80
	}
	return width
}

func (p debugPanel) viewHeight() int {
	return debugBodyHeight(p.height)
}

func debugBodyHeight(terminalHeight int) int {
	h := terminalHeight - 4
	if h < 3 {
		h = 3
	}
	return h
}

// lines renders the buffered entries that pass the current level filter.
// Multi-line messages are split so scrolling stays accurate.
func (p debugPanel) lines() []string {
	entries := p.log.Entries()
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Level < p.level {
			continue
		}
		lines = append(lines, strings.Split(debugEntryText(entry), "\n")...)
	}
	return lines
}

func debugEntryText(entry util.Entry) string {
	var b strings.Builder
	b.WriteString(dimStyle.Render(entry.Time.Format("15:04:05.000")))
	b.WriteByte(' ')
	b.WriteString(coloredDebugLevel(entry.Level))
	if entry.Event != "" {
		b.WriteByte(' ')
		b.WriteString(entry.Event)
	}
	if entry.Message != "" {
		b.WriteByte(' ')
		b.WriteString(entry.Message)
	}
	for _, field := range entry.Fields {
		b.WriteByte(' ')
		b.WriteString(dimStyle.Render(field.Key + "="))
		b.WriteString(field.Value)
	}
	return b.String()
}

func coloredDebugLevel(level util.Level) string {
	switch level {
	case util.LevelError:
		return missingStyle.Render("ERROR")
	case util.LevelWarn:
		return writtenStyle.Render("WARN ")
	case util.LevelInfo:
		return accentStyle.Render("INFO ")
	default:
		return dimStyle.Render("DEBUG")
	}
}

func nextDebugLevel(level util.Level) util.Level {
	switch level {
	case util.LevelDebug:
		return util.LevelInfo
	case util.LevelInfo:
		return util.LevelWarn
	case util.LevelWarn:
		return util.LevelError
	default:
		return util.LevelDebug
	}
}

func debugLevelLabel(level util.Level) string {
	if level == util.LevelDebug {
		return "all"
	}
	return strings.ToLower(level.String())
}

// view renders the overlay for a terminal of the given size.
func (p debugPanel) view(width, height int) string {
	entries := p.log.Entries()
	lines := p.lines()

	bodyHeight := debugBodyHeight(height)
	maxOffset := len(lines) - bodyHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := p.offset
	if p.follow {
		offset = maxOffset
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	verbose := "off"
	if p.verbose {
		verbose = "on"
	}

	var b strings.Builder
	b.WriteString(clipStyled(titleStyle.Render("Debug log")+dimStyle.Render(fmt.Sprintf("  ·  %d entries  ·  %s", len(entries), p.log.Path())), p.viewWidth(width)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", p.viewWidth(width))))
	b.WriteString("\n")

	end := offset + bodyHeight
	if end > len(lines) {
		end = len(lines)
	}
	for i := offset; i < end; i++ {
		b.WriteString(clipStyled(lines[i], p.viewWidth(width)))
		b.WriteString("\n")
	}
	for i := end - offset; i < bodyHeight; i++ {
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render(strings.Repeat("─", p.viewWidth(width))))
	b.WriteString("\n")
	help := fmt.Sprintf("level: %s · verbose: %s · ↑/↓ scroll · l level · v verbose · c clear · esc close",
		debugLevelLabel(p.level), verbose)
	b.WriteString(clipStyled(dimStyle.Render(help), p.viewWidth(width)))
	return b.String()
}
