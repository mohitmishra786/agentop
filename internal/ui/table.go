package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Header   string
	Width    int
	Align    string
	MinWidth int
}

func FullColumns(termWidth int) []Column {
	fixed := 22 + 8 + 8 + 6 + 10 + 9
	barWidth := termWidth - fixed
	if barWidth < 20 {
		barWidth = 20
	}

	return []Column{
		{Header: "session", Width: 22, Align: "left"},
		{Header: "token mix", Width: barWidth, Align: "left"},
		{Header: "cache%", Width: 8, Align: "right"},
		{Header: "msgs", Width: 6, Align: "right", MinWidth: 80},
		{Header: "tools", Width: 6, Align: "right", MinWidth: 100},
		{Header: "cost", Width: 9, Align: "right"},
		{Header: "time", Width: 8, Align: "right"},
	}
}

func RenderHeader(cols []Column) string {
	var parts []string
	for _, c := range cols {
		if c.Width <= 0 {
			continue
		}
		parts = append(parts, PadRight(c.Header, c.Width))
	}
	return StyleHeader.Render(strings.Join(parts, " "))
}

func RenderRow(cols []Column, cells []string) string {
	var parts []string
	for i, c := range cols {
		if c.Width <= 0 || i >= len(cells) {
			continue
		}
		val := cells[i]
		if c.Align == "right" {
			parts = append(parts, PadLeft(val, c.Width))
		} else {
			parts = append(parts, PadRight(val, c.Width))
		}
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(parts, " "))
}
