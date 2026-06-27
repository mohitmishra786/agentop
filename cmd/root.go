package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/agentop-dev/agentop/internal/adapter"
	claudeAdapter "github.com/agentop-dev/agentop/internal/adapter/claude"
	codexAdapter "github.com/agentop-dev/agentop/internal/adapter/codex"
	copilotAdapter "github.com/agentop-dev/agentop/internal/adapter/copilot"
	geminiAdapter "github.com/agentop-dev/agentop/internal/adapter/gemini"
	kiroAdapter "github.com/agentop-dev/agentop/internal/adapter/kiro"
	opencodeAdapter "github.com/agentop-dev/agentop/internal/adapter/opencode"
	cursorAdapter "github.com/agentop-dev/agentop/internal/adapter/cursor"
	continueAdapter "github.com/agentop-dev/agentop/internal/adapter/continueadapter"
	jetbrainsAdapter "github.com/agentop-dev/agentop/internal/adapter/jetbrains"
	windsurfAdapter "github.com/agentop-dev/agentop/internal/adapter/windsurf"
	grokAdapter "github.com/agentop-dev/agentop/internal/adapter/grok"
)

var (
	claudeDir  string
	since      string
	until      string
	project    string
	model      string
	agentFlag  string
	listAgents bool
	jsonOut    bool
	compact    bool
	noColor    bool
	watch      bool
	refresh    int
	themeOpt   string
	layout     string

	registry = adapter.NewRegistry()
)

func init() {
	registry.Register(&claudeAdapter.Adapter{})
	registry.Register(&codexAdapter.Adapter{})
	registry.Register(&copilotAdapter.Adapter{})
	registry.Register(&geminiAdapter.Adapter{})
	registry.Register(&kiroAdapter.Adapter{})
	registry.Register(&opencodeAdapter.Adapter{})
	registry.Register(&cursorAdapter.Adapter{})
	registry.Register(&continueAdapter.Adapter{})
	registry.Register(&jetbrainsAdapter.Adapter{})
	registry.Register(&windsurfAdapter.Adapter{})
	registry.Register(&grokAdapter.Adapter{})
}

var rootCmd = &cobra.Command{
	Use:   "agentop",
	Short: "Terminal dashboard for AI coding assistant sessions",
	Long: `agentop reads AI assistant session data and shows token usage,
cost, and cache efficiency in a duf-style terminal dashboard.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if listAgents {
			fmt.Print(registry.ListAgentsText())
			return fmt.Errorf("__list_agents__")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runToday(cmd, args)
	},
}

func init() {
	defaultDir := findClaudeDir()

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
	rootCmd.PersistentFlags().StringVar(&layout, "layout", "default", "Layout style: default, table")
	rootCmd.PersistentFlags().StringVar(&agentFlag, "agent", "all", `Agent(s) to query: "all", or comma-separated IDs like "claude,codex"`)
	rootCmd.PersistentFlags().BoolVar(&listAgents, "list-agents", false, "List all registered agents and their availability")

	rootCmd.Version = printVersion()

	rootCmd.AddCommand(todayCmd, dailyCmd, monthlyCmd, sessionCmd, blocksCmd, doctorCmd, configCmd, mcpCmd, budgetCmd)
}

func resolveAgentIDs() []adapter.AgentID {
	return adapter.ParseAgentFlag(agentFlag, registry)
}

// findClaudeDir returns the first existing .claude directory from a list of
// candidate locations. On Windows, Claude Code may place data under
// %APPDATA%\Claude or %LOCALAPPDATA%\Claude in addition to the standard
// %USERPROFILE%\.claude path.
func findClaudeDir() string {
	var candidates []string

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".claude"))
	}

	// Windows-specific fallbacks (%APPDATA% and %LOCALAPPDATA%)
	if appData := os.Getenv("APPDATA"); appData != "" {
		candidates = append(candidates, filepath.Join(appData, "Claude"))
		candidates = append(candidates, filepath.Join(appData, ".claude"))
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "Claude"))
		candidates = append(candidates, filepath.Join(localAppData, ".claude"))
	}

	for _, dir := range candidates {
		projectsDir := filepath.Join(dir, "projects")
		if info, err := os.Stat(projectsDir); err == nil && info.IsDir() {
			return dir
		}
	}

	// Fall back to the first candidate (standard ~/.claude) even if it doesn't exist yet
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ".claude"
}

func Execute() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		if err.Error() != "__list_agents__" {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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
