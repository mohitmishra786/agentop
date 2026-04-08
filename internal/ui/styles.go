package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Theme holds the color palette for all UI elements
type Theme struct {
	// Bar segment colors — semantic meaning drives color choice
	BarInput       string // amber:  you're spending on input
	BarOutput      string // coral:  you're spending on output (priciest per token)
	BarCacheCreate string // yellow: one-time cache creation cost
	BarCacheRead   string // teal:   savings — calm, efficient, good
	BarEmpty       string // dim:    unused bar space

	// Panel structure
	Border     string // panel border color
	HeaderBg   string // header row background
	HeaderFg   string // header row text
	RowSep     string // thin row separator line

	// Text hierarchy
	TextPrimary   string // session names, costs — brightest
	TextSecondary string // paths, token breakdowns
	TextDim       string // separators, labels

	// Status colors (same meaning across all terminals)
	Green string
	Amber string
	Red   string

	// Model tag backgrounds + foregrounds
	TagGLMBg     string
	TagGLMFg     string
	TagOpusBg    string
	TagOpusFg    string
	TagSonnetBg  string
	TagSonnetFg  string
	TagHaikuBg   string
	TagHaikuFg   string
	TagUnknownBg string
	TagUnknownFg string
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
			// Bar segments: warm=cost, cool=savings
			BarInput:       "#E8A838", // amber
			BarOutput:      "#D65D3A", // coral/red
			BarCacheCreate: "#C9A227", // golden yellow
			BarCacheRead:   "#3A9BB5", // teal — CALM because high cache read = GOOD
			BarEmpty:       "#2A2A3A", // near-black

			// Structure
			Border:   "#3D3D5C",
			HeaderBg: "#1E1E2E",
			HeaderFg: "#8888BB",
			RowSep:   "#2D2D45",

			// Text
			TextPrimary:   "#DDDDEE",
			TextSecondary: "#7777AA",
			TextDim:       "#444466",

			// Status
			Green: "#73D673",
			Amber: "#E8A838",
			Red:   "#E05555",

			// GLM — green (Zhipu AI brand)
			TagGLMBg: "#1A3A22", TagGLMFg: "#5EC989",
			// Anthropic models
			TagOpusBg:   "#2A1A3A", TagOpusFg: "#C090F0",
			TagSonnetBg: "#1A2434", TagSonnetFg: "#70B0E0",
			TagHaikuBg:  "#1A2A1A", TagHaikuFg: "#70C880",
			// Unknown
			TagUnknownBg: "#252525", TagUnknownFg: "#606080",
		},
		"light": {
			BarInput:       "#D4891A",
			BarOutput:      "#B84020",
			BarCacheCreate: "#A88A10",
			BarCacheRead:   "#1A7A99",
			BarEmpty:       "#DDDDDD",

			Border:   "#AAAACC",
			HeaderBg: "#EEEEF8",
			HeaderFg: "#6666AA",
			RowSep:   "#CCCCDD",

			TextPrimary:   "#222233",
			TextSecondary: "#555577",
			TextDim:       "#AAAACC",

			Green: "#226622",
			Amber: "#884400",
			Red:   "#AA2222",

			TagGLMBg: "#C8EED4", TagGLMFg: "#1A5E2A",
			TagOpusBg:   "#E8D8F8", TagOpusFg: "#5A2090",
			TagSonnetBg: "#D0E4F8", TagSonnetFg: "#1A4A80",
			TagHaikuBg:  "#D0EED0", TagHaikuFg: "#1A5A1A",
			TagUnknownBg: "#EEEEEE", TagUnknownFg: "#666688",
		},
	}
	t, ok := themes[name]
	if !ok {
		return InitTheme("dark")
	}
	theme = t
	initStyles()
	return nil
}

func Init() error {
	return InitTheme(defaultThemeName())
}

// ---------------------------------------------------------------------------
// Lipgloss style vars — initialized by initStyles()
// ---------------------------------------------------------------------------

