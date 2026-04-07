package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/claude"
	"github.com/agentop-dev/agentop/internal/pricing"
	"github.com/agentop-dev/agentop/internal/ui"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's Claude Code sessions",
	RunE:  runToday,
}

func runToday(cmd *cobra.Command, args []string) error {
	if err := ui.Init(); err != nil {
		return err
	}

	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	if jsonOut {
		return outputJSON(sessions)
	}

	if watch {
		return runWatchMode(sessions)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	fmt.Println(ui.RenderToday(sessions, width))
	return nil
}

func loadSessions() ([]*aggregator.SessionStats, error) {
	files, err := claude.Discover(claudeDir)
	if err != nil {
		return nil, err
	}

	pricer := pricing.DefaultPricer{}
	var sessions []*aggregator.SessionStats

	for _, f := range files {
		events, err := claude.ParseSession(f.Path)
		if err != nil {
			continue
		}

		stats := aggregator.AggregateSession(events, nil, pricer)
		if stats == nil {
			continue
		}

		stats.ProjectHash = f.ProjectHash
		if stats.ID == "" {
			stats.ID = f.SessionID
		}

		if len(f.SubagentFiles) > 0 {
			subTokens, subCount, subCost := aggregator.AggregateSubagents(f.SubagentFiles, pricer)
			stats.SubagentTokens = subTokens
			stats.SubagentCount = subCount
			stats.CostUSD += subCost
			stats.TotalTokens += subTokens
		}

		sessions = append(sessions, stats)
	}

	return sessions, nil
}

func filterSessions(sessions []*aggregator.SessionStats) []*aggregator.SessionStats {
	startTime, err := parseSince(since)
	if err != nil {
		startTime = time.Now().AddDate(0, 0, -1)
	}

	var filtered []*aggregator.SessionStats
	for _, s := range sessions {
		if !s.EndedAt.IsZero() && s.EndedAt.Before(startTime) {
			continue
		}
		if project != "" && s.ProjectPath != "" {
			matched := false
			for _, p := range []string{s.ProjectPath, s.ProjectName} {
				if containsFold(p, project) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if model != "" && s.Model != "" {
			if !containsFold(s.Model, model) {
				continue
			}
		}
		filtered = append(filtered, s)
	}

	return filtered
}

func parseSince(s string) (time.Time, error) {
	now := time.Now()
	switch s {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location()), nil
	}

	if len(s) > 1 && s[len(s)-1] == 'd' {
		days := 0
		_, err := fmt.Sscanf(s, "%dd", &days)
		if err != nil {
			return time.Time{}, err
		}
		return now.AddDate(0, 0, -days), nil
	}

	return time.Parse("2006-01-02", s)
}

func containsFold(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsFoldSimple(s, substr))
}

func containsFoldSimple(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func outputJSON(sessions []*aggregator.SessionStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(sessions)
}

func runWatchMode(sessions []*aggregator.SessionStats) error {
	m := ui.NewModel(sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
