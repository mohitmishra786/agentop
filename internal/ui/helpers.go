package ui

import (
	"fmt"
	"time"
)

func FormatCost(usd float64) string {
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
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}

func FormatCostPerMessage(cost float64, msgs int) string {
	if msgs == 0 {
		return "N/A"
	}
	return fmt.Sprintf("$%.2f/msg", cost/float64(msgs))
}

func FormatBurnRate(rate float64) string {
	if rate >= 1_000_000 {
		return fmt.Sprintf("%.1fM/m", rate/1_000_000)
	}
	if rate >= 1_000 {
		return fmt.Sprintf("%.1fk/m", rate/1_000)
	}
	return fmt.Sprintf("%.0f/m", rate)
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

func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + spaces(width-len(s))
}

func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return spaces(width-len(s)) + s
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
