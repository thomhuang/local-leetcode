package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomhuang/local-leetcode/internal/assistant"
)

// chatStreamReadyMsg reports that the streaming response is open, or why it
// failed to open.
type chatStreamReadyMsg struct {
	reader *sseReader
	closer io.Closer
	cancel context.CancelFunc
	err    error
}

// chatChunkMsg carries one streamed delta from the assistant.
type chatChunkMsg struct {
	content   string
	reasoning string
	done      bool
	err       error
}

type chatTurn struct {
	role string // "you" or "assistant"
	text string
}

// chatModel is the bubbletea sub-model for the assistant chat screen. It is
// owned by browseModel and is only live while the user is chatting.
type chatModel struct {
	app   *App
	cfg   assistant.Config
	title string

	messages []assistant.Message // full conversation sent to the API
	turns    []chatTurn          // rendered conversation

	input  string
	scroll int
	follow bool

	width, height int

	streaming        bool
	thinking         bool
	assistantStarted bool
	status           string
	exited           bool

	reader *sseReader
	closer io.Closer
	cancel context.CancelFunc

	// Per-turn diagnostics surfaced in the debug log.
	startedAt      time.Time
	contentChars   int
	reasoningChars int
	chunks         int
}

// newChatModel starts an empty conversation with a system message. system is
// the system prompt with the problem context already appended.
func newChatModel(app *App, cfg assistant.Config, system, title string, width, height int) chatModel {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	return chatModel{
		app:      app,
		cfg:      cfg,
		title:    title,
		messages: []assistant.Message{{Role: "system", Content: system}},
		follow:   true,
		width:    width,
		height:   height,
	}
}

// reasoningEfforts are the values cycled by ctrl+r. "none" disables thinking
// mode; the rest map to DeepSeek's reasoning_effort levels (medium maps to high
// server-side, but is kept so it matches the config default).
var reasoningEfforts = []string{"none", "low", "medium", "high", "max"}

// chatResetKey starts a fresh conversation for the current problem. The browse
// screen intercepts it to rebuild the grounded context; the standalone ask
// screen clears the conversation in place.
const chatResetKey = "ctrl+n"

func (m chatModel) currentEffort() string {
	if !m.cfg.Thinking {
		return "none"
	}
	if m.cfg.ReasoningEffort == "" {
		return assistant.DefaultReasoningEffort
	}
	return m.cfg.ReasoningEffort
}

// cycleEffort advances the reasoning effort for the next request. It is
// session-only; edit config.json to change the default.
func (m *chatModel) cycleEffort() {
	index := indexOf(reasoningEfforts, m.currentEffort())
	next := reasoningEfforts[(index+1)%len(reasoningEfforts)]
	m.cfg.ReasoningEffort = next
	m.cfg.Thinking = next != "none"
	m.status = "reasoning effort: " + next
}

