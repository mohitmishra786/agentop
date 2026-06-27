package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/ui"
	"github.com/spf13/cobra"
)

type BudgetConfig struct {
	MonthlyLimit float64 `json:"monthly_limit"`
	WarnAt       float64 `json:"warn_at"`
	DailyLimit   float64 `json:"daily_limit"`
}

type budgetAgentSummary struct {
	AgentID  string  `json:"agent_id"`
	Cost     float64 `json:"cost"`
	Sessions int     `json:"sessions"`
}

type budgetProjectSummary struct {
	Project  string  `json:"project"`
	Cost     float64 `json:"cost"`
	Sessions int     `json:"sessions"`
}

type budgetJSONOutput struct {
	MonthlyLimit    float64                 `json:"monthly_limit"`
	WarnAt          float64                 `json:"warn_at"`
	DailyLimit      float64                 `json:"daily_limit"`
	TotalCost       float64                 `json:"total_cost"`
	UsedPercent     float64                 `json:"used_percent"`
	SessionCount    int                     `json:"session_count"`
	DailyRunRate    float64                 `json:"daily_run_rate"`
	EstimatedMonthly float64                `json:"estimated_monthly"`
	DayOfMonth      int                     `json:"day_of_month"`
	DaysInMonth     int                     `json:"days_in_month"`
	ByAgent         []budgetAgentSummary    `json:"by_agent"`
	ByProject       []budgetProjectSummary  `json:"by_project"`
	Status          string                  `json:"status"`
}

var (
	budgetLimit    float64
	budgetSetLimit float64
)

var budgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Show budget status and spending against monthly limits",
	RunE:  runBudget,
}

func init() {
	budgetCmd.Flags().Float64Var(&budgetLimit, "limit", 0, "Override monthly budget limit")
	budgetCmd.Flags().Float64Var(&budgetSetLimit, "set-limit", 0, "Persist a new monthly limit to config")
}

func loadBudgetConfig() BudgetConfig {
	cfg := BudgetConfig{
		MonthlyLimit: 50.0,
		WarnAt:       0.8,
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "agentop", "budget.json"))
	if err != nil {
		return cfg
	}
	var fc BudgetConfig
	if json.Unmarshal(data, &fc) != nil {
		return cfg
	}
	if fc.MonthlyLimit != 0 {
		cfg.MonthlyLimit = fc.MonthlyLimit
	}
	if fc.WarnAt != 0 {
		cfg.WarnAt = fc.WarnAt
	}
	cfg.DailyLimit = fc.DailyLimit
	return cfg
}

func persistBudgetLimit(limit float64) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "agentop")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	cfg := loadBudgetConfig()
	cfg.MonthlyLimit = limit
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(configDir, "budget.json"), data, 0644); err != nil {
		return err
	}
	fmt.Printf("  %s Budget limit set to $%.2f\n", ui.StyleGreen.Render("✓"), limit)
	return nil
}

