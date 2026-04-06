# Build Prompt: agentop (Claude Code Edition)
## Paste this into Claude Code as your initial prompt in a fresh project directory

---

You are building `agentop` — a terminal dashboard for Claude Code sessions, inspired by `duf` (the disk usage formatter). It reads `~/.claude/projects/` JSONL files and renders a beautiful, panel-based TUI showing token usage, cost, cache efficiency, and insights — without any server, account, or API key required.

**Language: Go 1.22+**
**UI: Charm ecosystem (bubbletea + lipgloss + bubbles)**
**CLI: Cobra**
**Distribution: goreleaser (single static binary)**

---

## Phase 1: Project Scaffold (do this first)

Initialize the Go module and create the full directory structure:

```
agentop/
├── main.go
├── go.mod                       # module: github.com/yourname/agentop
├── go.sum
├── .goreleaser.yml
├── Makefile
├── README.md
├── cmd/
│   ├── root.go
│   ├── today.go
│   ├── daily.go
│   ├── monthly.go
│   ├── session.go
│   ├── blocks.go
│   ├── doctor.go
│   └── config.go
├── internal/
│   ├── claude/
│   │   ├── discover.go
│   │   ├── jsonl.go
│   │   ├── schema.go
│   │   └── index.go
│   ├── aggregator/
│   │   ├── session.go
│   │   ├── daily.go
│   │   ├── blocks.go
│   │   ├── monthly.go
│   │   └── doctor.go
│   ├── pricing/
│   │   ├── models.go
│   │   └── embedded.go
│   ├── ui/
│   │   ├── app.go
│   │   ├── styles.go
│   │   ├── bars.go
│   │   ├── panels.go
│   │   ├── summary.go
│   │   ├── table.go
│   │   └── helpers.go
│   ├── watcher/
│   │   └── watcher.go
│   └── config/
│       └── config.go
├── assets/
│   └── pricing.json
└── testdata/
    └── sessions/
        ├── good-cache.jsonl
        ├── cold-start.jsonl
        └── in-progress.jsonl
```

Run `go mod init github.com/yourname/agentop` then add all dependencies:

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/spf13/cobra@latest
go get github.com/fsnotify/fsnotify@latest
go get github.com/dustin/go-humanize@latest
go get github.com/adrg/xdg@latest
```

---

## Phase 2: Data Layer (build in this order)

### 2.1 Schema (internal/claude/schema.go)

Define these exact Go structs to match Claude Code's JSONL format:

```go
package claude

import (
    "encoding/json"
    "time"
)

type RawEvent struct {
    Type       string          `json:"type"`
    Message    *RawMessage     `json:"message,omitempty"`
    ToolName   string          `json:"toolName,omitempty"`
    ToolInput  json.RawMessage `json:"toolInput,omitempty"`
    Summary    string          `json:"summary,omitempty"`
    SessionID  string          `json:"sessionId"`
    UUID       string          `json:"uuid"`
    ParentUUID string          `json:"parentUuid,omitempty"`
    Timestamp  time.Time       `json:"timestamp"`
    CostUSD    float64         `json:"costUSD,omitempty"`
    CWD        string          `json:"cwd,omitempty"`
    IsSidechain bool           `json:"isSidechain,omitempty"`
}

type RawMessage struct {
    ID      string    `json:"id"`
    Role    string    `json:"role"`
    Model   string    `json:"model,omitempty"`
    Usage   *Usage    `json:"usage,omitempty"`
    Content []Content `json:"content,omitempty"`
}