func (m chatModel) Update(msg tea.Msg) (chatModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}
		if msg.Height > 0 {
			m.height = msg.Height
		}
		m.clampScroll()
		return m, nil
	case chatStreamReadyMsg:
		if msg.err != nil {
			m.app.log().Error("chat.stream", "failed to open stream", "error", msg.err.Error())
			m.finishStream()
			m.status = "Assistant error: " + msg.err.Error()
			return m, nil
		}
		m.app.log().Debug("chat.stream", "stream ready")
		m.reader, m.closer, m.cancel = msg.reader, msg.closer, msg.cancel
		return m, m.readCmd()
	case chatChunkMsg:
		if msg.err != nil {
			m.app.log().Error("chat.stream", "stream read failed",
				"error", msg.err.Error(),
				"chunks", m.chunks,
				"content_chars", m.contentChars)
			m.finishStream()
			m.status = "Assistant error: " + msg.err.Error()
			return m, nil
		}
		if msg.done {
			m.finishStream()
			return m, nil
		}
		m.appendChunk(msg.content, msg.reasoning)
		return m, m.readCmd()
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey only intercepts navigation and editing keys. Everything else,
// including letters such as q and j, is typed into the message.
func (m chatModel) handleKey(msg tea.KeyMsg) (chatModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.close()
		m.exited = true
	case tea.KeyEnter:
		return m.send()
	case tea.KeyBackspace:
		if runes := []rune(m.input); len(runes) > 0 {
			m.input = string(runes[:len(runes)-1])
		}
	case tea.KeyUp:
		m.scrollBy(-1)
	case tea.KeyDown:
		m.scrollBy(1)
	case tea.KeyPgUp:
		m.scrollBy(-m.transcriptHeight())
	case tea.KeyPgDown:
		m.scrollBy(m.transcriptHeight())
	case tea.KeyHome:
		m.follow = false
		m.scroll = 0
	case tea.KeyEnd:
		m.follow = true
		m.clampScroll()
	case tea.KeyCtrlR:
		m.cycleEffort()
	case tea.KeyCtrlN:
		m.reset()
	case tea.KeyCtrlE:
		m.dumpTranscript()
	case tea.KeySpace:
		m.input += " "
	case tea.KeyRunes:
		m.input += string(msg.Runes)
	}
	m.clampScroll()
	return m, nil
}

func (m chatModel) send() (chatModel, tea.Cmd) {
	text := strings.TrimSpace(m.input)
	if text == "" || m.streaming {
		return m, nil
	}

	m.input = ""
	m.status = ""
	m.messages = append(m.messages, assistant.Message{Role: "user", Content: text})
	m.turns = append(m.turns, chatTurn{role: "you", text: text})

	// Placeholder for the reply. It joins the API history only once the first
	// content delta arrives, so we never send an empty assistant turn.
	m.turns = append(m.turns, chatTurn{role: "assistant"})
	m.assistantStarted = false
	m.streaming = true
	m.thinking = m.cfg.Thinking
	m.follow = true
	m.clampScroll()

	m.startedAt = time.Now()
	m.contentChars = 0
	m.reasoningChars = 0
	m.chunks = 0
	m.app.log().Info("chat.send", "user message queued",
		"model", m.cfg.Model,
		"effort", m.currentEffort(),
		"chars", len(text),
		"history", len(m.messages))
	m.app.log().VerboseDebug("chat.send.body", "user message", "text", text)

	return m, m.startStreamCmd()
}

// startStreamCmd opens the streaming request. It runs off the UI goroutine and
// reports back with chatStreamReadyMsg.
func (m chatModel) startStreamCmd() tea.Cmd {
	messages := append([]assistant.Message(nil), m.messages...)
	cfg := m.cfg
	app := m.app
	effort := m.currentEffort()
	return func() tea.Msg {
		app.log().Debug("chat.request", "opening stream",
			"model", cfg.Model,
			"effort", effort,
			"thinking", cfg.Thinking,
			"messages", len(messages))
		ctx, cancel := context.WithCancel(context.Background())
		body, err := app.startChatStream(ctx, cfg, messages)
		if err != nil {
			app.log().Error("chat.request", "stream request failed", "error", err.Error())
			cancel()
			return chatStreamReadyMsg{err: err}
		}
		return chatStreamReadyMsg{reader: newSSEReader(body), closer: body, cancel: cancel}
	}
}

// readCmd reads a single stream event, then schedules itself again.
func (m chatModel) readCmd() tea.Cmd {
	reader := m.reader
	return func() tea.Msg {
		if reader == nil {
			return chatChunkMsg{done: true}
		}
		chunk, done, err := reader.Next()
		if err != nil {
			return chatChunkMsg{err: err}
		}
		if done {
			return chatChunkMsg{done: true}
		}

		var content, reasoning string
		if len(chunk.Choices) > 0 {
			content = chunk.Choices[0].Delta.Content
			reasoning = chunk.Choices[0].Delta.ReasoningContent
		}
		return chatChunkMsg{content: content, reasoning: reasoning}
	}
}

