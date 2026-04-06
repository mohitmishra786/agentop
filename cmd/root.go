package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	claudeDir string
	since     string
	until     string
	project   string
	model     string
	jsonOut   bool
	compact   bool
	noColor   bool
	watch     bool
	refresh   int
)

var rootCmd = &cobra.Command{
	Use:   "agentop",
	Short: "Terminal dashboard for Claude Code sessions",
	Long: `agentop reads ~/.claude/projects/ and shows token usage,
cost, and cache efficiency in a duf-style terminal dashboard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runToday(cmd, args)
	},
}

func init() {
	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".claude")

	rootCmd.PersistentFlags().StringVar(&claudeDir, "claude-dir", defaultDir, "Path to Claude data directory")
	rootCmd.PersistentFlags().StringVar(&since, "since", "today", `Date filter: "today", "7d", "30d", or "2026-04-01"`)
	rootCmd.PersistentFlags().StringVar(&until, "until", "", "End date filter")
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "Filter by project path (partial match)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "", "Filter by model: opus, sonnet, haiku")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&compact, "compact", false, "Force compact layout")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colors")
	rootCmd.PersistentFlags().BoolVarP(&watch, "watch", "w", false, "Live refresh mode")
	rootCmd.PersistentFlags().IntVar(&refresh, "refresh", 5, "Refresh interval in seconds (--watch mode)")

	rootCmd.AddCommand(todayCmd, dailyCmd, monthlyCmd, sessionCmd, blocksCmd, doctorCmd, configCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

func getVersionInfo() string {
	return fmt.Sprintf("agentop %s (commit: %s, built: %s)", version, commit, date)
}