var (
	StyleBorder    lipgloss.Style
	StyleHeader    lipgloss.Style
	StyleColHeader lipgloss.Style
	StyleDim       lipgloss.Style
	StyleSecondary lipgloss.Style
	StylePrimary   lipgloss.Style
	StyleGreen     lipgloss.Style
	StyleAmber     lipgloss.Style
	StyleRed       lipgloss.Style
	StyleBold      lipgloss.Style
	StyleRowSep    lipgloss.Style
)

func initStyles() {
	StyleBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Border))

	StyleHeader = lipgloss.NewStyle().
		Background(lipgloss.Color(theme.HeaderBg)).
		Foreground(lipgloss.Color(theme.HeaderFg)).
		Padding(0, 1)

	StyleColHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.HeaderFg))

	StyleDim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextDim))

	StyleSecondary = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextSecondary))

	StylePrimary = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextPrimary))

	StyleGreen = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Green)).Bold(true)

	StyleAmber = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Amber)).Bold(true)

	StyleRed = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Red)).Bold(true)

	StyleBold = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextPrimary)).Bold(true)

	StyleRowSep = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.RowSep))
}

// ---------------------------------------------------------------------------
// Model tags
// ---------------------------------------------------------------------------

func ModelTag(model string) string {
	ml := strings.ToLower(model)
	switch {
	case strings.HasPrefix(ml, "glm"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagGLMBg)).
			Foreground(lipgloss.Color(theme.TagGLMFg)).
			Padding(0, 1).Render("glm")
	case strings.Contains(ml, "opus"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagOpusBg)).
			Foreground(lipgloss.Color(theme.TagOpusFg)).
			Padding(0, 1).Render("opus")
	case strings.Contains(ml, "sonnet"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagSonnetBg)).
			Foreground(lipgloss.Color(theme.TagSonnetFg)).
			Padding(0, 1).Render("sonnet")
	case strings.Contains(ml, "haiku"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagHaikuBg)).
			Foreground(lipgloss.Color(theme.TagHaikuFg)).
			Padding(0, 1).Render("haiku")
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagUnknownBg)).
			Foreground(lipgloss.Color(theme.TagUnknownFg)).
			Padding(0, 1).Render("?")
	}
}

// ---------------------------------------------------------------------------
// Cache efficiency color helper
// ---------------------------------------------------------------------------

func CacheEfficiencyColor(eff float64) string {
	switch {
	case eff >= 0.80:
		return theme.Green
	case eff >= 0.40:
		return theme.Amber
	default:
		return theme.Red
	}
}

// ---------------------------------------------------------------------------
// TokenBar — duf-style solid colored block bar
// ---------------------------------------------------------------------------

// TokenBar renders a duf-style token composition bar.
// Segments: amber=input, coral=output, yellow=cacheCreate, TEAL=cacheRead
// A 98% teal bar means 98% cache reads = efficient = calm looking.
type TokenBar struct {
	Input, Output, CacheCreate, CacheRead int64
	Width                                 int
}

func (b TokenBar) Render() string {
	total := b.Input + b.Output + b.CacheCreate + b.CacheRead
	innerW := b.Width - 2 // subtract 2 for [ and ]
	if innerW <= 0 {
		innerW = 1
	}

	if total == 0 || b.Width <= 0 {
		empty := strings.Repeat("░", innerW)
		return StyleDim.Render("[") + lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.BarEmpty)).
			Render(empty) + StyleDim.Render("]")
	}

	type seg struct {
		count int64
		color string
	}
	segs := []seg{
		{b.Input, theme.BarInput},
		{b.Output, theme.BarOutput},
		{b.CacheCreate, theme.BarCacheCreate},
		{b.CacheRead, theme.BarCacheRead}, // teal — NOT magenta
	}

	// Count non-zero segments
	nonZeroSegs := 0
	for _, s := range segs {
		if s.count > 0 {
			nonZeroSegs++
		}
	}

	widths := make([]int, len(segs))
	tw := 0
	for i, s := range segs {
		if s.count == 0 {
			widths[i] = 0
			continue
		}
		// Reserve at least 1 for each non-zero segment
		w := int(float64(s.count) / float64(total) * float64(innerW))
		if w < 1 {
			w = 1
		}
		widths[i] = w
		tw += w
	}
	// Distribute remainder to largest segments
	if diff := innerW - tw; diff > 0 {
		for i := 0; i < diff; i++ {
			// Find segment with largest count that's not the last one
			mi := 0
			maxCount := segs[0].count
			for j := 1; j < len(segs); j++ {
				if segs[j].count > maxCount && widths[j] > 0 {
					mi = j
					maxCount = segs[j].count
				}
			}
			widths[mi]++
		}
	}

	var parts []string
	for i, s := range segs {
		if widths[i] <= 0 {
			continue
		}
		// Use foreground color with block characters for better visibility
		bar := strings.Repeat("█", widths[i])
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.Color(s.color)).
			Render(bar))
	}

	bar := strings.Join(parts, "")
	return StyleDim.Render("[") + bar + StyleDim.Render("]")
}

