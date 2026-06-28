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

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Show daily session reports",
	RunE:  runDaily,
}

func runDaily(_ *cobra.Command, _ []string) error {
	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	dailyGroups := groupByDay(sessions)

	if jsonOut {
		return outputJSONDaily(dailyGroups)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	for day, daySessions := range dailyGroups {
		fmt.Printf("\n%s · %d sessions\n", day.Format("Mon Jan 2"), len(daySessions))
		fmt.Println(ui.RenderToday(daySessions, width))
	}

	return nil
}

func groupByDay(sessions []*aggregator.SessionStats) map[time.Time][]*aggregator.SessionStats {
	groups := make(map[time.Time][]*aggregator.SessionStats)
	for _, s := range sessions {
		day := time.Date(s.StartedAt.Year(), s.StartedAt.Month(), s.StartedAt.Day(), 0, 0, 0, 0, s.StartedAt.Location())
		groups[day] = append(groups[day], s)
	}
	return groups
}

func outputJSONDaily(groups map[time.Time][]*aggregator.SessionStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(groups)
}
