package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

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
	themeOpt  string
)

var rootCmd = &cobra.Command{
	Use:   "agentop",
	Short: "Terminal dashboard for AI coding assistant sessions",
	Long: `agentop reads AI assistant session data and shows token usage,
cost, and cache efficiency in a duf-style terminal dashboard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runToday(cmd, args)
	},
}

func init() {
	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".claude")

	rootCmd.PersistentFlags().StringVar(&claudeDir, "claude-dir", defaultDir, "Path to AI assistant data directory")
	rootCmd.PersistentFlags().StringVar(&since, "since", "today", `Date filter: "today", "7d", "30d", or "2026-04-01"`)
	rootCmd.PersistentFlags().StringVar(&until, "until", "", "End date filter")
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "Filter by project path (partial match)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "", "Filter by model: opus, sonnet, haiku")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&compact, "compact", false, "Force compact layout")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colors")
	rootCmd.PersistentFlags().BoolVarP(&watch, "watch", "w", false, "Live refresh mode")
	rootCmd.PersistentFlags().IntVar(&refresh, "refresh", 5, "Refresh interval in seconds (--watch mode)")
	rootCmd.PersistentFlags().StringVar(&themeOpt, "theme", defaultThemeName(), "Color themes: dark, light, ansi")

	rootCmd.Version = printVersion()

	rootCmd.AddCommand(todayCmd, dailyCmd, monthlyCmd, sessionCmd, blocksCmd, doctorCmd, configCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	Version   = ""
	CommitSHA = ""
)

func SetVersionInfo(v, c string) {
	Version = v
	CommitSHA = c
}

func printVersion() string {
	info, ok := debug.ReadBuildInfo()
	var buildTime time.Time
	var modified bool
	if ok {
		if len(Version) == 0 {
			vs := info.Main.Version
			if vs != "" && vs != "(devel)" {
				Version = vs
			}
		}

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(CommitSHA) == 0 {
					CommitSHA = setting.Value
					if len(CommitSHA) > 12 {
						CommitSHA = CommitSHA[:12]
					}
				}
			case "vcs.time":
				buildTime, _ = time.Parse(time.RFC3339, setting.Value)
			case "vcs.modified":
				modified, _ = strconv.ParseBool(setting.Value)
			}
		}
	}

	if Version == "" || Version == "(devel)" {
		Version = "(built from source)"
	}

	ver := fmt.Sprintf("agentop %s", Version)
	if len(CommitSHA) > 0 {
		if modified {
			CommitSHA += "+modified"
		}
		ver += fmt.Sprintf(" (%s)", CommitSHA)
	}
	if !buildTime.IsZero() {
		ver += fmt.Sprintf(" (built on %s)", buildTime.Format("2006-01-02"))
	}

	return ver
}
