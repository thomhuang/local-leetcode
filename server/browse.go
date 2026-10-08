package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomhuang/local-leetcode/internal/assistant"
	"github.com/thomhuang/local-leetcode/internal/question"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	acceptedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	writtenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	todoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	premiumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("135"))
	missingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type browseMode int

const (
	modeList browseMode = iota
	modeDetail
	modeText
	modeChat
)

type browseRow struct {
	category string
	problem  question.ProblemSetProblem
	state    problemState
}

// browseModel is the bubbletea model behind the "Browse problems" action.
type browseModel struct {
	app     *App
	set     question.ProblemSet
	rows    []browseRow
	visible []int

	cursor   int
	offset   int
	category string
	status   string

	width, height int
	mode          browseMode
	detailIndex   int
	textTitle     string
	textLines     []string
	textOffset    int
	textImages    []string
	lastResult    string
	busy          bool
	message       string

	chat chatModel
}

type testResultMsg struct {
	text string
	err  error
}

type submitResultMsg struct {
	text string
	err  error
}

type statementMsg struct {
	title string
	text  string
	err   error
}

type importResultMsg struct {
	message string
}

func newBrowseModel(app *App, set question.ProblemSet) browseModel {
	model := browseModel{app: app, set: set, height: 30, width: 80}
	model.reload()
	return model
}

func (m *browseModel) reload() {
	m.rows = m.rows[:0]
	for _, category := range m.set.Categories {
		for _, problem := range category.Problems {
			m.rows = append(m.rows, browseRow{
				category: category.Name,
				problem:  problem,
				state:    m.app.problemState(problem),
			})
		}
	}
	m.applyFilter()
}

// refresh recomputes problem states after an action such as submit or import.
func (m *browseModel) refresh() {
	for i := range m.rows {
		m.rows[i].state = m.app.problemState(m.rows[i].problem)
	}
	m.applyFilter()
}

func (m *browseModel) applyFilter() {
	m.visible = m.visible[:0]
	for i, row := range m.rows {
		if m.category != "" && row.category != m.category {
			continue
		}
		if m.status != "" && row.state.StatusKey() != m.status {
			continue
		}
		m.visible = append(m.visible, i)
	}
	m.move(0)
}

func (m browseModel) Init() tea.Cmd {
	return nil
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.mode == modeChat {
			m.chat, _ = m.chat.Update(msg)
		}
		m.clampOffset()
		m.clampTextOffset()
		return m, nil
	case chatStreamReadyMsg, chatChunkMsg:
		if m.mode != modeChat {
			return m, nil
		}
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		if m.chat.exited {
			m.chat.exited = false
			m.mode = modeDetail
		}
		return m, cmd
	case testResultMsg:
		m.busy = false
		m.lastResult = m.renderResult(msg.text, msg.err)
		m.showText("Test result", m.lastResult)
		return m, nil
	case submitResultMsg:
		m.busy = false
		m.refresh()
		m.lastResult = m.renderResult(msg.text, msg.err)
		m.showText("Submission result", m.lastResult)
		return m, nil
	case statementMsg:
		m.busy = false
		m.showStatement(msg.title, msg.text, msg.err)
		return m, nil
	case importResultMsg:
		m.busy = false
		m.refresh()
		m.message = msg.message
		return m, nil
	case tea.KeyMsg:
		if m.mode == modeChat {
			var cmd tea.Cmd
			m.chat, cmd = m.chat.Update(msg)
			if m.chat.exited {
				m.chat.exited = false
				m.mode = modeDetail
			}
			return m, cmd
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m browseModel) renderResult(text string, err error) string {
	if err != nil {
		if text != "" {
			return text + "\n\nError: " + err.Error()
		}
		return "Error: " + err.Error()
	}
	return text
}

func (m browseModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeText:
		return m.handleTextKey(msg)
	case modeDetail:
		return m.handleDetailKey(msg)
	default:
		return m.handleListKey(msg)
	}
}

func (m browseModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, func() tea.Msg { return browseExitMsg{} }
	case "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "pgup":
		m.move(-m.listViewHeight())
	case "pgdown":
		m.move(m.listViewHeight())
	case "home":
		m.cursor = 0
		m.clampOffset()
	case "end":
		m.cursor = len(m.visible) - 1
		m.clampOffset()
	case "enter":
		if len(m.visible) > 0 {
			m.detailIndex = m.visible[m.cursor]
			m.mode = modeDetail
			m.message = ""
			m.lastResult = ""
		}
	case "c":
		m.cycleCategory()
	case "s":
		m.cycleStatus()
	case "n":
		m.nextTodo()
	}
	return m, nil
}

