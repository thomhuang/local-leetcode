package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tuiScreen identifies which view the root TUI is showing.
type tuiScreen int

const (
	screenMenu tuiScreen = iota
	screenSelect
	screenInput
	screenConfirm
	screenBusy
	screenResult
	screenBrowse
)

// mainMenu is the label of each action, in the order they are shown.
var mainMenu = []string{
	"Add a new question",
	"Add a problem set",
	"Authenticate user",
	"Test code",
	"Submit code",
	"Browse problems",
	"Exit",
}

const (
	// panelWidth is the lipgloss width of a card, including its horizontal
	// padding but excluding the border.
	panelWidth = 64
	// panelInner is the usable text width inside a card.
	panelInner = panelWidth - 4
)

var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
	accentStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	taglineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	itemStyle    = lipgloss.NewStyle().Padding(0, 1).Width(panelInner)
	itemSelStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Width(panelInner).
			Bold(true).
			Foreground(lipgloss.Color("232")).
			Background(lipgloss.Color("212"))
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// tuiResultMsg carries the outcome of an asynchronous action back to the root
// model, which renders it on the scrollable result screen.
type tuiResultMsg struct {
	title string
	body  string
}

// browseExitMsg is returned by the browse model when the user backs out to the
// main menu.
type browseExitMsg struct{}

// openBrowseMsg asks the root model to open the checklist for a named set.
type openBrowseMsg struct {
	name string
}

// spinnerTickMsg advances the busy-screen spinner.
type spinnerTickMsg struct{}

type confirmHandlerFunc func(m tuiModel, yes bool) (tea.Model, tea.Cmd)

// tuiModel is the root bubbletea model. It owns the main menu and the small
// input/select/confirm/result screens that the menu actions need.
type tuiModel struct {
	app *App

	width, height int
	screen        tuiScreen

	menuIndex int

	selectPrompt  string
	selectChoices []string
	selectIndex   int
	selectBusy    string
	selectPick    func(string) tea.Cmd

	inputPrompt   string
	inputValue    string
	inputError    string
	inputBusy     string
	inputValidate func(string) string
	inputSubmit   func(string) tea.Cmd

	confirmPrompt string
	confirmAnswer confirmHandlerFunc

	busyText     string
	spinnerFrame int

	resultTitle  string
	resultLines  []string
	resultOffset int

	browse   browseModel
	browsing bool
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.browsing {
			updated, cmd := m.browse.Update(msg)
			m.browse = updated.(browseModel)
			return m, cmd
		}
		return m, nil
	case spinnerTickMsg:
		if m.screen == screenBusy {
			m.spinnerFrame++
			return m, tickSpinnerCmd()
		}
		return m, nil
	case browseExitMsg:
		m.browsing = false
		m.screen = screenMenu
		return m, nil
	case openBrowseMsg:
		return m.openBrowse(msg.name)
	case tuiResultMsg:
		m.showResult(msg.title, msg.body)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// The browse checklist owns its own key handling and result messages.
	if m.browsing {
		switch msg.(type) {
		case tea.KeyMsg, testResultMsg, submitResultMsg, statementMsg, importResultMsg:
			updated, cmd := m.browse.Update(msg)
			m.browse = updated.(browseModel)
			return m, cmd
		}
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenInput:
		return m.updateInput(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	case screenResult:
		return m.updateResult(msg)
	default:
		return m.updateMenu(msg)
	}
}

func (m tuiModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.menuIndex = (m.menuIndex + len(mainMenu) - 1) % len(mainMenu)
	case "down", "j":
		m.menuIndex = (m.menuIndex + 1) % len(mainMenu)
	case "enter":
		return m.runMenuAction(m.menuIndex)
	}
	return m, nil
}

func (m tuiModel) runMenuAction(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0:
		return m.startAddQuestion()
	case 1:
		return m.startAddProblemSet()
	case 2:
		return m.startAuthenticate()
	case 3:
		return m.startTestCode()
	case 4:
		return m.startSubmitCode()
	case 5:
		return m.startBrowse()
	case 6:
		return m, tea.Quit
	}
	return m, nil
}

// startSelect moves to the dropdown screen. The busy text is shown while the
// picked option runs.
func (m *tuiModel) startSelect(prompt string, choices []string, busy string, pick func(string) tea.Cmd) {
	m.screen = screenSelect
	m.selectPrompt = prompt
	m.selectChoices = choices
	m.selectIndex = 0
	m.selectBusy = busy
	m.selectPick = pick
}

