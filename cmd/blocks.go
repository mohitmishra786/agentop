package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/ui"
)

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Show 5-hour billing window report",
	RunE:  runBlocks,
}

func runBlocks(cmd *cobra.Command, args []string) error {
	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	blocks := groupByBlocks(sessions)

	if jsonOut {
		return outputJSONBlocks(blocks)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	var blockTimes []time.Time
	for t := range blocks {
		blockTimes = append(blockTimes, t)
	}
	sort.Slice(blockTimes, func(i, j int) bool {
		return blockTimes[i].Before(blockTimes[j])
	})

	for _, bt := range blockTimes {
		block := blocks[bt]
		fmt.Printf("\n%s - %s · %d sessions\n",
			bt.Format("15:04"),
			bt.Add(5*time.Hour).Format("15:04"),
			len(block.Sessions))
		fmt.Println(ui.RenderToday(block.Sessions, width))
	}

	return nil
}

type BlockStats struct {
	Start       time.Time
	End         time.Time
	Sessions    []*aggregator.SessionStats
	TotalCost   float64
	TotalTokens int64
}

func groupByBlocks(sessions []*aggregator.SessionStats) map[time.Time]*BlockStats {
	blocks := make(map[time.Time]*BlockStats)
	for _, s := range sessions {
		blockStart := blockStartForTime(s.StartedAt)
		b, ok := blocks[blockStart]
		if !ok {
			b = &BlockStats{
				Start: blockStart,
				End:   blockStart.Add(5 * time.Hour),
			}
			blocks[blockStart] = b
		}
		b.Sessions = append(b.Sessions, s)
		b.TotalCost += s.CostUSD
		b.TotalTokens += s.TotalTokens
	}
	return blocks
}

func blockStartForTime(t time.Time) time.Time {
	utc := t.UTC()
	blockHour := (utc.Hour() / 5) * 5
	return time.Date(utc.Year(), utc.Month(), utc.Day(), int(blockHour), 0, 0, 0, time.UTC)
}

func outputJSONBlocks(blocks map[time.Time]*BlockStats) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(blocks)
}
