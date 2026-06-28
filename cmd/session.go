package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/pricing"
	"github.com/agentop-dev/agentop/internal/ui"
)

var sessionCmd = &cobra.Command{
	Use:   "session [id]",
	Short: "Deep-dive into a single session",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSession,
}

func runSession(_ *cobra.Command, args []string) error {
	sessionID := args[0]

	agentIDs := resolveAgentIDs()
	files := registry.DiscoverSelected(agentIDs)

	pricer := pricing.DefaultPricer{}

	for _, f := range files {
		if f.SessionID != sessionID {
			continue
		}

		ad := registry.Get(f.AgentID)
		if ad == nil {
			continue
		}

		result, err := ad.ParseSession(f.Path)
		if err != nil {
			return fmt.Errorf("parsing session: %w", err)
		}

		stats := aggregator.AggregateSession(result.Events, result.Meta, pricer)
		if stats == nil {
			return fmt.Errorf("no data for session %s", sessionID)
		}

		stats.ProjectHash = f.ProjectHash

		if jsonOut {
			return outputJSON([]*aggregator.SessionStats{stats})
		}

		width, _, _ := term.GetSize(int(os.Stdout.Fd()))
		if width < 40 {
			width = 80
		}

		fmt.Println(renderSessionDetail(stats, width))
		return nil
	}

	return fmt.Errorf("session %s not found", sessionID)
}

func renderSessionDetail(s *aggregator.SessionStats, width int) string {
	var lines []string

	lines = append(lines, ui.StyleBold.Render(fmt.Sprintf("session: %s", s.Summary)))
	if s.ProjectPath != "" {
		lines = append(lines, ui.StyleDim.Render(fmt.Sprintf("project: %s", s.ProjectPath)))
	}
	if s.GitBranch != "" {
		lines = append(lines, ui.StyleDim.Render(fmt.Sprintf("branch: %s", s.GitBranch)))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("model: %s  ·  started: %s  ·  duration: %s",
		ui.ModelTag(s.Model),
		s.StartedAt.Format("15:04"),
		ui.FormatDuration(s.Duration)))
	lines = append(lines, "")

	toolParts := []string{}
	for tool, count := range s.ToolCalls {
		toolParts = append(toolParts, fmt.Sprintf("%s: %d", tool, count))
	}

	lines = append(lines, ui.StyleDim.Render("tokens:"))
	lines = append(lines, fmt.Sprintf("  input:     %s", ui.HumanizeTokens(s.InputTokens)))
	lines = append(lines, fmt.Sprintf("  output:    %s", ui.HumanizeTokens(s.OutputTokens)))
	lines = append(lines, fmt.Sprintf("  cache cr:  %s", ui.HumanizeTokens(s.CacheReadTokens)))
	lines = append(lines, fmt.Sprintf("  cache cc:  %s", ui.HumanizeTokens(s.CacheCreateTokens)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("total cost: %s  (verified: %s)",
		ui.FormatCost(s.CostUSD),
		ui.FormatCost(s.CostUSDCalculated)))
	lines = append(lines, fmt.Sprintf("cache efficiency: %.0f%%", s.CacheEfficiency*100))
	lines = append(lines, fmt.Sprintf("messages: %d  ·  cost/msg: %s",
		s.MessageCount,
		ui.FormatCostPerMessage(s.CostUSD, s.MessageCount)))
	lines = append(lines, "")
	lines = append(lines, ui.StyleDim.Render("tools: "+joinStrings(toolParts, "  ")))

	return ui.Panel("session: "+s.Summary[:minInt(len(s.Summary), 40)], joinStrings(lines, "\n"), width-2)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