// startInput moves to the text input screen. validate returns an error message
// to show, or an empty string when the value is acceptable.
func (m *tuiModel) startInput(prompt, busy string, validate func(string) string, submit func(string) tea.Cmd) {
	m.screen = screenInput
	m.inputPrompt = prompt
	m.inputValue = ""
	m.inputError = ""
	m.inputBusy = busy
	m.inputValidate = validate
	m.inputSubmit = submit
}

// withBusy switches to the busy screen and starts both the action and the
// spinner.
func (m *tuiModel) withBusy(text string, action tea.Cmd) (tea.Model, tea.Cmd) {
	m.screen = screenBusy
	m.busyText = text
	m.spinnerFrame = 0
	return *m, tea.Batch(tickSpinnerCmd(), action)
}

func tickSpinnerCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} })
}

func (m tuiModel) startAddQuestion() (tea.Model, tea.Cmd) {
	m.startInput(
		"What problem number are you interested in?",
		"Fetching the problem statement",
		func(input string) string {
			id, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil {
				return "Please input an integer problem number."
			}
			if _, ok := m.app.Questions[id]; !ok {
				return fmt.Sprintf("Problem %d was not found in the problem list.", id)
			}
			return ""
		},
		func(input string) tea.Cmd {
			id, _ := strconv.Atoi(strings.TrimSpace(input))
			return m.app.addQuestionCmd(id)
		},
	)
	return m, nil
}

func (m tuiModel) startAddProblemSet() (tea.Model, tea.Cmd) {
	names := availableProblemSets()
	if len(names) == 0 {
		m.showResult("Add a problem set", fmt.Sprintf("No problem sets found under %s.", problemSetsDir))
		return m, nil
	}
	m.startSelect("Which problem set would you like to add?", names, "Importing the problem set", func(name string) tea.Cmd {
		return m.app.importSetCmd(name)
	})
	return m, nil
}

func (m tuiModel) startAuthenticate() (tea.Model, tea.Cmd) {
	if fresh, days := m.app.authenticationFreshness(); fresh {
		m.screen = screenConfirm
		m.confirmPrompt = fmt.Sprintf("Your session has %d days left. Authenticate anyway?", days)
		m.confirmAnswer = func(model tuiModel, yes bool) (tea.Model, tea.Cmd) {
			if !yes {
				model.screen = screenMenu
				return model, nil
			}
			return model.startCookieInput()
		}
		return m, nil
	}
	return m.startCookieInput()
}

func (m tuiModel) startCookieInput() (tea.Model, tea.Cmd) {
	m.startInput(
		"Paste your authenticated LeetCode cookies",
		"Saving and verifying your session",
		func(input string) string {
			if _, _, ok := parseCookies(input); !ok {
				return "Cookies must contain LEETCODE_SESSION and csrftoken."
			}
			return ""
		},
		func(input string) tea.Cmd {
			session, csrfToken, _ := parseCookies(input)
			return m.app.authenticateCmd(session, csrfToken)
		},
	)
	return m, nil
}

func (m tuiModel) startTestCode() (tea.Model, tea.Cmd) {
	m.startInput(
		"What problem number would you like to run?",
		"Running the example test cases",
		m.app.problemNumberValidator(),
		func(input string) tea.Cmd {
			id, _ := strconv.Atoi(strings.TrimSpace(input))
			return m.app.testProblemCmd(id)
		},
	)
	return m, nil
}

func (m tuiModel) startSubmitCode() (tea.Model, tea.Cmd) {
	m.startInput(
		"What problem number would you like to submit?",
		"Submitting your solution",
		m.app.problemNumberValidator(),
		func(input string) tea.Cmd {
			id, _ := strconv.Atoi(strings.TrimSpace(input))
			return m.app.submitProblemCmd(id)
		},
	)
	return m, nil
}

func (m tuiModel) startBrowse() (tea.Model, tea.Cmd) {
	names := availableProblemSets()
	if len(names) == 0 {
		m.showResult("Browse problems", fmt.Sprintf("No problem sets found under %s.", problemSetsDir))
		return m, nil
	}
	m.startSelect("Choose a problem set to browse", names, "Loading the problem set", func(name string) tea.Cmd {
		return func() tea.Msg { return openBrowseMsg{name: name} }
	})
	return m, nil
}

