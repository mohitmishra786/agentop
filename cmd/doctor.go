package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/ui"
)

type Anomaly struct {
	Severity    string
	SessionID   string
	ProjectName string
	Code        string
	Title       string
	Detail      string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Anomaly detection and insights",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	anomalies := detectAnomalies(sessions)

	if jsonOut {
		return outputJSONAnomalies(anomalies)
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	if len(anomalies) == 0 {
		fmt.Println(ui.StyleGreen.Render("✓ All sessions look healthy. No anomalies detected."))
		return nil
	}

	var lines []string
	for _, a := range anomalies {
		switch a.Severity {
		case "warn":
			lines = append(lines, fmt.Sprintf("⚠  WARN  %s (%s)", a.Title, a.ProjectName))
			lines = append(lines, "     "+a.Detail)
		case "info":
			lines = append(lines, fmt.Sprintf("✓  INFO  %s", a.Title))
			lines = append(lines, "     "+a.Detail)
		case "tip":
			lines = append(lines, fmt.Sprintf("ℹ  TIP   %s", a.Title))
			lines = append(lines, "     "+a.Detail)
		}
		lines = append(lines, "")
	}

	fmt.Println(ui.Panel("anomalies & insights", joinStrings(lines, "\n"), width-2))
	return nil
}

func detectAnomalies(sessions []*aggregator.SessionStats) []Anomaly {
	var anomalies []Anomaly

	for _, s := range sessions {
		name := s.Summary
		if name == "" {
			name = s.ID[:8]
		}

		if s.CacheEfficiency < 0.15 && s.TotalTokens > 500_000 {
			anomalies = append(anomalies, Anomaly{
				Severity:    "warn",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "CACHE_MISS_HIGH",
				Title:       name,
				Detail: fmt.Sprintf("Cache efficiency %.0f%% — cold-start session with large context rebuild. %d messages cost %s (%s/msg).",
					s.CacheEfficiency*100, s.MessageCount,
					ui.FormatCost(s.CostUSD),
					ui.FormatCostPerMessage(s.CostUSD, s.MessageCount)),
			})
		}

		if s.MessageCount < 5 && s.CostUSD > 3.0 {
			anomalies = append(anomalies, Anomaly{
				Severity:    "warn",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "SHORT_SESSION_COST",
				Title:       name,
				Detail: fmt.Sprintf("High cost for a short session — %d messages cost %s.",
					s.MessageCount, ui.FormatCost(s.CostUSD)),
			})
		}

		if s.Model != "" && containsFold(s.Model, "opus") && s.MessageCount < 10 {
			anomalies = append(anomalies, Anomaly{
				Severity:    "warn",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "MODEL_MISMATCH",
				Title:       name,
				Detail:      "Opus used for a short session — consider Sonnet for quick tasks.",
			})
		}

		if s.CacheEfficiency >= 0.80 && s.MessageCount > 10 {
			anomalies = append(anomalies, Anomaly{
				Severity:    "info",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "GOOD_CACHE",
				Title:       fmt.Sprintf("%s — %.0f%% cache efficiency", name, s.CacheEfficiency*100),
				Detail: fmt.Sprintf("Over %s. Cost/message: %s.",
					ui.FormatDuration(s.Duration),
					ui.FormatCostPerMessage(s.CostUSD, s.MessageCount)),
			})
		}
	}

	anomalies = append(anomalies, Anomaly{
		Severity: "tip",
		Code:     "CACHE_WARMUP",
		Title:    "CLAUDE.md cache warm-up cost",
		Detail:   "First turn of each session pays cache-creation premium. Run fewer, longer sessions to amortise this cost.",
	})

	return anomalies
}

func outputJSONAnomalies(anomalies []Anomaly) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(anomalies)
}