type Usage struct {
    InputTokens              int `json:"input_tokens"`
    OutputTokens             int `json:"output_tokens"`
    CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
    CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type Content struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

// SessionMeta comes from sessions-index.json
type SessionMeta struct {
    ID               string    `json:"id"`
    Summary          string    `json:"summary"`
    FirstUserMessage string    `json:"firstUserMessage"`
    MessageCount     int       `json:"messageCount"`
    CreatedAt        time.Time `json:"createdAt"`
    UpdatedAt        time.Time `json:"updatedAt"`
    GitBranch        string    `json:"gitBranch"`
    CWD              string    `json:"cwd"`
    ProjectPath      string    `json:"projectPath"`
}
```

### 2.2 Directory Discovery (internal/claude/discover.go)

Scan `~/.claude/projects/` for all session JSONL files:

```go
package claude

import (
    "os"
    "path/filepath"
    "sort"
    "strings"
)

var ErrClaudeNotFound = errors.New("~/.claude/projects/ not found — is Claude Code installed?")

type SessionFile struct {
    Path        string
    ProjectHash string  // the directory name (hash of project path)
    SessionID   string  // filename without .jsonl
    ModTime     time.Time
}

func Discover(claudeDir string) ([]SessionFile, error) {
    projectsDir := filepath.Join(claudeDir, "projects")
    
    projectEntries, err := os.ReadDir(projectsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrClaudeNotFound
        }
        return nil, fmt.Errorf("reading projects dir: %w", err)
    }

    var sessions []SessionFile
    for _, projectEntry := range projectEntries {
        if !projectEntry.IsDir() { continue }
        
        sessionsDir := filepath.Join(projectsDir, projectEntry.Name(), "sessions")
        fileEntries, err := os.ReadDir(sessionsDir)
        if err != nil { continue } // project may have no sessions dir
        
        for _, f := range fileEntries {
            if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") { continue }
            info, _ := f.Info()
            sessions = append(sessions, SessionFile{
                Path:        filepath.Join(sessionsDir, f.Name()),
                ProjectHash: projectEntry.Name(),
                SessionID:   strings.TrimSuffix(f.Name(), ".jsonl"),
                ModTime:     info.ModTime(),
            })
        }
    }

    sort.Slice(sessions, func(i, j int) bool {
        return sessions[i].ModTime.After(sessions[j].ModTime)
    })

    return sessions, nil
}

// ReadSessionsIndex reads the sessions-index.json for a project hash
func ReadSessionsIndex(claudeDir, projectHash string) ([]SessionMeta, error) {
    path := filepath.Join(claudeDir, "projects", projectHash, "sessions-index.json")
    data, err := os.ReadFile(path)
    if err != nil { return nil, err }
    
    var result struct {
        Sessions []SessionMeta `json:"sessions"`
    }
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    return result.Sessions, nil
}
```

### 2.3 JSONL Parser (internal/claude/jsonl.go)

Streaming parser — never loads entire file into memory:

```go
package claude

import (
    "bufio"
    "encoding/json"
    "os"
)

const maxLineBytes = 10 * 1024 * 1024 // 10MB (tool results can be large)

func ParseSession(path string) ([]RawEvent, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    scanner.Buffer(make([]byte, maxLineBytes), maxLineBytes)

    var events []RawEvent
    lineNum := 0
    skipped := 0

    for scanner.Scan() {
        lineNum++
        line := bytes.TrimSpace(scanner.Bytes())
        if len(line) == 0 { continue }

        var e RawEvent
        if err := json.Unmarshal(line, &e); err != nil {
            skipped++
            continue // silently skip malformed lines (truncated in-progress writes)
        }
        events = append(events, e)
    }

    if err := scanner.Err(); err != nil {
        // ErrTooLong means a single line exceeded our buffer — skip it
        if errors.Is(err, bufio.ErrTooLong) {
            return events, nil // return what we got
        }
        return events, err
    }

    return events, nil
}
```

### 2.4 Session Aggregation (internal/aggregator/session.go)

This is the heart of the tool. Aggregate a parsed session into a `SessionStats`:

```go
package aggregator

import (
    "math"
    "sort"
    "time"
    "github.com/yourname/agentop/internal/claude"
)

type SessionStats struct {
    ID          string
    ProjectHash string
    ProjectPath string
    ProjectName string // basename of ProjectPath
    Summary     string
    GitBranch   string
    Model       string        // dominant model
    AllModels   []string      // all models seen (deduplicated)

    StartedAt  time.Time
    EndedAt    time.Time
    Duration   time.Duration

    InputTokens       int64
    OutputTokens      int64
    CacheCreateTokens int64
    CacheReadTokens   int64
    TotalTokens       int64

    CostUSD           float64
    CostUSDCalculated float64

    MessageCount int   // human turns
    TurnCount    int   // all turns
    ToolCalls    map[string]int

    CacheEfficiency float64
    CostPerMessage  float64
    BurnRate        float64 // tokens per minute
    WasCompacted    bool
}