func (m tuiModel) openBrowse(name string) (tea.Model, tea.Cmd) {
	set, err := m.app.LoadProblemSet(name)
	if err != nil {
		m.showResult("Browse problems", "Failed to load problem set: "+err.Error())
		return m, nil
	}
	if len(set.Problems()) == 0 {
		m.showResult("Browse problems", fmt.Sprintf("Problem set %q has no problems.", name))
		return m, nil
	}

	m.browse = newBrowseModel(m.app, set)
	m.browsing = true
	m.screen = screenBrowse
	if m.width > 0 {
		updated, _ := m.browse.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		m.browse = updated.(browseModel)
	}
	return m, nil
}

func (m tuiModel) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "q":
		m.screen = screenMenu
	case "up", "k":
		if len(m.selectChoices) > 0 {
			m.selectIndex = (m.selectIndex + len(m.selectChoices) - 1) % len(m.selectChoices)
		}
	case "down", "j":
		if len(m.selectChoices) > 0 {
			m.selectIndex = (m.selectIndex + 1) % len(m.selectChoices)
		}
	case "enter":
		if len(m.selectChoices) == 0 || m.selectPick == nil {
			return m, nil
		}
		choice := m.selectChoices[m.selectIndex]
		return m.withBusy(m.selectBusy, m.selectPick(choice))
	}
	return m, nil
}

func (m tuiModel) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyEsc:
		m.screen = screenMenu
	case tea.KeyEnter:
		value := strings.TrimSpace(m.inputValue)
		if m.inputValidate != nil {
			if errMsg := m.inputValidate(value); errMsg != "" {
				m.inputError = errMsg
				return m, nil
			}
		}
		m.inputError = ""
		if m.inputSubmit != nil {
			return m.withBusy(m.inputBusy, m.inputSubmit(value))
		}
	case tea.KeyBackspace:
		if len(m.inputValue) > 0 {
			runes := []rune(m.inputValue)
			m.inputValue = string(runes[:len(runes)-1])
		}
	case tea.KeySpace:
		m.inputValue += " "
	case tea.KeyRunes:
		m.inputValue += string(key.Runes)
	}
	return m, nil
}

func (m tuiModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "y", "Y":
		if m.confirmAnswer != nil {
			return m.confirmAnswer(m, true)
		}
	case "n", "N", "esc", "q":
		if m.confirmAnswer != nil {
			return m.confirmAnswer(m, false)
		}
	}
	return m, nil
}

func (m tuiModel) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc", "q", "enter":
		m.screen = screenMenu
	case "up", "k":
		m.resultOffset--
	case "down", "j":
		m.resultOffset++
	}
	m.clampResult()
	return m, nil
}

func (m *tuiModel) showResult(title, body string) {
	m.resultTitle = title
	body = strings.ReplaceAll(body, "\r\n", "\n")
	m.resultLines = strings.Split(body, "\n")
	m.resultOffset = 0
	m.screen = screenResult
}

