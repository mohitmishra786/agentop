package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

// Theme holds the color palette for all UI elements
type Theme struct {
	BarInput       string
	BarOutput      string
	BarCacheCreate string
	BarCacheRead   string
	BarEmpty       string

	Border   string
	HeaderBg string
	HeaderFg string
	RowSep   string

	TextPrimary   string
	TextSecondary string
	TextDim       string

	Green string
	Amber string
	Red   string

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

	TagClaudeBg  string
	TagClaudeFg  string
	TagCodexBg   string
	TagCodexFg   string
	TagCursorBg  string
	TagCursorFg  string
	TagCopilotBg string
	TagCopilotFg string
	TagGeminiBg  string
	TagGeminiFg  string
	TagKiroBg    string
	TagKiroFg    string
	TagOpenCodeBg  string
	TagOpenCodeFg  string
	TagJetBrainsBg string
	TagJetBrainsFg string
	TagContinueBg  string
	TagContinueFg  string
	TagWindsurfBg  string
	TagWindsurfFg  string
	TagGrokBg      string
	TagGrokFg      string
	TagDeepseekBg  string
	TagDeepseekFg  string
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
			BarInput:       "#E8A838",
			BarOutput:      "#D65D3A",
			BarCacheCreate: "#C9A227",
			BarCacheRead:   "#3A9BB5",
			BarEmpty:       "#2A2A3A",

			Border:   "#3D3D5C",
			HeaderBg: "#1E1E2E",
			HeaderFg: "#8888BB",
			RowSep:   "#2D2D45",

			TextPrimary:   "#DDDDEE",
			TextSecondary: "#7777AA",
			TextDim:       "#444466",

			Green: "#73D673",
			Amber: "#E8A838",
			Red:   "#E05555",

			TagGLMBg: "#1A3A22", TagGLMFg: "#5EC989",
			TagOpusBg: "#2A1A3A", TagOpusFg: "#C090F0",
			TagSonnetBg: "#1A2434", TagSonnetFg: "#70B0E0",
			TagHaikuBg: "#1A2A1A", TagHaikuFg: "#70C880",
			TagUnknownBg: "#252525", TagUnknownFg: "#606080",

			TagClaudeBg: "#1A1A3A", TagClaudeFg: "#9090E0",
			TagCodexBg: "#2A1A1A", TagCodexFg: "#E08060",
			TagCursorBg: "#1A2A2A", TagCursorFg: "#60C0B0",
			TagCopilotBg: "#221A2A", TagCopilotFg: "#B080D0",
			TagGeminiBg: "#1A2A1A", TagGeminiFg: "#80C880",
			TagKiroBg: "#2A2A1A", TagKiroFg: "#C8B860",
			TagOpenCodeBg: "#1A1A2A", TagOpenCodeFg: "#8080D0",
			TagJetBrainsBg: "#2A1A2A", TagJetBrainsFg: "#D080B0",
			TagContinueBg: "#1A221A", TagContinueFg: "#70B070",
			TagWindsurfBg: "#222A1A", TagWindsurfFg: "#A0C060",
			TagGrokBg: "#1A1A2A", TagGrokFg: "#C0C0E0",
			TagDeepseekBg: "#1A221A", TagDeepseekFg: "#80C080",
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
			TagOpusBg: "#E8D8F8", TagOpusFg: "#5A2090",
			TagSonnetBg: "#D0E4F8", TagSonnetFg: "#1A4A80",
			TagHaikuBg: "#D0EED0", TagHaikuFg: "#1A5A1A",
			TagUnknownBg: "#EEEEEE", TagUnknownFg: "#666688",

			TagClaudeBg: "#D0D0F0", TagClaudeFg: "#202080",
			TagCodexBg: "#F0D0C0", TagCodexFg: "#803020",
			TagCursorBg: "#C0E8E0", TagCursorFg: "#106050",
			TagCopilotBg: "#E0D0F0", TagCopilotFg: "#603080",
			TagGeminiBg: "#C8E8C8", TagGeminiFg: "#106010",
			TagKiroBg: "#F0F0C8", TagKiroFg: "#606010",
			TagOpenCodeBg: "#D0D0F0", TagOpenCodeFg: "#3030A0",
			TagJetBrainsBg: "#F0D0E0", TagJetBrainsFg: "#802060",
			TagContinueBg: "#C8E0C8", TagContinueFg: "#206020",
			TagWindsurfBg: "#E8F0C8", TagWindsurfFg: "#506010",
			TagGrokBg: "#D0D0E8", TagGrokFg: "#303060",
			TagDeepseekBg: "#C8E8C8", TagDeepseekFg: "#106010",
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
// Lipgloss style vars
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

// modelVersion extracts the version string from a Claude model ID.
// "claude-opus-4-6" → "4.6", "claude-haiku-4-5-20251001" → "4.5".
func modelVersion(model string) string {
	ml := strings.ToLower(model)
	parts := strings.Split(ml, "-")
	for i, p := range parts {
		switch p {
		case "opus", "sonnet", "haiku":
			if i+2 < len(parts) && isDigitStr(parts[i+1]) && isDigitStr(parts[i+2]) {
				return parts[i+1] + "." + parts[i+2]
			}
			return ""
		}
	}
	// "deepseek-v4-flash-free" → "v4"
	if strings.Contains(ml, "deepseek") {
		for _, p := range parts {
			if p == "v4" || p == "v5" || p == "v3" {
				return p
			}
		}
	}
	// "grok-composer-2.5-fast" → "2.5"
	if strings.Contains(ml, "grok") {
		for i, p := range parts {
			if i > 0 && isDigitStr(p) && i+1 < len(parts) && isDigitStr(parts[i+1]) {
				return p + "." + parts[i+1]
			}
		}
	}
	return ""
}

func isDigitStr(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ModelTag returns a coloured badge that includes the version when available.
// e.g. "claude-sonnet-4-6" → " sonnet 4.6 ", "claude-opus-4-7" → " opus 4.7 ".
func ModelTag(model string) string {
	ml := strings.ToLower(model)
	ver := modelVersion(model)
	label := func(family string) string {
		if ver != "" {
			return family + " " + ver
		}
		return family
	}
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
			Padding(0, 1).Render(label("opus"))
	case strings.Contains(ml, "sonnet"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagSonnetBg)).
			Foreground(lipgloss.Color(theme.TagSonnetFg)).
			Padding(0, 1).Render(label("sonnet"))
	case strings.Contains(ml, "haiku"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagHaikuBg)).
			Foreground(lipgloss.Color(theme.TagHaikuFg)).
			Padding(0, 1).Render(label("haiku"))
	case strings.Contains(ml, "deepseek"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagDeepseekBg)).
			Foreground(lipgloss.Color(theme.TagDeepseekFg)).
			Padding(0, 1).Render(label("deepseek"))
	case strings.Contains(ml, "grok"):
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagGrokBg)).
			Foreground(lipgloss.Color(theme.TagGrokFg)).
			Padding(0, 1).Render(label("grok"))
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagUnknownBg)).
			Foreground(lipgloss.Color(theme.TagUnknownFg)).
			Padding(0, 1).Render("?")
	}
}

