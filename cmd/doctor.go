package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/claude"
	"github.com/agentop-dev/agentop/internal/pricing"
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
	Use:   "doctor [session-id]",
	Short: "Anomaly detection and session optimization insights",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	sessions = filterSessions(sessions)

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width < 40 {
		width = 80
	}

	// Session-specific mode — search all sessions (ignore date filter) so any ID works.
	if len(args) == 1 {
		prefix := strings.ToLower(args[0])
		allSessions, err2 := loadSessions()
		if err2 != nil {
			allSessions = sessions
		}
		var target *aggregator.SessionStats
		for _, s := range allSessions {
			if strings.HasPrefix(strings.ToLower(s.ID), prefix) {
				target = s
				break
			}
		}
		if target == nil {
			return fmt.Errorf("no session found matching %q", args[0])
		}
		anomalies := detectAnomalies([]*aggregator.SessionStats{target})
		if jsonOut {
			return outputJSONAnomalies(anomalies)
		}
		fmt.Println(renderDoctorDetail(target, anomalies, width))
		return nil
	}

	anomalies := detectAnomalies(sessions)

	if jsonOut {
		return outputJSONAnomalies(anomalies)
	}

	if len(anomalies) == 0 {
		fmt.Println(ui.StyleGreen.Render("  ✓ All sessions look healthy. No anomalies detected."))
		return nil
	}

	// Auto-focus: if only one session has issues, go straight to detail view
	sessionsWithIssues := uniqueSessionIDs(anomalies)
	if len(sessionsWithIssues) == 1 {
		for _, s := range sessions {
			if s.ID == sessionsWithIssues[0] {
				fmt.Println(renderDoctorDetail(s, anomalies, width))
				return nil
			}
		}
	}

	fmt.Println(renderDoctorSummary(sessions, anomalies, width))
	return nil
}

func uniqueSessionIDs(anomalies []Anomaly) []string {
	seen := map[string]bool{}
	var out []string
	for _, a := range anomalies {
		if a.SessionID != "" && !seen[a.SessionID] {
			seen[a.SessionID] = true
			out = append(out, a.SessionID)
		}
	}
	return out
}

// severityTag returns a padded colored label for a severity level.
func severityTag(sev string) string {
	switch sev {
	case "crit":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#AA2222")).Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).Bold(true).Render("CRIT")
	case "warn":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#884400")).Foreground(lipgloss.Color("#FFDD88")).
			Padding(0, 1).Bold(true).Render("WARN")
	case "info":
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#1A3050")).Foreground(lipgloss.Color("#70B0E0")).
			Padding(0, 1).Render("INFO")
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#252525")).Foreground(lipgloss.Color("#888888")).
			Padding(0, 1).Render(" TIP")
	}
}

// stripSessionPrefix removes leading "name — " or "id — " from anomaly titles.
func stripSessionPrefix(title, name, shortID string) string {
	for _, pfx := range []string{name + " — ", shortID + " — "} {
		title = strings.TrimPrefix(title, pfx)
	}
	return title
}

