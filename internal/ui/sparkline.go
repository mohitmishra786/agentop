package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Sparkline renders a mini sparkline chart
func Sparkline(data []float64, width int, color string) string {
	if len(data) == 0 || width <= 0 {
		return ""
	}

	min, max := minMax(data)
	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}

	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	result := make([]rune, width)

	for i := 0; i < width; i++ {
		dataIndex := int(float64(i) / float64(width) * float64(len(data)))
		if dataIndex >= len(data) {
			dataIndex = len(data) - 1
		}

		normalized := (data[dataIndex] - min) / range_
		barIndex := int(normalized * float64(len(bars)-1))
		if barIndex < 0 {
			barIndex = 0
		} else if barIndex >= len(bars) {
			barIndex = len(bars) - 1
		}

		result[i] = bars[barIndex]
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(string(result))
}

// TrendIndicator shows a trend arrow
func TrendIndicator(current, previous float64) string {
	if previous == 0 {
		return "—"
	}

	diff := ((current - previous) / previous) * 100
	switch {
	case diff > 10:
		return StyleGreen.Render(fmt.Sprintf("↑%.0f%%", diff))
	case diff > 0:
		return StyleGreen.Render(fmt.Sprintf("↑%.1f%%", diff))
	case diff < -10:
		return StyleRed.Render(fmt.Sprintf("↓%.0f%%", diff))
	case diff < 0:
		return StyleRed.Render(fmt.Sprintf("↓%.1f%%", diff))
	default:
		return "—"
	}
}

// Gauge renders a horizontal gauge/bar
func Gauge(value, max float64, width int, color string) string {
	if width <= 0 || max == 0 {
		return ""
	}
	ratio := value / max
	if ratio > 1 {
		ratio = 1
	} else if ratio < 0 {
		ratio = 0
	}

	filled := int(ratio * float64(width))
	empty := width - filled

	filledStr := lipgloss.NewStyle().Background(lipgloss.Color(color)).Width(filled).Render("")
	emptyStr := lipgloss.NewStyle().Background(lipgloss.Color(theme.BarEmpty)).Width(empty).Render("")
	return filledStr + emptyStr
}

// MiniSpark renders a tiny sparkline for compact display
func MiniSpark(data []float64, width int) string {
	if len(data) == 0 || width <= 0 {
		return ""
	}

	min, max := minMax(data)
	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}

	result := ""
	for i := 0; i < width; i++ {
		if i >= len(data) {
			break
		}

		normalized := (data[i] - min) / range_
		if normalized < 0.25 {
			result += "▁"
		} else if normalized < 0.5 {
			result += "▂"
		} else if normalized < 0.75 {
			result += "▃"
		} else {
			result += "▄"
		}
	}

	return StyleDim.Render(result)
}

// ThresholdBadge returns a colored badge based on thresholds
func ThresholdBadge(value float64, thresholds struct{ Yellow, Red float64 }) string {
	var style lipgloss.Style
	switch {
	case value >= thresholds.Red:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Red)).
			Background(lipgloss.Color("#2d1b1b")).
			Bold(true)
	case value >= thresholds.Yellow:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Amber)).
			Background(lipgloss.Color("#2d2d1b")).
			Bold(true)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Green)).
			Background(lipgloss.Color("#1b2d1b")).
			Bold(true)
	}

	return style.Render(fmt.Sprintf("%.0f%%", value*100))
}

// ProgressRing renders a circular progress indicator
func ProgressRing(percent float64) string {
	segments := []string{
		"█", "▓", "▒", "░",
	}

	filled := int(percent / 100.0 * float64(len(segments)))
	if filled >= len(segments) {
		filled = len(segments) - 1
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Green)).
		Render(segments[filled])
}

func minMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