func (m *tuiModel) clampResult() {
	height := m.resultViewHeight()
	maxOffset := len(m.resultLines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.resultOffset < 0 {
		m.resultOffset = 0
	}
	if m.resultOffset > maxOffset {
		m.resultOffset = maxOffset
	}
}

func (m tuiModel) resultViewHeight() int {
	h := m.height - 8
	if h < 3 {
		h = 3
	}
	return h
}

func (m tuiModel) View() string {
	if m.browsing {
		return m.browse.View()
	}
	switch m.screen {
	case screenSelect:
		return m.selectView()
	case screenInput:
		return m.inputView()
	case screenConfirm:
		return m.confirmView()
	case screenBusy:
		return m.busyView()
	case screenResult:
		return m.resultView()
	default:
		return m.menuView()
	}
}

// card frames a title, an optional body, and an optional help line in a rounded
// panel.
func card(title, body, help string) string {
	sections := []string{accentStyle.Render(title)}
	if body != "" {
		sections = append(sections, body)
	}
	if help != "" {
		sections = append(sections, helpStyle.Render(help))
	}
	return panelStyle.Width(panelWidth).Render(strings.Join(sections, "\n\n"))
}

func (m tuiModel) center(view string) string {
	if m.width <= 0 || m.height <= 0 {
		return view
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
}

func (m tuiModel) placeTop(view string) string {
	if m.width <= 0 {
		return view
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, view)
}

// fitLine truncates a line to width runes, adding an ellipsis when clipped.
func fitLine(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

// fitTail keeps the end of a line so the user can see what they last typed.
func fitTail(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	if width <= 1 {
		return string(runes[len(runes)-width:])
	}
	return "…" + string(runes[len(runes)-(width-1):])
}

// progressBar renders a filled/empty bar for the accepted count.
func progressBar(done, total, width int) string {
	if total <= 0 {
		total = 1
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	filled := done * width / total
	return acceptedStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", width-filled))
}

func menuItem(label string, selected bool) string {
	if selected {
		return itemSelStyle.Render("❯ " + label)
	}
	return itemStyle.Render("  " + label)
}

func (m tuiModel) menuView() string {
	items := make([]string, 0, len(mainMenu))
	for i, label := range mainMenu {
		items = append(items, menuItem(label, i == m.menuIndex))
	}
	body := strings.Join(items, "\n")
	view := card("local-leetcode", body, "↑/↓ move · enter select · q quit")
	return m.center(view)
}

func (m tuiModel) selectView() string {
	items := make([]string, 0, len(m.selectChoices))
	for i, choice := range m.selectChoices {
		items = append(items, menuItem(choice, i == m.selectIndex))
	}
	help := fmt.Sprintf("%d available · ↑/↓ move · enter select · esc back", len(m.selectChoices))
	view := card(m.selectPrompt, strings.Join(items, "\n"), help)
	return m.center(view)
}

func (m tuiModel) inputView() string {
	var b strings.Builder
	b.WriteString(fitTail(m.inputValue, panelInner-1))
	b.WriteString(accentStyle.Render("█"))
	if m.inputError != "" {
		b.WriteString("\n\n")
		b.WriteString(missingStyle.Render(fitLine(m.inputError, panelInner)))
	}
	view := card(m.inputPrompt, b.String(), "enter confirm · esc back")
	return m.center(view)
}

func (m tuiModel) confirmView() string {
	body := itemStyle.Render("  y  yes") + "\n" + itemStyle.Render("  n  no")
	view := card(m.confirmPrompt, body, "y yes · n no")
	return m.center(view)
}

func (m tuiModel) busyView() string {
	frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
	body := accentStyle.Render(frame) + "  " + m.busyText + "…"
	return m.center(card("Working", body, ""))
}

func (m tuiModel) resultView() string {
	height := m.resultViewHeight()
	end := m.resultOffset + height
	if end > len(m.resultLines) {
		end = len(m.resultLines)
	}

	lines := make([]string, 0, end-m.resultOffset)
	for i := m.resultOffset; i < end; i++ {
		lines = append(lines, fitLine(m.resultLines[i], panelInner))
	}

	help := "↑/↓ scroll · enter/esc back"
	if len(m.resultLines) > height {
		help = fmt.Sprintf("%d/%d · %s", m.resultOffset+1, len(m.resultLines), help)
	}

	view := card(m.resultTitle, strings.Join(lines, "\n"), help)
	return m.placeTop(view)
}

// problemNumberValidator returns a validator that checks a typed problem number
// against the loaded problem list.
func (app *App) problemNumberValidator() func(string) string {
	return func(input string) string {
		id, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return "Please input an integer problem number."
		}
		if _, ok := app.Questions[id]; !ok {
			return fmt.Sprintf("Problem %d was not found in the problem list.", id)
		}
		return ""
	}
}

// authenticationFreshness reports whether the saved session is still fresh and
// how many days remain before it should be refreshed.
func (app *App) authenticationFreshness() (fresh bool, daysLeft int) {
	lastUpdated := app.UserAuth.LastUpdated
	fiveDaysPrev := time.Now().AddDate(0, 0, -authFreshnessDays)
	if fiveDaysPrev.Before(lastUpdated) {
		return true, int(lastUpdated.Sub(fiveDaysPrev).Hours() / 24)
	}
	return false, 0
}

// RunTUI starts the full-screen interactive menu.
func (app *App) RunTUI() error {
	program := tea.NewProgram(tuiModel{app: app}, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