func AggregateSession(events []claude.RawEvent, meta *claude.SessionMeta, pricer Pricer) *SessionStats {
    if len(events) == 0 { return nil }

    s := &SessionStats{
        ToolCalls: make(map[string]int),
    }

    // Identify session ID from first event
    s.ID = events[0].SessionID

    // Track models seen (for per-model cost calculation)
    modelTokens := make(map[string]claude.Usage)
    
    var firstTime, lastTime time.Time

    for _, e := range events {
        if firstTime.IsZero() || e.Timestamp.Before(firstTime) {
            firstTime = e.Timestamp
        }
        if e.Timestamp.After(lastTime) {
            lastTime = e.Timestamp
        }

        switch e.Type {
        case "human":
            s.MessageCount++
            s.TurnCount++
            // Extract cwd from first human message
            if s.ProjectPath == "" && e.CWD != "" {
                s.ProjectPath = e.CWD
                s.ProjectName = filepath.Base(e.CWD)
            }

        case "assistant":
            s.TurnCount++
            if e.Message == nil { continue }
            
            usage := e.Message.Usage
            if usage == nil { continue }

            model := e.Message.Model
            if model == "" { model = "unknown" }

            // Accumulate tokens
            s.InputTokens       += int64(usage.InputTokens)
            s.OutputTokens      += int64(usage.OutputTokens)
            s.CacheCreateTokens += int64(usage.CacheCreationInputTokens)
            s.CacheReadTokens   += int64(usage.CacheReadInputTokens)

            // Accumulate cost from Anthropic's pre-calculated value
            s.CostUSD += e.CostUSD

            // Track per-model usage for calculated cost
            existing := modelTokens[model]
            existing.InputTokens              += usage.InputTokens
            existing.OutputTokens             += usage.OutputTokens
            existing.CacheCreationInputTokens += usage.CacheCreationInputTokens
            existing.CacheReadInputTokens     += usage.CacheReadInputTokens
            modelTokens[model] = existing

        case "tool":
            s.ToolCalls[e.ToolName]++

        case "summary":
            if s.Summary == "" {
                s.Summary = e.Summary
            }
        }
    }

    // Compute derived fields
    s.TotalTokens = s.InputTokens + s.OutputTokens + s.CacheCreateTokens + s.CacheReadTokens
    s.StartedAt = firstTime
    s.EndedAt = lastTime
    s.Duration = lastTime.Sub(firstTime)

    // Cache efficiency
    denom := s.InputTokens + s.CacheCreateTokens + s.CacheReadTokens
    if denom > 0 {
        s.CacheEfficiency = float64(s.CacheReadTokens) / float64(denom)
    }

    // Calculated cost (independent verification)
    for model, usage := range modelTokens {
        s.CostUSDCalculated += pricer.Calculate(usage, model)
    }

    // Dominant model (by input token count)
    maxTokens := 0
    for model, usage := range modelTokens {
        if usage.InputTokens > maxTokens {
            maxTokens = usage.InputTokens
            s.Model = model
        }
        s.AllModels = append(s.AllModels, model)
    }
    sort.Strings(s.AllModels)

    // Cost per message
    if s.MessageCount > 0 {
        s.CostPerMessage = s.CostUSD / float64(s.MessageCount)
    }

    // Burn rate (tokens per minute)
    if s.Duration >= time.Minute {
        minutes := s.Duration.Minutes()
        s.BurnRate = float64(s.TotalTokens) / minutes
    }

    // Overlay sessions-index metadata if available
    if meta != nil {
        if s.Summary == "" { s.Summary = meta.Summary }
        if s.Summary == "" { s.Summary = meta.FirstUserMessage }
        if s.GitBranch == "" { s.GitBranch = meta.GitBranch }
        if s.ProjectPath == "" { s.ProjectPath = meta.CWD }
    }

    // Truncate summary for display
    if len(s.Summary) > 80 {
        s.Summary = s.Summary[:77] + "..."
    }

    return s
}
```

---

## Phase 3: Pricing Engine (internal/pricing/)

### embedded.go

```go
package pricing