func (m browseModel) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+c":
		m.mode = modeList
		m.message = ""
	case "t":
		if !m.busy {
			m.busy = true
			m.message = ""
			return m, m.testCmd()
		}
	case "s":
		if !m.busy {
			m.busy = true
			m.message = ""
			return m, m.submitCmd()
		}
	case "v":
		if !m.busy {
			m.busy = true
			m.message = ""
			return m, m.statementCmd()
		}
	case "i":
		if !m.busy {
			m.busy = true
			m.message = ""
			return m, m.importCmd()
		}
	case "o":
		m.openFile()
	case "a":
		if !m.busy {
			m.openChat()
		}
	case "m":
		m.toggleAccepted()
	}
	return m, nil
}

func (m browseModel) handleTextKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+c":
		m.mode = modeDetail
	case "up", "k":
		m.textOffset--
	case "down", "j":
		m.textOffset++
	case "pgup":
		m.textOffset -= m.textViewHeight()
	case "pgdown":
		m.textOffset += m.textViewHeight()
	case "home":
		m.textOffset = 0
	case "end":
		m.textOffset = len(m.textLines)
	case "i":
		m.openImages()
	}
	m.clampTextOffset()
	return m, nil
}

func (m *browseModel) move(delta int) {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	m.clampOffset()
}

func (m browseModel) listViewHeight() int {
	h := m.height - 8
	if h < 3 {
		h = 3
	}
	return h
}

func (m browseModel) textViewHeight() int {
	h := m.height - 4
	if h < 3 {
		h = 3
	}
	return h
}

func (m *browseModel) clampOffset() {
	if len(m.visible) == 0 {
		m.offset = 0
		return
	}
	height := m.listViewHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+height {
		m.offset = m.cursor - height + 1
	}
	maxOffset := len(m.visible) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *browseModel) clampTextOffset() {
	maxOffset := len(m.textLines) - m.textViewHeight()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.textOffset < 0 {
		m.textOffset = 0
	}
	if m.textOffset > maxOffset {
		m.textOffset = maxOffset
	}
}

func (m *browseModel) cycleCategory() {
	options := append([]string{""}, categoryNames(m.set)...)
	m.category = options[(indexOf(options, m.category)+1)%len(options)]
	m.cursor, m.offset = 0, 0
	m.applyFilter()
}

func (m *browseModel) cycleStatus() {
	options := []string{"", "todo", "written", "accepted"}
	m.status = options[(indexOf(options, m.status)+1)%len(options)]
	m.cursor, m.offset = 0, 0
	m.applyFilter()
}

func (m *browseModel) nextTodo() {
	if len(m.visible) == 0 {
		return
	}
	for i := 1; i <= len(m.visible); i++ {
		pos := (m.cursor + i) % len(m.visible)
		if m.rows[m.visible[pos]].state == stateTodo {
			m.cursor = pos
			m.clampOffset()
			return
		}
	}
}

// openFile launches the problem's solution file in the default application.
func (m *browseModel) openFile() {
	row := m.rows[m.detailIndex]
	starter := starterFilePath(problemsDir, row.problem)
	if starter == "" {
		m.app.log().Warn("file.open", "no starter file", "slug", row.problem.Slug)
		m.message = "No .go file yet - press i to import."
		return
	}
	if err := openInDefaultApp(starter); err != nil {
		m.app.log().Error("file.open", "open failed", "path", starter, "error", err.Error())
		m.message = "Unable to open file: " + err.Error()
		return
	}
	m.app.log().Debug("file.open", "opened file", "path", starter)
	m.message = "Opened " + starter
}

func (m *browseModel) toggleAccepted() {
	row := m.rows[m.detailIndex]
	title := row.problem.Title
	accepted := row.state != stateAccepted

	if err := m.app.setAccepted(row.problem.Slug, accepted); err != nil {
		m.app.log().Error("progress.save", "save failed", "slug", row.problem.Slug, "error", err.Error())
		m.message = "Unable to save progress: " + err.Error()
		return
	}
	m.refresh()

	m.app.log().Info("progress.mark", "accepted toggled", "slug", row.problem.Slug, "accepted", accepted)

	if accepted {
		m.message = "Marked accepted: " + title
	} else {
		m.message = "Cleared accepted: " + title
	}
}

// openChat starts an assistant session grounded in the current problem.
func (m *browseModel) openChat() {
	if !hasAssistantKey() {
		m.app.log().Warn("assistant.open", "missing API key", "env", deepseekAPIKeyEnv)
		m.message = "Set " + deepseekAPIKeyEnv + " to use the assistant."
		return
	}

	row := m.rows[m.detailIndex]
	cfg, prompt, err := assistant.Load(assistantDir)
	if err != nil {
		m.app.log().Error("assistant.open", "load config failed", "error", err.Error())
		m.message = "Unable to load assistant config: " + err.Error()
		return
	}

	system := strings.TrimRight(prompt, "\n")
	if context := buildProblemContext(row.problem, m.lastResult); strings.TrimSpace(context) != "" {
		system += "\n\n" + context
	}

	m.app.log().Info("assistant.open", "chat session started",
		"problem", row.problem.ID,
		"title", row.problem.Title,
		"model", cfg.Model,
		"system_chars", len(system))
	m.app.log().VerboseDebug("assistant.system", "system prompt", "text", system)

	m.chat = newChatModel(m.app, cfg, system, fmt.Sprintf("%d. %s", row.problem.ID, row.problem.Title), m.width, m.height)
	m.mode = modeChat
	m.message = ""
}

