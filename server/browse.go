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
	modeFilterStatus
	modeFilterCategory
	modeVariant
	modeVariantName
)

var filterStates = []problemState{
	stateTodo,
	stateWritten,
	stateAccepted,
	statePremium,
	stateMissing,
}

type filterOption struct {
	label   string
	checked bool
	count   int
}

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

	cursor       int
	offset       int
	categories   map[string]bool
	statuses     map[problemState]bool
	filterCursor int
	filterOffset int

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
	confirmReset  bool

	chat     chatModel
	chatSlug string

	variants      []solutionVariant
	variantCursor int
	variantOffset int
	variantInput  string
	confirmDelete bool
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

type resetResultMsg struct {
	message string
	err     error
}

type variantResultMsg struct {
	message string
	err     error
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
		if !m.matchesCategory(row.category) {
			continue
		}
		if !m.matchesStatus(row.state) {
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
		m.clampFilterOffset()
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
	case resetResultMsg:
		m.busy = false
		m.confirmReset = false
		m.refresh()
		if msg.err != nil {
			m.message = "Reset failed: " + msg.err.Error()
		} else {
			m.message = msg.message
		}
		return m, nil
	case variantResultMsg:
		m.busy = false
		m.openVariantPicker()
		if msg.err != nil {
			m.message = "Create variant failed: " + msg.err.Error()
		} else {
			m.message = msg.message
		}
		return m, nil
	case tea.KeyMsg:
		if m.mode == modeChat {
			if msg.String() == chatResetKey {
				m.resetChat()
				return m, nil
			}
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
	case modeFilterStatus, modeFilterCategory:
		return m.handleFilterKey(msg)
	case modeVariant:
		return m.handleVariantKey(msg)
	case modeVariantName:
		return m.handleVariantNameKey(msg)
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
		m.openCategoryFilter()
	case "s":
		m.openStatusFilter()
	case "n":
		m.nextTodo()
	}
	return m, nil
}

func (m browseModel) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmReset {
		switch msg.String() {
		case "y":
			m.confirmReset = false
			m.message = ""
			if !m.busy {
				m.busy = true
				return m, m.resetCmd()
			}
		case "n", "esc", "q", "ctrl+c":
			m.confirmReset = false
			m.message = ""
		}
		return m, nil
	}

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
	case "x":
		m.openVariantPicker()
	case "a":
		if !m.busy {
			m.openChat()
		}
	case "m":
		m.toggleAccepted()
	case "r":
		if !m.busy && starterFilePath(problemsDir, m.rows[m.detailIndex].problem) != "" {
			m.confirmReset = true
			m.message = ""
		}
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

func (m *browseModel) openCategoryFilter() {
	m.mode = modeFilterCategory
	m.filterCursor = 0
	for i, name := range categoryNames(m.set) {
		if m.categories[name] {
			m.filterCursor = i
			break
		}
	}
	m.filterOffset = 0
	m.clampFilterOffset()
	m.message = ""
}

func (m *browseModel) openStatusFilter() {
	m.mode = modeFilterStatus
	m.filterCursor = 0
	for i, state := range filterStates {
		if m.statuses[state] {
			m.filterCursor = i
			break
		}
	}
	m.filterOffset = 0
	m.clampFilterOffset()
	m.message = ""
}

func (m browseModel) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := m.filterOptionCount()
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = modeList
	case "ctrl+c":
		return m, tea.Quit
	case "s":
		if m.mode == modeFilterStatus {
			m.mode = modeList
		} else {
			m.openStatusFilter()
		}
	case "c":
		if m.mode == modeFilterCategory {
			m.mode = modeList
		} else {
			m.openCategoryFilter()
		}
	case "up", "k":
		m.moveFilter(-1)
	case "down", "j":
		m.moveFilter(1)
	case "pgup":
		m.moveFilter(-m.filterViewHeight())
	case "pgdown":
		m.moveFilter(m.filterViewHeight())
	case "home":
		m.filterCursor = 0
		m.clampFilterOffset()
	case "end":
		if count > 0 {
			m.filterCursor = count - 1
		}
		m.clampFilterOffset()
	case " ", "space", "x":
		m.toggleFilterOption()
	case "a":
		m.toggleAllFilterOptions()
	}
	return m, nil
}