// renderDoctorDetail renders a per-session doctor view styled like the main table.
func renderDoctorDetail(s *aggregator.SessionStats, anomalies []Anomaly, width int) string {
	if width > 120 {
		width = 120
	}

	shortID := s.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	name := sessionName(s)

	// ── Header: session info row ─────────────────────────────────────────────
	tab := table.NewWriter()
	tab.SetAllowedRowLength(width)
	tab.SetStyle(doctorTableStyle())

	effColor := ui.CacheEfficiencyColor(s.CacheEfficiency)
	effVal := lipgloss.NewStyle().Foreground(lipgloss.Color(effColor)).Bold(true).
		Render(fmt.Sprintf("%.0f%%", s.CacheEfficiency*100))

	sessionCell := ui.StylePrimary.Render(name) + "\n" +
		ui.StyleDim.Render(shortID) + "  " + ui.StyleSecondary.Render(s.ProjectName)
	metaCell := ui.FormatCost(s.CostUSD) + "\n" +
		ui.StyleSecondary.Render(ui.FormatDuration(s.Duration))
	statsCell := effVal + ui.StyleDim.Render(" cache") + "\n" +
		ui.StyleDim.Render(fmt.Sprintf("%d msgs", s.MessageCount))

	const colWSession = 26
	const colWModel = 12
	const colWMeta = 12
	const colWStats = 14

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: colWSession, WidthMax: colWSession},
		{Number: 2, WidthMin: colWModel, WidthMax: colWModel},
		{Number: 3, WidthMin: colWMeta, WidthMax: colWMeta, Align: text.AlignRight},
		{Number: 4, WidthMin: colWStats, WidthMax: colWStats, Align: text.AlignRight},
	})
	tab.AppendHeader(table.Row{
		ui.StyleColHeader.Render("SESSION"),
		ui.StyleColHeader.Render("MODEL"),
		ui.StyleColHeader.Render("COST"),
		ui.StyleColHeader.Render("CACHE"),
	})
	tab.AppendRow(table.Row{sessionCell, ui.ModelTag(s.Model), metaCell, statsCell})

	titleWord := "issue"
	issueCount := 0
	for _, a := range anomalies {
		if a.Severity == "crit" || a.Severity == "warn" {
			issueCount++
		}
	}
	if issueCount != 1 {
		titleWord = "issues"
	}
	tab.SetTitle(fmt.Sprintf(" doctor · %s · %d %s ", name, issueCount, titleWord))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(tab.Render())
	sb.WriteString("\n\n")

	// ── Anomaly rows ─────────────────────────────────────────────────────────
	// SEV col: 1(border)+1(pad)+6+1(pad)+1(sep) = 10; right border = 1; inner padding = 2
	// FINDING visible width = width - 10 - 1 - 2 = width - 13; subtract 2 for safety
	findingColMax := width - 13
	if findingColMax < 40 {
		findingColMax = 40
	}
	detailW := findingColMax - 2

	aTab := table.NewWriter()
	aTab.SetAllowedRowLength(width)
	aTab.SetStyle(doctorTableStyle())
	aTab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 6, WidthMax: 6},
		{Number: 2, WidthMax: findingColMax},
	})
	aTab.AppendHeader(table.Row{
		ui.StyleColHeader.Render("SEV"),
		ui.StyleColHeader.Render("FINDING"),
	})

	for _, a := range anomalies {
		title := stripSessionPrefix(a.Title, name, shortID)
		// Style each wrapped line individually so ANSI codes don't span line breaks,
		// preventing go-pretty from re-wrapping styled multi-line content.
		var detailLines []string
		for _, line := range wrapText(a.Detail, detailW) {
			detailLines = append(detailLines, ui.StyleSecondary.Render(line))
		}
		finding := ui.StylePrimary.Render(title) + "\n" + strings.Join(detailLines, "\n")
		aTab.AppendRow(table.Row{severityTag(a.Severity), finding})
	}

	sb.WriteString(aTab.Render())
	sb.WriteString("\n\n")
	sb.WriteString("  " + ui.StyleDim.Render("agentop doctor "+shortID+" · run without session-id to see all sessions") + "\n")

	return sb.String()
}

