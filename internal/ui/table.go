package ui

import (
	"fmt"

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

// RenderSessionsTable renders sessions as a formatted table.
// Columns: Session | Model | Bar | Cache | in/out/cc/cr | Cost | Time
func RenderSessionsTable(sessions []*aggregator.SessionStats, opts TableOptions) string {
	if len(sessions) == 0 {
		return StyleDim.Render("No sessions found.")
	}

	var active, empty []*aggregator.SessionStats
	for _, s := range sessions {
		if s.TotalTokens == 0 {
			empty = append(empty, s)
		} else {
			active = append(active, s)
		}
	}

	if len(active) == 0 {
		return StyleDim.Render(fmt.Sprintf("No active sessions. (%d empty hidden)", len(empty)))
	}

	tab := table.NewWriter()
	tab.SetOutputMirror(nil)
	tab.SetAllowedRowLength(opts.Width)
	style := table.StyleLight
	style.Options.SeparateRows = true
	tab.SetStyle(style)

	title := fmt.Sprintf(" %d session", len(active))
	if len(active) != 1 {
		title += "s"
	}
	if len(empty) > 0 {
		title += fmt.Sprintf("  (+%d empty hidden)", len(empty))
	}
	tab.SetTitle(title + " ")

	// Pass *SessionStats for every column so transformers have full context.
	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Name: "Session", WidthMax: 22, Align: text.AlignLeft},
		{Number: 2, Name: "Model", WidthMax: 8, Align: text.AlignLeft, Transformer: modelTransformer},
		{Number: 3, Name: "Tokens", WidthMax: 22, Align: text.AlignLeft, Transformer: tokenBarTransformer(opts.Width)},
		{Number: 4, Name: "Cache", WidthMax: 6, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: cacheTransformer},
		{Number: 5, Name: "in / out / cc / cr", WidthMax: 38, Align: text.AlignLeft, Transformer: breakdownTransformer},
		{Number: 6, Name: "Cost", WidthMax: 8, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: costTransformer},
		{Number: 7, Name: "Time", WidthMax: 7, Align: text.AlignRight, AlignHeader: text.AlignRight, Transformer: timeTransformer},
	})

	tab.AppendHeader(table.Row{"Session", "Model", "Tokens", "Cache", "in / out / cc / cr", "Cost", "Time"})

	for _, s := range active {
		tab.AppendRow(table.Row{tableSessionName(s), s, s, s, s, s, s})
	}

	return tab.Render()
}

func modelTransformer(val interface{}) string {
	s, ok := val.(*aggregator.SessionStats)
	if !ok {
		return ""
	}
	return ModelTag(s.Model)
}

func tokenBarTransformer(width int) func(interface{}) string {
	return func(val interface{}) string {
		s, ok := val.(*aggregator.SessionStats)
		if !ok {
			return ""
		}
		if s.InputTokens+s.OutputTokens+s.CacheCreateTokens+s.CacheReadTokens == 0 {
			return StyleDim.Render("~")
		}
		return TokenBar{
			Input:       s.InputTokens,
			Output:      s.OutputTokens,
			CacheCreate: s.CacheCreateTokens,
			CacheRead:   s.CacheReadTokens,
			Width:       20,
		}.Render()
	}
}

func breakdownTransformer(val interface{}) string {
	s, ok := val.(*aggregator.SessionStats)
	if !ok {
		return ""
	}
	if s.TotalTokens == 0 {
		return StyleDim.Render("~")
	}
	return TokenSubtitleColored(s.InputTokens, s.OutputTokens, s.CacheCreateTokens, s.CacheReadTokens)
}

func cacheTransformer(val interface{}) string {
	s, ok := val.(*aggregator.SessionStats)
	if !ok {
		return ""
	}
	if s.TotalTokens == 0 {
		return StyleDim.Render("~")
	}
	eff := s.CacheEfficiency * 100
	var style lipgloss.Style
	switch {
	case eff >= 80:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Green)).Bold(true)
	case eff >= 40:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Amber)).Bold(true)
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Red)).Bold(true)
	}
	return style.Render(fmt.Sprintf("%.0f%%", eff))
}

func costTransformer(val interface{}) string {
	s, ok := val.(*aggregator.SessionStats)
	if !ok {
		return ""
	}
	if s.CostUSD <= 0 {
		return StyleDim.Render("~")
	}
	return FormatCost(s.CostUSD)
}

func timeTransformer(val interface{}) string {
	s, ok := val.(*aggregator.SessionStats)
	if !ok {
		return ""
	}
	return FormatDuration(s.Duration)
}

func tableSessionName(s *aggregator.SessionStats) string {
	name := s.Summary
	if name == "" {
		if s.ProjectName != "" {
			name = s.ProjectName
		} else if len(s.ID) >= 8 {
			name = s.ID[:8]
		} else {
			name = "session"
		}
	}
	return Truncate(name, 20)
}

// sessionName is kept for compatibility with other call sites.
func sessionName(s *aggregator.SessionStats) string {
	return tableSessionName(s)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