// AgentTag returns a coloured badge for the AI coding assistant that produced the session.
func AgentTag(agentID string) string {
	type agentBadge struct {
		bg, fg string
		label  string
	}
	badges := map[string]agentBadge{
		"claude":    {theme.TagClaudeBg, theme.TagClaudeFg, "claude"},
		"codex":     {theme.TagCodexBg, theme.TagCodexFg, "codex"},
		"cursor":    {theme.TagCursorBg, theme.TagCursorFg, "cursor"},
		"copilot":   {theme.TagCopilotBg, theme.TagCopilotFg, "copilot"},
		"gemini":    {theme.TagGeminiBg, theme.TagGeminiFg, "gemini"},
		"kiro":      {theme.TagKiroBg, theme.TagKiroFg, "kiro"},
		"opencode":  {theme.TagOpenCodeBg, theme.TagOpenCodeFg, "opencode"},
		"jetbrains": {theme.TagJetBrainsBg, theme.TagJetBrainsFg, "jetbrains"},
		"continue":  {theme.TagContinueBg, theme.TagContinueFg, "continue"},
		"windsurf":  {theme.TagWindsurfBg, theme.TagWindsurfFg, "windsurf"},
		"grok":      {theme.TagGrokBg, theme.TagGrokFg, "grok"},
	}
	b, ok := badges[agentID]
	if !ok {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(theme.TagUnknownBg)).
			Foreground(lipgloss.Color(theme.TagUnknownFg)).
			Padding(0, 1).Render("?")
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(b.bg)).
		Foreground(lipgloss.Color(b.fg)).
		Padding(0, 1).Render(b.label)
}