// BarLegend returns the one-line color key to show under the panel header
func BarLegend() string {
	dot := func(col, label string) string {
		sq := lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render("█")
		return sq + StyleSecondary.Render(label)
	}
	return strings.Join([]string{
		dot(theme.BarInput, " in"),
		dot(theme.BarOutput, " out"),
		dot(theme.BarCacheCreate, " cc"),
		dot(theme.BarCacheRead, " cr"),
	}, " ")
}

// ---------------------------------------------------------------------------
// MiniBar — small summary bar for the top summary strip
// ---------------------------------------------------------------------------

func MiniBar(ratio float64, width int, fillColor string) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	f := int(ratio * float64(width))
	if f > width {
		f = width
	}
	e := width - f
	filled := lipgloss.NewStyle().Background(lipgloss.Color(fillColor)).Width(f).Render("")
	empty := lipgloss.NewStyle().Background(lipgloss.Color(theme.BarEmpty)).Width(e).Render("")
	return filled + empty
}

// ---------------------------------------------------------------------------
// Token subtitle line
// ---------------------------------------------------------------------------

func TokenSubtitle(input, output, cc, cr int64) string {
	parts := []string{
		HumanizeTokens(input) + " in",
		HumanizeTokens(output) + " out",
	}
	if cc > 0 {
		parts = append(parts, HumanizeTokens(cc)+" cc")
	}
	parts = append(parts, HumanizeTokens(cr)+" cr")
	return StyleSecondary.Render(strings.Join(parts, " · "))
}

// ---------------------------------------------------------------------------
// Formatters
// ---------------------------------------------------------------------------

func HumanizeTokens(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
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
		return StyleDim.Render("~")
	}
	if usd >= 10 {
		return StyleAmber.Render(fmt.Sprintf("$%.2f", usd))
	}
	return StylePrimary.Render(fmt.Sprintf("$%.2f", usd))
}

func FormatCostRaw(usd float64) string {
	if usd <= 0 {
		return "~"
	}
	return fmt.Sprintf("$%.2f", usd)
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return StyleDim.Render("0s")
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	if h > 10000 {
		return StyleDim.Render("N/A")
	}
	return fmt.Sprintf("%dh%dm", h, int(d.Minutes())%60)
}

