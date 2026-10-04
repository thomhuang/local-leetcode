package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// accentStyle highlights interactive text such as the prompt and headings.
var accentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

// fitLine truncates a line to width runes, adding an ellipsis when clipped.
func fitLine(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}
	if width <= 0 {
		return ""
	}
	if width == 1 {
		return string(runes[:1])
	}
	return string(runes[:width-1]) + "…"
}

// termWidth returns a usable terminal width, falling back to 80 columns when a
// model has not yet received its initial WindowSizeMsg.
func termWidth(width int) int {
	if width <= 0 {
		return 80
	}
	return width
}

// clipStyled truncates a line that may contain ANSI color codes to a display
// width, measured in terminal cells rather than runes.
func clipStyled(line string, width int) string {
	if width <= 0 || lipgloss.Width(line) <= width {
		return line
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(line)
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
