package ui

import (
	"fmt"
	"strings"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/lipgloss"
)

func RenderPanel(title string, content string, width int) string {
	header := StyleHeader.Render(" " + title + " ")
	body := lipgloss.NewStyle().Padding(0, 1).Render(content)
	panel := StyleBorder.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
	return panel
}

func RenderSessionRow(s *aggregator.SessionStats, barWidth int) string {
	model := s.Model
	if model == "" {
		model = "unknown"
	}

	sessionName := s.Summary
	if sessionName == "" {
		if len(s.ID) >= 8 {
			sessionName = s.ID[:8]
		} else if s.ProjectName != "" {
			sessionName = s.ProjectName
		} else {
			sessionName = "session"
		}
	}
	sessionName = Truncate(sessionName, 20)

	cachePct := fmt.Sprintf("%.0f%%", s.CacheEfficiency*100)
	costStr := FormatCost(s.CostUSD)
	durStr := FormatDuration(s.Duration)

	bar := TokenBar{
		Input:       s.InputTokens,
		Output:      s.OutputTokens,
		CacheCreate: s.CacheCreateTokens,
		CacheRead:   s.CacheReadTokens,
		Width:       barWidth,
	}.Render()

	effStyle := CacheEfficiencyStyle(s.CacheEfficiency)

	subagentNote := ""
	if s.SubagentCount > 0 {
		subagentNote = StyleDim.Render(fmt.Sprintf(" (+%d subagents, %s tokens)", s.SubagentCount, HumanizeTokens(s.SubagentTokens)))
	}

	row := fmt.Sprintf("%s %s  %s  %s  %s",
		sessionName+" "+ModelTag(model),
		bar,
		effStyle.Render(cachePct),
		costStr,
		durStr,
	)

	subtitle := TokenSubtitle(s.InputTokens, s.OutputTokens, s.CacheCreateTokens, s.CacheReadTokens)

	return row + subagentNote + "\n" + subtitle
}

func RenderToday(sessions []*aggregator.SessionStats, termWidth int) string {
	if len(sessions) == 0 {
		return StyleDim.Render("No sessions found for today.")
	}

	var totalCost float64
	var totalTokens int64
	var totalCacheEff float64
	var sessionsWithCache int

	for _, s := range sessions {
		totalCost += s.CostUSD
		totalTokens += s.TotalTokens
		if s.InputTokens+s.CacheCreateTokens+s.CacheReadTokens > 0 {
			totalCacheEff += s.CacheEfficiency
			sessionsWithCache++
		}
	}

	avgCacheEff := 0.0
	if sessionsWithCache > 0 {
		avgCacheEff = totalCacheEff / float64(sessionsWithCache)
	}

	barWidth := 30
	summaryCards := renderSummaryCards(totalCost, totalTokens, avgCacheEff, barWidth)

	var sessionRows []string
	for _, s := range sessions {
		sessionRows = append(sessionRows, RenderSessionRow(s, termWidth-60))
	}

	sessionContent := strings.Join(sessionRows, "\n"+strings.Repeat("─", termWidth-4)+"\n")

	sessionPanel := RenderPanel(
		fmt.Sprintf("claude code · %d sessions", len(sessions)),
		sessionContent,
		termWidth-2,
	)

	anomalies := renderAnomalies(sessions)

	output := summaryCards + "\n\n" + sessionPanel
	if anomalies != "" {
		output += "\n\n" + anomalies
	}

	return output
}

func renderSummaryCards(totalCost float64, totalTokens int64, cacheEff float64, barWidth int) string {
	cards := []string{
		renderCard("total cost", FormatCost(totalCost), ColGreen),
		renderCard("tokens", HumanizeTokens(totalTokens), ColBlue),
		renderCard("cache eff", fmt.Sprintf("%.0f%%", cacheEff*100), ColTeal),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func renderCard(label, value, color string) string {
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Padding(0, 1)
	labelStyle := StyleDim.Padding(0, 1)
	bar := MiniBar(0.6, 10, color)
	return lipgloss.JoinVertical(lipgloss.Center,
		labelStyle.Render(label),
		valStyle.Render(value),
		bar,
	)
}

func renderAnomalies(sessions []*aggregator.SessionStats) string {
	var lines []string
	for _, s := range sessions {
		if s.CacheEfficiency < 0.15 && s.TotalTokens > 500_000 {
			lines = append(lines, fmt.Sprintf("⚠  %s: %.0f%% cache — cold start. %s (%d msgs, %s)",
				Truncate(s.Summary, 20),
				s.CacheEfficiency*100,
				FormatCostPerMessage(s.CostUSD, s.MessageCount),
				s.MessageCount,
				FormatCost(s.CostUSD),
			))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return RenderPanel("anomalies", strings.Join(lines, "\n"), 100)
}