func FormatCostPerMessage(cost float64, msgs int) string {
	if msgs == 0 || cost <= 0 {
		return StyleDim.Render("N/A")
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
	vis := lipgloss.Width(s) // use lipgloss width to handle ANSI escape codes
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

func PadLeft(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return s
	}
	return strings.Repeat(" ", w-vis) + s
}

func Separator(width int) string {
	return StyleRowSep.Render(strings.Repeat("─", width))
}

// ---------------------------------------------------------------------------
// Panel wrapper — duf-style rounded border with header
// ---------------------------------------------------------------------------

func Panel(title, content string, width int) string {
	// Constrain width to fit terminal
	maxWidth := width
	if maxWidth > 120 {
		maxWidth = 120
	}

	lines := strings.Split(content, "\n")
	// Process each line to fit within max width
	var processedLines []string
	for _, line := range lines {
		// Calculate actual visual width (accounting for ANSI codes)
		visWidth := lipgloss.Width(line)
		if visWidth > maxWidth-4 {
			// Truncate if too long
			processedLines = append(processedLines, Truncate(line, maxWidth-6))
		} else {
			processedLines = append(processedLines, line)
		}
	}

	// Find max line length for panel sizing
	maxLen := 0
	for _, line := range processedLines {
		visWidth := lipgloss.Width(line)
		if visWidth > maxLen {
			maxLen = visWidth
		}
	}

	innerWidth := maxLen + 2
	if innerWidth < len(title)+4 {
		innerWidth = len(title) + 4
	}
	if innerWidth > maxWidth-4 {
		innerWidth = maxWidth - 4
	}

	// Build duf-style panel
	var result []string
	titlePadding := innerWidth - len(title) - 4
	if titlePadding < 0 {
		titlePadding = 0
	}
	result = append(result, "╭─ "+StyleHeader.Render(title)+" "+strings.Repeat("─", titlePadding)+"╮")

	for _, line := range processedLines {
		visWidth := lipgloss.Width(line)
		padding := innerWidth - 2 - visWidth
		if padding < 0 {
			padding = 0
		}
		result = append(result, "│ "+line+strings.Repeat(" ", padding)+"│")
	}
	result = append(result, "╰"+strings.Repeat("─", innerWidth)+"╯")

	return strings.Join(result, "\n")
}

// ---------------------------------------------------------------------------
// Session row — duf-style two-line row with proper column layout
// ---------------------------------------------------------------------------

func RenderSessionRow(s *aggregator.SessionStats, barWidth int, termWidth int) string {
	// ── Line 1: ID + tag | bar | cache% | cost | duration ──────────────────

	// Session identifier (8-char hash or truncated summary)
	sessionID := s.ID
	if len(sessionID) >= 8 {
		sessionID = sessionID[:8]
	}
	name := s.Summary
	if name == "" {
		name = sessionID
	}
	name = Truncate(name, 16) // Shorter to fit better
	nameStr := StylePrimary.Render(name)

	// Model badge
	model := s.Model
	if model == "" {
		model = "?"
	}
	tag := ModelTag(model)

	// Left cell: name + tag, padded to fixed width
	leftWidth := 24 // Reduced width
	left := PadRight(nameStr+" "+tag, leftWidth)

	// Token bar
	bar := TokenBar{
		Input:       s.InputTokens, Output: s.OutputTokens,
		CacheCreate: s.CacheCreateTokens, CacheRead: s.CacheReadTokens,
		Width:       barWidth,
	}.Render()

	// Cache efficiency — right-aligned 4 chars + color
	var cacheStr string
	if s.TotalTokens == 0 {
		cacheStr = StyleDim.Render(PadLeft("~", 4))
	} else {
		eff := s.CacheEfficiency * 100
		effStyled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(CacheEfficiencyColor(s.CacheEfficiency))).
			Bold(true).
			Render(fmt.Sprintf("%.0f%%", eff))
		cacheStr = PadLeft(effStyled, 4)
	}

	// Cost — right-aligned 6 chars
	costStr := PadLeft(FormatCostRaw(s.CostUSD), 6)
	if s.CostUSD <= 0 {
		costStr = PadLeft(StyleDim.Render("~"), 6)
	} else if s.CostUSD >= 10 {
		costStr = PadLeft(StyleAmber.Render(fmt.Sprintf("$%.2f", s.CostUSD)), 6)
	} else {
		costStr = PadLeft(StylePrimary.Render(fmt.Sprintf("$%.2f", s.CostUSD)), 6)
	}

	// Duration — right-aligned 5 chars
	durStr := PadLeft(FormatDuration(s.Duration), 5)

	line1 := left + " " + bar + " " + cacheStr + " " + costStr + " " + durStr

	// ── Line 2: project path + token breakdown (dim) ────────────────────────

	// Project path — truncated, dimmer
	projectStr := ""
	if s.ProjectPath != "" {
		projectStr = StyleSecondary.Render(Truncate(s.ProjectPath, 30))
	} else if s.ProjectName != "" {
		projectStr = StyleSecondary.Render(Truncate(s.ProjectName, 30))
	}

	tokenBreak := TokenSubtitle(s.InputTokens, s.OutputTokens, s.CacheCreateTokens, s.CacheReadTokens)

	// Indent line 2 to align with bar start
	indent := strings.Repeat(" ", leftWidth+1)
	line2 := indent + tokenBreak
	if projectStr != "" {
		line2 = StyleSecondary.Render(Truncate(s.ProjectPath, 30)) + "  " + tokenBreak
	}

	// ── Line 3: subagent note (if any) ─────────────────────────────────────
	line3 := ""
	if s.SubagentCount > 0 {
		line3 = StyleDim.Render(
			fmt.Sprintf("  +%d subagents, %s tokens", s.SubagentCount, HumanizeTokens(s.SubagentTokens)))
	}

	result := line1 + "\n" + line2
	if line3 != "" {
		result += "\n" + line3
	}
	return result
}