import _ "embed"

//go:embed ../../assets/pricing.json
var embeddedPricing []byte
```

### models.go

```go
package pricing

import (
    "encoding/json"
    "strings"
)

type ModelPrice struct {
    Input       float64 `json:"input"`       // USD per 1M tokens
    Output      float64 `json:"output"`
    CacheCreate float64 `json:"cacheCreate"`
    CacheRead   float64 `json:"cacheRead"`
}

type PricingDB struct {
    Version string                `json:"version"`
    Models  map[string]ModelPrice `json:"models"`
}

var db *PricingDB

func init() {
    db = &PricingDB{}
    if err := json.Unmarshal(embeddedPricing, db); err != nil {
        panic("failed to parse embedded pricing: " + err.Error())
    }
}

func Get(model string) ModelPrice {
    // Exact match
    if p, ok := db.Models[model]; ok { return p }
    
    // Prefix match (e.g. "claude-opus-4" matches "claude-opus-4-6")
    modelLower := strings.ToLower(model)
    for name, price := range db.Models {
        if strings.HasPrefix(modelLower, strings.ToLower(name)) {
            return price
        }
    }
    
    // Fallback to Sonnet pricing
    return db.Models["claude-sonnet-4-6"]
}

func Calculate(u claude.Usage, model string) float64 {
    p := Get(model)
    return float64(u.InputTokens)              * p.Input       / 1e6 +
           float64(u.OutputTokens)             * p.Output      / 1e6 +
           float64(u.CacheCreationInputTokens) * p.CacheCreate / 1e6 +
           float64(u.CacheReadInputTokens)     * p.CacheRead   / 1e6
}
```

---

## Phase 4: UI Layer (internal/ui/)

### 4.1 styles.go — All colors and lipgloss styles

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — match the mockup exactly
const (
    ColGreen     = "#50c87a"
    ColAmber     = "#e5a040"
    ColRed       = "#e05050"
    ColBlue      = "#4a90d9"
    ColTeal      = "#7ecfb3"
    ColPurple    = "#b36de0"
    ColDim       = "#666680"
    ColBorder    = "#3a3a4a"
    ColHeader    = "#9898b8"
    ColText      = "#c8c8c8"
    ColTextBold  = "#c8c8e8"
)

var (
    StyleBorder = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(ColBorder))

    StyleHeader = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColHeader)).
        Background(lipgloss.Color("#222230")).
        Padding(0, 1)

    StyleDim = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColDim))

    StyleGreen = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColGreen)).
        Bold(true)

    StyleAmber = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColAmber)).
        Bold(true)

    StyleRed = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColRed)).
        Bold(true)

    StyleBold = lipgloss.NewStyle().
        Foreground(lipgloss.Color(ColTextBold))

    // Model tags
    StyleTagOpus = lipgloss.NewStyle().
        Background(lipgloss.Color("#3a2050")).
        Foreground(lipgloss.Color("#c090f0")).
        Padding(0, 1)

    StyleTagSonnet = lipgloss.NewStyle().
        Background(lipgloss.Color("#1a3050")).
        Foreground(lipgloss.Color("#70b0e0")).
        Padding(0, 1)

    StyleTagHaiku = lipgloss.NewStyle().
        Background(lipgloss.Color("#1e3520")).
        Foreground(lipgloss.Color("#90d080")).
        Padding(0, 1)
)

func ModelTag(model string) string {
    switch {
    case strings.Contains(model, "opus"):
        return StyleTagOpus.Render("opus")
    case strings.Contains(model, "sonnet"):
        return StyleTagSonnet.Render("sonnet")
    case strings.Contains(model, "haiku"):
        return StyleTagHaiku.Render("haiku")
    default:
        return StyleDim.Render(model)
    }
}

func CacheEfficiencyStyle(eff float64) lipgloss.Style {
    switch {
    case eff >= 0.80: return StyleGreen
    case eff >= 0.40: return StyleAmber
    default:          return StyleRed
    }
}
```