func (m *browseModel) moveFilter(delta int) {
	count := m.filterOptionCount()
	if count == 0 {
		m.filterCursor = 0
		m.filterOffset = 0
		return
	}
	m.filterCursor += delta
	if m.filterCursor < 0 {
		m.filterCursor = 0
	}
	if m.filterCursor >= count {
		m.filterCursor = count - 1
	}
	m.clampFilterOffset()
}

func (m browseModel) filterViewHeight() int {
	h := m.height - 6
	if h < 3 {
		h = 3
	}
	return h
}

func (m *browseModel) clampFilterOffset() {
	count := m.filterOptionCount()
	if count == 0 {
		m.filterOffset = 0
		return
	}
	height := m.filterViewHeight()
	if m.filterCursor < m.filterOffset {
		m.filterOffset = m.filterCursor
	}
	if m.filterCursor >= m.filterOffset+height {
		m.filterOffset = m.filterCursor - height + 1
	}
	maxOffset := count - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.filterOffset > maxOffset {
		m.filterOffset = maxOffset
	}
	if m.filterOffset < 0 {
		m.filterOffset = 0
	}
}

func (m browseModel) filterOptionCount() int {
	if m.mode == modeFilterStatus {
		return len(filterStates)
	}
	return len(m.set.Categories)
}

func (m browseModel) filterOptions() []filterOption {
	if m.mode == modeFilterStatus {
		counts := make(map[problemState]int, len(filterStates))
		for _, row := range m.rows {
			if m.matchesCategory(row.category) {
				counts[row.state]++
			}
		}
		options := make([]filterOption, 0, len(filterStates))
		for _, state := range filterStates {
			options = append(options, filterOption{
				label:   state.Label(),
				checked: m.statuses[state],
				count:   counts[state],
			})
		}
		return options
	}

	counts := make(map[string]int, len(m.set.Categories))
	for _, row := range m.rows {
		if m.matchesStatus(row.state) {
			counts[row.category]++
		}
	}
	options := make([]filterOption, 0, len(m.set.Categories))
	for _, cat := range m.set.Categories {
		options = append(options, filterOption{
			label:   cat.Name,
			checked: m.categories[cat.Name],
			count:   counts[cat.Name],
		})
	}
	return options
}

func (m *browseModel) setCategoryFlag(category string, on bool) {
	if !on {
		delete(m.categories, category)
		return
	}
	if m.categories == nil {
		m.categories = make(map[string]bool)
	}
	m.categories[category] = true
}

func (m *browseModel) setStatusFlag(state problemState, on bool) {
	if !on {
		delete(m.statuses, state)
		return
	}
	if m.statuses == nil {
		m.statuses = make(map[problemState]bool)
	}
	m.statuses[state] = true
}

func (m *browseModel) toggleFilterOption() {
	if m.mode == modeFilterStatus {
		if m.filterCursor < 0 || m.filterCursor >= len(filterStates) {
			return
		}
		state := filterStates[m.filterCursor]
		next := !m.statuses[state]
		m.setStatusFlag(state, next)
		m.app.log().Debug("browse.filter", "status flag toggled", "status", state.StatusKey(), "enabled", next)
	} else {
		names := categoryNames(m.set)
		if m.filterCursor < 0 || m.filterCursor >= len(names) {
			return
		}
		name := names[m.filterCursor]
		next := !m.categories[name]
		m.setCategoryFlag(name, next)
		m.app.log().Debug("browse.filter", "category flag toggled", "category", name, "enabled", next)
	}
	m.cursor, m.offset = 0, 0
	m.applyFilter()
}

func (m *browseModel) toggleAllFilterOptions() {
	if m.mode == modeFilterStatus {
		if m.hasStatusFilter() {
			m.statuses = nil
		} else {
			m.statuses = make(map[problemState]bool, len(filterStates))
			for _, state := range filterStates {
				m.statuses[state] = true
			}
		}
	} else {
		if m.hasCategoryFilter() {
			m.categories = nil
		} else {
			names := categoryNames(m.set)
			m.categories = make(map[string]bool, len(names))
			for _, name := range names {
				m.categories[name] = true
			}
		}
	}
	m.cursor, m.offset = 0, 0
	m.applyFilter()
}

func (m browseModel) hasCategoryFilter() bool {
	for _, on := range m.categories {
		if on {
			return true
		}
	}
	return false
}

func (m browseModel) hasStatusFilter() bool {
	for _, on := range m.statuses {
		if on {
			return true
		}
	}
	return false
}

func (m browseModel) matchesCategory(category string) bool {
	if !m.hasCategoryFilter() {
		return true
	}
	return m.categories[category]
}