// ---------------------------------------------------------------------------
// Today view — the main output
// ---------------------------------------------------------------------------

func RenderToday(sessions []*aggregator.SessionStats, termWidth int) string {
	if len(sessions) == 0 {
		return StyleDim.Render("  No sessions found.")
	}

	// ── Aggregate summary stats ──────────────────────────────────────────────
	var totalCost float64
	var totalTokens int64
	var totalCacheRead int64
	var totalCacheCreate int64
	var totalInput int64
	var sessCount int

	for _, s := range sessions {
		totalCost += s.CostUSD
		totalTokens += s.TotalTokens
		totalCacheRead += s.CacheReadTokens
		totalCacheCreate += s.CacheCreateTokens
		totalInput += s.InputTokens
		sessCount++
	}

	// Overall cache efficiency = cache read / (input + cacheCreate + cacheRead)
	denom := totalInput + totalCacheCreate + totalCacheRead
	overallCacheEff := 0.0
	if denom > 0 {
		overallCacheEff = float64(totalCacheRead) / float64(denom)
	}

	// ── Summary strip ────────────────────────────────────────────────────────
	summary := renderSummaryStrip(totalCost, totalTokens, overallCacheEff, sessCount)

	// ── Session panel ────────────────────────────────────────────────────────
	// Use fixed, reasonable column widths
	leftWidth := 24
	barWidth := 20 // Fixed, reasonable bar width
	cacheWidth := 4
	costWidth := 6
	timeWidth := 5

	// Calculate total content width
	contentWidth := leftWidth + 2 + barWidth + 2 + cacheWidth + 2 + costWidth + 2 + timeWidth
	maxContentWidth := termWidth - 8 // Leave room for panel borders
	if contentWidth > maxContentWidth {
		contentWidth = maxContentWidth
		// Adjust bar width if needed
		barWidth = contentWidth - leftWidth - cacheWidth - costWidth - timeWidth - 10
		if barWidth < 15 {
			barWidth = 15
		}
	}

	headerCols := PadRight(StyleColHeader.Render("Session"), leftWidth) +
		PadRight(StyleColHeader.Render("Tokens"), barWidth+2) +
		PadLeft(StyleColHeader.Render("Cache"), cacheWidth) +
		PadLeft(StyleColHeader.Render("Cost"), costWidth) +
		PadLeft(StyleColHeader.Render("Time"), timeWidth)

	var rowParts []string
	// Add legend + column headers before rows
	rowParts = append(rowParts, BarLegend())
	rowParts = append(rowParts, StyleRowSep.Render(strings.Repeat("─", contentWidth)))
	rowParts = append(rowParts, headerCols)
	rowParts = append(rowParts, StyleRowSep.Render(strings.Repeat("─", contentWidth)))

	for i, s := range sessions {
		rowParts = append(rowParts, RenderSessionRow(s, barWidth, termWidth))
		if i < len(sessions)-1 {
			rowParts = append(rowParts, StyleRowSep.Render(strings.Repeat("─", contentWidth)))
		}
	}

	sessionContent := strings.Join(rowParts, "\n")
	panelTitle := fmt.Sprintf("claude code · %d sessions", len(sessions))
	sessionPanel := Panel(panelTitle, sessionContent, termWidth-2)

	// ── Anomalies panel (only if issues found) ───────────────────────────────
	anomalies := renderAnomalies(sessions)

	output := summary + "\n\n" + sessionPanel
	if anomalies != "" {
		output += "\n\n" + anomalies
	}
	return output
}