### 4.2 bars.go — Token bar rendering

```go
package ui

import (
    "strings"
    "github.com/charmbracelet/lipgloss"
)

type TokenBar struct {
    Input       int64
    Output      int64
    CacheCreate int64
    CacheRead   int64
    Width       int
}

func (b TokenBar) Render() string {
    total := b.Input + b.Output + b.CacheCreate + b.CacheRead
    if total == 0 || b.Width <= 0 {
        return lipgloss.NewStyle().
            Background(lipgloss.Color(ColBorder)).
            Width(b.Width).
            Render("")
    }

    type seg struct {
        count int64
        color string
    }
    segs := []seg{
        {b.Input, ColBlue},
        {b.Output, ColTeal},
        {b.CacheCreate, ColAmber},
        {b.CacheRead, ColPurple},
    }

    // Calculate widths proportionally, ensure they sum to exactly b.Width
    widths := make([]int, len(segs))
    totalWidth := 0
    for i, s := range segs {
        w := int(float64(s.count) / float64(total) * float64(b.Width))
        widths[i] = w
        totalWidth += w
    }
    // Distribute rounding error to largest segment
    if diff := b.Width - totalWidth; diff != 0 {
        maxIdx := 0
        for i := range widths {
            if widths[i] > widths[maxIdx] { maxIdx = i }
        }
        widths[maxIdx] += diff
    }

    var parts []string
    for i, s := range segs {
        if widths[i] <= 0 { continue }
        parts = append(parts, lipgloss.NewStyle().
            Background(lipgloss.Color(s.color)).
            Width(widths[i]).
            Render(""))
    }

    return strings.Join(parts, "")
}

// MiniBar renders a 6-char high summary bar (for summary strip cards)
func MiniBar(ratio float64, width int, color string) string {
    if ratio < 0 { ratio = 0 }
    if ratio > 1 { ratio = 1 }
    filled := int(ratio * float64(width))
    empty := width - filled

    filledStr := lipgloss.NewStyle().
        Background(lipgloss.Color(color)).
        Width(filled).Render("")
    emptyStr := lipgloss.NewStyle().
        Background(lipgloss.Color(ColBorder)).
        Width(empty).Render("")

    return filledStr + emptyStr
}

// TokenSubtitle returns "0.3M in · 1.2M out · 0.5M cc · 4.9M cr"
func TokenSubtitle(input, output, cc, cr int64) string {
    parts := []string{
        humanizeTokens(input) + " in",
        humanizeTokens(output) + " out",
        humanizeTokens(cc) + " cc",
        humanizeTokens(cr) + " cr",
    }
    return StyleDim.Render(strings.Join(parts, " · "))
}

func humanizeTokens(n int64) string {
    switch {
    case n >= 1_000_000: return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
    case n >= 1_000:     return fmt.Sprintf("%.0fk", float64(n)/1_000)
    default:             return fmt.Sprintf("%d", n)
    }
}
```

### 4.3 table.go — Grid row rendering

Define a flexible grid system that respects terminal width:

```go
package ui

// Column defines one column in the grid
type Column struct {
    Header    string
    Width     int     // in terminal chars
    Align     string  // "left", "right"
    MinWidth  int     // don't render below this (compact mode)
}

// FullColumns is the default column set for agentop today view
func FullColumns(termWidth int) []Column {
    // Bar column takes remaining space
    fixed := 22 + 8 + 8 + 6 + 10 + 9 // other columns + separators
    barWidth := termWidth - fixed
    if barWidth < 20 { barWidth = 20 }

    return []Column{
        {Header: "session",  Width: 22, Align: "left"},
        {Header: "token mix", Width: barWidth, Align: "left"},
        {Header: "cache%",   Width: 8,  Align: "right"},
        {Header: "msgs",     Width: 6,  Align: "right", MinWidth: 80},
        {Header: "tools",    Width: 6,  Align: "right", MinWidth: 100},
        {Header: "cost",     Width: 9,  Align: "right"},
        {Header: "time",     Width: 8,  Align: "right"},
    }
}
```

---

## Phase 5: Commands (cmd/)

### 5.1 root.go