// appendChunk adds a streamed delta to the current assistant turn. Reasoning
// content is consumed silently; it only flips the "thinking" indicator.
func (m *chatModel) appendChunk(content, reasoning string) {
	if content == "" {
		if reasoning != "" {
			m.thinking = true
			m.reasoningChars += len(reasoning)
			m.chunks++
			m.app.log().VerboseDebug("chat.reasoning", "reasoning delta", "delta", reasoning)
		}
		return
	}

	m.thinking = false
	m.contentChars += len(content)
	m.chunks++
	if !m.assistantStarted {
		m.app.log().Debug("chat.first_content", "first content chunk",
			"latency_ms", time.Since(m.startedAt).Milliseconds())
		m.messages = append(m.messages, assistant.Message{Role: "assistant", Content: content})
		m.assistantStarted = true
	} else if len(m.messages) > 0 {
		m.messages[len(m.messages)-1].Content += content
	}
	m.app.log().VerboseDebug("chat.delta", "content delta", "delta", content)

	if n := len(m.turns); n > 0 && m.turns[n-1].role == "assistant" {
		m.turns[n-1].text += content
	}
	m.clampScroll()
}

// finishStream closes the response and drops an empty assistant placeholder
// left behind when a request fails before producing any content.
func (m *chatModel) finishStream() {
	wasStreaming := m.streaming
	m.streaming = false
	m.thinking = false

	if m.closer != nil {
		m.closer.Close()
		m.closer = nil
	}
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.reader = nil

	if n := len(m.turns); n > 0 && m.turns[n-1].role == "assistant" && strings.TrimSpace(m.turns[n-1].text) == "" {
		m.turns = m.turns[:n-1]
	}
	m.assistantStarted = false
	m.clampScroll()

	if wasStreaming {
		m.app.log().Info("chat.stream.done", "turn finished",
			"content_chars", m.contentChars,
			"reasoning_chars", m.reasoningChars,
			"chunks", m.chunks,
			"duration_ms", time.Since(m.startedAt).Milliseconds())
		if n := len(m.turns); n > 0 && m.turns[n-1].role == "assistant" {
			m.app.log().VerboseDebug("chat.reply.body", "assistant reply", "text", m.turns[n-1].text)
		}
	}
}

// close aborts an in-flight request. It is safe to call at any time.
func (m *chatModel) close() {
	if m.streaming {
		m.app.log().Debug("chat.cancel", "cancelled in-flight reply",
			"content_chars", m.contentChars,
			"duration_ms", time.Since(m.startedAt).Milliseconds())
	}
	m.finishStream()
}

// reset clears the conversation and aborts any in-flight reply, keeping only
// the system prompt so the chat starts fresh for the same problem. The browse
// screen rebuilds the grounded context instead; this path is the standalone
// ask screen, where the system prompt is already fixed.
func (m *chatModel) reset() {
	m.close()
	m.turns = nil
	if len(m.messages) > 0 {
		m.messages = m.messages[:1]
	}
	m.input = ""
	m.scroll = 0
	m.follow = true
	m.assistantStarted = false
	m.startedAt = time.Time{}
	m.contentChars = 0
	m.reasoningChars = 0
	m.chunks = 0
	m.status = "conversation reset"
	m.app.log().Info("chat.reset", "conversation cleared")
}

func (m chatModel) maxScroll() int {
	max := len(m.transcriptLines()) - m.transcriptHeight()
	if max < 0 {
		return 0
	}
	return max
}

func (m *chatModel) clampScroll() {
	max := m.maxScroll()
	if m.follow {
		m.scroll = max
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
	if m.scroll > max {
		m.scroll = max
	}
}

func (m *chatModel) scrollBy(delta int) {
	m.follow = false
	m.scroll += delta
	if m.scroll < 0 {
		m.scroll = 0
	}
	max := m.maxScroll()
	if m.scroll > max {
		m.scroll = max
	}
	m.follow = m.scroll >= max
}

func (m chatModel) transcriptHeight() int {
	h := m.height - 4 - len(m.helpLines()) - len(m.visibleInputLines())
	if h < 1 {
		h = 1
	}
	return h
}

// rule is the horizontal divider between the sections of the chat view, sized to
// the terminal so it never wraps on a narrow window.
func (m chatModel) rule() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	return dimStyle.Render(strings.Repeat("─", w))
}