// renderSummaryStrip — the 4-metric cards at the top
func renderSummaryStrip(totalCost float64, totalTokens int64, cacheEff float64, sessCount int) string {
	// Normalize ratios for mini bars
	// Cost bar: scale to $100 as "full"
	costRatio := totalCost / 100.0
	if costRatio > 1 {
		costRatio = 1
	}

	// Tokens bar: scale to 1B as "full"
	tokenRatio := float64(totalTokens) / 1_000_000_000.0
	if tokenRatio > 1 {
		tokenRatio = 1
	}

	// Cache efficiency: direct ratio
	cacheRatio := cacheEff

	cardWidth := 18
	miniW := cardWidth - 2

	card := func(label, value, valColor string, ratio float64, barColor string) string {
		lbl := StyleColHeader.Width(cardWidth).Render(label)
		val := lipgloss.NewStyle().Foreground(lipgloss.Color(valColor)).Bold(true).Width(cardWidth).Render(value)
		bar := MiniBar(ratio, miniW, barColor)
		return lipgloss.JoinVertical(lipgloss.Left, lbl, val, bar)
	}

	costVal := "~"
	if totalCost > 0 {
		costVal = fmt.Sprintf("$%.2f", totalCost)
	}

	cacheColor := theme.Red
	if cacheEff >= 0.80 {
		cacheColor = theme.Green
	} else if cacheEff >= 0.40 {
		cacheColor = theme.Amber
	}

	cards := []string{
		card("  total cost", costVal, theme.Amber, costRatio, theme.BarInput),
		card("  tokens", HumanizeTokens(totalTokens), "#4A90D9", tokenRatio, "#4A90D9"),
		card("  cache eff", fmt.Sprintf("%.0f%%", cacheEff*100), cacheColor, cacheRatio, cacheColor),
		card("  sessions", fmt.Sprintf("%d", sessCount), theme.TextPrimary, float64(sessCount)/20.0, theme.TextSecondary),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

// renderAnomalies — anomaly panel, only shown if issues exist
func renderAnomalies(sessions []*aggregator.SessionStats) string {
	var lines []string
	for _, s := range sessions {
		// Cold start / low cache on non-trivial session
		if s.CacheEfficiency < 0.15 && s.TotalTokens > 500_000 {
			name := s.ID
			if len(name) > 8 {
				name = name[:8]
			}
			lines = append(lines, StyleAmber.Render("⚠ ")+StylePrimary.Render(name)+
				StyleSecondary.Render(fmt.Sprintf(": %.0f%% cache · %d msgs · %s",
					s.CacheEfficiency*100, s.MessageCount, FormatCostRaw(s.CostUSD))))
		}
		// Heavy model, short session
		model := strings.ToLower(s.Model)
		if strings.Contains(model, "opus") && s.MessageCount < 5 && s.CostUSD > 1.0 {
			name := s.ID
			if len(name) > 8 {
				name = name[:8]
			}
			lines = append(lines, StyleAmber.Render("⚠ ")+StylePrimary.Render(name)+
				StyleSecondary.Render(": opus used for "+fmt.Sprintf("%d", s.MessageCount)+
					"-message session · consider sonnet"))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return Panel("anomalies & insights", strings.Join(lines, "\n"), 100)
}