// ModelCell delegates to ModelTag (version is now embedded in the badge).
func ModelCell(model string) string {
	return ModelTag(model)
}

// TokenTilde returns "~" suffix if tokens are estimated, empty otherwise.
// Use it alongside token values to indicate estimated counts.
func TokenTilde(src adapter.TokenSource) string {
	if src == adapter.TokenEstimated {
		return StyleDim.Render("~")
	}
	return ""
}

// FormatTokenWithSource formats a token count with optional ~ suffix.
func FormatTokenWithSource(n int64, src adapter.TokenSource) string {
	suffix := ""
	if n > 0 && src == adapter.TokenEstimated {
		suffix = "~"
	}
	return HumanizeTokens(n) + suffix
}

// ---------------------------------------------------------------------------
// Cache efficiency
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
// TokenBar — proportional colored bar (duf-style)
// ---------------------------------------------------------------------------

type TokenBar struct {
	Input, Output, CacheCreate, CacheRead int64
	Width                                 int
}

func (b TokenBar) Render() string {
	total := b.Input + b.Output + b.CacheCreate + b.CacheRead
	innerW := b.Width - 2
	if innerW <= 0 {
		innerW = 1
	}

	if total == 0 || b.Width <= 0 {
		return StyleDim.Render("[") +
			lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BarEmpty)).Render(strings.Repeat("░", innerW)) +
			StyleDim.Render("]")
	}

	type seg struct {
		count int64
		color string
	}
	segs := []seg{
		{b.Input, theme.BarInput},
		{b.Output, theme.BarOutput},
		{b.CacheCreate, theme.BarCacheCreate},
		{b.CacheRead, theme.BarCacheRead},
	}

	widths := make([]int, len(segs))
	tw := 0
	for i, s := range segs {
		if s.count == 0 {
			continue
		}
		w := int(float64(s.count) / float64(total) * float64(innerW))
		if w < 1 {
			w = 1
		}
		widths[i] = w
		tw += w
	}

	// Trim overshoot (minimum-1 enforcement can push tw > innerW)
	for tw > innerW {
		mi := 0
		for j := 1; j < len(segs); j++ {
			if widths[j] > widths[mi] {
				mi = j
			}
		}
		if widths[mi] <= 1 {
			break
		}
		widths[mi]--
		tw--
	}
	// Distribute any remaining space to the largest segment
	if diff := innerW - tw; diff > 0 {
		mi := 0
		for j := 1; j < len(segs); j++ {
			if segs[j].count > segs[mi].count {
				mi = j
			}
		}
		widths[mi] += diff
	}

	var parts []string
	for i, s := range segs {
		if widths[i] > 0 {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(lipgloss.Color(s.color)).
				Render(strings.Repeat("█", widths[i])))
		}
	}
	return StyleDim.Render("[") + strings.Join(parts, "") + StyleDim.Render("]")
}

// BarLegend returns the color key line.
func BarLegend() string {
	dot := func(col, label string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render("█") +
			StyleDim.Render(label)
	}
	return strings.Join([]string{
		dot(theme.BarInput, " in"),
		dot(theme.BarOutput, " out"),
		dot(theme.BarCacheCreate, " cc"),
		dot(theme.BarCacheRead, " cr"),
	}, "  ")
}

// ---------------------------------------------------------------------------
// MiniBar — small bar for the summary strip
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
	filled := lipgloss.NewStyle().Background(lipgloss.Color(fillColor)).Width(f).Render("")
	empty := lipgloss.NewStyle().Background(lipgloss.Color(theme.BarEmpty)).Width(width - f).Render("")
	return filled + empty
}

// ---------------------------------------------------------------------------
// Token subtitles
// ---------------------------------------------------------------------------

// TokenSubtitle returns plain text (used for width measurement).
func TokenSubtitle(input, output, cc, cr int64) string {
	parts := []string{
		HumanizeTokens(input) + " in",
		HumanizeTokens(output) + " out",
	}
	if cc > 0 {
		parts = append(parts, HumanizeTokens(cc)+" cc")
	}
	parts = append(parts, HumanizeTokens(cr)+" cr")
	return strings.Join(parts, "  ")
}

