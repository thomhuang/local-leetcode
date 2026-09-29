package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	busy          bool
	message       string
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
		m.clampOffset()
		return m, nil
	case testResultMsg:
		m.busy = false
		m.showText("Test result", m.renderResult(msg.text, msg.err))
		return m, nil
	case submitResultMsg:
		m.busy = false
		m.refresh()
		m.showText("Submission result", m.renderResult(msg.text, msg.err))
		return m, nil
	case statementMsg:
		m.busy = false
		m.showText(msg.title, m.renderResult(msg.text, msg.err))
		return m, nil
	case importResultMsg:
		m.busy = false
		m.refresh()
		m.message = msg.message
		return m, nil
	case tea.KeyMsg:
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
	case "q", "esc", "ctrl+c":
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
	h := m.height - 6
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

func (m *browseModel) toggleAccepted() {
	row := m.rows[m.detailIndex]
	title := row.problem.Title
	accepted := row.state != stateAccepted

	if err := m.app.setAccepted(row.problem.Slug, accepted); err != nil {
		m.message = "Unable to save progress: " + err.Error()
		return
	}
	m.refresh()

	if accepted {
		m.message = "Marked accepted: " + title
	} else {
		m.message = "Cleared accepted: " + title
	}
}

func (m *browseModel) showText(title, body string) {
	m.textTitle = title
	body = strings.ReplaceAll(body, "\r\n", "\n")
	m.textLines = strings.Split(body, "\n")
	m.textOffset = 0
	m.mode = modeText
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
		return statementMsg{title: title, text: string(data), err: err}
	}
}

func (m browseModel) importCmd() tea.Cmd {
	problem := m.rows[m.detailIndex].problem
	return func() tea.Msg {
		imported, premium, err := m.app.ImportProblem(problem)
		switch {
		case err != nil:
			return importResultMsg{message: "Import failed: " + err.Error()}
		case !imported:
			return importResultMsg{message: "Already imported: " + problem.Title}
		case premium:
			return importResultMsg{message: "Premium placeholder created: " + problem.Title}
		default:
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
	default:
		return m.listView()
	}
}

func (m browseModel) listView() string {
	var b strings.Builder

	accepted, total := 0, len(m.rows)
	for _, row := range m.rows {
		if row.state == stateAccepted {
			accepted++
		}
	}

	b.WriteString(titleStyle.Render(m.set.Name))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  ·  %d/%d accepted", accepted, total)))
	b.WriteString("\n")

	category := "all"
	if m.category != "" {
		category = m.category
	}
	status := "all"
	if m.status != "" {
		status = m.status
	}
	b.WriteString(dimStyle.Render(fmt.Sprintf("category: %s (c)   status: %s (s)   showing %d", category, status, len(m.visible))))
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
		b.WriteString(line)
		b.WriteString("\n")
	}
	for pos := end - m.offset; pos < height; pos++ {
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n" + dimStyle.Render(m.message) + "\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render("↑/↓ move · enter open · n next todo · c category · s status · q quit"))
	if m.busy {
		b.WriteString(dimStyle.Render("   working..."))
	}
	return b.String()
}

func (m browseModel) detailView() string {
	row := m.rows[m.detailIndex]

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("%d. %s", row.problem.ID, row.problem.Title)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("%s · %s", row.problem.Difficulty, row.category)))
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
	b.WriteString(dimStyle.Render(strings.Join([]string{"t test", "s submit", "v statement", "i import", mark, "esc back"}, " · ")))

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
	b.WriteString(titleStyle.Render(m.textTitle))
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
	b.WriteString(dimStyle.Render("↑/↓ scroll · esc back"))
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

// BrowseProblems opens the checklist TUI for a problem set.
func (app *App) BrowseProblems(setName string) error {
	set, err := app.LoadProblemSet(setName)
	if err != nil {
		return err
	}
	if len(set.Problems()) == 0 {
		return fmt.Errorf("problem set %q has no problems", setName)
	}

	program := tea.NewProgram(newBrowseModel(app, set), tea.WithAltScreen())
	_, err = program.Run()
	return err
}