```go
package cmd

import (
    "os"
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
```

### 5.2 today.go (the main view)

Implement `runToday()` which:
1. Calls `claude.Discover(claudeDir)` to list all session files
2. Filters to sessions modified today (or matching --since)
3. For each matching session, calls `claude.ParseSession(path)`
4. Calls `aggregator.AggregateSession(events, meta, pricer)` for each
5. Sorts sessions by start time, newest first
6. If `--json`: marshals and prints JSON, exits
7. If `--watch`: starts bubbletea app with fsnotify
8. Otherwise: calls `ui.RenderToday(sessions, termWidth)` and prints

### 5.3 blocks.go

5-hour billing window logic:

```go
// BlockStart returns the UTC start of the 5-hour billing block containing t
func BlockStart(t time.Time) time.Time {
    utc := t.UTC()
    blockHour := (utc.Hour() / 5) * 5
    return time.Date(utc.Year(), utc.Month(), utc.Day(), blockHour, 0, 0, 0, time.UTC)
}

// Group all session messages by their billing block
// A message's block is determined by its timestamp
func GroupByBlocks(allSessions []*aggregator.SessionStats) map[time.Time]*BlockStats {
    blocks := make(map[time.Time]*BlockStats)
    // ... aggregate per block
}
```

---

## Phase 6: Main Entry (main.go)

```go
package main

import "github.com/yourname/agentop/cmd"

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    cmd.SetVersionInfo(version, commit, date)
    cmd.Execute()
}
```

---

## Phase 7: Test Data (testdata/)

Create realistic test JSONL files:

### testdata/sessions/good-cache.jsonl
A session with 47 messages and 91% cache efficiency. Include:
- 1 human message with `type: "human"`
- 1 assistant response with high `cache_read_input_tokens` (e.g. 78000), lower input (8500), some output (1200), some cache creation (5000)
- Repeat pattern for 47 turns, escalating cache_read as session progresses
- Include tool use events for Read, Edit, Bash
- Include `costUSD` on each assistant message
- `model: "claude-opus-4-6"`

### testdata/sessions/cold-start.jsonl
A new session, first 5 messages, cache_read = 0 throughout.

### testdata/sessions/in-progress.jsonl
A session where the last line is truncated (simulates a session currently running):
```
{"type":"human","sessionId":"test-123",...}
{"type":"assistant","sessionId":"test-123",... (truncated here)
```
Parser must handle this gracefully (skip last line if JSON fails to parse).

---

## Phase 8: Output Format

### Static output (no --watch)

Print to stdout, use lipgloss to render panels. The exact visual should match this:

```
$ agentop

 today · apr 7 2026 · 4 sessions · ~/.claude/projects/
 ┌──────────────────────────────────────────────────────────────────────┐
 │  total cost      tokens        cache eff     5h block    burn rate  │
 │   $18.42          42.3M          61%           23%        2.1k/m   │
 │  ████████░░    ████████░░     ██████░░░░     ███░░░░     ████░░░░  │
 └──────────────────────────────────────────────────────────────────────┘

 ┌─ claude code · 4 sessions ─────────────────────────────────────────┐
 │  session          token mix                        cache  cost  dur │
 ├──────────────────────────────────────────────────────────────────── │
 │  frontend-redesign[opus]  ████████████████████████  91%  $7.20  2h18m │
 │  ~/work/acme-app          0.3M in·1.2M out·0.5M cc·4.9M cr        │
 ├──────────────────────────────────────────────────────────────────── │
 │  api-refactor [opus]      ██████████████████████░░   3%  $5.81  22m  │
 │  ~/work/acme-api          8.3M in·4.8M out·2.4M cc·0.5M cr        │
 └────────────────────────────────────────────────────────────────────┘

 ┌─ anomalies ────────────────────────────────────────────────────────┐
 │  ⚠  api-refactor: 3% cache — cold start. $0.73/msg (8 msgs, $5.81) │
 │  ✓  frontend-redesign: 91% cache efficiency. $0.15/msg.             │
 └────────────────────────────────────────────────────────────────────┘
```

---

## Implementation Order

Build in this exact order (each step is runnable/testable):