// renderDoctorSummary renders a multi-session overview as a go-pretty table.
func renderDoctorSummary(sessions []*aggregator.SessionStats, anomalies []Anomaly, width int) string {
	if width > 120 {
		width = 120
	}

	// Group anomalies by session
	bySession := map[string][]Anomaly{}
	var globalTips []Anomaly
	for _, a := range anomalies {
		if a.SessionID == "" {
			globalTips = append(globalTips, a)
		} else {
			bySession[a.SessionID] = append(bySession[a.SessionID], a)
		}
	}

	// Count sessions with issues
	issueCount := 0
	for _, s := range sessions {
		if len(bySession[s.ID]) > 0 {
			issueCount++
		}
	}

	tab := table.NewWriter()
	tab.SetAllowedRowLength(width)
	tab.SetStyle(doctorTableStyle())

	const colWSession = 26
	const colWModel = 12
	const colWIssues = 20

	previewW := width - colWSession - colWModel - colWIssues - 20
	if previewW < 20 {
		previewW = 20
	}

	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: colWSession, WidthMax: colWSession},
		{Number: 2, WidthMin: colWModel, WidthMax: colWModel},
		{Number: 3, WidthMin: colWIssues, WidthMax: colWIssues},
		{Number: 4, WidthMax: previewW},
	})
	tab.AppendHeader(table.Row{
		ui.StyleColHeader.Render("SESSION"),
		ui.StyleColHeader.Render("MODEL"),
		ui.StyleColHeader.Render("ISSUES"),
		ui.StyleColHeader.Render("TOP FINDING"),
	})

	sessionWord := "sessions"
	if issueCount == 1 {
		sessionWord = "session"
	}
	tab.SetTitle(fmt.Sprintf(" doctor · %d %s with findings ", issueCount, sessionWord))

	for _, s := range sessions {
		sa := bySession[s.ID]
		if len(sa) == 0 {
			continue
		}

		shortID := s.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		name := sessionName(s)

		sessionCell := ui.StylePrimary.Render(name) + "\n" +
			ui.StyleDim.Render(shortID)
		if s.ProjectName != "" {
			sessionCell += "  " + ui.StyleSecondary.Render(ui.Truncate(s.ProjectName, 14))
		}

		// Issue count badges
		counts := map[string]int{}
		for _, a := range sa {
			counts[a.Severity]++
		}
		var badges []string
		if n := counts["crit"]; n > 0 {
			badges = append(badges, ui.StyleRed.Render(fmt.Sprintf("%d crit", n)))
		}
		if n := counts["warn"]; n > 0 {
			badges = append(badges, ui.StyleAmber.Render(fmt.Sprintf("%d warn", n)))
		}
		if n := counts["info"]; n > 0 {
			badges = append(badges, ui.StyleSecondary.Render(fmt.Sprintf("%d info", n)))
		}
		if n := counts["tip"]; n > 0 {
			badges = append(badges, ui.StyleDim.Render(fmt.Sprintf("%d tip", n)))
		}
		issuesCell := strings.Join(badges, ui.StyleDim.Render("  ")) + "\n" +
			ui.StyleDim.Render("doctor "+shortID)

		// Top warn/crit preview
		var preview string
		for _, a := range sa {
			if a.Severity == "crit" || a.Severity == "warn" {
				preview = stripSessionPrefix(a.Title, name, shortID)
				break
			}
		}
		if preview == "" && len(sa) > 0 {
			preview = stripSessionPrefix(sa[0].Title, name, shortID)
		}
		previewCell := ui.StyleSecondary.Render(ui.Truncate(preview, previewW-2))

		tab.AppendRow(table.Row{sessionCell, ui.ModelTag(s.Model), issuesCell, previewCell})
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(tab.Render())

	// Global tips below table
	if len(globalTips) > 0 {
		sb.WriteString("\n\n")
		tipW := width - 10
		for _, a := range globalTips {
			sb.WriteString("  " + severityTag("tip") + "  " + ui.StylePrimary.Render(a.Title) + "\n")
			for _, line := range wrapText(a.Detail, tipW) {
				sb.WriteString("         " + ui.StyleSecondary.Render(line) + "\n")
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString("  " + ui.StyleDim.Render("run  agentop doctor <session-id>  for full per-session analysis") + "\n")

	return sb.String()
}

// doctorTableStyle returns a go-pretty table style matching the main view.
func doctorTableStyle() table.Style {
	s := table.StyleRounded
	s.Options.SeparateRows = true
	s.Options.SeparateColumns = true
	s.Options.DrawBorder = true
	return s
}

// wrapText breaks text into lines of at most maxW visible characters.
func wrapText(text string, maxW int) []string {
	if maxW < 20 {
		maxW = 20
	}
	words := strings.Fields(text)
	var lines []string
	var cur strings.Builder
	curLen := 0
	for _, w := range words {
		wl := len(w) // ASCII only in our detail strings
		if curLen+wl+1 > maxW && curLen > 0 {
			lines = append(lines, cur.String())
			cur.Reset()
			curLen = 0
		}
		if curLen > 0 {
			cur.WriteByte(' ')
			curLen++
		}
		cur.WriteString(w)
		curLen += wl
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

func label(a Anomaly) string {
	if a.ProjectName != "" {
		return fmt.Sprintf("%s  (%s)", a.Title, a.ProjectName)
	}
	return a.Title
}

// ─── helpers ────────────────────────────────────────────────────────────────

var pricer = pricing.DefaultPricer{}

// tokensToCost returns the dollar cost of cache-creation tokens at the given model rate.
func coldStartCost(tokens int64, model string) float64 {
	return pricer.Calculate(claude.Usage{CacheCreationInputTokens: int(tokens)}, model)
}

// counterfactualCost returns what the session would have cost on Sonnet.
func counterfactualCost(s *aggregator.SessionStats) float64 {
	u := claude.Usage{
		InputTokens:              int(s.InputTokens),
		OutputTokens:             int(s.OutputTokens),
		CacheCreationInputTokens: int(s.CacheCreateTokens),
		CacheReadInputTokens:     int(s.CacheReadTokens),
	}
	return pricer.Calculate(u, "claude-sonnet-4-6")
}

func sessionName(s *aggregator.SessionStats) string {
	if s.Summary != "" {
		return s.Summary
	}
	if s.ProjectName != "" {
		return s.ProjectName
	}
	if len(s.ID) >= 8 {
		return s.ID[:8]
	}
	return s.ID
}

func isOpus(model string) bool {
	return strings.Contains(strings.ToLower(model), "opus")
}

// ─── anomaly detection ───────────────────────────────────────────────────────

func detectAnomalies(sessions []*aggregator.SessionStats) []Anomaly {
	var anomalies []Anomaly

	for _, s := range sessions {
		anomalies = append(anomalies, checkCacheHealth(s)...)
		anomalies = append(anomalies, checkContextGrowth(s)...)
		anomalies = append(anomalies, checkToolPatterns(s)...)
		anomalies = append(anomalies, checkModelRouting(s)...)
		anomalies = append(anomalies, checkGeneralHealth(s)...)
	}

	anomalies = append(anomalies, checkResumeOpportunity(sessions)...)

	anomalies = append(anomalies, Anomaly{
		Severity: "tip",
		Code:     "CACHE_WARMUP",
		Title:    "First turn always pays cache-creation premium",
		Detail:   "Run fewer, longer sessions to amortise the cold-start cost. Resume an existing session instead of starting a new one when possible.",
	})

	return anomalies
}

// checkCacheHealth fires on COLD_START_COST, CACHE_STALL, and CACHE_BREAK.
func checkCacheHealth(s *aggregator.SessionStats) []Anomaly {
	var out []Anomaly
	name := sessionName(s)

	// COLD_START_COST — Turn 1 cache creation > 80k tokens is a large context tax.
	if s.Turn1CacheCreate > 80_000 {
		cost := coldStartCost(s.Turn1CacheCreate, s.Model)
		out = append(out, Anomaly{
			Severity:    "info",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "COLD_START_COST",
			Title:       fmt.Sprintf("%s — %s cold-start tax per session", name, ui.FormatCost(cost)),
			Detail: fmt.Sprintf(
				"Turn 1 loaded %s tokens in cache creation (%s at current model rates). "+
					"This is your CLAUDE.md + system context, paid fresh every session. "+
					"Keep CLAUDE.md under ~500 lines; move task-specific instructions into skills.",
				formatTokens(s.Turn1CacheCreate), ui.FormatCost(cost)),
		})
	}

	// CACHE_STALL — cache creation still high at turn 5 means the cache never warmed.
	if len(s.Turns) >= 5 && s.Turns[0].CacheCreate > 20_000 {
		ratio := float64(s.Turns[4].CacheCreate) / float64(s.Turns[0].CacheCreate)
		if ratio > 0.4 {
			out = append(out, Anomaly{
				Severity:    "warn",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "CACHE_STALL",
				Title:       fmt.Sprintf("%s — cache never stabilized", name),
				Detail: fmt.Sprintf(
					"Cache creation was still %s tokens at turn 5 (%.0f%% of turn 1's %s). "+
						"Cache should drop to near zero by turn 3. "+
						"Possible causes: model switch mid-session, >5min idle gap, or CLAUDE.md edited between turns.",
					formatTokens(s.Turns[4].CacheCreate),
					ratio*100,
					formatTokens(s.Turns[0].CacheCreate)),
			})
		}
	}

	// CACHE_BREAK — cache_read dropped >60% in a single turn while cache_create spiked.
	for i := 1; i < len(s.Turns); i++ {
		prev := s.Turns[i-1]
		curr := s.Turns[i]
		if prev.CacheRead < 20_000 {
			continue
		}
		dropRatio := 1.0 - float64(curr.CacheRead)/float64(prev.CacheRead)
		spiked := curr.CacheCreate > prev.CacheCreate*2 || curr.CacheCreate > 50_000
		if dropRatio >= 0.60 && spiked {
			out = append(out, Anomaly{
				Severity:    "warn",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "CACHE_BREAK",
				Title:       fmt.Sprintf("%s — cache collapsed at turn %d", name, i+1),
				Detail: fmt.Sprintf(
					"cache_read: %s → %s (−%.0f%%), cache_create spiked to %s. "+
						"All subsequent turns re-paid full context cost. "+
						"Cause: model switch or >5min idle gap.",
					formatTokens(prev.CacheRead), formatTokens(curr.CacheRead), dropRatio*100,
					formatTokens(curr.CacheCreate)),
			})
			break // report the first break only
		}
	}

	return out
}

// checkContextGrowth fires on CONTEXT_BLOAT and LARGE_TOOL_RESULT.
func checkContextGrowth(s *aggregator.SessionStats) []Anomaly {
	var out []Anomaly
	name := sessionName(s)

	// CONTEXT_BLOAT — input tokens growing faster than ~8k/turn.
	if len(s.Turns) >= 6 {
		first := s.Turns[0].InputTokens
		last := s.Turns[len(s.Turns)-1].InputTokens
		n := int64(len(s.Turns) - 1)
		if n > 0 && last > first {
			slope := (last - first) / n
			if slope > 8_000 {
				// Rough optimal compact point: when input first exceeds 2x its starting value.
				compactTurn := 0
				for i, t := range s.Turns {
					if t.InputTokens > first*2 {
						compactTurn = i
						break
					}
				}
				detail := fmt.Sprintf(
					"Input grew from %s → %s over %d turns (+%s/turn). "+
						"Context is accumulating. Run /compact after completing major milestones.",
					formatTokens(first), formatTokens(last), len(s.Turns), formatTokens(slope))
				if compactTurn > 0 {
					detail = fmt.Sprintf(
						"Input grew from %s → %s over %d turns (+%s/turn). "+
							"Optimal /compact point was around turn %d (when context first doubled). "+
							"Run /compact after completing major milestones.",
						formatTokens(first), formatTokens(last), len(s.Turns), formatTokens(slope), compactTurn+1)
				}
				out = append(out, Anomaly{
					Severity:    "warn",
					SessionID:   s.ID,
					ProjectName: s.ProjectName,
					Code:        "CONTEXT_BLOAT",
					Title:       fmt.Sprintf("%s — context growing fast (+%s/turn)", name, formatTokens(slope)),
					Detail:      detail,
				})
			}
		}
	}

	// LARGE_TOOL_RESULT — a single tool result >= 500 KB may bloat all subsequent turns.
	if s.MaxToolResultBytes >= 500_000 {
		approxTokens := s.MaxToolResultBytes / 4
		out = append(out, Anomaly{
			Severity:    "warn",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "LARGE_TOOL_RESULT",
			Title:       fmt.Sprintf("%s — huge tool result (%s)", name, formatBytes(s.MaxToolResultBytes)),
			Detail: fmt.Sprintf(
				"Largest tool result was %s (~%s tokens). "+
					"This inflated context for all subsequent turns. "+
					"Pipe large Bash outputs with | head -100, or redirect to a temp file and read selectively.",
				formatBytes(s.MaxToolResultBytes), formatTokens(approxTokens)),
		})
	}

	return out
}

// checkToolPatterns fires on REPEATED_FILE_READS, EXPLORE_HEAVY, and BASH_BLOAT_RISK.
func checkToolPatterns(s *aggregator.SessionStats) []Anomaly {
	var out []Anomaly
	name := sessionName(s)

	// REPEATED_FILE_READS — same file read 4+ times.
	if len(s.FilesRead) > 0 {
		type fileCount struct {
			path  string
			count int
		}
		var top []fileCount
		for path, n := range s.FilesRead {
			if n >= 4 {
				top = append(top, fileCount{path, n})
			}
		}
		sort.Slice(top, func(i, j int) bool { return top[i].count > top[j].count })
		if len(top) > 0 {
			// Report the most re-read file.
			worst := top[0]
			shortPath := worst.path
			if len(shortPath) > 40 {
				shortPath = "…" + shortPath[len(shortPath)-39:]
			}
			out = append(out, Anomaly{
				Severity:    "info",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "REPEATED_FILE_READS",
				Title:       fmt.Sprintf("%s — %s re-read %d times", name, shortPath, worst.count),
				Detail: fmt.Sprintf(
					"%s was read %d times. Each re-read re-adds the file content to context. "+
						"Reference the file once at session start, or pin it with @%s to keep it in context.",
					worst.path, worst.count, worst.path),
			})
		}
	}

	// EXPLORE_HEAVY — Read:Edit+Write ratio > 8.
	reads := s.ToolCalls["Read"]
	edits := s.ToolCalls["Edit"] + s.ToolCalls["Write"]
	if reads > 0 && edits > 0 {
		ratio := float64(reads) / float64(edits)
		if ratio > 8 && reads > 16 { // at least 8 actual reads (counts may be doubled)
			out = append(out, Anomaly{
				Severity:    "info",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "EXPLORE_HEAVY",
				Title:       fmt.Sprintf("%s — %.0f:1 read-to-edit ratio", name, ratio),
				Detail: fmt.Sprintf(
					"%d reads, %d edits. Main context carried all explored files. "+
						"For exploration phases, use subagents — they read in a child context and "+
						"return a summary, leaving your main session context clean.",
					reads, edits),
			})
		}
	}

	// BASH_BLOAT_RISK — Bash > 50% of tool calls AND large session.
	totalCalls := 0
	for _, n := range s.ToolCalls {
		totalCalls += n
	}
	bashCalls := s.ToolCalls["Bash"]
	if totalCalls > 0 && s.TotalTokens > 500_000 {
		bashFrac := float64(bashCalls) / float64(totalCalls)
		if bashFrac > 0.50 {
			out = append(out, Anomaly{
				Severity:    "info",
				SessionID:   s.ID,
				ProjectName: s.ProjectName,
				Code:        "BASH_BLOAT_RISK",
				Title:       fmt.Sprintf("%s — Bash is %.0f%% of tool calls", name, bashFrac*100),
				Detail: "Bash output is unbounded and lands in context unchanged. " +
					"Pipe large outputs with | head -100, redirect build/test output to a file, " +
					"or add heavy build directories to .claudeignore.",
			})
		}
	}

	return out
}

// checkModelRouting fires on OPUS_OVERKILL, MID_SESSION_SWITCH, THINKING_SPIKE,
// and SUBAGENT_MODEL_COST.
func checkModelRouting(s *aggregator.SessionStats) []Anomaly {
	var out []Anomaly
	name := sessionName(s)

	// OPUS_OVERKILL — Opus used but average output was low (< 400 tokens/turn).
	if isOpus(s.Model) && s.MessageCount > 0 {
		avgOutput := s.OutputTokens / int64(s.MessageCount)
		if avgOutput < 400 {
			sonnetCost := counterfactualCost(s)
			if sonnetCost < s.CostUSD*0.8 { // only flag if saving > 20%
				out = append(out, Anomaly{
					Severity:    "warn",
					SessionID:   s.ID,
					ProjectName: s.ProjectName,
					Code:        "OPUS_OVERKILL",
					Title: fmt.Sprintf("%s — Opus for low-output work (%s → %s on Sonnet)",
						name, ui.FormatCost(s.CostUSD), ui.FormatCost(sonnetCost)),
					Detail: fmt.Sprintf(
						"%d messages, avg %s output tokens/turn. "+
							"Same session on claude-sonnet-4-6: ~%s (%.0f%% savings). "+
							"Opus is most valuable for complex multi-step reasoning; "+
							"short, mechanical tasks run equally well on Sonnet.",
						s.MessageCount, formatTokens(avgOutput),
						ui.FormatCost(sonnetCost),
						(1-sonnetCost/s.CostUSD)*100),
				})
			}
		}
	}

	// MID_SESSION_SWITCH — model changed mid-session, invalidating the cache.
	if len(s.Turns) > 1 {
		for i := 1; i < len(s.Turns); i++ {
			prev := s.Turns[i-1].Model
			curr := s.Turns[i].Model
			if prev != "" && curr != "" && prev != "unknown" && curr != "unknown" && prev != curr {
				out = append(out, Anomaly{
					Severity:    "warn",
					SessionID:   s.ID,
					ProjectName: s.ProjectName,
					Code:        "MID_SESSION_SWITCH",
					Title:       fmt.Sprintf("%s — model switched at turn %d", name, i+1),
					Detail: fmt.Sprintf(
						"Model changed from %s → %s. "+
							"Every model switch invalidates the prompt cache, "+
							"forcing full cache re-creation on the next turn.",
						prev, curr),
				})
				break // report the first switch only
			}
		}
	} else if len(s.AllModels) > 1 {
		// Fallback: no per-turn data but AllModels shows diversity.
		out = append(out, Anomaly{
			Severity:    "warn",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "MID_SESSION_SWITCH",
			Title:       fmt.Sprintf("%s — multiple models used", name),
			Detail: fmt.Sprintf(
				"Models seen: %s. Each switch invalidates the prompt cache.",
				strings.Join(s.AllModels, ", ")),
		})
	}

	// THINKING_SPIKE — a single turn's output >> session median (extended thinking fired).
	if len(s.Turns) >= 5 {
		outputs := make([]int64, len(s.Turns))
		for i, t := range s.Turns {
			outputs[i] = t.OutputTokens
		}
		sort.Slice(outputs, func(i, j int) bool { return outputs[i] < outputs[j] })
		median := outputs[len(outputs)/2]
		if median > 200 {
			for i, t := range s.Turns {
				if t.OutputTokens > median*5 {
					out = append(out, Anomaly{
						Severity:    "info",
						SessionID:   s.ID,
						ProjectName: s.ProjectName,
						Code:        "THINKING_SPIKE",
						Title: fmt.Sprintf("%s — thinking spike at turn %d (%s tokens)",
							name, i+1, formatTokens(t.OutputTokens)),
						Detail: fmt.Sprintf(
							"Turn %d output was %s tokens vs session median of %s. "+
								"Extended thinking likely fired. "+
								"For mechanical tasks use /effort low or set MAX_THINKING_TOKENS=2000.",
							i+1, formatTokens(t.OutputTokens), formatTokens(median)),
					})
					break // report the first spike only
				}
			}
		}
	}

	// SUBAGENT_MODEL_COST — many subagents, likely inheriting heavy parent model.
	if s.SubagentCount >= 3 && isOpus(s.Model) {
		out = append(out, Anomaly{
			Severity:    "info",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "SUBAGENT_MODEL_COST",
			Title:       fmt.Sprintf("%s — %d subagents likely running Opus", name, s.SubagentCount),
			Detail: fmt.Sprintf(
				"%d subagents, each inheriting the parent model by default. "+
					"Set CLAUDE_CODE_SUBAGENT_MODEL=claude-sonnet-4-6 to cut subagent cost ~80%%. "+
					"Most subagent tasks (file reads, focused edits) don't need Opus.",
				s.SubagentCount),
		})
	}

	return out
}

// checkGeneralHealth fires on the existing CACHE_MISS_HIGH, SHORT_SESSION_COST,
// and GOOD_CACHE signals.
func checkGeneralHealth(s *aggregator.SessionStats) []Anomaly {
	var out []Anomaly
	name := sessionName(s)

	if s.CacheEfficiency < 0.15 && s.TotalTokens > 500_000 {
		out = append(out, Anomaly{
			Severity:    "warn",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "CACHE_MISS_HIGH",
			Title:       name,
			Detail: fmt.Sprintf(
				"Cache efficiency %.0f%% — cold-start session with large context rebuild. "+
					"%d messages cost %s (%s/msg).",
				s.CacheEfficiency*100, s.MessageCount,
				ui.FormatCost(s.CostUSD),
				ui.FormatCostPerMessage(s.CostUSD, s.MessageCount)),
		})
	}

	if s.MessageCount < 5 && s.CostUSD > 3.0 {
		out = append(out, Anomaly{
			Severity:    "warn",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "SHORT_SESSION_COST",
			Title:       name,
			Detail: fmt.Sprintf(
				"High cost for a short session — %d messages cost %s.",
				s.MessageCount, ui.FormatCost(s.CostUSD)),
		})
	}

	if s.CacheEfficiency >= 0.80 && s.MessageCount > 10 {
		out = append(out, Anomaly{
			Severity:    "info",
			SessionID:   s.ID,
			ProjectName: s.ProjectName,
			Code:        "GOOD_CACHE",
			Title: fmt.Sprintf(
				"%s — %.0f%% cache efficiency", name, s.CacheEfficiency*100),
			Detail: fmt.Sprintf(
				"Over %s. Cost/message: %s.",
				ui.FormatDuration(s.Duration),
				ui.FormatCostPerMessage(s.CostUSD, s.MessageCount)),
		})
	}

	return out
}

// checkResumeOpportunity flags sessions that paid a cold-start cost when a recent
// session on the same project was available to resume.
func checkResumeOpportunity(sessions []*aggregator.SessionStats) []Anomaly {
	var out []Anomaly

	// Group sessions by project hash, sorted by start time.
	byProject := make(map[string][]*aggregator.SessionStats)
	for _, s := range sessions {
		if s.ProjectHash == "" {
			continue
		}
		byProject[s.ProjectHash] = append(byProject[s.ProjectHash], s)
	}

	for _, group := range byProject {
		sort.Slice(group, func(i, j int) bool {
			return group[i].StartedAt.Before(group[j].StartedAt)
		})
		for i := 1; i < len(group); i++ {
			prev := group[i-1]
			curr := group[i]

			if curr.Turn1CacheCreate < 80_000 {
				continue // cold-start cost wasn't significant
			}
			if prev.EndedAt.IsZero() || curr.StartedAt.IsZero() {
				continue
			}
			gap := curr.StartedAt.Sub(prev.EndedAt)
			if gap < 0 || gap > 24*time.Hour {
				continue
			}

			cost := coldStartCost(curr.Turn1CacheCreate, curr.Model)
			if cost < 0.10 {
				continue
			}

			prevID := prev.ID
			if len(prevID) > 8 {
				prevID = prevID[:8]
			}
			out = append(out, Anomaly{
				Severity:    "tip",
				SessionID:   curr.ID,
				ProjectName: curr.ProjectName,
				Code:        "RESUME_OPPORTUNITY",
				Title:       fmt.Sprintf("Could have resumed session %s (saved %s)", prevID, ui.FormatCost(cost)),
				Detail: fmt.Sprintf(
					"Session %s on %s ended %s before this one started. "+
						"Starting fresh paid %s tokens in cache creation (%s). "+
						"Use claude --resume %s to continue an existing session.",
					prevID, curr.ProjectName,
					formatDuration(gap),
					formatTokens(curr.Turn1CacheCreate), ui.FormatCost(cost),
					prev.ID),
			})
		}
	}

	return out
}

// ─── formatting helpers ──────────────────────────────────────────────────────

func formatTokens(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.0fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatBytes(n int64) string {
	if n >= 1_048_576 {
		return fmt.Sprintf("%.1f MB", float64(n)/1_048_576)
	}
	if n >= 1_024 {
		return fmt.Sprintf("%.0f KB", float64(n)/1_024)
	}
	return fmt.Sprintf("%d B", n)
}

func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// ─── output ──────────────────────────────────────────────────────────────────

func outputJSONAnomalies(anomalies []Anomaly) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(anomalies)
}