// TokenSubtitleColored renders each metric in its bar-segment color with label.
// If src is TokenEstimated, token values get a ~ suffix.
func TokenSubtitleColored(input, output, cc, cr int64, src adapter.TokenSource) string {
	col := func(color, num, label string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(num) +
			StyleDim.Render(" "+label)
	}
	tilde := func(n int64) string {
		if n > 0 && src == adapter.TokenEstimated {
			return StyleDim.Render("~")
		}
		return ""
	}
	parts := []string{
		col(theme.BarInput, HumanizeTokens(input)+tilde(input), "in"),
		col(theme.BarOutput, HumanizeTokens(output)+tilde(output), "out"),
	}
	if cc > 0 {
		parts = append(parts, col(theme.BarCacheCreate, HumanizeTokens(cc)+tilde(cc), "cc"))
	}
	parts = append(parts, col(theme.BarCacheRead, HumanizeTokens(cr)+tilde(cr), "cr"))
	return strings.Join(parts, StyleDim.Render("  "))
}

// TokenCompact renders just the colored token values (no labels) for tight columns.
// Colors match the bar legend in the title so the mapping is self-evident.
func TokenCompact(input, output, cc, cr int64, src adapter.TokenSource) string {
	col := func(color, num string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(num)
	}
	tilde := func(n int64) string {
		if n > 0 && src == adapter.TokenEstimated {
			return StyleDim.Render("~")
		}
		return ""
	}
	parts := []string{
		col(theme.BarInput, HumanizeTokens(input)+tilde(input)),
		col(theme.BarOutput, HumanizeTokens(output)+tilde(output)),
	}
	if cc > 0 {
		parts = append(parts, col(theme.BarCacheCreate, HumanizeTokens(cc)+tilde(cc)))
	}
	parts = append(parts, col(theme.BarCacheRead, HumanizeTokens(cr)+tilde(cr)))
	return strings.Join(parts, StyleDim.Render("  "))
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
	vis := lipgloss.Width(s)
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

// abbreviatePath replaces $HOME with ~ and trims to fit maxLen showing the tail.
func abbreviatePath(path string, maxLen int) string {
	if path == "" {
		return ""
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
	}
	// Normalise to forward slashes for display (handles Windows paths).
	path = filepath.ToSlash(path)
	if len(path) <= maxLen {
		return path
	}
	tail := path[len(path)-(maxLen-1):]
	if idx := strings.Index(tail, "/"); idx >= 0 {
		tail = tail[idx:]
	}
	return "…" + tail
}

// sessionDisplayName returns the best display name for a session (max 22 chars).
func sessionDisplayName(s *aggregator.SessionStats) string {
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
	return Truncate(name, 22)
}

// ---------------------------------------------------------------------------
// Panel — simple bordered box used only for anomalies
// ---------------------------------------------------------------------------

func Panel(title, content string, width int) string {
	maxWidth := width
	if maxWidth > 120 {
		maxWidth = 120
	}

	lines := strings.Split(content, "\n")
	var processedLines []string
	for _, line := range lines {
		if lipgloss.Width(line) > maxWidth-4 {
			processedLines = append(processedLines, Truncate(line, maxWidth-6))
		} else {
			processedLines = append(processedLines, line)
		}
	}

	maxLen := 0
	for _, line := range processedLines {
		if w := lipgloss.Width(line); w > maxLen {
			maxLen = w
		}
	}

	innerWidth := maxLen + 2
	titleVisWidth := lipgloss.Width(title)
	if innerWidth < titleVisWidth+6 {
		innerWidth = titleVisWidth + 6
	}
	if innerWidth > maxWidth-2 {
		innerWidth = maxWidth - 2
	}

	// StyleHeader has Padding(0,1) so rendered width = titleVisWidth+2
	// Top total = 2(╭─) + 1( ) + (titleVisWidth+2) + 1( ) + pad + 1(╮) = titleVisWidth+pad+7
	// Need = innerWidth+2  →  pad = innerWidth - titleVisWidth - 5
	titlePadding := innerWidth - titleVisWidth - 5
	if titlePadding < 0 {
		titlePadding = 0
	}
	bc := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Border))
	top := bc.Render("╭─") + " " + StyleHeader.Render(title) + " " +
		bc.Render(strings.Repeat("─", titlePadding)+"╮")

	var rows []string
	rows = append(rows, top)
	for _, line := range processedLines {
		padding := innerWidth - 2 - lipgloss.Width(line)
		if padding < 0 {
			padding = 0
		}
		rows = append(rows, bc.Render("│")+" "+line+strings.Repeat(" ", padding)+" "+bc.Render("│"))
	}
	rows = append(rows, bc.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(rows, "\n")
}