func runBudget(cmd *cobra.Command, args []string) error {
	if err := ui.Init(); err != nil {
		return err
	}

	if cmd.Flags().Changed("set-limit") {
		return persistBudgetLimit(budgetSetLimit)
	}

	config := loadBudgetConfig()
	if cmd.Flags().Changed("limit") {
		config.MonthlyLimit = budgetLimit
	}

	sessions, err := loadSessions()
	if err != nil {
		return err
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var monthSessions []*aggregator.SessionStats
	for _, s := range sessions {
		if !s.StartedAt.IsZero() && !s.StartedAt.Before(monthStart) {
			monthSessions = append(monthSessions, s)
		}
	}

	totalCost := 0.0
	for _, s := range monthSessions {
		totalCost += s.CostUSD
	}

	dayOfMonth := now.Day()
	if dayOfMonth < 1 {
		dayOfMonth = 1
	}
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	dailyRunRate := totalCost / float64(dayOfMonth)
	estimatedMonthly := dailyRunRate * float64(daysInMonth)
	usedPct := 0.0
	if config.MonthlyLimit > 0 {
		usedPct = totalCost / config.MonthlyLimit
	}

	byAgent := map[string]*budgetAgentSummary{}
	for _, s := range monthSessions {
		aid := string(s.AgentID)
		if aid == "" {
			aid = "?"
		}
		if byAgent[aid] == nil {
			byAgent[aid] = &budgetAgentSummary{AgentID: aid}
		}
		byAgent[aid].Cost += s.CostUSD
		byAgent[aid].Sessions++
	}
	var agentSummaries []budgetAgentSummary
	for _, v := range byAgent {
		agentSummaries = append(agentSummaries, *v)
	}
	sort.Slice(agentSummaries, func(i, j int) bool {
		return agentSummaries[i].Cost > agentSummaries[j].Cost
	})

	byProject := map[string]*budgetProjectSummary{}
	for _, s := range monthSessions {
		proj := s.ProjectName
		if proj == "" {
			proj = "(unknown)"
		}
		if byProject[proj] == nil {
			byProject[proj] = &budgetProjectSummary{Project: proj}
		}
		byProject[proj].Cost += s.CostUSD
		byProject[proj].Sessions++
	}
	var projectSummaries []budgetProjectSummary
	for _, v := range byProject {
		projectSummaries = append(projectSummaries, *v)
	}
	sort.Slice(projectSummaries, func(i, j int) bool {
		return projectSummaries[i].Cost > projectSummaries[j].Cost
	})
	topN := 3
	if len(projectSummaries) < topN {
		topN = len(projectSummaries)
	}
	topProjects := projectSummaries[:topN]

	var status string
	var barColor string
	switch {
	case usedPct >= 1.0:
		status = "OVER BUDGET"
		barColor = "#E05555"
	case usedPct >= 0.8:
		status = "NEARING LIMIT"
		barColor = "#E05555"
	case usedPct >= 0.5:
		status = "CAUTION"
		barColor = "#E8A838"
	default:
		status = "ON TRACK"
		barColor = "#73D673"
	}

	if jsonOut {
		out := budgetJSONOutput{
			MonthlyLimit:     config.MonthlyLimit,
			WarnAt:           config.WarnAt,
			DailyLimit:       config.DailyLimit,
			TotalCost:        totalCost,
			UsedPercent:      usedPct,
			SessionCount:     len(monthSessions),
			DailyRunRate:     dailyRunRate,
			EstimatedMonthly: estimatedMonthly,
			DayOfMonth:       dayOfMonth,
			DaysInMonth:      daysInMonth,
			ByAgent:          agentSummaries,
			ByProject:        projectSummaries,
			Status:           status,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	var sb strings.Builder

	monthLabel := now.Format("Jan 2006")
	sb.WriteString(fmt.Sprintf("\n  %s  %s\n\n",
		ui.StyleBold.Render("Budget  "+monthLabel),
		ui.StyleDim.Render("● "+status)))

	limitStr := fmt.Sprintf("$%.2f", config.MonthlyLimit)
	costStr := fmt.Sprintf("$%.2f", totalCost)
	bar := ui.MiniBar(usedPct, 20, barColor)

	var pctStr string
	switch {
	case usedPct >= 1.0:
		pctStr = ui.StyleRed.Render(fmt.Sprintf("%.0f%%", usedPct*100))
	case usedPct >= config.WarnAt:
		pctStr = ui.StyleAmber.Render(fmt.Sprintf("%.0f%%", usedPct*100))
	default:
		pctStr = ui.StyleGreen.Render(fmt.Sprintf("%.0f%%", usedPct*100))
	}

	sb.WriteString(fmt.Sprintf("  %s  %s  %s  %s\n",
		ui.StylePrimary.Render(costStr),
		ui.StyleDim.Render("of"),
		ui.StylePrimary.Render(limitStr),
		bar))
	sb.WriteString(fmt.Sprintf("  %s\n\n",
		pctStr))

	estStr := fmt.Sprintf("$%.2f", estimatedMonthly)
	rateStr := fmt.Sprintf("$%.2f/day", dailyRunRate)
	sb.WriteString(fmt.Sprintf("  %s  %s\n",
		ui.StyleDim.Render("Estimated monthly:"),
		ui.StylePrimary.Render(estStr)))
	sb.WriteString(fmt.Sprintf("  %s  %s\n\n",
		ui.StyleDim.Render("Daily run rate:   "),
		ui.StyleSecondary.Render(rateStr)))

	if len(agentSummaries) > 1 {
		sb.WriteString(fmt.Sprintf("  %s\n", ui.StyleColHeader.Render("By agent")))
		for _, a := range agentSummaries {
			aPct := 0.0
			if totalCost > 0 {
				aPct = a.Cost / totalCost
			}
			aCostStr := fmt.Sprintf("$%.2f", a.Cost)
			aPctStr := fmt.Sprintf("%.1f%%", aPct*100)
			tag := ui.AgentTag(a.AgentID)
			sb.WriteString(fmt.Sprintf("  %s  %s  %s\n", tag, ui.StylePrimary.Render(aCostStr), ui.StyleDim.Render(aPctStr)))
		}
		sb.WriteString("\n")
	}

	if len(topProjects) > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", ui.StyleColHeader.Render("By project (top 3)")))
		for _, p := range topProjects {
			pPct := 0.0
			if totalCost > 0 {
				pPct = p.Cost / totalCost
			}
			pCostStr := fmt.Sprintf("$%.2f", p.Cost)
			pPctStr := fmt.Sprintf("%.1f%%", pPct*100)
			projName := p.Project
			if len(projName) > 28 {
				projName = projName[:28] + "…"
			}
			sb.WriteString(fmt.Sprintf("  %-28s  %s  %s\n",
				ui.StylePrimary.Render(projName),
				ui.StyleSecondary.Render(pCostStr),
				ui.StyleDim.Render(pPctStr)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("  %s\n", ui.StyleDim.Render("agentop budget · --help for flags")))
	fmt.Print(sb.String())
	return nil
}