func (m browseModel) matchesStatus(state problemState) bool {
	if !m.hasStatusFilter() {
		return true
	}
	return m.statuses[state]
}

func (m browseModel) activeCategories() []string {
	var selected []string
	for _, cat := range m.set.Categories {
		if m.categories[cat.Name] {
			selected = append(selected, cat.Name)
		}
	}
	return selected
}

func (m browseModel) activeStatuses() []problemState {
	var selected []problemState
	for _, state := range filterStates {
		if m.statuses[state] {
			selected = append(selected, state)
		}
	}
	return selected
}

func (m browseModel) categoryFilterSummary() string {
	selected := m.activeCategories()
	if len(selected) == 0 || len(selected) == len(m.set.Categories) {
		return "all"
	}
	if len(selected) == 1 {
		return selected[0]
	}
	return fmt.Sprintf("%d selected", len(selected))
}

func (m browseModel) statusFilterSummary() string {
	selected := m.activeStatuses()
	if len(selected) == 0 || len(selected) == len(filterStates) {
		return "all"
	}
	if len(selected) == 1 {
		return selected[0].StatusKey()
	}
	if len(selected) == 2 {
		return selected[0].StatusKey() + ", " + selected[1].StatusKey()
	}
	return fmt.Sprintf("%d selected", len(selected))
}

