package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Panel(title string, content string, width int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColHeader)).
		Background(lipgloss.Color("#222230")).
		Padding(0, 1)

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColBorder))

	header := headerStyle.Render(" " + title + " ")
	body := lipgloss.NewStyle().Padding(0, 1).Render(content)

	inner := lipgloss.JoinVertical(lipgloss.Left, header, body)
	return borderStyle.Width(width).Render(inner)
}

func Bordered(content string) string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColBorder)).
		Render(content)
}

func Separator(width int) string {
	return strings.Repeat("─", width)
}
