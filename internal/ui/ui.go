package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type Theme struct {
	Red, Yellow, Green, Blue, Gray, Magenta, Cyan string
	BgRed, BgYellow, BgGreen, BgBlue              string
}

var theme Theme

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}

func InitTheme(name string) error {
	themes := map[string]Theme{
		"dark": {
			Red: "#E88388", Yellow: "#DBAB79", Green: "#A8CC8C",
			Blue: "#71BEF2", Gray: "#B9BFCA", Magenta: "#D290E4",
			Cyan: "#66C2CD", BgRed: "#2d1b1b", BgYellow: "#2d2d1b",
			BgGreen: "#1b2d1b", BgBlue: "#1b1b2d",
		},
		"light": {
			Red: "#D70000", Yellow: "#FFAF00", Green: "#005F00",
			Blue: "#000087", Gray: "#303030", Magenta: "#AF00FF",
			Cyan: "#0087FF", BgRed: "#ffdede", BgYellow: "#fff4d0",
			BgGreen: "#e6ffe6", BgBlue: "#dedeff",
		},
		"ansi": {
			Red: "#FF5555", Yellow: "#FFFF55", Green: "#55FF55",
			Blue: "#5555FF", Gray: "#AAAAAA", Magenta: "#FF55FF",
			Cyan: "#55FFFF", BgRed: "#AA0000", BgYellow: "#AAAA00",
			BgGreen: "#00AA00", BgBlue: "#0000AA",
		},
	}
	t, ok := themes[name]
	if !ok {
		return nil
	}
	theme = t
	initStyles()
	return nil
}

func Init() error {
	return InitTheme(defaultThemeName())
}

var (
	StyleBorder, StyleHeader, StyleDim, StyleGreen, StyleAmber, StyleRed, StyleBold lipgloss.Style
	StyleTagOpus, StyleTagSonnet, StyleTagHaiku, StyleTagGLM                        lipgloss.Style
)

func initStyles() {
	StyleBorder = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(theme.Gray))
	StyleHeader = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Gray)).Background(lipgloss.Color("#222230")).Padding(0, 1)
	StyleDim = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Gray))
	StyleGreen = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Green)).Bold(true)
	StyleAmber = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Yellow)).Bold(true)
	StyleRed = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Red)).Bold(true)
	StyleBold = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Gray)).Bold(true)
	StyleTagOpus = lipgloss.NewStyle().Background(lipgloss.Color(theme.BgYellow)).Foreground(lipgloss.Color(theme.Magenta)).Padding(0, 1)
	StyleTagSonnet = lipgloss.NewStyle().Background(lipgloss.Color(theme.BgBlue)).Foreground(lipgloss.Color(theme.Blue)).Padding(0, 1)
	StyleTagHaiku = lipgloss.NewStyle().Background(lipgloss.Color(theme.BgGreen)).Foreground(lipgloss.Color(theme.Green)).Padding(0, 1)
	StyleTagGLM = lipgloss.NewStyle().Background(lipgloss.Color(theme.BgYellow)).Foreground(lipgloss.Color(theme.Magenta)).Padding(0, 1)
}

func ModelTag(model string) string {
	switch {
	case strings.Contains(model, "opus"):
		return StyleTagOpus.Render("opus")
	case strings.Contains(model, "sonnet"):
		return StyleTagSonnet.Render("sonnet")
	case strings.Contains(model, "haiku"):
		return StyleTagHaiku.Render("haiku")
	case strings.Contains(model, "glm"):
		return StyleTagGLM.Render("glm")
	default:
		return StyleDim.Render(model)
	}
}

func CacheEfficiencyColor(eff float64) string {
	switch {
	case eff >= 0.80:
		return theme.Green
	case eff >= 0.40:
		return theme.Yellow
	default:
		return theme.Red
	}
}

type TokenBar struct {
	Input, Output, CacheCreate, CacheRead int64
	Width                                 int
}