func (m *browseModel) showText(title, body string) {
	m.textTitle = title
	body = strings.ReplaceAll(body, "\r\n", "\n")
	m.textLines = strings.Split(body, "\n")
	m.textOffset = 0
	m.textImages = nil
	m.mode = modeText
}

// showStatement renders a problem's Markdown statement for the text view. On a
// render failure it falls back to showing the raw Markdown so the content is
// never lost.
func (m *browseModel) showStatement(title, markdown string, err error) {
	images := extractImageURLs(markdown)
	if err != nil {
		m.showText(title, m.renderResult(markdown, err))
		m.textImages = images
		return
	}
	rendered, renderErr := renderMarkdown(markdown, termWidth(m.width))
	if renderErr != nil {
		m.app.log().Warn("statement.render", "markdown render failed", "error", renderErr.Error())
		m.showText(title, markdown)
		m.textImages = images
		return
	}
	m.showText(title, rendered)
	m.textImages = images
}

// imageLinkRe matches Markdown image links and captures their URL. LeetCode
// statements store diagrams as `![](url)` with no alt text.
var imageLinkRe = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`)

// extractImageURLs returns the deduplicated image URLs referenced by a Markdown
// document, in order of first appearance.
func extractImageURLs(markdown string) []string {
	matches := imageLinkRe.FindAllStringSubmatch(markdown, -1)
	seen := make(map[string]bool, len(matches))
	urls := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 && m[1] != "" && !seen[m[1]] {
			seen[m[1]] = true
			urls = append(urls, m[1])
		}
	}
	return urls
}

// openImages opens every image referenced by the current statement in the
// default application (the browser, for these URLs).
func (m *browseModel) openImages() {
	if len(m.textImages) == 0 {
		return
	}
	for _, url := range m.textImages {
		if err := openInDefaultApp(url); err != nil {
			m.app.log().Error("statement.image", "open failed", "url", url, "error", err.Error())
		}
	}
	m.app.log().Info("statement.image", "opened images", "count", len(m.textImages))
}

// renderMarkdown renders Markdown to ANSI-styled terminal output, word-wrapped
// to width, using glamour's auto-detected style.
func renderMarkdown(markdown string, width int) (string, error) {
	if width <= 0 {
		width = 80
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}
	defer renderer.Close()
	rendered, err := renderer.Render(markdown)
	if err != nil {
		return "", err
	}
	return strings.Trim(rendered, "\n"), nil
}

func (m browseModel) testCmd() tea.Cmd {
	id := m.rows[m.detailIndex].problem.ID
	return func() tea.Msg {
		text, err := m.app.TestProblem(id)
		return testResultMsg{text: text, err: err}
	}
}

func (m browseModel) submitCmd() tea.Cmd {
	id := m.rows[m.detailIndex].problem.ID
	return func() tea.Msg {
		text, err := m.app.SubmitProblem(id)
		return submitResultMsg{text: text, err: err}
	}
}

func (m browseModel) statementCmd() tea.Cmd {
	row := m.rows[m.detailIndex]
	title := fmt.Sprintf("%d. %s", row.problem.ID, row.problem.Title)
	path := problemsDir + "/" + row.problem.Slug + "/" + row.problem.Slug + ".md"
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			m.app.log().Error("statement.read", "read failed", "path", path, "error", err.Error())
		}
		return statementMsg{title: title, text: string(data), err: err}
	}
}

func (m browseModel) importCmd() tea.Cmd {
	problem := m.rows[m.detailIndex].problem
	return func() tea.Msg {
		imported, premium, err := m.app.ImportProblem(problem)
		switch {
		case err != nil:
			m.app.log().Error("problem.import", "import failed", "slug", problem.Slug, "error", err.Error())
			return importResultMsg{message: "Import failed: " + err.Error()}
		case !imported:
			m.app.log().Debug("problem.import", "already imported", "slug", problem.Slug)
			return importResultMsg{message: "Already imported: " + problem.Title}
		case premium:
			m.app.log().Info("problem.import", "premium placeholder created", "slug", problem.Slug)
			return importResultMsg{message: "Premium placeholder created: " + problem.Title}
		default:
			m.app.log().Info("problem.import", "problem imported", "slug", problem.Slug)
			return importResultMsg{message: "Imported: " + problem.Title}
		}
	}
}

func (m browseModel) View() string {
	switch m.mode {
	case modeText:
		return m.textView()
	case modeDetail:
		return m.detailView()
	case modeChat:
		return m.chat.View()
	default:
		return m.listView()
	}
}

func (m browseModel) listView() string {
	var b strings.Builder

	width := termWidth(m.width)
	barWidth := 32
	if width < barWidth {
		barWidth = width
	}

	accepted, total := 0, len(m.rows)
	for _, row := range m.rows {
		if row.state == stateAccepted {
			accepted++
		}
	}

	b.WriteString(titleStyle.Render(m.set.Name))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  ·  %d/%d accepted", accepted, total)))
	b.WriteString("\n")
	b.WriteString(progressBar(accepted, total, barWidth))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	category := "all"
	if m.category != "" {
		category = m.category
	}
	status := "all"
	if m.status != "" {
		status = m.status
	}
	b.WriteString(clipStyled(dimStyle.Render(fmt.Sprintf("category: %s (c)   status: %s (s)   showing %d", category, status, len(m.visible))), width))
	b.WriteString("\n\n")

	height := m.listViewHeight()
	end := m.offset + height
	if end > len(m.visible) {
		end = len(m.visible)
	}
	showCategory := m.category == ""
	for pos := m.offset; pos < end; pos++ {
		row := m.rows[m.visible[pos]]
		line := fmt.Sprintf("%s %s", coloredSymbol(row.state), problemRowText(row.problem, row.category, showCategory))
		if pos == m.cursor {
			line = cursorStyle.Render("> ") + line
		} else {
			line = "  " + line
		}
		b.WriteString(clipStyled(line, width))
		b.WriteString("\n")
	}
	for pos := end - m.offset; pos < height; pos++ {
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n" + clipStyled(dimStyle.Render(m.message), width) + "\n")
	} else {
		b.WriteString("\n")
	}

	footer := "↑/↓ move · enter open · n next todo · c category · s status · q back"
	if m.busy {
		footer += "   working..."
	}
	b.WriteString(clipStyled(dimStyle.Render(footer), width))
	return b.String()
}

func (m browseModel) detailView() string {
	row := m.rows[m.detailIndex]
	width := termWidth(m.width)

	var b strings.Builder
	b.WriteString(clipStyled(titleStyle.Render(fmt.Sprintf("%d. %s", row.problem.ID, row.problem.Title)), width))
	b.WriteString("\n")
	b.WriteString(clipStyled(dimStyle.Render(fmt.Sprintf("%s · %s", row.problem.Difficulty, row.category)), width))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("status: %s %s\n", coloredSymbol(row.state), row.state.Label()))
	if starter := starterFilePath(problemsDir, row.problem); starter != "" {
		b.WriteString("file:   " + starter + "\n")
	} else {
		b.WriteString("file:   (not imported)\n")
	}

	b.WriteString("\n")
	mark := "m mark accepted"
	if row.state == stateAccepted {
		mark = "m clear accepted"
	}
	b.WriteString(clipStyled(dimStyle.Render(strings.Join([]string{"t test", "s submit", "v statement", "o open file", "a assistant", "i import", mark, "esc back"}, " · ")), width))

	if m.busy {
		b.WriteString("\n\n" + dimStyle.Render("working..."))
	}
	if m.message != "" {
		b.WriteString("\n\n" + dimStyle.Render(m.message))
	}
	return b.String()
}

func (m browseModel) textView() string {
	var b strings.Builder
	width := termWidth(m.width)
	b.WriteString(clipStyled(titleStyle.Render(m.textTitle), width))
	b.WriteString("\n\n")

	height := m.textViewHeight()
	end := m.textOffset + height
	if end > len(m.textLines) {
		end = len(m.textLines)
	}
	for i := m.textOffset; i < end; i++ {
		b.WriteString(m.textLines[i])
		b.WriteString("\n")
	}
	b.WriteString("\n")
	footer := "↑/↓ scroll · esc back"
	if len(m.textImages) > 0 {
		footer = fmt.Sprintf("↑/↓ scroll · i open images (%d) · esc back", len(m.textImages))
	}
	b.WriteString(clipStyled(dimStyle.Render(footer), width))
	return b.String()
}

func coloredSymbol(state problemState) string {
	switch state {
	case stateAccepted:
		return acceptedStyle.Render("[x]")
	case stateWritten:
		return writtenStyle.Render("[-]")
	case statePremium:
		return premiumStyle.Render("[p]")
	case stateMissing:
		return missingStyle.Render("[?]")
	default:
		return todoStyle.Render("[ ]")
	}
}

func indexOf(values []string, want string) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}