func (m browseModel) filterSelectionDetail() string {
	if m.mode == modeFilterStatus {
		selected := m.activeStatuses()
		if len(selected) == 0 {
			return "all (none checked = show all)"
		}
		if len(selected) == len(filterStates) {
			return "all"
		}
		labels := make([]string, len(selected))
		for i, s := range selected {
			labels[i] = s.Label()
		}
		return strings.Join(labels, ", ")
	}

	selected := m.activeCategories()
	if len(selected) == 0 {
		return "all (none checked = show all)"
	}
	if len(selected) == len(m.set.Categories) {
		return "all"
	}
	return strings.Join(selected, ", ")
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

// openFile launches the problem's active implementation in the default
// application.
func (m *browseModel) openFile() {
	row := m.rows[m.detailIndex]
	active := activeVariantName(problemsDir, row.problem.ID, row.problem.Slug)
	path := solutionPath(problemsDir, row.problem.ID, row.problem.Slug, active)
	if !fileExists(path) {
		m.app.log().Warn("file.open", "no solution file", "slug", row.problem.Slug)
		m.message = "No .go file yet - press i to import."
		return
	}
	if err := openInDefaultApp(path); err != nil {
		m.app.log().Error("file.open", "open failed", "path", path, "error", err.Error())
		m.message = "Unable to open file: " + err.Error()
		return
	}
	m.app.log().Debug("file.open", "opened file", "path", path)
	m.message = "Opened " + path
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

// openChat starts an assistant session grounded in the current problem. When
// the assistant is reopened for the same problem, the existing conversation is
// reused so its context survives leaving and returning to the chat. Opening a
// different problem starts a fresh session.
func (m *browseModel) openChat() {
	if !hasAssistantKey() {
		m.app.log().Warn("assistant.open", "missing API key", "env", deepseekAPIKeyEnv)
		m.message = "Set " + deepseekAPIKeyEnv + " to use the assistant."
		return
	}

	row := m.rows[m.detailIndex]
	if m.chatSlug == row.problem.Slug {
		m.app.log().Debug("assistant.open", "reusing chat session", "slug", row.problem.Slug)
		m.mode = modeChat
		m.message = ""
		return
	}

	chat, ok := m.buildChat(row, "assistant.open")
	if !ok {
		return
	}
	m.chat = chat
	m.chatSlug = row.problem.Slug
	m.mode = modeChat
	m.message = ""
}

// resetChat starts a fresh assistant session for the current problem. It
// re-reads the statement and current solution so the new conversation sees the
// latest context.
func (m *browseModel) resetChat() {
	row := m.rows[m.detailIndex]
	chat, ok := m.buildChat(row, "assistant.reset")
	if !ok {
		return
	}
	chat.status = "conversation reset"
	m.chat = chat
	m.chatSlug = row.problem.Slug
	m.app.log().Info("assistant.reset", "conversation reset", "slug", row.problem.Slug)
}

// buildChat loads the assistant config and assembles a fresh chatModel grounded
// in the given problem row. On failure it records the error in m.message and
// reports ok=false.
func (m *browseModel) buildChat(row browseRow, event string) (chatModel, bool) {
	cfg, prompt, err := assistant.Load(assistantDir)
	if err != nil {
		m.app.log().Error(event, "load config failed", "error", err.Error())
		m.message = "Unable to load assistant config: " + err.Error()
		return chatModel{}, false
	}

	system := strings.TrimRight(prompt, "\n")
	if context := buildProblemContext(row.problem, m.lastResult); strings.TrimSpace(context) != "" {
		system += "\n\n" + context
	}

	m.app.log().Info(event, "chat session started",
		"problem", row.problem.ID,
		"title", row.problem.Title,
		"model", cfg.Model,
		"system_chars", len(system))
	m.app.log().VerboseDebug(event+".system", "system prompt", "text", system)

	return newChatModel(m.app, cfg, system, fmt.Sprintf("%d. %s", row.problem.ID, row.problem.Title), m.width, m.height), true
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
		text, err := m.app.TestProblem(id, "")
		return testResultMsg{text: text, err: err}
	}
}

func (m browseModel) submitCmd() tea.Cmd {
	id := m.rows[m.detailIndex].problem.ID
	return func() tea.Msg {
		text, err := m.app.SubmitProblem(id, "")
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

func (m browseModel) resetCmd() tea.Cmd {
	problem := m.rows[m.detailIndex].problem
	return func() tea.Msg {
		message, err := m.app.ResetProblem(problem)
		return resetResultMsg{message: message, err: err}
	}
}

// openVariantPicker loads a problem's implementations and enters the picker,
// placing the cursor on the active one.
func (m *browseModel) openVariantPicker() {
	row := m.rows[m.detailIndex]
	m.variants = listSolutionVariants(problemsDir, row.problem.ID, row.problem.Slug)
	m.variantCursor = 0
	m.variantOffset = 0
	for i, v := range m.variants {
		if v.Selected {
			m.variantCursor = i
			break
		}
	}
	m.confirmDelete = false
	m.variantInput = ""
	m.mode = modeVariant
	m.message = ""
	m.clampVariantCursor()
}

// handleVariantKey manages the implementation picker.
func (m *browseModel) handleVariantKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmDelete {
		switch msg.String() {
		case "y":
			m.deleteSelectedVariant()
		case "n", "esc", "q", "ctrl+c":
			m.confirmDelete = false
			m.message = ""
		}
		return m, nil
	}

	switch msg.String() {
	case "esc", "q", "ctrl+c":
		m.mode = modeDetail
		m.message = ""
	case "up", "k":
		m.variantCursor--
	case "down", "j":
		m.variantCursor++
	case "enter":
		m.selectActiveVariant()
	case "n":
		m.variantInput = ""
		m.mode = modeVariantName
	case "d":
		if m.variantCursor >= 0 && m.variantCursor < len(m.variants) && !m.variants[m.variantCursor].Primary {
			m.confirmDelete = true
			m.message = ""
		}
	case "o":
		m.openVariantFile()
	}
	m.clampVariantCursor()
	return m, nil
}

// handleVariantNameKey collects the name of a new implementation.
func (m *browseModel) handleVariantNameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeVariant
	case tea.KeyEnter:
		name := strings.TrimSpace(m.variantInput)
		if name != "" && !m.busy {
			m.busy = true
			m.variantInput = ""
			return m, m.newVariantCmd(name)
		}
	case tea.KeyBackspace:
		if runes := []rune(m.variantInput); len(runes) > 0 {
			m.variantInput = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes:
		m.variantInput += string(msg.Runes)
	case tea.KeySpace:
		m.variantInput += " "
	}
	return m, nil
}

// selectActiveVariant sets the highlighted implementation as active and returns
// to the detail view.
func (m *browseModel) selectActiveVariant() {
	if m.variantCursor < 0 || m.variantCursor >= len(m.variants) {
		return
	}
	v := m.variants[m.variantCursor]
	row := m.rows[m.detailIndex]
	name := "primary"
	if !v.Primary {
		name = v.Name
	}
	if err := setActiveVariant(problemsDir, row.problem.ID, row.problem.Slug, name); err != nil {
		m.message = "Unable to select: " + err.Error()
		return
	}
	m.app.log().Info("variant.select", "active variant set", "problem", row.problem.ID, "variant", name)
	m.message = "Active implementation: " + name
	m.mode = modeDetail
}

// deleteSelectedVariant removes the highlighted variant (never the primary) and
// refreshes the picker.
func (m *browseModel) deleteSelectedVariant() {
	if m.variantCursor < 0 || m.variantCursor >= len(m.variants) {
		return
	}
	v := m.variants[m.variantCursor]
	if v.Primary {
		return
	}
	row := m.rows[m.detailIndex]
	message, err := m.app.deleteVariant(row.problem.ID, v.Name)
	if err != nil {
		m.message = "Delete failed: " + err.Error()
		return
	}
	m.app.log().Info("variant.delete", "variant deleted", "problem", row.problem.ID, "variant", v.Name)
	m.openVariantPicker()
	m.message = message
}

// openVariantFile opens the highlighted implementation in the default app.
func (m *browseModel) openVariantFile() {
	if m.variantCursor < 0 || m.variantCursor >= len(m.variants) {
		return
	}
	path := m.variants[m.variantCursor].Path
	if err := openInDefaultApp(path); err != nil {
		m.message = "Unable to open file: " + err.Error()
		return
	}
	m.message = "Opened " + path
}

// newVariantCmd creates a fresh implementation off the UI goroutine.
func (m browseModel) newVariantCmd(name string) tea.Cmd {
	id := m.rows[m.detailIndex].problem.ID
	return func() tea.Msg {
		message, err := m.app.createVariant(id, name)
		return variantResultMsg{message: message, err: err}
	}
}

func (m *browseModel) clampVariantCursor() {
	if len(m.variants) == 0 {
		m.variantCursor = 0
		m.variantOffset = 0
		return
	}
	if m.variantCursor < 0 {
		m.variantCursor = 0
	}
	if m.variantCursor >= len(m.variants) {
		m.variantCursor = len(m.variants) - 1
	}
	height := m.variantViewHeight()
	if m.variantCursor < m.variantOffset {
		m.variantOffset = m.variantCursor
	}
	if m.variantCursor >= m.variantOffset+height {
		m.variantOffset = m.variantCursor - height + 1
	}
	maxOffset := len(m.variants) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.variantOffset > maxOffset {
		m.variantOffset = maxOffset
	}
	if m.variantOffset < 0 {
		m.variantOffset = 0
	}
}

func (m browseModel) variantViewHeight() int {
	h := m.height - 8
	if h < 3 {
		h = 3
	}
	return h
}

func (m browseModel) variantView() string {
	var b strings.Builder
	width := termWidth(m.width)
	row := m.rows[m.detailIndex]

	b.WriteString(clipStyled(titleStyle.Render(fmt.Sprintf("%d. %s · implementations", row.problem.ID, row.problem.Title)), width))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n\n")

	if len(m.variants) == 0 {
		b.WriteString(clipStyled(dimStyle.Render("  No implementations yet. Press i to import, then x to add a variant."), width))
		b.WriteString("\n")
	} else {
		height := m.variantViewHeight()
		end := m.variantOffset + height
		if end > len(m.variants) {
			end = len(m.variants)
		}
		for i := m.variantOffset; i < end; i++ {
			v := m.variants[i]
			marker := "  "
			if v.Selected {
				marker = accentStyle.Render("* ")
			}
			line := marker + v.label()
			if i == m.variantCursor {
				line = cursorStyle.Render("> ") + line
			} else {
				line = "  " + line
			}
			b.WriteString(clipStyled(line, width))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.confirmDelete {
		b.WriteString(dimStyle.Render("Delete this implementation? (y/n)"))
		b.WriteString("\n")
	} else {
		b.WriteString(clipStyled(dimStyle.Render("↑/↓ move · enter select · n new · d delete · o open · esc back"), width))
	}
	if m.message != "" {
		b.WriteString("\n" + clipStyled(dimStyle.Render(m.message), width))
	}
	return b.String()
}

func (m browseModel) variantNameView() string {
	var b strings.Builder
	width := termWidth(m.width)
	b.WriteString(clipStyled(titleStyle.Render("New implementation name"), width))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n\n")
	b.WriteString(accentStyle.Render("> " + m.variantInput + "█"))
	b.WriteString("\n\n")
	b.WriteString(clipStyled(dimStyle.Render("enter create · esc back"), width))
	return b.String()
}

func (m browseModel) View() string {
	switch m.mode {
	case modeText:
		return m.textView()
	case modeDetail:
		return m.detailView()
	case modeChat:
		return m.chat.View()
	case modeFilterStatus, modeFilterCategory:
		return m.filterView()
	case modeVariant:
		return m.variantView()
	case modeVariantName:
		return m.variantNameView()
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

	b.WriteString(clipStyled(
		titleStyle.Render(m.set.Name)+dimStyle.Render(fmt.Sprintf("  ·  %d/%d accepted", accepted, total)),
		width,
	))
	b.WriteString("\n")
	b.WriteString(progressBar(accepted, total, barWidth))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	category := m.categoryFilterSummary()
	status := m.statusFilterSummary()
	b.WriteString(clipStyled(dimStyle.Render(fmt.Sprintf("category: %s (c)   status: %s (s)   showing %d", category, status, len(m.visible))), width))
	b.WriteString("\n\n")

	height := m.listViewHeight()
	if len(m.visible) == 0 {
		b.WriteString(clipStyled(dimStyle.Render("  No problems match the current filters."), width))
		b.WriteString("\n")
		for pos := 1; pos < height; pos++ {
			b.WriteString("\n")
		}
	} else {
		end := m.offset + height
		if end > len(m.visible) {
			end = len(m.visible)
		}
		showCategory := len(m.activeCategories()) != 1
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

func (m browseModel) filterView() string {
	var b strings.Builder
	width := termWidth(m.width)

	title := "Filter by category"
	hasAny := m.hasCategoryFilter()
	switchHint := "s status"
	if m.mode == modeFilterStatus {
		title = "Filter by status"
		hasAny = m.hasStatusFilter()
		switchHint = "c category"
	}

	b.WriteString(clipStyled(
		titleStyle.Render(title)+dimStyle.Render(fmt.Sprintf("  ·  showing %d/%d", len(m.visible), len(m.rows))),
		width,
	))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")
	b.WriteString(clipStyled(dimStyle.Render("active: "+m.filterSelectionDetail()), width))
	b.WriteString("\n\n")

	options := m.filterOptions()
	labelWidth := 0
	for _, opt := range options {
		if n := len([]rune(opt.label)); n > labelWidth {
			labelWidth = n
		}
	}
	maxLabel := width - 11
	if maxLabel < 4 {
		maxLabel = 4
	}
	if maxLabel > 46 {
		maxLabel = 46
	}
	if labelWidth > maxLabel {
		labelWidth = maxLabel
	}

	height := m.filterViewHeight()
	end := m.filterOffset + height
	if end > len(options) {
		end = len(options)
	}
	for pos := m.filterOffset; pos < end; pos++ {
		opt := options[pos]
		flag := dimStyle.Render("[ ]")
		if opt.checked {
			flag = accentStyle.Render("[x]")
		}
		label := fitLine(opt.label, labelWidth)
		padded := fmt.Sprintf("%-*s", labelWidth, label)
		count := dimStyle.Render(fmt.Sprintf("%4d", opt.count))
		line := fmt.Sprintf("%s %s %s", flag, padded, count)
		if pos == m.filterCursor {
			line = cursorStyle.Render("> ") + line
		} else {
			line = "  " + line
		}
		b.WriteString(clipStyled(line, width))
		b.WriteString("\n")
	}
	for pos := end - m.filterOffset; pos < height; pos++ {
		b.WriteString("\n")
	}

	b.WriteString("\n")
	allAction := "a check all"
	if hasAny {
		allAction = "a clear"
	}
	footer := fmt.Sprintf("↑/↓ move · space toggle · %s · %s · enter/esc done", allAction, switchHint)
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
	active := activeVariantName(problemsDir, row.problem.ID, row.problem.Slug)
	file := solutionPath(problemsDir, row.problem.ID, row.problem.Slug, active)
	if fileExists(file) {
		suffix := ""
		if active != "" {
			suffix = "  [variant: " + active + "]"
		}
		b.WriteString("file:   " + file + suffix + "\n")
	} else {
		b.WriteString("file:   (not imported)\n")
	}

	b.WriteString("\n")
	mark := "m mark accepted"
	if row.state == stateAccepted {
		mark = "m clear accepted"
	}
	b.WriteString(clipStyled(dimStyle.Render(strings.Join([]string{"t test", "s submit", "v statement", "o open file", "x variants", "a assistant", "i import", "r reset", mark, "esc back"}, " · ")), width))

	if m.confirmReset {
		b.WriteString("\n\n" + dimStyle.Render("Reset implementation to starter? (y/n)"))
	}
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