func (b TokenBar) Render() string {
	total := b.Input + b.Output + b.CacheCreate + b.CacheRead
	if total == 0 || b.Width <= 0 {
		return lipgloss.NewStyle().Background(lipgloss.Color(theme.Gray)).Width(b.Width).Render("")
	}
	segs := []struct {
		count int64
		color string
	}{
		{b.Input, theme.Blue},
		{b.Output, theme.Cyan},
		{b.CacheCreate, theme.Yellow},
		{b.CacheRead, theme.Magenta},
	}
	widths := make([]int, len(segs))
	tw := 0
	for i, s := range segs {
		w := int(float64(s.count) / float64(total) * float64(b.Width))
		widths[i] = w
		tw += w
	}
	if d := b.Width - tw; d != 0 {
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
		parts = append(parts, lipgloss.NewStyle().Background(lipgloss.Color(s.color)).Width(widths[i]).Render(""))
	}
	return strings.Join(parts, "")
}

func MiniBar(ratio float64, width int, color string) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	f := int(ratio * float64(width))
	e := width - f
	return lipgloss.NewStyle().Background(lipgloss.Color(color)).Width(f).Render("") +
		lipgloss.NewStyle().Background(lipgloss.Color(theme.Gray)).Width(e).Render("")
}

func TokenSubtitle(input, output, cc, cr int64) string {
	return StyleDim.Render(fmt.Sprintf("%s in · %s out · %s cc · %s cr",
		HumanizeTokens(input), HumanizeTokens(output), HumanizeTokens(cc), HumanizeTokens(cr)))
}

func HumanizeTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.0fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func FormatCost(usd float64) string {
	if usd <= 0 {
		return "~"
	}
	return fmt.Sprintf("$%.2f", usd)
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	if h > 10000 {
		return "N/A"
	}
	return fmt.Sprintf("%dh%dm", h, int(d.Minutes())%60)
}

func FormatCostPerMessage(cost float64, msgs int) string {
	if msgs == 0 {
		return "N/A"
	}
	return fmt.Sprintf("$%.2f/msg", cost/float64(msgs))
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func PadRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func PadLeft(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-len(s)) + s
}

func Panel(title, content string, width int) string {
	header := StyleHeader.Render(" " + title + " ")
	body := lipgloss.NewStyle().Padding(0, 1).Render(content)
	return StyleBorder.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, header, body))
}

func Bordered(content string) string {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(theme.Gray)).Render(content)
}

func Separator(width int) string {
	return strings.Repeat("─", width)
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
		Input: s.InputTokens, Output: s.OutputTokens,
		CacheCreate: s.CacheCreateTokens, CacheRead: s.CacheReadTokens,
		Width: barWidth,
	}.Render()
	effColor := CacheEfficiencyColor(s.CacheEfficiency)
	subagentNote := ""
	if s.SubagentCount > 0 {
		subagentNote = StyleDim.Render(fmt.Sprintf(" (+%d subagents, %s tokens)", s.SubagentCount, HumanizeTokens(s.SubagentTokens)))
	}
	row := fmt.Sprintf("%s %s  %s  %s  %s",
		sessionName+" "+ModelTag(model),
		bar,
		lipgloss.NewStyle().Foreground(lipgloss.Color(effColor)).Bold(true).Render(cachePct),
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
	sessionContent := strings.Join(sessionRows, "\n"+Separator(termWidth-4)+"\n")
	sessionPanel := Panel(
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
		renderCard("total cost", FormatCost(totalCost), theme.Green),
		renderCard("tokens", HumanizeTokens(totalTokens), theme.Blue),
		renderCard("cache eff", fmt.Sprintf("%.0f%%", cacheEff*100), theme.Cyan),
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
			lines = append(lines, fmt.Sprintf("[!] %s: %.0f%% cache - cold start. %s (%d msgs, %s)",
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
	return Panel("anomalies", strings.Join(lines, "\n"), 100)
}
