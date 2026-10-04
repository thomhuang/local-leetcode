package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// browseExitMsg is returned by the browse model when the user backs out to the
// prompt; the screen host turns it into a program quit.
type browseExitMsg struct{}

// runScreen runs a full-screen bubbletea view. It is a package variable so
// tests can substitute a stub instead of opening a real terminal program.
var runScreen = runScreenProgram

func runScreenProgram(model tea.Model) error {
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// screenHost wraps a full-screen sub-view (the browse checklist or the chat) and
// adds the global ctrl+o debug overlay. It also turns the sub-view's exit signal
// into a program quit.
type screenHost struct {
	inner  tea.Model
	debug  debugPanel
	width  int
	height int
}

func newScreenHost(app *App, inner tea.Model) screenHost {
	return screenHost{inner: inner, debug: newDebugPanel(app.log())}
}

func (h screenHost) Init() tea.Cmd {
	return h.inner.Init()
}

func (h screenHost) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width, h.height = m.Width, m.Height
		h.debug.setHeight(m.Height)
	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			return h, tea.Quit
		}
		if m.String() == debugOpenKey {
			cmd := h.debug.toggle()
			return h, cmd
		}
		if h.debug.open {
			h.debug.handleKey(m)
			return h, nil
		}
	case debugTickMsg:
		if h.debug.open {
			return h, tickDebugCmd()
		}
		return h, nil
	case browseExitMsg:
		return h, tea.Quit
	}

	var cmd tea.Cmd
	h.inner, cmd = h.inner.Update(msg)
	return h, cmd
}

func (h screenHost) View() string {
	if h.debug.open {
		return h.debug.view(h.width, h.height)
	}
	return h.inner.View()
}

// chatHost adapts chatModel to tea.Model and quits the program when the user
// leaves the chat with esc.
type chatHost struct {
	chat chatModel
}

func (h chatHost) Init() tea.Cmd {
	return nil
}

func (h chatHost) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.chat, cmd = h.chat.Update(msg)
	if h.chat.exited {
		return h, tea.Quit
	}
	return h, cmd
}

func (h chatHost) View() string {
	return h.chat.View()
}
