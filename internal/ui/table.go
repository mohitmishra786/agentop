package ui

import (
	"fmt"
	"strings"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// TableOptions configures table rendering
type TableOptions struct {
	SortBy       string
	ShowProgress bool
	Theme        string
	Width        int
}

// RenderSessionsTable renders sessions as a formatted table
func RenderSessionsTable(sessions []*aggregator.SessionStats, opts TableOptions) string {
	if len(sessions) == 0 {
		return StyleDim.Render("No sessions found.")
	}

	tab := table.NewWriter()
	tab.SetOutputMirror(nil)
	tab.SetAllowedRowLength(opts.Width)
	tab.SetStyle(table.Style{
		Name:    "StyleDefault",
		Options: table.Options{DrawBorder: false, SeparateColumns: true, SeparateHeader: true},
	})
	tab.SetTitle(fmt.Sprintf(" %d sessions ", len(sessions)))

	// Configure columns
	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Name: "Session", WidthMax: 25},
		{Number: 2, Name: "Model", WidthMax: 10, Align: text.AlignLeft},
		{Number: 3, Name: "Tokens", WidthMax: 30, Align: text.AlignLeft, Transformer: tokenBarTransformer(opts.Width)},
		{Number: 4, Name: "Cache", WidthMax: 6, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: cacheTransformer},
		{Number: 5, Name: "Cost", WidthMax: 8, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: costTransformer},
		{Number: 6, Name: "Time", WidthMax: 6, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: timeTransformer},
	})

	// Header
	tab.AppendHeader(table.Row{"Session", "Model", "Tokens", "Cache", "Cost", "Time"})

	// Rows
	for _, s := range sessions {
		tab.AppendRow(table.Row{
			sessionName(s),
			s.Model,
			s,
			s,
			s,
			s,
		})
	}

	return tab.Render()
}

func tokenBarTransformer(width int) func(interface{}) string {
	return func(val interface{}) string {
		s := val.(*aggregator.SessionStats)

		barWidth := min(20, max(10, width/4))
		total := s.InputTokens + s.OutputTokens + s.CacheCreateTokens + s.CacheReadTokens
		if total == 0 {
			return StyleDim.Render("~")
		}

		segs := []struct {
			count int64
			color string
		}{
			{s.InputTokens, theme.Blue},
			{s.OutputTokens, theme.Cyan},
			{s.CacheCreateTokens, theme.Yellow},
			{s.CacheReadTokens, theme.Magenta},
		}

		widths := make([]int, len(segs))
		tw := 0
		for i, s := range segs {
			w := int(float64(s.count) / float64(total) * float64(barWidth))
			widths[i] = w
			tw += w
		}
		if d := barWidth - tw; d != 0 {
			mi := 0
			for i := range widths {
				if widths[i] > widths[mi] {
					mi = i
				}
			}
			widths[mi] += d
		}

		var parts []string
		for i, s := range segs {
			if widths[i] <= 0 {
				continue
			}
			bar := strings.Repeat("|", widths[i])
			coloredBar := lipgloss.NewStyle().Foreground(lipgloss.Color(s.color)).Render(bar)
			parts = append(parts, coloredBar)
		}

		barStr := strings.Join(parts, "")
		infoStr := fmt.Sprintf("%s in · %s out", HumanizeTokens(s.InputTokens), HumanizeTokens(s.OutputTokens))

		return barStr + " " + StyleDim.Render(infoStr)
	}
}

func cacheTransformer(val interface{}) string {
	s := val.(*aggregator.SessionStats)
	if s.TotalTokens == 0 {
		return "~"
	}

	eff := s.CacheEfficiency * 100
	var style lipgloss.Style
	switch {
	case eff >= 80:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Green)).Bold(true)
	case eff >= 40:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Yellow)).Bold(true)
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Red)).Bold(true)
	}

	return style.Render(fmt.Sprintf("%.0f%%", eff))
}

func costTransformer(val interface{}) string {
	s := val.(*aggregator.SessionStats)
	if s.CostUSD <= 0 {
		return "~"
	}
	return FormatCost(s.CostUSD)
}

func timeTransformer(val interface{}) string {
	s := val.(*aggregator.SessionStats)
	return FormatDuration(s.Duration)
}

func sessionName(s *aggregator.SessionStats) string {
	name := s.Summary
	if name == "" {
		if len(s.ID) >= 8 {
			name = s.ID[:8]
		} else if s.ProjectName != "" {
			name = s.ProjectName
		} else {
			name = "session"
		}
	}
	return Truncate(name, 20)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