// ---------------------------------------------------------------------------
// duf-style table
// ---------------------------------------------------------------------------

// dufStyle returns a table style with rounded corners and row separators
// matching duf's clean column-grid aesthetic.
func dufStyle() table.Style {
	s := table.StyleRounded
	s.Options.SeparateRows = true
	s.Options.SeparateColumns = true
	s.Options.DrawBorder = true
	// Dim the border color by embedding it in a no-op; go-pretty uses ANSI for colors
	// separately, so we just keep the default border chars.
	return s
}

// ---------------------------------------------------------------------------
// Today view — duf-style grid table
// ---------------------------------------------------------------------------

func RenderToday(sessions []*aggregator.SessionStats, termWidth int) string {
	if len(sessions) == 0 {
		return StyleDim.Render("  No sessions found.")
	}

	// Split active vs empty
	var active, empty []*aggregator.SessionStats
	for _, s := range sessions {
		if s.TotalTokens == 0 {
			empty = append(empty, s)
		} else {
			active = append(active, s)
		}
	}

	// Aggregate totals
	var totalCost float64
	var totalTokens, totalCacheRead, totalCacheCreate, totalInput int64
	for _, s := range sessions {
		totalCost += s.CostUSD
		totalTokens += s.TotalTokens
		totalCacheRead += s.CacheReadTokens
		totalCacheCreate += s.CacheCreateTokens
		totalInput += s.InputTokens
	}
	denom := totalInput + totalCacheCreate + totalCacheRead
	overallCacheEff := 0.0
	if denom > 0 {
		overallCacheEff = float64(totalCacheRead) / float64(denom)
	}

	summary := renderSummaryStrip(totalCost, totalTokens, overallCacheEff, len(sessions))

	// ── Build table ───────────────────────────────────────────────────────────
	tab := table.NewWriter()
	tab.SetAllowedRowLength(termWidth)
	tab.SetStyle(dufStyle())

	// Title row: panel name + color legend
	sessionWord := "sessions"
	if len(active) == 1 {
		sessionWord = "session"
	}
	emptyHint := ""
	if len(empty) > 0 {
		emptyHint = fmt.Sprintf("  +%d empty", len(empty))
	}
	// Detect unique agent IDs among sessions.
	agentLabel := "all agents"
	agentIDs := map[string]bool{}
	for _, s := range sessions {
		if string(s.AgentID) != "" {
			agentIDs[string(s.AgentID)] = true
		}
	}
	if len(agentIDs) == 1 {
		for id := range agentIDs {
			agentLabel = AgentTag(id)
		}
	}
	panelTitle := fmt.Sprintf(" %s · %d %s%s   %s ",
		agentLabel, len(active), sessionWord, emptyHint, BarLegend())
	tab.SetTitle(panelTitle)

	// 5-column layout matching --layout table:
	//   SESSION(24) │ MODEL(8) │ TOKENS(bar+breakdown) │ CACHE(6) │ COST(8)
	//
	// Overhead for 5 cols: 1(left) + 5×2(pad) + 4×1(sep) + 1(right) = 16 chars
	const (
		colWSession = 24
		colWModel   = 12
		colWCache   = 6
		colWCost    = 8
		colOverhead = 16
	)
	// barWidth fills leftover space; TOKENS also shows breakdown on line 2
	barWidth := termWidth - colOverhead - colWSession - colWModel - colWCache - colWCost
	if barWidth < 16 {
		barWidth = 16
	}
	if barWidth > 28 {
		barWidth = 28
	}
	// TOKENS column: bar on line 1, labeled breakdown on line 2.
	// WidthMax = max(barWidth, breakdown_max_width). Breakdown ≤ ~38 chars.
	colWTokens := barWidth
	if colWTokens < 38 {
		colWTokens = 38
	}

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: colWSession, WidthMax: colWSession},
		{Number: 2, WidthMin: colWModel, WidthMax: colWModel, Align: text.AlignLeft},
		{Number: 3, WidthMax: colWTokens, Align: text.AlignLeft},
		{Number: 4, WidthMin: colWCache, WidthMax: colWCache, Align: text.AlignRight},
		{Number: 5, WidthMin: colWCost, WidthMax: colWCost, Align: text.AlignRight},
	})

	tab.AppendHeader(table.Row{
		StyleColHeader.Render("SESSION"),
		StyleColHeader.Render("MODEL"),
		StyleColHeader.Render("TOKENS"),
		StyleColHeader.Render("CACHE"),
		StyleColHeader.Render("COST"),
	})

	if len(active) == 0 {
		msg := "No active sessions"
		if len(empty) > 0 {
			msg = fmt.Sprintf("No sessions with token data  (%d found with 0 tokens — sessions may still be initializing or use --since 7d)", len(empty))
		}
		tab.AppendRow(table.Row{StyleDim.Render(msg), "", "", "", ""})
	} else {
		for _, s := range active {
			// ── SESSION cell ─────────────────────────────────────────────
			// Line 1: agent badge + name  Line 2: id + path  [Line 3: branch]  [Line 4: subagents]
			agentBadge := ""
			if string(s.AgentID) != "" {
				agentBadge = AgentTag(string(s.AgentID)) + " "
			}
			name := sessionDisplayName(s)
			shortID := s.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			pathStr := abbreviatePath(s.ProjectPath, 13)
			meta := StyleDim.Render(shortID)
			if pathStr != "" {
				meta += "  " + StyleSecondary.Render(pathStr)
			}
			if s.GitBranch != "" && s.GitBranch != "main" && s.GitBranch != "master" {
				meta += "\n" + StyleDim.Render("⎇ "+Truncate(s.GitBranch, 20))
			}
			if s.SubagentCount > 0 {
				noun := "subagents"
				if s.SubagentCount == 1 {
					noun = "subagent"
				}
				meta += "\n" + StyleDim.Render(fmt.Sprintf("↳ %d %s", s.SubagentCount, noun))
			}
			sessionCell := agentBadge + StylePrimary.Render(name) + "\n" + meta

			// ── MODEL cell ───────────────────────────────────────────────
			modelCell := ModelCell(s.Model)

			// ── TOKENS cell ──────────────────────────────────────────────
			// Line 1: colored bar
			// Line 2: labeled colored token breakdown (in/out/cc/cr)
			bar := TokenBar{
				Input:       s.InputTokens,
				Output:      s.OutputTokens,
				CacheCreate: s.CacheCreateTokens,
				CacheRead:   s.CacheReadTokens,
				Width:       barWidth,
			}.Render()
			tokensCell := bar + "\n" +
				TokenSubtitleColored(s.InputTokens, s.OutputTokens, s.CacheCreateTokens, s.CacheReadTokens, s.TokenSource)

			// ── CACHE cell ───────────────────────────────────────────────
			var cacheCell string
			if s.TotalTokens == 0 {
				cacheCell = StyleDim.Render("~")
			} else {
				eff := s.CacheEfficiency * 100
				cacheCell = lipgloss.NewStyle().
					Foreground(lipgloss.Color(CacheEfficiencyColor(s.CacheEfficiency))).
					Bold(true).
					Render(fmt.Sprintf("%.0f%%", eff))
			}

			// ── COST cell ────────────────────────────────────────────────
			// Line 1: cost  Line 2: duration
			costCell := FormatCost(s.CostUSD) + "\n" + StyleSecondary.Render(FormatDuration(s.Duration))

			tab.AppendRow(table.Row{sessionCell, modelCell, tokensCell, cacheCell, costCell})
		}
	}

	// Per-agent breakdown bar (shown only when multiple agents are present)
	agentBreakdown := ""
	if len(agentIDs) > 1 {
		agentBreakdown = "\n" + renderAgentBreakdown(sessions)
	}

	sessionPanel := tab.Render()

	anomalies := renderAnomalies(active)
	output := summary + agentBreakdown + "\n\n" + sessionPanel
	if anomalies != "" {
		output += "\n\n" + anomalies
	}
	return output
}