func (m chatModel) textWidth() int {
	w := m.width - 4
	if w < 10 {
		w = 10
	}
	return w
}

// inputWidth is the number of runes that fit on one input line after the "> "
// prompt.
func (m chatModel) inputWidth() int {
	w := m.textWidth() - 2
	if w < 1 {
		w = 1
	}
	return w
}

// inputLines wraps the message input to the view width. The first line carries
// the "> " prompt; continuation lines are indented to line up with it. The
// cursor block trails the last line unless a reply is streaming.
func (m chatModel) inputLines() []string {
	wrapped := wrapText(m.input, m.inputWidth())
	last := len(wrapped) - 1
	lines := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		prefix := "  "
		if i == 0 {
			prefix = "> "
		}
		text := prefix + line
		if i == last && !m.streaming {
			text += "█"
		}
		lines = append(lines, accentStyle.Render(text))
	}
	return lines
}

// visibleInputLines caps the input to the lines that fit on screen. When the
// input is taller than that, the tail is shown so the cursor stays visible.
func (m chatModel) visibleInputLines() []string {
	lines := m.inputLines()
	max := m.height - 8
	if max < 1 {
		max = 1
	}
	if len(lines) > max {
		lines = lines[len(lines)-max:]
	}
	return lines
}

// transcriptLines renders the conversation into display lines, wrapping long
// lines to the view width.
func (m chatModel) transcriptLines() []string {
	if len(m.turns) == 0 {
		return []string{dimStyle.Render("Ask a question about this problem."), ""}
	}

	wrapWidth := m.textWidth() - 2
	var lines []string

	for i, turn := range m.turns {
		if turn.role == "you" {
			lines = append(lines, accentStyle.Render("you"))
		} else {
			lines = append(lines, acceptedStyle.Render("assistant"))
		}

		text := turn.text
		if text == "" && turn.role == "assistant" && m.streaming && i == len(m.turns)-1 {
			if m.thinking {
				text = "thinking…"
			} else {
				text = "…"
			}
		}
		for _, line := range wrapText(text, wrapWidth) {
			lines = append(lines, "  "+line)
		}
		lines = append(lines, "")
	}

	return lines
}

// wrapText breaks text into lines of at most width runes, preserving explicit
// newlines. Lines are broken at spaces so words stay whole; a word wider than a
// full line is split only as a last resort.
func wrapText(text string, width int) []string {
	if width < 1 {
		width = 1
	}

	var lines []string
	for _, raw := range strings.Split(text, "\n") {
		lines = append(lines, wrapLine(raw, width)...)
	}
	return lines
}

// wrapLine wraps one line at spaces, keeping leading indentation on the first
// output line. A single word longer than width is hard-split.
func wrapLine(line string, width int) []string {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return []string{""}
	}
	indent := line[:len(line)-len(trimmed)]

	var lines []string
	current := indent
	currentWidth := runeLen(indent)
	hasWord := false

	for _, word := range strings.Fields(trimmed) {
		wordWidth := runeLen(word)

		// A word wider than a full line has to be split somewhere.
		for wordWidth > width {
			if hasWord {
				lines = append(lines, current)
				current, currentWidth, hasWord = "", 0, false
			}
			runes := []rune(word)
			lines = append(lines, string(runes[:width]))
			word = string(runes[width:])
			wordWidth = runeLen(word)
		}

		sep := 0
		if hasWord {
			sep = 1
		}
		if hasWord && currentWidth+sep+wordWidth > width {
			lines = append(lines, current)
			current, currentWidth, hasWord = "", 0, false
			sep = 0
		}
		if hasWord {
			current += " "
			currentWidth++
		}
		current += word
		currentWidth += wordWidth
		hasWord = true
	}
	lines = append(lines, current)

	// Safety net: split anything still too wide, such as a remainder left next
	// to leading indentation.
	var clamped []string
	for _, l := range lines {
		for runeLen(l) > width {
			runes := []rune(l)
			clamped = append(clamped, string(runes[:width]))
			l = string(runes[width:])
		}
		clamped = append(clamped, l)
	}
	return clamped
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

