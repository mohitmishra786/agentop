package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type TokenBar struct {
	Input       int64
	Output      int64
	CacheCreate int64
	CacheRead   int64
	Width       int
}

func (b TokenBar) Render() string {
	total := b.Input + b.Output + b.CacheCreate + b.CacheRead
	if total == 0 || b.Width <= 0 {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(ColBorder)).
			Width(b.Width).
			Render("")
	}

	type seg struct {
		count int64
		color string
	}
	segs := []seg{
		{b.Input, ColBlue},
		{b.Output, ColTeal},
		{b.CacheCreate, ColAmber},
		{b.CacheRead, ColPurple},
	}

	widths := make([]int, len(segs))
	totalWidth := 0
	for i, s := range segs {
		w := int(float64(s.count) / float64(total) * float64(b.Width))
		widths[i] = w
		totalWidth += w
	}
	if diff := b.Width - totalWidth; diff != 0 {
		maxIdx := 0
		for i := range widths {
			if widths[i] > widths[maxIdx] {
				maxIdx = i
			}
		}
		widths[maxIdx] += diff
	}

	var parts []string
	for i, s := range segs {
		if widths[i] <= 0 {
			continue
		}
		parts = append(parts, lipgloss.NewStyle().
			Background(lipgloss.Color(s.color)).
			Width(widths[i]).
			Render(""))
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
	filled := int(ratio * float64(width))
	empty := width - filled

	filledStr := lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		Width(filled).Render("")
	emptyStr := lipgloss.NewStyle().
		Background(lipgloss.Color(ColBorder)).
		Width(empty).Render("")

	return filledStr + emptyStr
}

func TokenSubtitle(input, output, cc, cr int64) string {
	parts := []string{
		HumanizeTokens(input) + " in",
		HumanizeTokens(output) + " out",
		HumanizeTokens(cc) + " cc",
		HumanizeTokens(cr) + " cr",
	}
	return StyleDim.Render(strings.Join(parts, " · "))
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
