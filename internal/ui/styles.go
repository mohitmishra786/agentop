package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColGreen    = "#50c87a"
	ColAmber    = "#e5a040"
	ColRed      = "#e05050"
	ColBlue     = "#4a90d9"
	ColTeal     = "#7ecfb3"
	ColPurple   = "#b36de0"
	ColDim      = "#666680"
	ColBorder   = "#3a3a4a"
	ColHeader   = "#9898b8"
	ColText     = "#c8c8c8"
	ColTextBold = "#c8c8e8"
)

var (
	StyleBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColBorder))

	StyleHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColHeader)).
			Background(lipgloss.Color("#222230")).
			Padding(0, 1)

	StyleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColDim))

	StyleGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColGreen)).
			Bold(true)

	StyleAmber = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColAmber)).
			Bold(true)

	StyleRed = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColRed)).
			Bold(true)

	StyleBold = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColTextBold))

	StyleTagOpus = lipgloss.NewStyle().
			Background(lipgloss.Color("#3a2050")).
			Foreground(lipgloss.Color("#c090f0")).
			Padding(0, 1)

	StyleTagSonnet = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a3050")).
			Foreground(lipgloss.Color("#70b0e0")).
			Padding(0, 1)

	StyleTagHaiku = lipgloss.NewStyle().
			Background(lipgloss.Color("#1e3520")).
			Foreground(lipgloss.Color("#90d080")).
			Padding(0, 1)
)

func ModelTag(model string) string {
	switch {
	case strings.Contains(model, "opus"):
		return StyleTagOpus.Render("opus")
	case strings.Contains(model, "sonnet"):
		return StyleTagSonnet.Render("sonnet")
	case strings.Contains(model, "haiku"):
		return StyleTagHaiku.Render("haiku")
	default:
		return StyleDim.Render(model)
	}
}

func CacheEfficiencyStyle(eff float64) lipgloss.Style {
	switch {
	case eff >= 0.80:
		return StyleGreen
	case eff >= 0.40:
		return StyleAmber
	default:
		return StyleRed
	}
}