// ---------------------------------------------------------------------------
// Summary strip — 4 metric cards
// ---------------------------------------------------------------------------

func renderSummaryStrip(totalCost float64, totalTokens int64, cacheEff float64, sessCount int) string {
	costRatio := totalCost / 100.0
	if costRatio > 1 {
		costRatio = 1
	}
	tokenRatio := float64(totalTokens) / 1_000_000_000.0
	if tokenRatio > 1 {
		tokenRatio = 1
	}

	cardWidth := 20
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
		card("  cache eff", fmt.Sprintf("%.0f%%", cacheEff*100), cacheColor, cacheEff, cacheColor),
		card("  sessions", fmt.Sprintf("%d", sessCount), theme.TextPrimary, float64(sessCount)/20.0, theme.TextSecondary),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

// ---------------------------------------------------------------------------
// Per-agent breakdown
// ---------------------------------------------------------------------------

// AgentTotals aggregates token counts and cost per agent.
type agentTotals struct {
	AgentID     string
	Tokens      int64
	Cost        float64
	SessionCount int
}

func renderAgentBreakdown(sessions []*aggregator.SessionStats) string {
	byAgent := map[string]*agentTotals{}
	for _, s := range sessions {
		id := string(s.AgentID)
		if id == "" {
			id = "unknown"
		}
		t, ok := byAgent[id]
		if !ok {
			t = &agentTotals{AgentID: id}
			byAgent[id] = t
		}
		t.Tokens += s.TotalTokens
		t.Cost += s.CostUSD
		t.SessionCount++
	}

	totalTokens := int64(0)
	totalCost := 0.0
	totalSessions := 0
	for _, t := range byAgent {
		totalTokens += t.Tokens
		totalCost += t.Cost
		totalSessions += t.SessionCount
	}

	var parts []string
	for _, t := range byAgent {
		tokPct := 0.0
		if totalTokens > 0 {
			tokPct = float64(t.Tokens) / float64(totalTokens)
		}
		barW := 40
		filled := int(tokPct * float64(barW))
		if filled < 1 && t.Tokens > 0 {
			filled = 1
		}
		bar := lipgloss.NewStyle().
			Background(lipgloss.Color("#4A90D9")).
			Width(filled).Render("")
		emptyW := barW - filled
		if emptyW < 0 {
			emptyW = 0
		}
		emptyBar := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.BarEmpty)).
			Width(emptyW).Render("")
		costStr := StyleSecondary.Render(fmt.Sprintf("$%.2f", t.Cost))
		tokStr := StyleSecondary.Render(HumanizeTokens(t.Tokens))
		countStr := StyleDim.Render(fmt.Sprintf("%d sess", t.SessionCount))
		line := fmt.Sprintf("  %s  %s  %s  %s  %s%s",
			AgentTag(t.AgentID), bar+emptyBar, tokStr, costStr, countStr, StyleDim.Render(fmt.Sprintf(" (%.0f%%)", tokPct*100)))
		parts = append(parts, line)
	}

	heading := StyleBold.Render("agenda") + StyleDim.Render(" · per agent breakdown")
	return "\n" + heading + "\n" + strings.Join(parts, "\n")
}

// ---------------------------------------------------------------------------
// Anomaly panel
// ---------------------------------------------------------------------------

func renderAnomalies(sessions []*aggregator.SessionStats) string {
	var lines []string
	for _, s := range sessions {
		if s.CacheEfficiency < 0.15 && s.TotalTokens > 500_000 {
			name := s.ID
			if len(name) > 8 {
				name = name[:8]
			}
			lines = append(lines, StyleAmber.Render("⚠ ")+StylePrimary.Render(name)+
				StyleSecondary.Render(fmt.Sprintf(": %.0f%% cache · %d msgs · %s",
					s.CacheEfficiency*100, s.MessageCount, FormatCostRaw(s.CostUSD))))
		}
		model := strings.ToLower(s.Model)
		if strings.Contains(model, "opus") && s.MessageCount < 5 && s.CostUSD > 1.0 {
			name := s.ID
			if len(name) > 8 {
				name = name[:8]
			}
			lines = append(lines, StyleAmber.Render("⚠ ")+StylePrimary.Render(name)+
				StyleSecondary.Render(": opus on a "+fmt.Sprintf("%d", s.MessageCount)+"-msg session · consider sonnet"))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return Panel("anomalies & insights", strings.Join(lines, "\n"), 100)
}