// dumpTranscript writes the current conversation to a timestamped markdown file
// under chatLogsDir and reports the path in the status line.
func (m *chatModel) dumpTranscript() {
	if len(m.turns) == 0 {
		m.status = "Nothing to export yet."
		return
	}

	var b strings.Builder
	b.WriteString("# " + m.title + "\n\n")
	b.WriteString("Exported " + time.Now().Format(time.RFC3339) + "\n\n")
	for _, turn := range m.turns {
		if strings.TrimSpace(turn.text) == "" {
			continue
		}
		b.WriteString("## " + turn.role + "\n\n")
		b.WriteString(turn.text + "\n\n")
	}

	name := chatFileName(m.title) + ".md"
	if err := os.MkdirAll(chatLogsDir, dirPerm); err != nil {
		m.app.log().Error("chat.export", "mkdir failed", "dir", chatLogsDir, "error", err.Error())
		m.status = "Export failed: " + err.Error()
		return
	}
	path := filepath.Join(chatLogsDir, name)
	if err := os.WriteFile(path, []byte(b.String()), filePerm); err != nil {
		m.app.log().Error("chat.export", "write failed", "path", path, "error", err.Error())
		m.status = "Export failed: " + err.Error()
		return
	}
	m.app.log().Info("chat.export", "conversation exported", "path", path, "turns", len(m.turns))
	m.status = "Exported to " + filepath.ToSlash(path)
}

// chatFileName turns a problem title such as "1. Two Sum" into a filename-safe
// slug like "1-two-sum".
func chatFileName(title string) string {
	lower := strings.ToLower(title)
	var b strings.Builder
	prevDash := false
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func (m chatModel) statusLine() string {
	if m.status != "" {
		return m.status
	}
	if m.streaming && !m.assistantStarted {
		return "thinking…"
	}
	return ""
}

func (m chatModel) helpText() string {
	if m.streaming {
		return "streaming… · ctrl+r effort · ctrl+n reset · ctrl+e export · esc cancel · ↑/↓"
	}
	return "enter send · ctrl+r effort · ctrl+n reset · ctrl+e export · esc back · ↑/↓"
}

// helpLines word-wraps the footer help to the view width, so a narrow window
// keeps the fixed-height layout intact instead of overflowing. It is capped so
// the footer cannot crowd out the transcript on a very short window.
func (m chatModel) helpLines() []string {
	lines := wrapText(m.helpText(), m.textWidth())
	const maxHelpLines = 3
	if len(lines) > maxHelpLines {
		lines = lines[:maxHelpLines]
	}
	return lines
}

func (m chatModel) View() string {
	var b strings.Builder

	header := titleStyle.Render(m.title) +
		dimStyle.Render(fmt.Sprintf("  ·  assistant (%s · effort: %s)", m.cfg.Model, m.currentEffort()))
	b.WriteString(clipStyled(header, m.width))
	b.WriteString("\n")
	b.WriteString(m.rule())
	b.WriteString("\n")

	height := m.transcriptHeight()
	lines := m.transcriptLines()
	end := m.scroll + height
	if end > len(lines) {
		end = len(lines)
	}
	for i := m.scroll; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}
	for i := end - m.scroll; i < height; i++ {
		b.WriteString("\n")
	}

	b.WriteString(m.rule())
	b.WriteString("\n")
	for _, line := range m.visibleInputLines() {
		b.WriteString(line)
		b.WriteString("\n")
	}

	if status := m.statusLine(); status != "" {
		b.WriteString(dimStyle.Render(fitLine(status, m.textWidth())))
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render(strings.Join(m.helpLines(), "\n")))

	return b.String()
}