1. `internal/claude/schema.go` — just structs, no logic
2. `internal/claude/jsonl.go` — parser with test files
3. `internal/claude/discover.go` — directory scanner
4. `internal/pricing/` — embedded pricing + Calculate()
5. `internal/aggregator/session.go` — AggregateSession()
6. `internal/ui/styles.go` — colors and lipgloss styles
7. `internal/ui/bars.go` — TokenBar.Render()
8. `internal/ui/helpers.go` — humanizeTokens, formatCost, formatDuration
9. `cmd/root.go` + `cmd/today.go` — wire it all up (print static output)
10. `internal/ui/summary.go` — 5-card summary strip
11. `internal/ui/panels.go` — bordered section panels
12. `internal/aggregator/blocks.go` + `cmd/blocks.go`
13. `internal/aggregator/daily.go` + `cmd/daily.go`
14. `internal/aggregator/monthly.go` + `cmd/monthly.go`
15. `internal/aggregator/doctor.go` + `cmd/doctor.go`
16. `cmd/session.go` — session detail view
17. `internal/watcher/watcher.go` + `--watch` support in today.go
18. `--json` output for all commands
19. `internal/config/config.go` + `cmd/config.go`
20. `.goreleaser.yml` + `Makefile`

---

## Critical Implementation Details

### Handling `costUSD`
- Always prefer `costUSD` from the JSONL when it's > 0 (Anthropic's authoritative calculation)
- Sum `costUSD` across all assistant messages in the session
- Use `CostUSDCalculated` (from your pricing table) as a fallback and for verification
- Show both in session detail: "cost: $7.20 (verified: $7.18)"

### Terminal width detection
```go
import "golang.org/x/term"

width, _, err := term.GetSize(int(os.Stdout.Fd()))
if err != nil || width < 40 {
    width = 80 // safe default
}
```

### Date filter parsing
```go
func ParseSince(s string) (time.Time, error) {
    now := time.Now()
    switch s {
    case "today":     return StartOfDay(now), nil
    case "yesterday": return StartOfDay(now.AddDate(0,0,-1)), nil
    }
    if strings.HasSuffix(s, "d") {
        days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
        if err != nil { return time.Time{}, err }
        return now.AddDate(0, 0, -days), nil
    }
    return time.Parse("2006-01-02", s) // YYYY-MM-DD
}
```

### No TTY detection (pipe-friendly)
```go
if !term.IsTerminal(int(os.Stdout.Fd())) {
    jsonOut = true // auto-JSON when piped
}
```

---

## assets/pricing.json (embed this exactly)

```json
{
  "version": "2026-04-07",
  "models": {
    "claude-opus-4-6": {
      "input": 15.00, "output": 75.00,
      "cacheCreate": 18.75, "cacheRead": 1.50
    },
    "claude-sonnet-4-6": {
      "input": 3.00, "output": 15.00,
      "cacheCreate": 3.75, "cacheRead": 0.30
    },
    "claude-haiku-4-5": {
      "input": 0.80, "output": 4.00,
      "cacheCreate": 1.00, "cacheRead": 0.08
    },
    "claude-opus-4": {
      "input": 15.00, "output": 75.00,
      "cacheCreate": 18.75, "cacheRead": 1.50
    },
    "claude-sonnet-4": {
      "input": 3.00, "output": 15.00,
      "cacheCreate": 3.75, "cacheRead": 0.30
    },
    "claude-haiku-4": {
      "input": 0.80, "output": 4.00,
      "cacheCreate": 1.00, "cacheRead": 0.08
    }
  }
}
```

---

## Makefile

```makefile
.PHONY: build test run clean install

build:
	go build -ldflags="-s -w" -o ./bin/agentop .

test:
	go test ./... -v

run:
	go run . 

run-watch:
	go run . --watch

lint:
	golangci-lint run

install:
	go install .

release-dry:
	goreleaser release --snapshot --clean

clean:
	rm -rf ./bin ./dist
```

---

## Start building now. Begin with Phase 1 (scaffold + go.mod), then proceed through each phase in order. After each phase, run `go build .` to verify it compiles before moving to the next.
