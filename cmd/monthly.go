package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/ui"
)

var monthlyCmd = &cobra.Command{
	Use:   "monthly",
	Short: "Show monthly session summaries",
	RunE:  runMonthly,
}

func runMonthly(_ *cobra.Command, _ []string) error {
	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	monthlyGroups := groupByMonth(sessions)

	if jsonOut {
		return outputJSONMonthly(monthlyGroups)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	for month, monthSessions := range monthlyGroups {
		fmt.Printf("\n%s · %d sessions\n", month.Format("January 2006"), len(monthSessions))
		fmt.Println(ui.RenderToday(monthSessions, width))
	}

	return nil
}

func groupByMonth(sessions []*aggregator.SessionStats) map[time.Time][]*aggregator.SessionStats {
	groups := make(map[time.Time][]*aggregator.SessionStats)
	for _, s := range sessions {
		month := time.Date(s.StartedAt.Year(), s.StartedAt.Month(), 1, 0, 0, 0, 0, s.StartedAt.Location())
		groups[month] = append(groups[month], s)
	}
	return groups
}

func outputJSONMonthly(groups map[time.Time][]*aggregator.SessionStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(groups)
}
